package bridge

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	aq "auronq/internal/auronq"
)

const (
	mobileVersion       = "0.3.0-alpha"
	mainnetNetworkID    = "44e62c2ace002a6660c14e252173c1aa303529c68e40c998e92da2b453f44f30b1e58c94d533587e2186004593fb856c433fcdb5418ed430ec8617e29529365c"
	bootstrapManifestURL = "https://raw.githubusercontent.com/promirmir/AuronQ/main/bootstrap.json"
	fallbackBootstrap    = "https://mir.taild63f46.ts.net"
)

type bootstrapManifest struct {
	NetworkID string   `json:"network_id"`
	Peers     []string `json:"peers"`
	ExpiresAt int64    `json:"expires_at,omitempty"`
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
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	if !strings.HasPrefix(base, "https://") {
		return errors.New("mobile wallet requires an HTTPS node")
	}
	cl := aq.NewClient(base)
	st, err := cl.Status()
	if err != nil {
		return err
	}
	if st.NetworkID.String() != mainnetNetworkID {
		return errors.New("node Network ID mismatch")
	}
	return nil
}

func DiscoverNode() (string, error) {
	peers, manifestErr := fetchManifest()
	seen := map[string]bool{}
	candidates := make([]string, 0, len(peers)+1)
	for _, p := range peers {
		p = strings.TrimRight(strings.TrimSpace(p), "/")
		if p != "" && !seen[p] {
			seen[p] = true
			candidates = append(candidates, p)
		}
	}
	if !seen[fallbackBootstrap] {
		candidates = append(candidates, fallbackBootstrap)
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
		return "", fmt.Errorf("no reachable AuronQ Mainnet node; manifest: %v; last node: %v", manifestErr, lastErr)
	}
	if lastErr != nil {
		return "", fmt.Errorf("no reachable AuronQ Mainnet node: %v", lastErr)
	}
	return "", errors.New("no reachable AuronQ Mainnet node")
}

func Status(nodeURL string) (string, error) {
	if err := checkNode(nodeURL); err != nil {
		return "", err
	}
	st, err := aq.NewClient(strings.TrimRight(nodeURL, "/")).Status()
	if err != nil {
		return "", err
	}
	out := map[string]any{
		"network": st.Network,
		"network_id": st.NetworkID.String(),
		"height": st.Height,
		"tip": st.Tip.String(),
		"peers": st.Peers,
	}
	b, _ := json.Marshal(out)
	return string(b), nil
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


func NetworkSnapshot(nodeURL string, recent int) (string, error) {
	nodeURL = strings.TrimRight(strings.TrimSpace(nodeURL), "/")
	if recent < 1 {
		recent = 1
	}
	if recent > 12 {
		recent = 12
	}
	if err := checkNode(nodeURL); err != nil {
		return "", err
	}
	st, err := aq.NewClient(nodeURL).Status()
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
