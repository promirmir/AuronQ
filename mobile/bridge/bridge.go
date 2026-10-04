package bridge

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	aq "auronq/internal/auronq"
)

const (
	mobileVersion       = "0.4.2-alpha"
	mainnetNetworkID    = "44e62c2ace002a6660c14e252173c1aa303529c68e40c998e92da2b453f44f30b1e58c94d533587e2186004593fb856c433fcdb5418ed430ec8617e29529365c"
	bootstrapManifestURL = "https://raw.githubusercontent.com/promirmir/AuronQ/main/bootstrap.json"
)

var bundledBootstrapPeers = []string{
	"https://mir.taild63f46.ts.net",
	"https://desktop-4nifg1j.taild63f46.ts.net",
}

type bootstrapManifest struct {
	NetworkID string   `json:"network_id"`
	Peers     []string `json:"peers"`
	ExpiresAt int64    `json:"expires_at,omitempty"`
}

const (
	maxMobileKnownCandidates = 8
	maxMobileProbeCandidates = 16
)

type mobileNodeObservation struct {
	Node      string `json:"node"`
	Height    uint64 `json:"height"`
	Tip       string `json:"tip"`
	ChainWork string `json:"chain_work"`
	Peers     int    `json:"peers"`
}

func normalizeMobileNode(raw string) string {
	raw = strings.TrimRight(strings.TrimSpace(raw), "/")
	if raw == "" || !strings.HasPrefix(raw, "https://") {
		return ""
	}
	return raw
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
	var wg sync.WaitGroup
	for _, node := range candidates {
		node := node
		wg.Add(1)
		go func() {
			defer wg.Done()
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
	wg.Wait()
	close(ch)
	out := make([]mobileNodeObservation, 0, len(candidates))
	for r := range ch {
		if r.ok {
			out = append(out, r.obs)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Node < out[j].Node })
	return out
}

func mobileStateKey(o mobileNodeObservation) string {
	return fmt.Sprintf("%d|%s|%s", o.Height, o.Tip, o.ChainWork)
}

func chainWorkCmp(a, b string) int {
	aa, okA := new(big.Int).SetString(strings.TrimSpace(a), 10)
	bb, okB := new(big.Int).SetString(strings.TrimSpace(b), 10)
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
	if len(obs) >= 2 {
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
		p = strings.TrimRight(strings.TrimSpace(p), "/")
		if p == "" || seen[p] || !strings.HasPrefix(p, "https://") {
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

func QuorumBalance(knownNodesJSON, address string) (string, error) {
	netByte, _, _, err := aq.DecodeAddress(strings.TrimSpace(address))
	if err != nil || netByte != aq.MainnetNetworkByte {
		return "", errors.New("invalid AuronQ Mainnet address")
	}
	obs, err := collectNodeObservations(knownNodesJSON)
	if err != nil {
		return "", err
	}
	_, agreeing, err := chooseNodeQuorum(obs)
	if err != nil {
		return "", err
	}
	type balanceObservation struct {
		Node      string
		Spendable uint64
		Total     uint64
	}
	var balances []balanceObservation
	for _, node := range agreeing {
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
	v := best[0]
	nodes := make([]string, 0, len(best))
	for _, b := range best {
		nodes = append(nodes, b.Node)
	}
	sort.Strings(nodes)
	out := map[string]any{
		"address": v.Node,
		"spendable_atoms": v.Spendable,
		"total_atoms": v.Total,
		"spendable": aq.FormatAmount(v.Spendable),
		"total": aq.FormatAmount(v.Total),
		"peer_observed": len(balances),
		"peer_agreement": len(best),
		"multi_peer_confirmed": len(best) >= 2,
		"agreement_nodes": nodes,
	}
	out["address"] = strings.TrimSpace(address)
	raw, _ := json.Marshal(out)
	return string(raw), nil
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
	for _, p := range bundledBootstrapPeers {
		addMobileCandidate(&candidates, seen, p)
	}
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
	if recent < 1 {
		recent = 1
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
		p = strings.TrimRight(strings.TrimSpace(p), "/")
		if p == "" || seen[p] || !strings.HasPrefix(p, "https://") {
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
