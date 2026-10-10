package bridge

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	aq "auronq/internal/auronq"
)

const (
	mobileVersion       = "0.5.3-alpha"
	mainnetNetworkID    = "44e62c2ace002a6660c14e252173c1aa303529c68e40c998e92da2b453f44f30b1e58c94d533587e2186004593fb856c433fcdb5418ed430ec8617e29529365c"
	bootstrapManifestURL = "https://raw.githubusercontent.com/promirmir/AuronQ/main/bootstrap.json"
)

var bundledBootstrapPeers = []string{
	"https://mir.taild63f46.ts.net",
	"https://desktop-4nifg1j.taild63f46.ts.net",
	"http://45.88.201.77:18444",
	"http://54.38.81.30:18444",
	"http://69.173.206.211:18444",
}

type bootstrapManifest struct {
	NetworkID string   `json:"network_id"`
	Peers     []string `json:"peers"`
	ExpiresAt int64    `json:"expires_at,omitempty"`
}

const (
	maxMobileKnownCandidates = 8
	maxMobileProbeCandidates = 16
	headerCacheVersion       = 1
	headerCacheKeep          = 128
	headerBatchLimit         = 64
	mainnetGenesisHash       = "5750a455c04bfe93c9edfef1a12744b05e29ac6da1a9dd5b790566629dea2080c581beb2f0324efba2067c9efb113ed7a29598fd3ffbed965ced31f265d0cec4"
	mainnetGenesisHeaderJSON = `{"version":2,"pow_algo":1,"height":0,"prev_hash":"00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000","merkle_root":"4ec627ff8d4b6ec5957fd476d60d0b5cbc533476206983bdda1dcea08554198e84a987036fd223aca6b330a926fcba54365e98008a154061f35b46f49bff60be","timestamp":1790951480,"target":"003fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff","nonce":169}`
)

type mobileNodeObservation struct {
	Node      string `json:"node"`
	Height    uint64 `json:"height"`
	Tip       string `json:"tip"`
	ChainWork string `json:"chain_work"`
	Peers     int    `json:"peers"`
}

func mobileNonPublicIP(ip net.IP) bool {
	if ip == nil || !ip.IsGlobalUnicast() || ip.IsLoopback() || ip.IsUnspecified() ||
		ip.IsMulticast() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}
	if v4 := ip.To4(); v4 != nil {
		if v4[0] == 100 && v4[1] >= 64 && v4[1] <= 127 {
			return true
		}
		if (v4[0] == 192 && v4[1] == 0 && v4[2] == 2) ||
			(v4[0] == 198 && v4[1] == 51 && v4[2] == 100) ||
			(v4[0] == 203 && v4[1] == 0 && v4[2] == 113) {
			return true
		}
	}
	if v6 := ip.To16(); v6 != nil && ip.To4() == nil &&
		v6[0] == 0x20 && v6[1] == 0x01 && v6[2] == 0x0d && v6[3] == 0xb8 {
		return true
	}
	return false
}

func safeMobileDNSHost(host string) bool {
	host = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(host), "."))
	return host != "" && strings.Contains(host, ".") &&
		host != "localhost" &&
		!strings.HasSuffix(host, ".localhost") &&
		!strings.HasSuffix(host, ".local") &&
		!strings.HasSuffix(host, ".internal") &&
		!strings.HasSuffix(host, ".home.arpa")
}

func normalizeMobileNode(raw string) string {
	raw = strings.TrimSpace(strings.TrimRight(raw, "/"))
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || u.User != nil || u.Fragment != "" || u.RawQuery != "" ||
		(u.Path != "" && u.Path != "/") || (u.Scheme != "http" && u.Scheme != "https") {
		return ""
	}
	host := strings.Trim(u.Hostname(), "[]")
	if ip := net.ParseIP(host); ip != nil {
		if mobileNonPublicIP(ip) {
			return ""
		}
		return u.Scheme + "://" + u.Host
	}
	// Cleartext is accepted only for literal globally routable IPs. DNS peers
	// must use HTTPS so a learned hostname cannot silently downgrade transport.
	if u.Scheme != "https" || !safeMobileDNSHost(host) {
		return ""
	}
	return "https://" + u.Host
}

func addMobileCandidate(out *[]string, seen map[string]bool, raw string) {
	p := normalizeMobileNode(raw)
	if p == "" || seen[p] {
		return
	}
	seen[p] = true
	*out = append(*out, p)
}

func mobileCandidates(knownJSON string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, maxMobileProbeCandidates)
	var known []string
	_ = json.Unmarshal([]byte(strings.TrimSpace(knownJSON)), &known)
	for i, p := range known {
		if i >= maxMobileKnownCandidates {
			break
		}
		addMobileCandidate(&out, seen, p)
	}
	for _, p := range bundledBootstrapPeers {
		addMobileCandidate(&out, seen, p)
	}
	if len(out) > maxMobileProbeCandidates {
		out = out[:maxMobileProbeCandidates]
	}
	return out
}

func mobileHTTP(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func statusFromNode(base string) (aq.Status, error) {
	var zero aq.Status
	base = normalizeMobileNode(base)
	if base == "" {
		return zero, errors.New("mobile wallet requires an HTTPS node")
	}
	cl := aq.NewClient(base)
	cl.HTTP = mobileHTTP(6 * time.Second)
	st, err := cl.Status()
	if err != nil {
		return zero, err
	}
	if st.NetworkID.String() != mainnetNetworkID {
		return zero, errors.New("node Network ID mismatch")
	}
	return st, nil
}

func observeNodes(candidates []string) []mobileNodeObservation {
	if len(candidates) > maxMobileProbeCandidates {
		candidates = candidates[:maxMobileProbeCandidates]
	}
	type result struct {
		obs mobileNodeObservation
		ok  bool
	}
	ch := make(chan result, len(candidates))
	for _, node := range candidates {
		node := node
		go func() {
			st, err := statusFromNode(node)
			if err != nil {
				ch <- result{}
				return
			}
			ch <- result{ok: true, obs: mobileNodeObservation{
				Node:      node,
				Height:    st.Height,
				Tip:       st.Tip.String(),
				ChainWork: st.ChainWork,
				Peers:     st.Peers,
			}}
		}()
	}

	// Do not make the mobile UI wait for every stale/dead endpoint once at least
	// one Mainnet peer answered. Keep a short grace window to collect a second
	// peer for quorum while preserving a six-second cold-start ceiling.
	deadline := time.NewTimer(6 * time.Second)
	defer deadline.Stop()
	var grace *time.Timer
	var graceC <-chan time.Time
	out := make([]mobileNodeObservation, 0, len(candidates))
	pending := len(candidates)
	finish := func() []mobileNodeObservation {
		sort.Slice(out, func(i, j int) bool { return out[i].Node < out[j].Node })
		return out
	}
	for pending > 0 {
		select {
		case r := <-ch:
			pending--
			if r.ok {
				out = append(out, r.obs)
				if len(out) == 1 && pending > 0 {
					grace = time.NewTimer(750 * time.Millisecond)
					graceC = grace.C
				}
			}
			if pending == 0 {
				if grace != nil {
					grace.Stop()
				}
				return finish()
			}
		case <-graceC:
			return finish()
		case <-deadline.C:
			if grace != nil {
				grace.Stop()
			}
			return finish()
		}
	}
	return finish()
}

func mobileStateKey(o mobileNodeObservation) string {
	return fmt.Sprintf("%d|%s|%s", o.Height, o.Tip, o.ChainWork)
}

func chainWorkCmp(a, b string) int {
	// Full-node ChainState serializes cumulative work as hexadecimal.
	aa, okA := new(big.Int).SetString(strings.TrimSpace(a), 16)
	bb, okB := new(big.Int).SetString(strings.TrimSpace(b), 16)
	if okA && okB {
		return aa.Cmp(bb)
	}
	return strings.Compare(a, b)
}

func chooseNodeQuorum(obs []mobileNodeObservation) (mobileNodeObservation, []string, error) {
	var zero mobileNodeObservation
	if len(obs) == 0 {
		return zero, nil, errors.New("no reachable AuronQ Mainnet nodes")
	}
	groups := map[string][]mobileNodeObservation{}
	for _, o := range obs {
		k := mobileStateKey(o)
		groups[k] = append(groups[k], o)
	}
	var best []mobileNodeObservation
	for _, g := range groups {
		if len(best) == 0 ||
			len(g) > len(best) ||
			(len(g) == len(best) && chainWorkCmp(g[0].ChainWork, best[0].ChainWork) > 0) ||
			(len(g) == len(best) && chainWorkCmp(g[0].ChainWork, best[0].ChainWork) == 0 && g[0].Height > best[0].Height) {
			best = g
		}
	}
	sort.Slice(best, func(i, j int) bool { return best[i].Node < best[j].Node })
	nodes := make([]string, 0, len(best))
	for _, o := range best {
		nodes = append(nodes, o.Node)
	}
	return best[0], nodes, nil
}

func collectNodeObservations(knownJSON string) ([]mobileNodeObservation, error) {
	candidates := mobileCandidates(knownJSON)
	obs := observeNodes(candidates)
	if len(obs) > 0 {
		return obs, nil
	}
	manifestPeers, manifestErr := fetchManifest()
	seen := map[string]bool{}
	for _, p := range candidates {
		seen[p] = true
	}
	for _, p := range manifestPeers {
		addMobileCandidate(&candidates, seen, p)
	}
	obs = observeNodes(candidates)
	if len(obs) > 0 {
		return obs, nil
	}
	if manifestErr != nil {
		return nil, fmt.Errorf("no reachable AuronQ Mainnet node; optional manifest: %v", manifestErr)
	}
	return nil, errors.New("no reachable AuronQ Mainnet node")
}

type headerCache struct {
	Version       int              `json:"version"`
	NetworkID     string           `json:"network_id"`
	VerifiedHeight uint64          `json:"verified_height"`
	VerifiedTip   string           `json:"verified_tip"`
	ChainWork     string           `json:"chain_work"`
	History       []aq.BlockHeader `json:"history"`
}

type headerVerification struct {
	Node      string `json:"node"`
	Height    uint64 `json:"height"`
	Tip       string `json:"tip"`
	ChainWork string `json:"chain_work"`
	HeadersChecked uint64 `json:"headers_checked"`
	Rebuilt   bool   `json:"rebuilt"`
}

func mainnetGenesisHeader() (aq.BlockHeader, error) {
	var h aq.BlockHeader
	if err := json.Unmarshal([]byte(mainnetGenesisHeaderJSON), &h); err != nil {
		return h, err
	}
	if h.Hash().String() != mainnetGenesisHash {
		return aq.BlockHeader{}, errors.New("embedded Mainnet genesis header mismatch")
	}
	return h, nil
}

func freshHeaderCache() (headerCache, error) {
	g, err := mainnetGenesisHeader()
	if err != nil {
		return headerCache{}, err
	}
	return headerCache{
		Version:        headerCacheVersion,
		NetworkID:      mainnetNetworkID,
		VerifiedHeight: 0,
		VerifiedTip:    g.Hash().String(),
		ChainWork:      aq.WorkForTarget(g.Target).Text(16),
		History:        []aq.BlockHeader{g},
	}, nil
}

func loadHeaderCache(path string) (headerCache, error) {
	fresh, err := freshHeaderCache()
	if err != nil {
		return headerCache{}, err
	}
	if strings.TrimSpace(path) == "" {
		return fresh, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fresh, nil
		}
		return headerCache{}, err
	}
	var h headerCache
	if json.Unmarshal(b, &h) != nil ||
		h.Version != headerCacheVersion ||
		h.NetworkID != mainnetNetworkID ||
		len(h.History) == 0 ||
		h.History[len(h.History)-1].Height != h.VerifiedHeight ||
		h.History[len(h.History)-1].Hash().String() != h.VerifiedTip {
		return fresh, nil
	}
	if _, ok := new(big.Int).SetString(h.ChainWork, 16); !ok {
		return fresh, nil
	}
	return h, nil
}

func saveHeaderCache(path string, h headerCache) error {
	if strings.TrimSpace(path) == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	b, err := json.Marshal(h)
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func fetchHeaderBatch(node string, start uint64, limit int) ([]aq.BlockHeader, error) {
	node = normalizeMobileNode(node)
	if node == "" {
		return nil, errors.New("invalid HTTPS AuronQ node")
	}
	if limit < 1 || limit > 256 {
		return nil, errors.New("invalid header batch limit")
	}
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/p2p/headers?start=%d&limit=%d", node, start, limit), nil)
	req.Header.Set("User-Agent", "AuronQ-Mobile/"+mobileVersion)
	resp, err := mobileHTTP(30 * time.Second).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("headers HTTP %s", resp.Status)
	}
	var body struct {
		Start   uint64           `json:"start"`
		Headers []aq.BlockHeader `json:"headers"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 2<<20)).Decode(&body); err != nil {
		return nil, err
	}
	if body.Start != start || len(body.Headers) > limit {
		return nil, errors.New("invalid header batch response")
	}
	return body.Headers, nil
}

func cachedTipCheck(cache headerCache, st aq.Status, at []aq.BlockHeader, fetchErr error) (bool, error) {
	if cache.VerifiedHeight > st.Height {
		return false, fmt.Errorf("peer is behind locally verified headers: peer=%d verified=%d", st.Height, cache.VerifiedHeight)
	}
	if fetchErr != nil {
		return false, fmt.Errorf("cannot confirm locally verified tip from this peer: %w", fetchErr)
	}
	if len(at) != 1 {
		return false, errors.New("peer did not return the locally verified tip header")
	}
	if at[0].Hash().String() != cache.VerifiedTip {
		return true, nil
	}
	return false, nil
}

func verifyHeaderChain(node, cachePath string) (headerVerification, error) {
	var out headerVerification
	node = normalizeMobileNode(node)
	st, err := statusFromNode(node)
	if err != nil {
		return out, err
	}
	cache, err := loadHeaderCache(cachePath)
	if err != nil {
		return out, err
	}
	rebuilt := false

	reset := func() error {
		fresh, err := freshHeaderCache()
		if err != nil {
			return err
		}
		cache = fresh
		rebuilt = true
		return nil
	}

	// Confirm that the remote chain still contains our locally verified tip.
	// Any reorg behind the cached tip triggers a full header replay from embedded
	// Mainnet genesis rather than trusting the remote rollback point.
	at, fetchErr := fetchHeaderBatch(node, cache.VerifiedHeight, 1)
	shouldReset, checkErr := cachedTipCheck(cache, st, at, fetchErr)
	if checkErr != nil {
		// A slow, temporarily unreachable, or lagging peer must never destroy a
		// valid local verification cache. The caller can immediately try another
		// observed peer instead of replaying hundreds of expensive AQM64 headers.
		return out, checkErr
	}
	if shouldReset {
		// A successfully fetched header at the same height disagrees with our
		// verified tip: this is a genuine chain-history mismatch/reorg signal.
		if err := reset(); err != nil {
			return out, err
		}
	}

	// A remote endpoint cannot replace the embedded genesis even if it lies in
	// /v1/status about Network ID.
	if cache.VerifiedHeight == 0 {
		gen, err := fetchHeaderBatch(node, 0, 1)
		if err != nil {
			return out, err
		}
		if len(gen) != 1 || gen[0].Hash().String() != mainnetGenesisHash {
			return out, errors.New("node does not serve the embedded AuronQ Mainnet genesis")
		}
	}

	work, ok := new(big.Int).SetString(cache.ChainWork, 16)
	if !ok {
		return out, errors.New("invalid local verified chain work")
	}
	checked := uint64(0)
	for next := cache.VerifiedHeight + 1; next <= st.Height; {
		remaining := st.Height - next + 1
		limit := headerBatchLimit
		if remaining < uint64(limit) {
			limit = int(remaining)
		}
		headers, err := fetchHeaderBatch(node, next, limit)
		if err != nil {
			return out, err
		}
		if len(headers) == 0 {
			return out, fmt.Errorf("node returned no headers at height %d", next)
		}
		for _, h := range headers {
			if h.Height != next {
				return out, fmt.Errorf("unexpected header height %d, expected %d", h.Height, next)
			}
			if err := aq.ValidateHeaderEnvelope(h, cache.History, time.Now().Unix()); err != nil {
				return out, fmt.Errorf("invalid AuronQ header %d: %w", h.Height, err)
			}
			work.Add(work, aq.WorkForTarget(h.Target))
			cache.History = append(cache.History, h)
			if len(cache.History) > headerCacheKeep {
				cache.History = append([]aq.BlockHeader(nil), cache.History[len(cache.History)-headerCacheKeep:]...)
			}
			cache.VerifiedHeight = h.Height
			cache.VerifiedTip = h.Hash().String()
			cache.ChainWork = work.Text(16)
			checked++
			next++
		}
		// Persist each successfully verified batch. A flaky peer or a process
		// restart can then resume from the last cryptographically verified height
		// instead of replaying AQM64 from genesis again.
		if err := saveHeaderCache(cachePath, cache); err != nil {
			return out, err
		}
	}

	if cache.VerifiedHeight != st.Height ||
		cache.VerifiedTip != st.Tip.String() ||
		!strings.EqualFold(cache.ChainWork, strings.TrimSpace(st.ChainWork)) {
		return out, errors.New("independently verified header chain does not match node status")
	}
	if err := saveHeaderCache(cachePath, cache); err != nil {
		return out, err
	}
	return headerVerification{
		Node: node, Height: cache.VerifiedHeight, Tip: cache.VerifiedTip,
		ChainWork: cache.ChainWork, HeadersChecked: checked, Rebuilt: rebuilt,
	}, nil
}

func QuorumSnapshotVerified(knownNodesJSON, headerCachePath string, recent int) (string, error) {
	obs, err := collectNodeObservations(knownNodesJSON)
	if err != nil {
		return "", err
	}
	selected, agreeing, err := chooseNodeQuorum(obs)
	if err != nil {
		return "", err
	}

	ordered := append([]string(nil), agreeing...)
	seen := map[string]bool{}
	for _, p := range ordered {
		seen[p] = true
	}
	// If the peer majority is invalid, independently try other observed chains.
	// Valid AQM64 headers outrank an unverified majority claim.
	sort.Slice(obs, func(i, j int) bool {
		cmp := chainWorkCmp(obs[i].ChainWork, obs[j].ChainWork)
		if cmp != 0 {
			return cmp > 0
		}
		return obs[i].Height > obs[j].Height
	})
	for _, o := range obs {
		if !seen[o.Node] {
			ordered = append(ordered, o.Node)
			seen[o.Node] = true
		}
	}

	var verified headerVerification
	var lastErr error
	for _, node := range ordered {
		verified, err = verifyHeaderChain(node, headerCachePath)
		if err == nil {
			break
		}
		lastErr = err
	}
	if err != nil {
		return "", fmt.Errorf("no observed peer supplied a valid independently verified AuronQ header chain: %v", lastErr)
	}

	// Agreement is recalculated against the chain we actually verified.
	agreeing = agreeing[:0]
	for _, o := range obs {
		if o.Height == verified.Height && o.Tip == verified.Tip && strings.EqualFold(o.ChainWork, verified.ChainWork) {
			agreeing = append(agreeing, o.Node)
		}
	}
	sort.Strings(agreeing)
	selected.Node = verified.Node
	raw, err := NetworkSnapshot(verified.Node, recent)
	if err != nil {
		return "", err
	}
	var snapshot map[string]any
	if err := json.Unmarshal([]byte(raw), &snapshot); err != nil {
		return "", err
	}
	snapshot["peer_observed"] = len(obs)
	snapshot["peer_agreement"] = len(agreeing)
	snapshot["multi_peer_confirmed"] = len(agreeing) >= 2
	snapshot["agreement_nodes"] = agreeing
	snapshot["observations"] = obs
	snapshot["header_verified"] = true
	snapshot["verified_height"] = verified.Height
	snapshot["verified_tip"] = verified.Tip
	snapshot["verified_chain_work"] = verified.ChainWork
	snapshot["headers_checked_now"] = verified.HeadersChecked
	snapshot["header_cache_rebuilt"] = verified.Rebuilt
	b, _ := json.Marshal(snapshot)
	return string(b), nil
}

func Version() string { return mobileVersion }
func MainnetNetworkID() string { return mainnetNetworkID }
func BootstrapURL() string { return bootstrapManifestURL }

func CreateWallet(path, password string) (string, error) {
	if len(password) < 12 {
		return "", errors.New("password must contain at least 12 characters")
	}
	if strings.TrimSpace(path) == "" {
		return "", errors.New("wallet path is empty")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return "", err
	}
	w, err := aq.NewWallet(path, password, aq.MainnetNetworkByte)
	if err != nil {
		return "", err
	}
	defer w.Close()
	return w.Address(), nil
}

func WalletAddress(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	var wf aq.WalletFile
	if err := json.Unmarshal(b, &wf); err != nil {
		return "", errors.New("invalid AuronQ wallet file")
	}
	if wf.Format != "auronq-wallet-v1" || wf.Scheme != aq.SchemeMLDSA87 {
		return "", errors.New("unsupported AuronQ wallet format")
	}
	if wf.NetworkByte != aq.MainnetNetworkByte {
		return "", errors.New("wallet belongs to another AuronQ network")
	}
	pub, err := hex.DecodeString(wf.PublicKey)
	if err != nil {
		return "", errors.New("wallet public key is invalid")
	}
	if len(pub) != aq.MLDSA87PublicKeySize {
		return "", errors.New("wallet public key has invalid length")
	}
	want := aq.AddressFromPub(pub, wf.Scheme, wf.NetworkByte)
	if want != wf.Address {
		return "", errors.New("wallet address/public key mismatch")
	}
	net, scheme, _, err := aq.DecodeAddress(wf.Address)
	if err != nil || net != aq.MainnetNetworkByte || scheme != aq.SchemeMLDSA87 {
		return "", errors.New("wallet address is not valid for AuronQ Mainnet")
	}
	return wf.Address, nil
}

func fetchManifest() ([]string, error) {
	cl := &http.Client{Timeout: 12 * time.Second}
	req, _ := http.NewRequest(http.MethodGet, bootstrapManifestURL, nil)
	req.Header.Set("User-Agent", "AuronQ-Mobile/"+mobileVersion)
	resp, err := cl.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bootstrap manifest HTTP %s", resp.Status)
	}
	var m bootstrapManifest
	if err := json.NewDecoder(io.LimitReader(resp.Body, 64<<10)).Decode(&m); err != nil {
		return nil, err
	}
	if !strings.EqualFold(strings.TrimSpace(m.NetworkID), mainnetNetworkID) {
		return nil, errors.New("bootstrap manifest Network ID mismatch")
	}
	if m.ExpiresAt > 0 && time.Now().Unix() > m.ExpiresAt {
		return nil, errors.New("bootstrap manifest expired")
	}
	if len(m.Peers) > 64 {
		return nil, errors.New("bootstrap manifest contains too many peers")
	}
	return m.Peers, nil
}

func checkNode(base string) error {
	_, err := statusFromNode(base)
	return err
}

func DiscoverNode() (string, error) {
	// Fresh installs first use the peer snapshot bundled into the APK. The GitHub
	// manifest is only an optional freshness source, not a runtime authority.
	// After the first successful contact, Android persists peers learned directly
	// from AuronQ /p2p/hello and tries those cached network peers before calling
	// this function again.
	manifestPeers, manifestErr := fetchManifest()
	seen := map[string]bool{}
	candidates := make([]string, 0, len(bundledBootstrapPeers)+len(manifestPeers))
	add := func(p string) {
		p = normalizeMobileNode(p)
		if p == "" || seen[p] {
			return
		}
		seen[p] = true
		candidates = append(candidates, p)
	}
	for _, p := range bundledBootstrapPeers {
		add(p)
	}
	for _, p := range manifestPeers {
		add(p)
	}

	// Rotate the starting point so one bundled address is not permanently favored.
	if len(candidates) > 1 {
		start := int(time.Now().UnixNano() % int64(len(candidates)))
		rotated := append([]string(nil), candidates[start:]...)
		rotated = append(rotated, candidates[:start]...)
		candidates = rotated
	}

	var lastErr error
	for _, p := range candidates {
		if err := checkNode(p); err == nil {
			return p, nil
		} else {
			lastErr = err
		}
	}
	if manifestErr != nil {
		return "", fmt.Errorf("no reachable AuronQ Mainnet node; optional manifest: %v; last node: %v", manifestErr, lastErr)
	}
	if lastErr != nil {
		return "", fmt.Errorf("no reachable AuronQ Mainnet node: %v", lastErr)
	}
	return "", errors.New("no reachable AuronQ Mainnet node")
}

func Status(nodeURL string) (string, error) {
	st, err := statusFromNode(nodeURL)
	if err != nil {
		return "", err
	}
	out := map[string]any{
		"network": st.Network,
		"network_id": st.NetworkID.String(),
		"height": st.Height,
		"tip": st.Tip.String(),
		"chain_work": st.ChainWork,
		"peers": st.Peers,
	}
	b, _ := json.Marshal(out)
	return string(b), nil
}

func QuorumSnapshot(knownNodesJSON string, recent int) (string, error) {
	obs, err := collectNodeObservations(knownNodesJSON)
	if err != nil {
		return "", err
	}
	selected, agreeing, err := chooseNodeQuorum(obs)
	if err != nil {
		return "", err
	}
	raw, err := NetworkSnapshot(selected.Node, recent)
	if err != nil {
		return "", err
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return "", err
	}
	out["peer_observed"] = len(obs)
	out["peer_agreement"] = len(agreeing)
	out["multi_peer_confirmed"] = len(agreeing) >= 2
	out["agreement_nodes"] = agreeing
	out["observations"] = obs
	b, _ := json.Marshal(out)
	return string(b), nil
}

func verifiedStateObservations(knownNodesJSON, verifiedTip, verifiedWork string, verifiedHeight int64) ([]mobileNodeObservation, error) {
	if verifiedHeight < 0 || strings.TrimSpace(verifiedTip) == "" || strings.TrimSpace(verifiedWork) == "" {
		return nil, errors.New("verified header state is missing")
	}
	obs, err := collectNodeObservations(knownNodesJSON)
	if err != nil {
		return nil, err
	}
	filtered := make([]mobileNodeObservation, 0, len(obs))
	for _, o := range obs {
		if o.Height == uint64(verifiedHeight) && o.Tip == verifiedTip && strings.EqualFold(o.ChainWork, verifiedWork) {
			filtered = append(filtered, o)
		}
	}
	if len(filtered) == 0 {
		return nil, errors.New("no reachable peer matches the independently verified header chain")
	}
	sort.Slice(filtered, func(i, j int) bool { return filtered[i].Node < filtered[j].Node })
	return filtered, nil
}

func historyFingerprint(items []aq.WalletHistoryItem) string {
	// Only confirmed/immature canonical history participates in state quorum.
	// Mempool propagation is intentionally asynchronous, so two honest peers may
	// temporarily disagree about pending transactions while sharing the exact
	// same canonical chain.
	cp := make([]aq.WalletHistoryItem, 0, len(items))
	for _, item := range items {
		if item.Status != "pending" {
			cp = append(cp, item)
		}
	}
	sort.Slice(cp, func(i, j int) bool {
		if cp[i].TXID != cp[j].TXID {
			return cp[i].TXID < cp[j].TXID
		}
		if cp[i].Status != cp[j].Status {
			return cp[i].Status < cp[j].Status
		}
		return cp[i].Type < cp[j].Type
	})
	b, _ := json.Marshal(cp)
	return string(b)
}

func mergeHistoryPending(group []historyObservation) []aq.WalletHistoryItem {
	if len(group) == 0 {
		return nil
	}
	confirmed := make([]aq.WalletHistoryItem, 0, len(group[0].Items))
	pending := map[string]aq.WalletHistoryItem{}
	for _, item := range group[0].Items {
		if item.Status != "pending" {
			confirmed = append(confirmed, item)
		}
	}
	for _, obs := range group {
		for _, item := range obs.Items {
			if item.Status == "pending" {
				pending[item.TXID] = item
			}
		}
	}
	pendingItems := make([]aq.WalletHistoryItem, 0, len(pending))
	for _, item := range pending {
		pendingItems = append(pendingItems, item)
	}
	sort.Slice(pendingItems, func(i, j int) bool {
		if pendingItems[i].Timestamp != pendingItems[j].Timestamp {
			return pendingItems[i].Timestamp > pendingItems[j].Timestamp
		}
		return pendingItems[i].TXID > pendingItems[j].TXID
	})
	out := make([]aq.WalletHistoryItem, 0, len(pendingItems)+len(confirmed))
	out = append(out, pendingItems...)
	out = append(out, confirmed...)
	return out
}

func utxoFingerprint(items []aq.UTXORecord) string {
	cp := append([]aq.UTXORecord(nil), items...)
	sort.Slice(cp, func(i, j int) bool {
		return cp[i].OutPoint.Key() < cp[j].OutPoint.Key()
	})
	b, _ := json.Marshal(cp)
	return string(b)
}

type historyObservation struct {
	Node  string
	Items []aq.WalletHistoryItem
	Key   string
}

func QuorumHistoryVerified(knownNodesJSON, address, verifiedTip, verifiedWork string, verifiedHeight int64, limit int) (string, error) {
	netByte, _, _, err := aq.DecodeAddress(strings.TrimSpace(address))
	if err != nil || netByte != aq.MainnetNetworkByte {
		return "", errors.New("invalid AuronQ Mainnet address")
	}
	if limit <= 0 {
		limit = 100
	}
	if limit > 250 {
		limit = 250
	}
	obs, err := verifiedStateObservations(knownNodesJSON, verifiedTip, verifiedWork, verifiedHeight)
	if err != nil {
		return "", err
	}
	results := make([]historyObservation, 0, len(obs))
	for _, o := range obs {
		cl := aq.NewClient(o.Node)
		cl.HTTP = mobileHTTP(8 * time.Second)
		items, err := cl.History(address, limit)
		if err == nil {
			results = append(results, historyObservation{Node: o.Node, Items: items, Key: historyFingerprint(items)})
		}
	}
	if len(results) == 0 {
		return "", errors.New("no verified-chain peer returned wallet history")
	}
	groups := map[string][]historyObservation{}
	for _, r := range results {
		groups[r.Key] = append(groups[r.Key], r)
	}
	var best []historyObservation
	for _, g := range groups {
		if len(g) > len(best) {
			best = g
		}
	}
	if len(results) >= 2 && len(best) < 2 {
		return "", errors.New("verified-chain peers returned conflicting wallet history")
	}
	nodes := make([]string, 0, len(best))
	for _, r := range best {
		nodes = append(nodes, r.Node)
	}
	sort.Strings(nodes)
	mergedItems := mergeHistoryPending(best)
	if len(mergedItems) > limit {
		mergedItems = mergedItems[:limit]
	}
	out := map[string]any{
		"address": strings.TrimSpace(address),
		"items": mergedItems,
		"peer_observed": len(results),
		"peer_agreement": len(best),
		"multi_peer_confirmed": len(best) >= 2,
		"agreement_nodes": nodes,
		"state_trust": "header-verified-peer-quorum",
		"pending_policy": "union-across-agreeing-chain-peers",
	}
	raw, _ := json.Marshal(out)
	return string(raw), nil
}

func verifiedUTXOQuorum(knownNodesJSON, address, verifiedTip, verifiedWork string, verifiedHeight int64) ([]aq.UTXORecord, []string, int, error) {
	netByte, _, _, err := aq.DecodeAddress(strings.TrimSpace(address))
	if err != nil || netByte != aq.MainnetNetworkByte {
		return nil, nil, 0, errors.New("invalid AuronQ Mainnet address")
	}
	obs, err := verifiedStateObservations(knownNodesJSON, verifiedTip, verifiedWork, verifiedHeight)
	if err != nil {
		return nil, nil, 0, err
	}
	type utxoObservation struct {
		Node string
		UTXO []aq.UTXORecord
		Key  string
	}
	results := make([]utxoObservation, 0, len(obs))
	for _, o := range obs {
		cl := aq.NewClient(o.Node)
		cl.HTTP = mobileHTTP(8 * time.Second)
		items, err := cl.UTXOs(address)
		if err == nil {
			results = append(results, utxoObservation{Node: o.Node, UTXO: items, Key: utxoFingerprint(items)})
		}
	}
	if len(results) == 0 {
		return nil, nil, 0, errors.New("no verified-chain peer returned wallet UTXOs")
	}
	groups := map[string][]utxoObservation{}
	for _, r := range results {
		groups[r.Key] = append(groups[r.Key], r)
	}
	var best []utxoObservation
	for _, g := range groups {
		if len(g) > len(best) {
			best = g
		}
	}
	if len(results) >= 2 && len(best) < 2 {
		return nil, nil, len(results), errors.New("verified-chain peers returned conflicting wallet UTXOs")
	}
	nodes := make([]string, 0, len(best))
	for _, r := range best {
		nodes = append(nodes, r.Node)
	}
	sort.Strings(nodes)
	return best[0].UTXO, nodes, len(results), nil
}

func SendMultiVerified(knownNodesJSON, walletPath, password, to, amount, verifiedTip, verifiedWork string, verifiedHeight int64) (string, error) {
	w, err := aq.LoadWallet(walletPath, password)
	if err != nil {
		return "", err
	}
	defer w.Close()
	if w.File.NetworkByte != aq.MainnetNetworkByte {
		return "", errors.New("wallet is not an AuronQ Mainnet wallet")
	}
	amt, err := aq.ParseAmount(strings.TrimSpace(amount))
	if err != nil {
		return "", err
	}
	utxos, agreeingNodes, observed, err := verifiedUTXOQuorum(
		knownNodesJSON, w.Address(), verifiedTip, verifiedWork, verifiedHeight,
	)
	if err != nil {
		return "", err
	}
	tx, fee, err := w.BuildTransaction(utxos, strings.TrimSpace(to), amt, aq.MainnetNetworkByte)
	if err != nil {
		return "", err
	}

	accepted := make([]string, 0, len(agreeingNodes))
	for _, node := range agreeingNodes {
		cl := aq.NewClient(node)
		cl.HTTP = mobileHTTP(10 * time.Second)
		if _, err := cl.SubmitTx(tx); err == nil {
			accepted = append(accepted, node)
		}
	}
	if len(accepted) == 0 {
		return "", errors.New("transaction was not accepted by any verified-chain peer")
	}
	out := map[string]any{
		"txid": tx.ID().String(),
		"fee_atoms": fee,
		"fee": aq.FormatAmount(fee),
		"broadcast_attempted": len(agreeingNodes),
		"direct_accepted": len(accepted),
		"accepted_nodes": accepted,
		"utxo_peer_observed": observed,
		"utxo_peer_agreement": len(agreeingNodes),
		"utxo_multi_peer_confirmed": len(agreeingNodes) >= 2,
		"state_trust": "header-verified-peer-quorum",
	}
	b, _ := json.Marshal(out)
	return string(b), nil
}

func QuorumBalanceVerified(knownNodesJSON, address string, verifiedTip, verifiedWork string, verifiedHeight int64) (string, error) {
	obs, err := verifiedStateObservations(knownNodesJSON, verifiedTip, verifiedWork, verifiedHeight)
	if err != nil {
		return "", err
	}
	nodes := make([]string, 0, len(obs))
	for _, o := range obs {
		nodes = append(nodes, o.Node)
	}
	return quorumBalanceFromNodes(nodes, address)
}

func quorumBalanceFromNodes(nodes []string, address string) (string, error) {
	netByte, _, _, err := aq.DecodeAddress(strings.TrimSpace(address))
	if err != nil || netByte != aq.MainnetNetworkByte {
		return "", errors.New("invalid AuronQ Mainnet address")
	}
	type balanceObservation struct {
		Node      string
		Spendable uint64
		Total     uint64
	}
	var balances []balanceObservation
	for _, node := range nodes {
		cl := aq.NewClient(node)
		cl.HTTP = mobileHTTP(8 * time.Second)
		res, err := cl.Balance(address)
		if err == nil {
			balances = append(balances, balanceObservation{Node: node, Spendable: res.Spendable, Total: res.Total})
		}
	}
	if len(balances) == 0 {
		return "", errors.New("no agreeing peer returned wallet balance")
	}
	groups := map[string][]balanceObservation{}
	for _, b := range balances {
		k := fmt.Sprintf("%d|%d", b.Spendable, b.Total)
		groups[k] = append(groups[k], b)
	}
	best := balances[:1]
	for _, g := range groups {
		if len(g) > len(best) {
			best = g
		}
	}
	if len(balances) >= 2 && len(best) < 2 {
		return "", errors.New("agreeing chain peers returned conflicting wallet balances")
	}
	v := best[0]
	agreeingNodes := make([]string, 0, len(best))
	for _, b := range best {
		agreeingNodes = append(agreeingNodes, b.Node)
	}
	sort.Strings(agreeingNodes)
	out := map[string]any{
		"address": strings.TrimSpace(address),
		"spendable_atoms": v.Spendable,
		"total_atoms": v.Total,
		"spendable": aq.FormatAmount(v.Spendable),
		"total": aq.FormatAmount(v.Total),
		"peer_observed": len(balances),
		"peer_agreement": len(best),
		"multi_peer_confirmed": len(best) >= 2,
		"agreement_nodes": agreeingNodes,
	}
	raw, _ := json.Marshal(out)
	return string(raw), nil
}

func QuorumBalance(knownNodesJSON, address string) (string, error) {
	obs, err := collectNodeObservations(knownNodesJSON)
	if err != nil {
		return "", err
	}
	_, agreeing, err := chooseNodeQuorum(obs)
	if err != nil {
		return "", err
	}
	return quorumBalanceFromNodes(agreeing, address)
}

func Balance(nodeURL, address string) (string, error) {
	net, _, _, err := aq.DecodeAddress(strings.TrimSpace(address))
	if err != nil || net != aq.MainnetNetworkByte {
		return "", errors.New("invalid AuronQ Mainnet address")
	}
	if err := checkNode(nodeURL); err != nil {
		return "", err
	}
	res, err := aq.NewClient(strings.TrimRight(nodeURL, "/")).Balance(address)
	if err != nil {
		return "", err
	}
	out := map[string]any{
		"address": res.Address,
		"spendable_atoms": res.Spendable,
		"total_atoms": res.Total,
		"spendable": aq.FormatAmount(res.Spendable),
		"total": aq.FormatAmount(res.Total),
	}
	b, _ := json.Marshal(out)
	return string(b), nil
}

func History(nodeURL, address string, limit int) (string, error) {
	net, _, _, err := aq.DecodeAddress(strings.TrimSpace(address))
	if err != nil || net != aq.MainnetNetworkByte {
		return "", errors.New("invalid AuronQ Mainnet address")
	}
	if err := checkNode(nodeURL); err != nil {
		return "", err
	}
	items, err := aq.NewClient(strings.TrimRight(nodeURL, "/")).History(address, limit)
	if err != nil {
		return "", err
	}
	out := map[string]any{"address": address, "items": items}
	b, _ := json.Marshal(out)
	return string(b), nil
}


func Send(nodeURL, walletPath, password, to, amount string) (string, error) {
	if err := checkNode(nodeURL); err != nil {
		return "", err
	}
	w, err := aq.LoadWallet(walletPath, password)
	if err != nil {
		return "", err
	}
	defer w.Close()
	if w.File.NetworkByte != aq.MainnetNetworkByte {
		return "", errors.New("wallet is not an AuronQ Mainnet wallet")
	}
	amt, err := aq.ParseAmount(strings.TrimSpace(amount))
	if err != nil {
		return "", err
	}
	client := aq.NewClient(strings.TrimRight(nodeURL, "/"))
	utxos, err := client.UTXOs(w.Address())
	if err != nil {
		return "", err
	}
	tx, fee, err := w.BuildTransaction(utxos, strings.TrimSpace(to), amt, aq.MainnetNetworkByte)
	if err != nil {
		return "", err
	}
	res, err := client.SubmitTx(tx)
	if err != nil {
		return "", err
	}
	out := map[string]any{
		"txid": res.TXID.String(),
		"fee_atoms": fee,
		"fee": aq.FormatAmount(fee),
	}
	b, _ := json.Marshal(out)
	return string(b), nil
}


func SendMulti(knownNodesJSON, nodeURL, walletPath, password, to, amount string) (string, error) {
	nodeURL = normalizeMobileNode(nodeURL)
	if nodeURL == "" {
		return "", errors.New("primary AuronQ node is missing")
	}
	if err := checkNode(nodeURL); err != nil {
		return "", err
	}
	w, err := aq.LoadWallet(walletPath, password)
	if err != nil {
		return "", err
	}
	defer w.Close()
	if w.File.NetworkByte != aq.MainnetNetworkByte {
		return "", errors.New("wallet is not an AuronQ Mainnet wallet")
	}
	amt, err := aq.ParseAmount(strings.TrimSpace(amount))
	if err != nil {
		return "", err
	}
	primary := aq.NewClient(nodeURL)
	primary.HTTP = mobileHTTP(12 * time.Second)
	utxos, err := primary.UTXOs(w.Address())
	if err != nil {
		return "", err
	}
	tx, fee, err := w.BuildTransaction(utxos, strings.TrimSpace(to), amt, aq.MainnetNetworkByte)
	if err != nil {
		return "", err
	}
	res, err := primary.SubmitTx(tx)
	if err != nil {
		return "", err
	}
	accepted := []string{nodeURL}
	attempted := 1

	seen := map[string]bool{nodeURL: true}
	candidates := mobileCandidates(knownNodesJSON)
	for _, p := range candidates {
		p = normalizeMobileNode(p)
		if p == "" || seen[p] {
			continue
		}
		seen[p] = true
		if attempted >= 8 {
			break
		}
		attempted++
		if _, err := statusFromNode(p); err != nil {
			continue
		}
		cl := aq.NewClient(p)
		cl.HTTP = mobileHTTP(8 * time.Second)
		if _, err := cl.SubmitTx(tx); err == nil {
			accepted = append(accepted, p)
		}
	}
	out := map[string]any{
		"txid": res.TXID.String(),
		"fee_atoms": fee,
		"fee": aq.FormatAmount(fee),
		"broadcast_attempted": attempted,
		"direct_accepted": len(accepted),
		"accepted_nodes": accepted,
	}
	b, _ := json.Marshal(out)
	return string(b), nil
}

func NetworkSnapshot(nodeURL string, recent int) (string, error) {
	nodeURL = strings.TrimRight(strings.TrimSpace(nodeURL), "/")
	// A three-peer network preview must not fetch a full block.
	// Wallet history and transaction amounts remain separate guarded reads.
	if recent < 0 {
		recent = 0
	}
	if recent > 12 {
		recent = 12
	}
	st, err := statusFromNode(nodeURL)
	if err != nil {
		return "", err
	}
	type recentBlock struct {
		Height       uint64 `json:"height"`
		Hash         string `json:"hash"`
		Timestamp    int64  `json:"timestamp"`
		TimeISO      string `json:"time_iso"`
		Transactions int    `json:"transactions"`
	}
	blocks := make([]recentBlock, 0, recent)
	client := &http.Client{Timeout: 8 * time.Second}
	for i := 0; i < recent; i++ {
		if st.Height < uint64(i) {
			break
		}
		h := st.Height - uint64(i)
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/p2p/getblock?height=%d", nodeURL, h), nil)
		req.Header.Set("User-Agent", "AuronQ-Mobile/"+mobileVersion)
		resp, err := client.Do(req)
		if err != nil {
			break
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			break
		}
		var b aq.Block
		err = json.NewDecoder(io.LimitReader(resp.Body, int64(aq.MaxBlockBytes)+64*1024)).Decode(&b)
		resp.Body.Close()
		if err != nil {
			break
		}
		blocks = append(blocks, recentBlock{
			Height:       b.Header.Height,
			Hash:         b.Hash().String(),
			Timestamp:    b.Header.Timestamp,
			TimeISO:      time.Unix(b.Header.Timestamp, 0).UTC().Format(time.RFC3339),
			Transactions: len(b.Transactions),
		})
	}
	out := map[string]any{
		"network":      st.Network,
		"network_id":   st.NetworkID.String(),
		"height":       st.Height,
		"tip":          st.Tip.String(),
		"chain_work":   st.ChainWork,
		"network_hashrate": st.NetworkHashrate,
		"issued_atoms": st.Issued,
		"issued_coins": st.IssuedCoins,
		"mempool":      st.Mempool,
		"peers":        st.Peers,
		"node":         nodeURL,
		"blocks":       blocks,
		"observed_at":  time.Now().UTC().Format(time.RFC3339),
	}
	b, _ := json.Marshal(out)
	return string(b), nil
}


func PeerCandidates(nodeURL string) (string, error) {
	nodeURL = strings.TrimRight(strings.TrimSpace(nodeURL), "/")
	if err := checkNode(nodeURL); err != nil {
		return "", err
	}
	cl := &http.Client{Timeout: 8 * time.Second}
	req, _ := http.NewRequest(http.MethodGet, nodeURL+"/p2p/hello", nil)
	req.Header.Set("User-Agent", "AuronQ-Mobile/"+mobileVersion)
	resp, err := cl.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("peer hello HTTP %s", resp.Status)
	}
	var hello aq.Hello
	if err := json.NewDecoder(io.LimitReader(resp.Body, 64<<10)).Decode(&hello); err != nil {
		return "", err
	}
	if hello.ProtocolVersion != 1 || hello.NetworkID.String() != mainnetNetworkID {
		return "", errors.New("peer hello Network ID mismatch")
	}
	seen := map[string]bool{}
	out := make([]string, 0, len(hello.Peers)+2)
	add := func(p string) {
		p = normalizeMobileNode(p)
		if p == "" || seen[p] {
			return
		}
		seen[p] = true
		out = append(out, p)
	}
	add(nodeURL)
	add(hello.Advertise)
	for _, p := range hello.Peers {
		add(p)
		if len(out) >= 32 {
			break
		}
	}
	b, _ := json.Marshal(out)
	return string(b), nil
}
