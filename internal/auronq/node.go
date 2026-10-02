package auronq

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type NodeConfig struct {
	Listen        string
	Advertise     string
	Peers         []string
	PeerStorePath string
	LookupHost    func(context.Context, string) ([]string, error)
}

type Node struct {
	Chain      *Chain
	cfg        NodeConfig
	server     *http.Server
	client     *http.Client
	pmu        sync.RWMutex
	peers      map[string]struct{}
	announced  map[string]struct{}
	bootstrap  map[string]struct{}
	failures   map[string]int
	syncCursor int
	blockSem   chan struct{}
	rateMu     sync.Mutex
	rate       map[string]requestRateWindow
}

const (
	maxKnownPeers                = 128
	maxPeersPerNetgroup          = 4
	maxSyncPeersPerRound         = 8
	maxSyncBlocksPerRound        = 256
	peerFailureDrop              = 6
	dnsRefreshInterval           = 5 * time.Minute
	bootstrapRefreshInterval     = 2 * time.Minute
	maxBootstrapManifestBytes    = 64 * 1024
	maxBootstrapManifestPeers    = 64
	maxBootstrapManifestSources  = 8
	smallP2PResponseLimit        = 64 * 1024
	maxConcurrentBlockValidation = 2
	requestRateWindowDuration    = 10 * time.Second
	maxRequestsPerIPWindow       = 320
	maxBlocksPerIPWindow         = 8
	maxRateEntries               = 8192
)

type requestRateWindow struct {
	Start time.Time
	Count int
}

type Hello struct {
	ProtocolVersion uint32   `json:"protocol_version"`
	NetworkID       Hash     `json:"network_id"`
	Height          uint64   `json:"height"`
	Tip             Hash     `json:"tip"`
	ChainWork       string   `json:"chain_work"`
	Advertise       string   `json:"advertise,omitempty"`
	Peers           []string `json:"peers,omitempty"`
}

type Status struct {
	Network     string  `json:"network"`
	NetworkID   Hash    `json:"network_id"`
	Height      uint64  `json:"height"`
	Tip         Hash    `json:"tip"`
	ChainWork   string  `json:"chain_work"`
	Issued      uint64  `json:"issued_atoms"`
	IssuedCoins float64 `json:"issued_coins"`
	MaxSupply   uint64  `json:"max_supply_atoms"`
	Mempool     int     `json:"mempool"`
	Peers       int     `json:"peers"`
}

type BalanceResponse struct {
	Address   string `json:"address"`
	Spendable uint64 `json:"spendable"`
	Total     uint64 `json:"total"`
}
type TxResponse struct {
	TXID Hash   `json:"txid"`
	Fee  uint64 `json:"fee"`
}
type BlockResponse struct {
	Hash   Hash   `json:"hash"`
	Height uint64 `json:"height"`
}
type MempoolResponse struct {
	Transactions []Transaction `json:"transactions"`
}

type PeerAnnounce struct {
	ProtocolVersion uint32 `json:"protocol_version"`
	NetworkID       Hash   `json:"network_id"`
	Advertise       string `json:"advertise,omitempty"`
	ListenPort      uint16 `json:"listen_port,omitempty"`
}

type BootstrapManifest struct {
	NetworkID Hash     `json:"network_id"`
	Peers     []string `json:"peers"`
	ExpiresAt int64    `json:"expires_at,omitempty"`
}

func NewNode(chain *Chain, cfg NodeConfig) *Node {
	if cfg.LookupHost == nil {
		cfg.LookupHost = net.DefaultResolver.LookupHost
	}
	n := &Node{Chain: chain, cfg: cfg, client: &http.Client{Timeout: 8 * time.Second}, peers: map[string]struct{}{}, announced: map[string]struct{}{}, bootstrap: map[string]struct{}{}, failures: map[string]int{}, blockSem: make(chan struct{}, maxConcurrentBlockValidation), rate: map[string]requestRateWindow{}}
	configured := append([]string(nil), chain.network.SeedPeers...)
	configured = append(configured, cfg.Peers...)
	for _, p := range configured {
		n.addBootstrapPeer(p)
	}
	n.loadPeerStore()
	return n
}
func normalizePeer(p string) string {
	p = strings.TrimSpace(strings.TrimRight(p, "/"))
	if p == "" {
		return ""
	}
	if !strings.HasPrefix(p, "http://") && !strings.HasPrefix(p, "https://") {
		p = "http://" + p
	}
	u, err := url.Parse(p)
	if err != nil || u.Host == "" {
		return ""
	}
	return u.Scheme + "://" + u.Host
}
func isNonPublicIP(ip net.IP) bool {
	if ip == nil || !ip.IsGlobalUnicast() || ip.IsLoopback() || ip.IsUnspecified() || ip.IsMulticast() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}
	// 100.64.0.0/10 is carrier-grade NAT (RFC 6598), not globally reachable.
	if v4 := ip.To4(); v4 != nil && v4[0] == 100 && v4[1] >= 64 && v4[1] <= 127 {
		return true
	}
	// Documentation-only address ranges are never usable public peers.
	if v4 := ip.To4(); v4 != nil {
		if (v4[0] == 192 && v4[1] == 0 && v4[2] == 2) ||
			(v4[0] == 198 && v4[1] == 51 && v4[2] == 100) ||
			(v4[0] == 203 && v4[1] == 0 && v4[2] == 113) {
			return true
		}
	}
	if v6 := ip.To16(); v6 != nil && ip.To4() == nil && v6[0] == 0x20 && v6[1] == 0x01 && v6[2] == 0x0d && v6[3] == 0xb8 {
		return true
	}
	return false
}

// Public peer gossip deliberately accepts only literal globally-routable IPs.
// DNS names are allowed as configured seed peers, but are not learned from an
// untrusted remote node. This prevents peer gossip from turning the node into
// an SSRF client for localhost/private networks or DNS-rebinding targets.
func isPublicAdvertisedPeer(p string) bool {
	p = normalizePeer(p)
	if p == "" {
		return false
	}
	u, err := url.Parse(p)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return false
	}
	ip := net.ParseIP(strings.Trim(u.Hostname(), "[]"))
	return ip != nil && !isNonPublicIP(ip)
}

func peerNetgroup(p string) string {
	p = normalizePeer(p)
	u, err := url.Parse(p)
	if err != nil {
		return ""
	}
	ip := net.ParseIP(strings.Trim(u.Hostname(), "[]"))
	if ip == nil {
		return ""
	}
	if v4 := ip.To4(); v4 != nil {
		return fmt.Sprintf("v4:%d.%d", v4[0], v4[1])
	}
	v6 := ip.To16()
	if v6 == nil {
		return ""
	}
	return fmt.Sprintf("v6:%02x%02x:%02x%02x", v6[0], v6[1], v6[2], v6[3])
}

func (n *Node) addPeer(p string) {
	p = normalizePeer(p)
	if p == "" || p == normalizePeer(n.cfg.Advertise) {
		return
	}
	n.pmu.Lock()
	if len(n.peers) < maxKnownPeers {
		n.peers[p] = struct{}{}
	}
	n.pmu.Unlock()
}

func (n *Node) addBootstrapPeer(p string) {
	p = normalizePeer(p)
	if p == "" || p == normalizePeer(n.cfg.Advertise) {
		return
	}
	n.pmu.Lock()
	if len(n.peers) < maxKnownPeers {
		n.peers[p] = struct{}{}
		n.bootstrap[p] = struct{}{}
	}
	n.pmu.Unlock()
}

func (n *Node) addDiscoveredPeer(p string) {
	p = normalizePeer(p)
	if !isPublicAdvertisedPeer(p) || p == normalizePeer(n.cfg.Advertise) {
		return
	}
	group := peerNetgroup(p)
	if group == "" {
		return
	}
	n.pmu.Lock()
	before := len(n.peers)
	if before < maxKnownPeers {
		if _, exists := n.peers[p]; !exists {
			groupCount := 0
			for existing := range n.peers {
				if peerNetgroup(existing) == group {
					groupCount++
				}
			}
			if groupCount < maxPeersPerNetgroup {
				n.peers[p] = struct{}{}
			}
		}
	}
	changed := len(n.peers) != before
	n.pmu.Unlock()
	if changed {
		n.savePeerStore()
	}
}

// AddPeer is intended for a peer learned from the public P2P network. Configured
// seed/manual peers are added through NodeConfig and may intentionally use a
// private address for development. Runtime gossip is restricted to public IPs.
func (n *Node) AddPeer(p string) { n.addDiscoveredPeer(p) }

// AddLocalPeer adds an explicitly discovered local/private peer for the current
// process. Local peers are deliberately not gossiped to the public network and
// are not written to the public peer store. This is used by the desktop
// application for zero-touch LAN discovery.
func (n *Node) AddLocalPeer(p string) { n.addBootstrapPeer(p) }

func (n *Node) peerList() []string {
	n.pmu.RLock()
	defer n.pmu.RUnlock()
	p := make([]string, 0, len(n.peers))
	for x := range n.peers {
		p = append(p, x)
	}
	sort.Strings(p)
	return p
}

func (n *Node) advertisedPeerList() []string {
	all := n.peerList()
	out := make([]string, 0, len(all))
	for _, p := range all {
		if isPublicAdvertisedPeer(p) {
			out = append(out, p)
		}
		if len(out) >= 32 {
			break
		}
	}
	return out
}

func (n *Node) loadPeerStore() {
	if n.cfg.PeerStorePath == "" {
		return
	}
	b, err := os.ReadFile(n.cfg.PeerStorePath)
	if err != nil {
		return
	}
	var peers []string
	if json.Unmarshal(b, &peers) != nil {
		return
	}
	for _, p := range peers {
		n.addDiscoveredPeer(p)
	}
}

func (n *Node) savePeerStore() {
	if n.cfg.PeerStorePath == "" {
		return
	}
	peers := n.advertisedPeerList()
	b, err := json.MarshalIndent(peers, "", "  ")
	if err != nil {
		return
	}
	_ = os.MkdirAll(filepath.Dir(n.cfg.PeerStorePath), 0700)
	tmp := n.cfg.PeerStorePath + ".tmp"
	if os.WriteFile(tmp, b, 0600) == nil {
		_ = os.Rename(tmp, n.cfg.PeerStorePath)
	}
}

func validBootstrapManifestURL(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || u.Fragment != "" {
		return "", false
	}
	// A bootstrap manifest is public metadata. Credentials, fragments and
	// non-HTTPS transports are deliberately rejected. Query strings are allowed
	// because immutable/raw hosting endpoints may use them.
	return u.String(), true
}

func isSafeManifestPeer(raw string) bool {
	p := normalizePeer(raw)
	if p == "" {
		return false
	}
	u, err := url.Parse(p)
	if err != nil || u.User != nil || u.Hostname() == "" {
		return false
	}
	host := strings.ToLower(strings.TrimSuffix(u.Hostname(), "."))
	if ip := net.ParseIP(strings.Trim(host, "[]")); ip != nil {
		return !isNonPublicIP(ip)
	}
	// DNS-based peers from a remotely hosted manifest are only accepted over
	// HTTPS. TLS hostname verification prevents a manifest from becoming a
	// general-purpose HTTP SSRF primitive even if its DNS is later changed.
	if u.Scheme != "https" {
		return false
	}
	if host == "localhost" || strings.HasSuffix(host, ".localhost") || strings.HasSuffix(host, ".local") || strings.HasSuffix(host, ".internal") || strings.HasSuffix(host, ".home.arpa") {
		return false
	}
	return strings.Contains(host, ".")
}

func (n *Node) addManifestPeer(raw string) bool {
	if !isSafeManifestPeer(raw) {
		return false
	}
	p := normalizePeer(raw)
	if p == "" || p == normalizePeer(n.cfg.Advertise) {
		return false
	}
	n.pmu.Lock()
	defer n.pmu.Unlock()
	if _, exists := n.peers[p]; exists || len(n.peers) >= maxKnownPeers {
		return false
	}
	n.peers[p] = struct{}{}
	// Manifest peers are intentionally not put in n.bootstrap: if a tunnel or
	// remote endpoint disappears, normal failure pruning must be able to remove it.
	return true
}

func (n *Node) refreshBootstrapManifests(ctx context.Context) {
	sources := n.Chain.network.BootstrapManifests
	if len(sources) > maxBootstrapManifestSources {
		sources = sources[:maxBootstrapManifestSources]
	}
	for _, raw := range sources {
		manifestURL, ok := validBootstrapManifestURL(raw)
		if !ok {
			log.Printf("bootstrap manifest rejected: %q", raw)
			continue
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, manifestURL, nil)
		if err != nil {
			continue
		}
		req.Header.Set("Accept", "application/json")
		resp, err := n.client.Do(req)
		if err != nil {
			log.Printf("bootstrap manifest %s failed: %v", manifestURL, err)
			continue
		}
		func() {
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
				log.Printf("bootstrap manifest %s: %s", manifestURL, resp.Status)
				return
			}
			lr := io.LimitReader(resp.Body, maxBootstrapManifestBytes+1)
			b, err := io.ReadAll(lr)
			if err != nil || len(b) > maxBootstrapManifestBytes {
				log.Printf("bootstrap manifest %s is unreadable or too large", manifestURL)
				return
			}
			var m BootstrapManifest
			dec := json.NewDecoder(bytes.NewReader(b))
			dec.DisallowUnknownFields()
			if err := dec.Decode(&m); err != nil {
				log.Printf("bootstrap manifest %s decode failed: %v", manifestURL, err)
				return
			}
			var trailing any
			if err := dec.Decode(&trailing); err != io.EOF {
				log.Printf("bootstrap manifest %s has trailing JSON data", manifestURL)
				return
			}
			if m.NetworkID != n.Chain.NetworkID() {
				log.Printf("bootstrap manifest %s is for a different network", manifestURL)
				return
			}
			if m.ExpiresAt != 0 && time.Now().Unix() > m.ExpiresAt {
				log.Printf("bootstrap manifest %s is expired", manifestURL)
				return
			}
			peers := m.Peers
			if len(peers) > maxBootstrapManifestPeers {
				peers = peers[:maxBootstrapManifestPeers]
			}
			added := 0
			for _, peer := range peers {
				if n.addManifestPeer(peer) {
					added++
				}
			}
			if added > 0 {
				n.savePeerStore()
			}
			log.Printf("bootstrap manifest %s accepted: %d peer(s)", manifestURL, len(peers))
		}()
	}
}

func parseSeedHost(s string) (string, int) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", DefaultP2PPort
	}
	if strings.Contains(s, "://") {
		u, err := url.Parse(s)
		if err != nil || u.Hostname() == "" {
			return "", DefaultP2PPort
		}
		port := DefaultP2PPort
		if u.Port() != "" {
			if _, err := fmt.Sscanf(u.Port(), "%d", &port); err != nil || port < 1 || port > 65535 {
				return "", DefaultP2PPort
			}
		}
		return u.Hostname(), port
	}
	if host, portText, err := net.SplitHostPort(s); err == nil {
		port := DefaultP2PPort
		if _, err := fmt.Sscanf(portText, "%d", &port); err != nil || port < 1 || port > 65535 {
			return "", DefaultP2PPort
		}
		return strings.Trim(host, "[]"), port
	}
	return strings.Trim(s, "[]"), DefaultP2PPort
}

func (n *Node) refreshDNSSeeds(ctx context.Context) {
	for _, seed := range n.Chain.network.DNSSeeds {
		host, port := parseSeedHost(seed)
		if host == "" {
			continue
		}
		ips, err := n.cfg.LookupHost(ctx, host)
		if err != nil {
			log.Printf("dns seed %s failed: %v", host, err)
			continue
		}
		for _, raw := range ips {
			ip := net.ParseIP(strings.TrimSpace(raw))
			if isNonPublicIP(ip) {
				continue
			}
			peer := "http://" + net.JoinHostPort(ip.String(), fmt.Sprintf("%d", port))
			n.addDiscoveredPeer(peer)
		}
	}
}

func listenPort(listen string) uint16 {
	_, p, err := net.SplitHostPort(listen)
	if err != nil {
		return 0
	}
	var port int
	if _, err = fmt.Sscanf(p, "%d", &port); err != nil || port < 1 || port > 65535 {
		return 0
	}
	return uint16(port)
}

func observedAnnouncementPeer(remoteAddr string, port uint16) string {
	if port == 0 {
		return ""
	}
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return ""
	}
	ip := net.ParseIP(strings.Trim(host, "[]"))
	if isNonPublicIP(ip) {
		return ""
	}
	return "http://" + net.JoinHostPort(ip.String(), fmt.Sprintf("%d", port))
}

func (n *Node) recordPeerSuccess(peer string) {
	n.pmu.Lock()
	delete(n.failures, peer)
	n.pmu.Unlock()
}

func (n *Node) recordPeerFailure(peer string) {
	n.pmu.Lock()
	defer n.pmu.Unlock()
	if _, keep := n.bootstrap[peer]; keep {
		return
	}
	n.failures[peer]++
	if n.failures[peer] >= peerFailureDrop {
		delete(n.peers, peer)
		delete(n.failures, peer)
		delete(n.announced, peer)
		go n.savePeerStore()
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func readJSON(w http.ResponseWriter, r *http.Request, max int64, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, max)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return false
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		writeJSON(w, 400, map[string]string{"error": "request must contain exactly one JSON value"})
		return false
	}
	return true
}

func (n *Node) acquireBlockValidation() bool {
	select {
	case n.blockSem <- struct{}{}:
		return true
	default:
		return false
	}
}

func (n *Node) releaseBlockValidation() {
	<-n.blockSem
}

func requestSourceIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return ""
	}
	ip := net.ParseIP(strings.Trim(host, "[]"))
	if ip == nil {
		return ""
	}
	return ip.String()
}

func (n *Node) allowRate(key string, max int) bool {
	if key == "" || max <= 0 {
		return true
	}
	now := time.Now()
	n.rateMu.Lock()
	defer n.rateMu.Unlock()
	if len(n.rate) >= maxRateEntries {
		for k, v := range n.rate {
			if now.Sub(v.Start) >= 2*requestRateWindowDuration {
				delete(n.rate, k)
			}
		}
		if len(n.rate) >= maxRateEntries {
			if _, ok := n.rate[key]; !ok {
				return false
			}
		}
	}
	w := n.rate[key]
	if w.Start.IsZero() || now.Sub(w.Start) >= requestRateWindowDuration {
		w = requestRateWindow{Start: now}
	}
	if w.Count >= max {
		n.rate[key] = w
		return false
	}
	w.Count++
	n.rate[key] = w
	return true
}

func (n *Node) allowRequest(r *http.Request) bool {
	ip := requestSourceIP(r.RemoteAddr)
	return n.allowRate(ip+"|all", maxRequestsPerIPWindow)
}

func (n *Node) allowBlockRequest(r *http.Request) bool {
	ip := requestSourceIP(r.RemoteAddr)
	return n.allowRate(ip+"|block", maxBlocksPerIPWindow)
}

func (n *Node) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(405)
			return
		}
		st := n.Chain.State()
		writeJSON(w, 200, Status{Network: n.Chain.network.Name, NetworkID: n.Chain.NetworkID(), Height: st.Height, Tip: st.Tip, ChainWork: st.ChainWork, Issued: st.Issued, IssuedCoins: float64(st.Issued) / float64(Coin), MaxSupply: MaxSupplyAtoms, Mempool: n.Chain.MempoolSize(), Peers: len(n.peerList())})
	})
	mux.HandleFunc("/v1/balance", func(w http.ResponseWriter, r *http.Request) {
		addr := r.URL.Query().Get("address")
		sp, total, err := n.Chain.Balance(addr)
		if err != nil {
			writeJSON(w, 400, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, 200, BalanceResponse{addr, sp, total})
	})
	mux.HandleFunc("/v1/utxos", func(w http.ResponseWriter, r *http.Request) {
		addr := r.URL.Query().Get("address")
		u, err := n.Chain.UTXOsForAddress(addr, false)
		if err != nil {
			writeJSON(w, 400, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, 200, u)
	})
	mux.HandleFunc("/v1/template", func(w http.ResponseWriter, r *http.Request) {
		addr := r.URL.Query().Get("address")
		b, err := n.Chain.BuildTemplate(addr)
		if err != nil {
			writeJSON(w, 400, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, 200, b)
	})
	mux.HandleFunc("/v1/tx", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(405)
			return
		}
		var tx Transaction
		if !readJSON(w, r, MaxTxBytes*2, &tx) {
			return
		}
		fee, err := n.Chain.AddMempool(tx)
		if err != nil {
			writeJSON(w, 400, map[string]string{"error": err.Error()})
			return
		}
		id := tx.ID()
		writeJSON(w, 200, TxResponse{id, fee})
		go n.broadcast("/p2p/tx", tx, "")
	})
	mux.HandleFunc("/v1/block", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(405)
			return
		}
		if !n.allowBlockRequest(r) {
			writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "block submission rate limit exceeded"})
			return
		}
		var b Block
		if !readJSON(w, r, MaxBlockBytes+64*1024, &b) {
			return
		}
		if !n.acquireBlockValidation() {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "proof-of-work validator busy"})
			return
		}
		err := n.Chain.AddBlock(&b)
		n.releaseBlockValidation()
		if err != nil {
			writeJSON(w, 400, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, 200, BlockResponse{b.Hash(), b.Header.Height})
		go n.broadcast("/p2p/block", b, "")
	})

	mux.HandleFunc("/p2p/hello", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(405)
			return
		}
		st := n.Chain.State()
		writeJSON(w, 200, Hello{1, n.Chain.NetworkID(), st.Height, st.Tip, st.ChainWork, n.cfg.Advertise, n.advertisedPeerList()})
	})
	mux.HandleFunc("/p2p/announce", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(405)
			return
		}
		var a PeerAnnounce
		if !readJSON(w, r, 16<<10, &a) {
			return
		}
		if a.ProtocolVersion != 1 || a.NetworkID != n.Chain.NetworkID() {
			writeJSON(w, 400, map[string]string{"error": "invalid peer announcement"})
			return
		}
		peer := ""
		if isPublicAdvertisedPeer(a.Advertise) {
			peer = normalizePeer(a.Advertise)
		} else {
			peer = observedAnnouncementPeer(r.RemoteAddr, a.ListenPort)
		}
		if peer == "" {
			writeJSON(w, 400, map[string]string{"error": "peer is not publicly reachable"})
			return
		}
		go n.verifyAndAddPeer(peer)
		writeJSON(w, 202, map[string]bool{"accepted": true})
	})
	mux.HandleFunc("/p2p/blockhash", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(405)
			return
		}
		h, err := parseHeight(r.URL.Query().Get("height"))
		if err != nil {
			writeJSON(w, 400, map[string]string{"error": err.Error()})
			return
		}
		b, ok := n.Chain.Block(h)
		if !ok {
			writeJSON(w, 404, map[string]string{"error": "height not found"})
			return
		}
		writeJSON(w, 200, map[string]any{"height": h, "hash": b.Hash()})
	})
	mux.HandleFunc("/p2p/getblock", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(405)
			return
		}
		h, err := parseHeight(r.URL.Query().Get("height"))
		if err != nil {
			writeJSON(w, 400, map[string]string{"error": err.Error()})
			return
		}
		b, ok := n.Chain.Block(h)
		if !ok {
			writeJSON(w, 404, map[string]string{"error": "height not found"})
			return
		}
		writeJSON(w, 200, b)
	})
	mux.HandleFunc("/p2p/mempool", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(405)
			return
		}
		// Keep a catch-up response comfortably below the normal block-response cap.
		writeJSON(w, 200, MempoolResponse{Transactions: n.Chain.MempoolTransactions(MaxBlockBytes / 2)})
	})
	mux.HandleFunc("/p2p/tx", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(405)
			return
		}
		var tx Transaction
		if !readJSON(w, r, MaxTxBytes*2, &tx) {
			return
		}
		fee, err := n.Chain.AddMempool(tx)
		if err != nil {
			writeJSON(w, 400, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, 200, TxResponse{tx.ID(), fee})
		go n.broadcast("/p2p/tx", tx, "")
	})
	mux.HandleFunc("/p2p/block", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(405)
			return
		}
		if !n.allowBlockRequest(r) {
			writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "block submission rate limit exceeded"})
			return
		}
		var b Block
		if !readJSON(w, r, MaxBlockBytes+64*1024, &b) {
			return
		}
		if !n.acquireBlockValidation() {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "proof-of-work validator busy"})
			return
		}
		err := n.Chain.AddBlock(&b)
		n.releaseBlockValidation()
		if err != nil {
			writeJSON(w, 400, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, 200, BlockResponse{b.Hash(), b.Header.Height})
		go n.broadcast("/p2p/block", b, "")
	})
	rateLimited := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !n.allowRequest(r) {
			writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "request rate limit exceeded"})
			return
		}
		mux.ServeHTTP(w, r)
	})
	return securityHeaders(limitConcurrent(rateLimited, 96))
}
func limitConcurrent(next http.Handler, max int) http.Handler {
	sem := make(chan struct{}, max)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case sem <- struct{}{}:
			defer func() { <-sem }()
			next.ServeHTTP(w, r)
		default:
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "node busy"})
		}
	})
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}
func parseHeight(s string) (uint64, error) {
	if s == "" {
		return 0, errors.New("height required")
	}
	h, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, errors.New("invalid height")
	}
	return h, nil
}

func (n *Node) Run(ctx context.Context) error {
	n.server = &http.Server{Addr: n.cfg.Listen, Handler: n.handler(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 20 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 16 * 1024}
	go n.syncLoop(ctx)
	errc := make(chan error, 1)
	go func() {
		log.Printf("AuronQ node listening on %s network=%s id=%s", n.cfg.Listen, n.Chain.network.Name, n.Chain.NetworkID().String()[:16])
		errc <- n.server.ListenAndServe()
	}()
	select {
	case <-ctx.Done():
		sdctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		return n.server.Shutdown(sdctx)
	case err := <-errc:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	}
}

func (n *Node) broadcast(path string, v any, skip string) {
	b, _ := json.Marshal(v)
	for _, p := range n.peerList() {
		if p == skip {
			continue
		}
		go func(peer string) {
			req, _ := http.NewRequest("POST", peer+path, bytes.NewReader(b))
			req.Header.Set("Content-Type", "application/json")
			resp, err := n.client.Do(req)
			if err == nil {
				io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
				resp.Body.Close()
			}
		}(p)
	}
}

func (n *Node) syncLoop(ctx context.Context) {
	t := time.NewTicker(10 * time.Second)
	defer t.Stop()
	dns := time.NewTicker(dnsRefreshInterval)
	defer dns.Stop()
	bootstrap := time.NewTicker(bootstrapRefreshInterval)
	defer bootstrap.Stop()
	n.refreshDNSSeeds(ctx)
	n.refreshBootstrapManifests(ctx)
	n.syncAll()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			n.syncAll()
		case <-dns.C:
			n.refreshDNSSeeds(ctx)
		case <-bootstrap.C:
			n.refreshBootstrapManifests(ctx)
		}
	}
}
func (n *Node) syncAll() {
	peers := n.peerList()
	if len(peers) == 0 {
		return
	}
	n.pmu.Lock()
	start := n.syncCursor % len(peers)
	count := len(peers)
	if count > maxSyncPeersPerRound {
		count = maxSyncPeersPerRound
	}
	n.syncCursor = (start + count) % len(peers)
	n.pmu.Unlock()
	for i := 0; i < count; i++ {
		p := peers[(start+i)%len(peers)]
		if err := n.syncPeer(p); err != nil {
			log.Printf("sync from %s failed: %v", p, err)
			n.recordPeerFailure(p)
			continue
		}
		n.recordPeerSuccess(p)
	}
}
func (n *Node) getJSONLimit(peer, path string, maxBytes int64, out any) error {
	resp, err := n.client.Get(peer + path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("peer %s: %s", resp.Status, strings.TrimSpace(string(b)))
	}
	dec := json.NewDecoder(io.LimitReader(resp.Body, maxBytes))
	if err := dec.Decode(out); err != nil {
		return err
	}
	return nil
}

func (n *Node) getJSON(peer, path string, out any) error {
	return n.getJSONLimit(peer, path, int64(MaxBlockBytes+64*1024), out)
}

// postPeerJSON sends a bounded P2P POST and requires a successful HTTP status.
// It is used by catch-up push so an outbound-only node can repair a public
// bootstrap peer that missed a previously broadcast block.
func (n *Node) postPeerJSON(peer, path string, v any) error {