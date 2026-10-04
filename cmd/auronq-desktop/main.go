package main

import (
	"context"
	"crypto/rand"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	aq "auronq/internal/auronq"
)

const desktopVersion = "1.7.7"

//go:embed web/*
var webFS embed.FS

type walletView struct {
	Name          string `json:"name"`
	Address       string `json:"address"`
	CreatedAt     int64  `json:"created_at"`
	Spendable     uint64 `json:"spendable_atoms"`
	Total         uint64 `json:"total_atoms"`
	SpendableText string `json:"spendable"`
	TotalText     string `json:"total"`
	Error         string `json:"error,omitempty"`
}

type minerState struct {
	Running     bool    `json:"running"`
	Threads     int     `json:"threads"`
	Wallet      string  `json:"wallet"`
	Address     string  `json:"address"`
	Hashrate    float64 `json:"hashrate"`
	Height      uint64  `json:"height"`
	BlocksFound uint64  `json:"blocks_found"`
	LastBlock   string  `json:"last_block,omitempty"`
	LastError   string  `json:"last_error,omitempty"`
	StartedAt   int64   `json:"started_at,omitempty"`
}

type appState struct {
	Version          string       `json:"version"`
	DataDir          string       `json:"data_dir"`
	NetworkLoaded    bool         `json:"network_loaded"`
	NetworkName      string       `json:"network_name,omitempty"`
	NetworkID        string       `json:"network_id,omitempty"`
	Founder          string       `json:"founder_address,omitempty"`
	NodeRunning      bool         `json:"node_running"`
	NodeError        string       `json:"node_error,omitempty"`
	Height           uint64       `json:"height"`
	Tip              string       `json:"tip,omitempty"`
	Peers            int          `json:"peers"`
	NetworkHashrate  float64      `json:"network_hashrate"`
	Mempool          int          `json:"mempool"`
	Issued           string       `json:"issued,omitempty"`
	CoinbaseMaturity uint64       `json:"coinbase_maturity"`
	BootstrapPeers   []string     `json:"bootstrap_peers"`
	Wallets          []walletView `json:"wallets"`
	Miner            minerState   `json:"miner"`
	CPUCount         int          `json:"cpu_count"`
	Logs             []string     `json:"logs"`
	WindowsCrypto    string       `json:"windows_crypto"`
	CryptoOK         bool         `json:"crypto_ok"`
	CryptoError      string       `json:"crypto_error,omitempty"`
}

type App struct {
	mu sync.RWMutex

	baseDir      string
	walletsDir   string
	nodeDir      string
	networkPath  string
	peerBookPath string
	network      *aq.NetworkConfig
	peerBook     map[string][]string

	node       *aq.Node
	chain      *aq.Chain
	nodeCancel context.CancelFunc
	nodeRun    bool
	nodeErr    string
	nodeURL    string

	miner       minerState
	minerCancel context.CancelFunc

	logs        []string
	token       string
	exit        chan struct{}
	cryptoOK    bool
	cryptoError string
}

func main() {
	log.SetFlags(log.LstdFlags | log.LUTC)
	app, err := newApp()
	if err != nil {
		log.Fatal(err)
	}
	defer app.shutdown()

	mux := http.NewServeMux()
	app.routes(mux)
	ln, err := net.Listen("tcp", "127.0.0.1:18445")
	if err != nil {
		// AuronQ Desktop intentionally keeps the node/miner alive if the UI
		// window is closed. Starting the EXE again reopens that existing instance.
		if reopenExisting("http://127.0.0.1:18445") {
			return
		}
		log.Fatal(err)
	}
	srv := &http.Server{Handler: securityHeaders(mux), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 30 * time.Second, WriteTimeout: 60 * time.Second, IdleTimeout: 60 * time.Second}
	url := "http://" + ln.Addr().String() + "/"
	go func() {
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			app.addLog("Błąd GUI HTTP: " + err.Error())
		}
	}()

	if app.network != nil {
		if err := app.startNode(); err != nil {
			app.addLog("Node nie wystartował: " + err.Error())
		} else {
			app.startLANDiscovery()
		}
	}
	if err := openDesktopWindow(url); err != nil {
		app.addLog("Nie udało się automatycznie otworzyć okna: " + err.Error())
	}

	<-app.exit
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}

func mergeUniqueStrings(dst []string, src []string) ([]string, bool) {
	out := append([]string(nil), dst...)
	seen := make(map[string]struct{}, len(out)+len(src))
	for _, v := range out {
		seen[v] = struct{}{}
	}
	changed := false
	for _, v := range src {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
		changed = true
	}
	return out, changed
}

func mergeBootstrapMetadata(existing, bundled *aq.NetworkConfig) bool {
	if existing == nil || bundled == nil || existing.NetworkID() != bundled.NetworkID() {
		return false
	}
	changed := false
	var c bool
	existing.SeedPeers, c = mergeUniqueStrings(existing.SeedPeers, bundled.SeedPeers)
	changed = changed || c
	existing.DNSSeeds, c = mergeUniqueStrings(existing.DNSSeeds, bundled.DNSSeeds)
	changed = changed || c
	existing.BootstrapManifests, c = mergeUniqueStrings(existing.BootstrapManifests, bundled.BootstrapManifests)
	changed = changed || c
	return changed
}

func newApp() (*App, error) {
	cfg, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	base := filepath.Join(cfg, "AuronQ")
	a := &App{baseDir: base, walletsDir: filepath.Join(base, "wallets"), nodeDir: filepath.Join(base, "nodes"), networkPath: filepath.Join(base, "network.json"), peerBookPath: filepath.Join(base, "peers.json"), peerBook: map[string][]string{}, nodeURL: "http://127.0.0.1:18444", exit: make(chan struct{}, 1)}
	for _, d := range []string{a.baseDir, a.walletsDir, a.nodeDir} {
		if err := os.MkdirAll(d, 0700); err != nil {
			return nil, err
		}
	}
	// Zero-touch public release bootstrap. A clean installation imports the
	// validated release-side network.json. On upgrades, immutable consensus
	// parameters are never replaced; only bootstrap metadata from a bundled file
	// with the exact same Network ID is merged into the existing configuration.
	// This lets an older installation learn newly published official bootstrap
	// manifests without deleting wallets, chain data, or user-configured peers.
	if exe, e := os.Executable(); e == nil {
		bundled := filepath.Join(filepath.Dir(exe), "network.json")
		if bundledNet, e := aq.LoadNetwork(bundled); e == nil {
			if _, statErr := os.Stat(a.networkPath); os.IsNotExist(statErr) {
				if b, readErr := os.ReadFile(bundled); readErr == nil {
					tmp := a.networkPath + ".tmp"
					if writeErr := os.WriteFile(tmp, b, 0600); writeErr == nil {
						if renameErr := os.Rename(tmp, a.networkPath); renameErr != nil {
							_ = os.Remove(tmp)
						}
					}
				}
			} else if existingNet, loadErr := aq.LoadNetwork(a.networkPath); loadErr == nil && existingNet.NetworkID() == bundledNet.NetworkID() {
				if mergeBootstrapMetadata(existingNet, bundledNet) {
					_ = aq.SaveNetwork(a.networkPath, existingNet)
				}
			}
		}
	}
	if b, e := os.ReadFile(a.peerBookPath); e == nil {
		_ = json.Unmarshal(b, &a.peerBook)
		if a.peerBook == nil {
			a.peerBook = map[string][]string{}
		}
	}
	var tb [32]byte
	if _, err := rand.Read(tb[:]); err != nil {
		return nil, err
	}
	a.token = hex.EncodeToString(tb[:])
	if _, err := os.Stat(a.networkPath); err == nil {
		n, e := aq.LoadNetwork(a.networkPath)
		if e != nil {
			a.nodeErr = "network.json: " + e.Error()
		} else {
			a.network = n
		}
	}
	if a.network != nil {
		a.loadBundledBootstrap()
	}
	if ok, e := cryptoSelfTest(); e != nil {
		a.cryptoError = e.Error()
		a.addLog("Błąd testu ML-DSA-87: " + e.Error())
	} else {
		a.cryptoOK = ok
		a.addLog("ML-DSA-87 self-test: OK")
	}
	a.addLog("AuronQ Desktop uruchomiony")
	return a, nil
}

func (a *App) loadBundledBootstrap() {
	exe, err := os.Executable()
	if err != nil {
		return
	}
	path := filepath.Join(filepath.Dir(exe), "bootstrap.json")
	b, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var m aq.BootstrapManifest
	dec := json.NewDecoder(strings.NewReader(string(b)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&m); err != nil {
		a.addLog("bootstrap.json odrzucony: " + err.Error())
		return
	}
	var trailing any
	if err := dec.Decode(&trailing); err != io.EOF {
		a.addLog("bootstrap.json odrzucony: dodatkowe dane JSON")
		return
	}
	n := a.currentNetwork()
	if n == nil || m.NetworkID != n.NetworkID() {
		a.addLog("bootstrap.json odrzucony: inny Network ID")
		return
	}
	if m.ExpiresAt != 0 && time.Now().Unix() > m.ExpiresAt {
		a.addLog("bootstrap.json wygasł")
		return
	}
	id := n.NetworkID().String()
	a.mu.RLock()
	existing := append([]string(nil), a.peerBook[id]...)
	a.mu.RUnlock()
	seen := map[string]bool{}
	merged := make([]string, 0, len(existing)+len(m.Peers))
	for _, p := range existing {
		if q, err := normalizeDesktopPeer(p); err == nil && q != "" && !seen[q] {
			seen[q] = true
			merged = append(merged, q)
		}
	}
	for _, p := range m.Peers {
		if len(merged) >= 64 {
			break
		}
		q, err := normalizeDesktopPeer(p)
		if err != nil || q == "" || seen[q] {
			continue
		}
		seen[q] = true
		merged = append(merged, q)
	}
	sort.Strings(merged)
	a.mu.Lock()
	a.peerBook[id] = merged
	a.mu.Unlock()
	if err := a.savePeerBook(); err != nil {
		a.addLog("Nie zapisano bootstrap peerów: " + err.Error())
		return
	}
	a.addLog(fmt.Sprintf("bootstrap.json: %d peerów startowych", len(merged)))
}

func (a *App) addLog(s string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.logs = append(a.logs, time.Now().Format("15:04:05")+"  "+s)
	if len(a.logs) > 200 {
		a.logs = a.logs[len(a.logs)-200:]
	}
}
func (a *App) shutdown() { a.stopMiner(); a.stopNode() }

func (a *App) routes(mux *http.ServeMux) {
	mux.HandleFunc("/", a.handleIndex)
	mux.HandleFunc("/api/state", a.handleState)
	mux.HandleFunc("/api/network/import", a.guard(a.handleNetworkImport))
	mux.HandleFunc("/api/network/peers", a.guard(a.handleNetworkPeers))
	mux.HandleFunc("/api/node/start", a.guard(a.handleNodeStart))
	mux.HandleFunc("/api/node/stop", a.guard(a.handleNodeStop))
	mux.HandleFunc("/api/wallet/create", a.guard(a.handleWalletCreate))
	mux.HandleFunc("/api/wallet/import", a.guard(a.handleWalletImport))
	mux.HandleFunc("/api/wallet/delete", a.guard(a.handleWalletDelete))
	mux.HandleFunc("/api/wallet/export", a.handleWalletExport)
	mux.HandleFunc("/api/wallet/history", a.handleWalletHistory)
	mux.HandleFunc("/api/send", a.guard(a.handleSend))
	mux.HandleFunc("/api/miner/start", a.guard(a.handleMinerStart))
	mux.HandleFunc("/api/miner/stop", a.guard(a.handleMinerStop))
	mux.HandleFunc("/api/open-data", a.guard(a.handleOpenData))
	mux.HandleFunc("/api/exit", a.guard(a.handleExit))
}

func (a *App) guard(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "method", 405)
			return
		}
		if r.Header.Get("X-AuronQ-Token") != a.token {
			http.Error(w, "forbidden", 403)
			return
		}
		next(w, r)
	}
}
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self'; frame-src http://127.0.0.1:18444; frame-ancestors 'none'")
		next.ServeHTTP(w, r)
	})
}

func (a *App) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	b, err := webFS.ReadFile("web/index.html")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	s := strings.ReplaceAll(string(b), "__AURONQ_TOKEN__", a.token)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	io.WriteString(w, s)
}
func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
func apiError(w http.ResponseWriter, code int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": err.Error()})
}
func readBody(r *http.Request, v any) error {
	d := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	d.DisallowUnknownFields()
	return d.Decode(v)
}

func (a *App) currentNetwork() *aq.NetworkConfig {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.network
}

func (a *App) state() appState {
	a.mu.RLock()
	n := a.network
	nodeRun := a.nodeRun
	nodeErr := a.nodeErr
	miner := a.miner
	logs := append([]string(nil), a.logs...)
	cryptoOK := a.cryptoOK
	cryptoError := a.cryptoError
	a.mu.RUnlock()
	st := appState{Version: desktopVersion, DataDir: a.baseDir, NetworkLoaded: n != nil, NodeRunning: nodeRun, NodeError: nodeErr, Miner: miner, CPUCount: runtime.NumCPU(), Logs: logs, WindowsCrypto: "ML-DSA-87 portable (FIPS 204)", CryptoOK: cryptoOK, CryptoError: cryptoError}
	if n != nil {
		st.NetworkName = n.Name
		st.NetworkID = n.NetworkID().String()
		st.Founder = n.FounderAddress
		st.CoinbaseMaturity = n.Maturity()
		a.mu.RLock()
		st.BootstrapPeers = append([]string(nil), a.peerBook[n.NetworkID().String()]...)
		a.mu.RUnlock()
	}
	if nodeRun {
		if x, err := aq.NewClient(a.nodeURL).Status(); err == nil {
			st.Height = x.Height
			st.Tip = x.Tip.String()
			st.Peers = x.Peers
			st.Mempool = x.Mempool
			st.Issued = aq.FormatAmount(x.Issued)
			st.NetworkHashrate = x.NetworkHashrate
		} else if st.NodeError == "" {
			st.NodeError = err.Error()
		}
	}
	st.Wallets = a.walletViews(nodeRun)
	return st
}
func (a *App) handleState(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "method", 405)
		return
	}
	writeJSON(w, a.state())
}

func (a *App) walletViews(withBalance bool) []walletView {
	files, _ := filepath.Glob(filepath.Join(a.walletsDir, "*.wallet"))
	sort.Strings(files)
	out := make([]walletView, 0, len(files))
	cl := aq.NewClient(a.nodeURL)
	for _, p := range files {
		b, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		var wf aq.WalletFile
		if json.Unmarshal(b, &wf) != nil {
			continue
		}
		if n := a.currentNetwork(); n != nil {
			wn, _, _, e := aq.DecodeAddress(wf.Address)
			if e != nil || wn != n.NetworkByte {
				continue
			}
		}
		v := walletView{Name: strings.TrimSuffix(filepath.Base(p), ".wallet"), Address: wf.Address, CreatedAt: wf.CreatedAt}
		if withBalance {
			bal, e := cl.Balance(wf.Address)
			if e != nil {
				v.Error = e.Error()
			} else {
				v.Spendable = bal.Spendable
				v.Total = bal.Total
				v.SpendableText = aq.FormatAmount(bal.Spendable)
				v.TotalText = aq.FormatAmount(bal.Total)
			}
		}
		out = append(out, v)
	}
	return out
}
func safeName(s string) string {
	s = strings.TrimSpace(s)
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
		}
	}
	x := b.String()
	if len(x) > 48 {
		x = x[:48]
	}
	return x
}
func (a *App) walletPath(name string) (string, error) {
	name = safeName(name)
	if name == "" {
		return "", errors.New("nieprawidłowa nazwa portfela")
	}
	return filepath.Join(a.walletsDir, name+".wallet"), nil
}

func (a *App) exactWalletPath(name string) (string, string, error) {
	raw := strings.TrimSpace(name)
	if raw == "" || safeName(raw) != raw {
		return "", "", errors.New("nieprawidłowa nazwa portfela")
	}
	p := filepath.Join(a.walletsDir, raw+".wallet")
	baseAbs, err := filepath.Abs(a.walletsDir)
	if err != nil {
		return "", "", err
	}
	pathAbs, err := filepath.Abs(p)
	if err != nil {
		return "", "", err
	}
	rel, err := filepath.Rel(baseAbs, pathAbs)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) || filepath.IsAbs(rel) {
		return "", "", errors.New("nieprawidłowa ścieżka portfela")
	}
	return raw, pathAbs, nil
}

func (a *App) handleWalletDelete(w http.ResponseWriter, r *http.Request) {
	var q struct {
		Name    string `json:"Name"`
		Confirm string `json:"Confirm"`
	}
	if err := readBody(r, &q); err != nil {
		apiError(w, 400, err)
		return
	}
	name, p, err := a.exactWalletPath(q.Name)
	if err != nil {
		apiError(w, 400, err)
		return
	}
	if q.Confirm != name {
		apiError(w, 400, errors.New("potwierdzenie nazwy portfela nie pasuje"))
		return
	}
	a.mu.RLock()
	miningThisWallet := a.miner.Running && a.miner.Wallet == name
	a.mu.RUnlock()
	if miningThisWallet {
		apiError(w, 409, errors.New("najpierw zatrzymaj kopanie na tym portfelu"))
		return
	}
	st, err := os.Lstat(p)
	if err != nil {
		if os.IsNotExist(err) {
			apiError(w, 404, errors.New("portfel nie istnieje"))
			return
		}
		apiError(w, 500, err)
		return
	}
	if !st.Mode().IsRegular() {
		apiError(w, 400, errors.New("odmowa usunięcia: plik portfela nie jest zwykłym plikiem"))
		return
	}
	if err := os.Remove(p); err != nil {
		apiError(w, 500, err)
		return
	}
	a.addLog("Usunięto lokalny plik portfela: " + name)
	writeJSON(w, map[string]any{"ok": true, "name": name})
}
func (a *App) handleNetworkImport(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(8 << 20); err != nil {
		apiError(w, 400, err)
		return
	}
	f, _, err := r.FormFile("file")
	if err != nil {
		apiError(w, 400, err)
		return
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, 8<<20))
	if err != nil {
		apiError(w, 400, err)
		return
	}
	tmp := filepath.Join(a.baseDir, "network.import.tmp")
	if err = os.WriteFile(tmp, b, 0600); err != nil {
		apiError(w, 500, err)
		return
	}
	defer os.Remove(tmp)
	n, err := aq.LoadNetwork(tmp)
	if err != nil {
		apiError(w, 400, err)
		return
	}
	a.stopMiner()
	a.stopNode()
	if err = aq.SaveNetwork(a.networkPath, n); err != nil {
		apiError(w, 500, err)
		return
	}
	a.mu.Lock()
	a.network = n
	a.nodeErr = ""
	a.mu.Unlock()
	a.addLog("Zaimportowano sieć " + n.Name)
	if err = a.startNode(); err != nil {
		a.addLog("Node: " + err.Error())
	}
	writeJSON(w, map[string]any{"ok": true, "network_id": n.NetworkID().String()})
}

func normalizeDesktopPeer(p string) (string, error) {
	p = strings.TrimSpace(strings.TrimRight(p, "/"))
	if p == "" {
		return "", nil
	}
	if !strings.HasPrefix(p, "http://") && !strings.HasPrefix(p, "https://") {
		p = "http://" + p
	}
	u, err := url.Parse(p)
	if err != nil || u.Host == "" {
		return "", errors.New("nieprawidłowy adres peer")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", errors.New("peer musi używać http:// lub https://")
	}
	return u.Scheme + "://" + u.Host, nil
}

func (a *App) savePeerBook() error {
	a.mu.RLock()
	b, err := json.MarshalIndent(a.peerBook, "", "  ")
	a.mu.RUnlock()
	if err != nil {
		return err
	}
	tmp := a.peerBookPath + ".tmp"
	if err := os.WriteFile(tmp, b, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, a.peerBookPath)
}

func (a *App) handleNetworkPeers(w http.ResponseWriter, r *http.Request) {
	var q struct {
		Peers []string `json:"Peers"`
	}
	if err := readBody(r, &q); err != nil {
		apiError(w, 400, err)
		return
	}
	n := a.currentNetwork()
	if n == nil {
		apiError(w, 400, errors.New("brak sieci"))
		return
	}
	if len(q.Peers) > 16 {
		apiError(w, 400, errors.New("maksymalnie 16 peerów"))
		return
	}
	seen := map[string]bool{}
	peers := make([]string, 0, len(q.Peers))
	for _, raw := range q.Peers {
		p, err := normalizeDesktopPeer(raw)
		if err != nil {
			apiError(w, 400, err)
			return
		}
		if p == "" || seen[p] {
			continue
		}
		seen[p] = true
		peers = append(peers, p)
	}
	sort.Strings(peers)
	wasRunning := false
	a.mu.RLock()
	wasRunning = a.nodeRun
	a.mu.RUnlock()
	if wasRunning {
		a.stopMiner()
		a.stopNode()
	}
	a.mu.Lock()
	a.peerBook[n.NetworkID().String()] = peers
	a.mu.Unlock()
	if err := a.savePeerBook(); err != nil {
		apiError(w, 500, err)
		return
	}
	a.addLog(fmt.Sprintf("Zapisano %d peerów bootstrap", len(peers)))
	if wasRunning {
		if err := a.startNode(); err != nil {
			apiError(w, 500, err)
			return
		}
	}
	writeJSON(w, map[string]any{"ok": true, "peers": peers})
}

func lan24Candidates(ip net.IP, port int) []string {
	v4 := ip.To4()
	if v4 == nil || !v4.IsPrivate() || port <= 0 || port > 65535 {
		return nil
	}
	out := make([]string, 0, 253)
	for host := 1; host <= 254; host++ {
		if byte(host) == v4[3] {
			continue
		}
		out = append(out, fmt.Sprintf("http://%d.%d.%d.%d:%d", v4[0], v4[1], v4[2], host, port))
	}
	return out
}

func localLANPeerCandidates(port int) []string {
	seen := map[string]bool{}
	var out []string
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			var ip net.IP
			switch x := a.(type) {
			case *net.IPNet:
				ip = x.IP
			case *net.IPAddr:
				ip = x.IP
			}
			for _, peer := range lan24Candidates(ip, port) {
				if !seen[peer] {
					seen[peer] = true
					out = append(out, peer)
				}
			}
		}
	}
	sort.Strings(out)
	return out
}

func probeAuronQPeerWithTimeout(peer string, networkID aq.Hash, timeout time.Duration) bool {
	cl := &http.Client{Timeout: timeout}
	resp, err := cl.Get(strings.TrimRight(peer, "/") + "/p2p/hello")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false
	}
	var h struct {
		NetworkID aq.Hash `json:"network_id"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 64<<10)).Decode(&h); err != nil {
		return false
	}
	return h.NetworkID == networkID
}

func (a *App) discoverLANPeersOnce() {
	a.mu.RLock()
	node := a.node
	netCfg := a.network
	a.mu.RUnlock()
	if node == nil || netCfg == nil {
		return
	}
	candidates := localLANPeerCandidates(18444)
	if len(candidates) == 0 {
		return
	}
	type hit struct{ peer string }
	hits := make(chan hit, len(candidates))
	sem := make(chan struct{}, 32)
	var wg sync.WaitGroup
	for _, peer := range candidates {
		peer := peer
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			ok := probeAuronQPeerWithTimeout(peer, netCfg.NetworkID(), 350*time.Millisecond)
			<-sem
			if ok {
				hits <- hit{peer: peer}
			}
		}()
	}
	wg.Wait()
	close(hits)
	for h := range hits {
		node.AddLocalPeer(h.peer)
		a.addLog("LAN peer wykryty automatycznie: " + h.peer)
	}
}

func (a *App) startLANDiscovery() {
	go func() {
		// Give the local node time to bind before probing the LAN, then refresh
		// periodically so two fresh desktops started in either order meet without
		// manual peer configuration.
		time.Sleep(1200 * time.Millisecond)
		for {
			a.discoverLANPeersOnce()
			select {
			case <-a.exit:
				return
			case <-time.After(45 * time.Second):
			}
		}
	}()
}

func probeAuronQPeer(peer string, networkID aq.Hash) bool {
	return probeAuronQPeerWithTimeout(peer, networkID, 1500*time.Millisecond)
}

func (a *App) startNode() error {
	a.mu.Lock()
	if a.nodeRun {
		a.mu.Unlock()
		return nil
	}
	n := a.network
	a.mu.Unlock()
	if n == nil {
		return errors.New("najpierw zaimportuj network.json")
	}
	networkNodeDir := filepath.Join(a.nodeDir, n.NetworkID().String())
	if err := os.MkdirAll(networkNodeDir, 0700); err != nil {
		return err
	}
	chain, err := aq.OpenChain(networkNodeDir, n)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	a.mu.RLock()
	peers := append([]string(nil), a.peerBook[n.NetworkID().String()]...)
	a.mu.RUnlock()
	peerStore := filepath.Join(networkNodeDir, "public-peers.json")
	advertise := strings.TrimSpace(os.Getenv("AURONQ_ADVERTISE"))
	node := aq.NewNode(chain, aq.NodeConfig{Listen: "[::]:18444", Advertise: advertise, Peers: peers, PeerStorePath: peerStore})
	if advertise != "" {
		a.addLog("Public advertise: " + advertise)
	}
	a.mu.Lock()
	a.node = node
	a.chain = chain
	a.nodeCancel = cancel
	a.nodeRun = true
	a.nodeErr = ""
	a.mu.Unlock()
	a.addLog("Uruchamianie pełnego noda na porcie 18444")
	if len(n.SeedPeers) > 0 || len(n.DNSSeeds) > 0 || len(n.BootstrapManifests) > 0 {
		a.addLog(fmt.Sprintf("Automatyczny bootstrap: %d stałych seedów + %d DNS seedów + %d manifestów", len(n.SeedPeers), len(n.DNSSeeds), len(n.BootstrapManifests)))
	} else {
		a.addLog("Brak publicznych seedów w network.json — node działa, ale świeża instalacja nie ma punktu startowego do globalnego P2P")
	}
	go func() {
		err := node.Run(ctx)
		a.mu.Lock()
		a.nodeRun = false
		if err != nil && ctx.Err() == nil {
			a.nodeErr = err.Error()
		}
		a.node = nil
		a.chain = nil
		a.nodeCancel = nil
		a.mu.Unlock()
		if err != nil && ctx.Err() == nil {
			a.addLog("Node zatrzymany z błędem: " + err.Error())
		} else {
			a.addLog("Node zatrzymany")
		}
	}()
	// Fail fast for common bind/startup errors.
	time.Sleep(250 * time.Millisecond)
	a.mu.RLock()
	run := a.nodeRun
	e := a.nodeErr
	a.mu.RUnlock()
	if !run && e != "" {
		return errors.New(e)
	}
	return nil
}
func (a *App) stopNode() {
	a.mu.Lock()
	c := a.nodeCancel
	a.mu.Unlock()
	if c != nil {
		c()
		time.Sleep(150 * time.Millisecond)
	}
}
func (a *App) handleNodeStart(w http.ResponseWriter, r *http.Request) {
	if err := a.startNode(); err != nil {
		apiError(w, 400, err)
		return
	}
	writeJSON(w, map[string]any{"ok": true})
}
func (a *App) handleNodeStop(w http.ResponseWriter, r *http.Request) {
	a.stopMiner()
	a.stopNode()
	writeJSON(w, map[string]any{"ok": true})
}

func (a *App) handleWalletCreate(w http.ResponseWriter, r *http.Request) {
	var q struct{ Name, Password string }
	if err := readBody(r, &q); err != nil {
		apiError(w, 400, err)
		return
	}
	if len(q.Password) < 12 {
		apiError(w, 400, errors.New("hasło musi mieć co najmniej 12 znaków"))
		return
	}
	n := a.currentNetwork()
	if n == nil {
		apiError(w, 400, errors.New("brak network.json"))
		return
	}
	p, err := a.walletPath(q.Name)
	if err != nil {
		apiError(w, 400, err)
		return
	}
	wa, err := aq.NewWallet(p, q.Password, n.NetworkByte)
	if err != nil {
		apiError(w, 400, err)
		return
	}
	addr := wa.Address()
	wa.Close()
	a.addLog("Utworzono portfel " + safeName(q.Name) + "  " + short(addr))
	writeJSON(w, map[string]any{"ok": true, "address": addr})
}

func (a *App) handleWalletImport(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(4 << 20); err != nil {
		apiError(w, 400, err)
		return
	}
	name := safeName(r.FormValue("name"))
	if name == "" {
		apiError(w, 400, errors.New("podaj nazwę portfela"))
		return
	}
	f, _, err := r.FormFile("file")
	if err != nil {
		apiError(w, 400, err)
		return
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, 4<<20))
	if err != nil {
		apiError(w, 400, err)
		return
	}
	var wf aq.WalletFile
	if err = json.Unmarshal(b, &wf); err != nil {
		apiError(w, 400, errors.New("nieprawidłowy plik portfela"))
		return
	}
	if wf.Format != "auronq-wallet-v1" || wf.Scheme != aq.SchemeMLDSA87 {
		apiError(w, 400, errors.New("nieobsługiwany format portfela"))
		return
	}
	net, scheme, _, err := aq.DecodeAddress(wf.Address)
	if err != nil || scheme != aq.SchemeMLDSA87 || net != wf.NetworkByte {
		apiError(w, 400, errors.New("niespójny adres w pliku portfela"))
		return
	}
	if n := a.currentNetwork(); n != nil && wf.NetworkByte != n.NetworkByte {
		apiError(w, 400, errors.New("portfel należy do innej sieci"))
		return
	}
	p, _ := a.walletPath(name)
	if _, err = os.Stat(p); err == nil {
		apiError(w, 409, errors.New("portfel o tej nazwie już istnieje"))
		return
	}
	if err = os.WriteFile(p, b, 0600); err != nil {
		apiError(w, 500, err)
		return
	}
	a.addLog("Zaimportowano portfel " + name)
	writeJSON(w, map[string]any{"ok": true, "address": wf.Address})
}

func (a *App) handleWalletExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "method", 405)
		return
	}
	if r.URL.Query().Get("token") != a.token {
		http.Error(w, "forbidden", 403)
		return
	}
	p, err := a.walletPath(r.URL.Query().Get("name"))
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filepath.Base(p)+"\"")
	w.Header().Set("Content-Type", "application/json")
	http.ServeFile(w, r, p)
}

func (a *App) handleWalletHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "method", http.StatusMethodNotAllowed)
		return
	}
	if r.Header.Get("X-AuronQ-Token") != a.token {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	name := safeName(r.URL.Query().Get("wallet"))
	if name == "" {
		apiError(w, 400, errors.New("wybierz portfel"))
		return
	}
	p, err := a.walletPath(name)
	if err != nil {
		apiError(w, 400, err)
		return
	}
	b, err := os.ReadFile(p)
	if err != nil {
		apiError(w, 404, errors.New("portfel nie istnieje"))
		return
	}
	var wf aq.WalletFile
	if err := json.Unmarshal(b, &wf); err != nil {
		apiError(w, 400, errors.New("nieprawidłowy plik portfela"))
		return
	}

	a.mu.RLock()
	chain := a.chain
	running := a.nodeRun
	a.mu.RUnlock()
	if !running || chain == nil {
		apiError(w, 400, errors.New("node nie jest uruchomiony"))
		return
	}
	items, err := chain.HistoryForAddress(wf.Address, 250)
	if err != nil {
		apiError(w, 400, err)
		return
	}
	writeJSON(w, map[string]any{
		"ok":      true,
		"wallet":  name,
		"address": wf.Address,
		"height":  chain.Height(),
		"items":   items,
	})
}

func (a *App) handleSend(w http.ResponseWriter, r *http.Request) {
	var q struct{ Wallet, Password, To, Amount string }
	if err := readBody(r, &q); err != nil {
		apiError(w, 400, err)
		return
	}
	a.mu.RLock()
	run := a.nodeRun
	a.mu.RUnlock()
	if !run {
		apiError(w, 400, errors.New("node nie jest uruchomiony"))
		return
	}
	n := a.currentNetwork()
	if n == nil {
		apiError(w, 400, errors.New("brak sieci"))
		return
	}
	p, err := a.walletPath(q.Wallet)
	if err != nil {
		apiError(w, 400, err)
		return
	}
	wa, err := aq.LoadWallet(p, q.Password)
	if err != nil {
		apiError(w, 400, err)
		return
	}
	defer wa.Close()
	amt, err := aq.ParseAmount(strings.TrimSpace(q.Amount))
	if err != nil {
		apiError(w, 400, err)
		return
	}
	cl := aq.NewClient(a.nodeURL)
	us, err := cl.UTXOs(wa.Address())
	if err != nil {
		apiError(w, 400, err)
		return
	}
	tx, fee, err := wa.BuildTransaction(us, strings.TrimSpace(q.To), amt, n.NetworkByte)
	if err != nil {
		apiError(w, 400, err)
		return
	}
	res, err := cl.SubmitTx(tx)
	if err != nil {
		apiError(w, 400, err)
		return
	}
	a.addLog("Wysłano " + aq.FormatAmount(amt) + " AURQ  tx " + short(res.TXID.String()))
	writeJSON(w, map[string]any{"ok": true, "txid": res.TXID.String(), "fee": aq.FormatAmount(fee)})
}

func (a *App) startMiner(wallet string, threads int) error {
	a.mu.Lock()
	if a.miner.Running {
		a.mu.Unlock()
		return errors.New("miner już działa")
	}
	if !a.nodeRun {
		a.mu.Unlock()
		return errors.New("najpierw uruchom node")
	}
	a.mu.Unlock()
	p, err := a.walletPath(wallet)
	if err != nil {
		return err
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return err
	}
	var wf aq.WalletFile
	if err = json.Unmarshal(b, &wf); err != nil {
		return err
	}
	if threads < 1 {
		threads = 1
	}
	if threads > runtime.NumCPU() {
		threads = runtime.NumCPU()
	}
	ctx, cancel := context.WithCancel(context.Background())
	a.mu.Lock()
	a.miner = minerState{Running: true, Threads: threads, Wallet: wallet, Address: wf.Address, StartedAt: time.Now().Unix()}
	a.minerCancel = cancel
	a.mu.Unlock()
	a.addLog(fmt.Sprintf("Mining start: %s, %d wątków", wallet, threads))
	go a.miningLoop(ctx, wf.Address, threads)
	return nil
}
func (a *App) miningLoop(ctx context.Context, address string, threads int) {
	cl := aq.NewClient(a.nodeURL)
	for {
		select {
		case <-ctx.Done():
			a.finishMiner("")
			return
		default:
		}
		tpl, err := cl.Template(address)
		if err != nil {
			a.finishMiner(err.Error())
			return
		}
		a.mu.Lock()
		a.miner.Height = tpl.Header.Height
		a.miner.Hashrate = 0
		a.mu.Unlock()
		res, err := aq.MineParallel(ctx, tpl, threads, func(h uint64, d time.Duration) {
			if d > 0 {
				a.mu.Lock()
				a.miner.Hashrate = float64(h) / d.Seconds()
				a.mu.Unlock()
			}
		})
		if err != nil {
			if ctx.Err() != nil {
				a.finishMiner("")
				return
			}
			a.finishMiner(err.Error())
			return
		}
		br, err := cl.SubmitBlock(res.Block)
		if err != nil {
			a.addLog("Blok odrzucony, odświeżam template: " + err.Error())
			continue
		}
		rate := float64(res.Hashes) / res.Duration.Seconds()
		a.mu.Lock()
		a.miner.Hashrate = rate
		a.miner.BlocksFound++
		a.miner.LastBlock = br.Hash.String()
		a.miner.Height = br.Height
		a.mu.Unlock()
		a.addLog(fmt.Sprintf("Wykopano blok %d  %s", br.Height, short(br.Hash.String())))
	}
}
func (a *App) finishMiner(errText string) {
	a.mu.Lock()
	if errText != "" {
		a.miner.LastError = errText
	}
	a.miner.Running = false
	a.miner.Hashrate = 0
	a.minerCancel = nil
	a.mu.Unlock()
	if errText != "" {
		a.addLog("Mining zatrzymany: " + errText)
	} else {
		a.addLog("Mining zatrzymany")
	}
}
func (a *App) stopMiner() {
	a.mu.Lock()
	c := a.minerCancel
	a.mu.Unlock()
	if c != nil {
		c()
	}
}
func (a *App) handleMinerStart(w http.ResponseWriter, r *http.Request) {
	var q struct {
		Wallet  string
		Threads int
	}
	if err := readBody(r, &q); err != nil {
		apiError(w, 400, err)
		return
	}
	if err := a.startMiner(q.Wallet, q.Threads); err != nil {
		apiError(w, 400, err)
		return
	}
	writeJSON(w, map[string]any{"ok": true})
}
func (a *App) handleMinerStop(w http.ResponseWriter, r *http.Request) {
	a.stopMiner()
	writeJSON(w, map[string]any{"ok": true})
}

func (a *App) handleOpenData(w http.ResponseWriter, r *http.Request) {
	if runtime.GOOS == "windows" {
		_ = exec.Command("explorer.exe", a.baseDir).Start()
	}
	writeJSON(w, map[string]any{"ok": true})
}
func (a *App) handleExit(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{"ok": true})
	go func() {
		time.Sleep(150 * time.Millisecond)
		select {
		case a.exit <- struct{}{}:
		default:
		}
	}()
}

func short(s string) string {
	if len(s) <= 18 {
		return s
	}
	return s[:9] + "…" + s[len(s)-8:]
}

func reopenExisting(base string) bool {
	c := &http.Client{Timeout: 750 * time.Millisecond}
	r, err := c.Get(base + "/api/state")
	if err != nil {
		return false
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		return false
	}
	var st appState
	if json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&st) != nil || st.Version == "" {
		return false
	}
	_ = openDesktopWindow(base + "/")
	return true
}

func openDesktopWindow(url string) error {
	if runtime.GOOS != "windows" {
		return exec.Command("xdg-open", url).Start()
	}
	candidates := []string{
		filepath.Join(os.Getenv("ProgramFiles(x86)"), "Microsoft", "Edge", "Application", "msedge.exe"),
		filepath.Join(os.Getenv("ProgramFiles"), "Microsoft", "Edge", "Application", "msedge.exe"),
		filepath.Join(os.Getenv("LocalAppData"), "Microsoft", "Edge", "Application", "msedge.exe"),
	}
	for _, p := range candidates {
		if p != "" {
			if _, err := os.Stat(p); err == nil {
				return exec.Command(p, "--app="+url, "--start-maximized").Start()
			}
		}
	}
	return exec.Command("rundll32.exe", "url.dll,FileProtocolHandler", url).Start()
}

func cryptoSelfTest() (bool, error) {
	seed, pub, err := aq.GenerateMLDSA87()
	if err != nil {
		return false, err
	}
	defer func() {
		for i := range seed {
			seed[i] = 0
		}
	}()
	msg := []byte("AURONQ_DESKTOP_CRYPTO_SELFTEST_V1")
	sig, err := aq.SignMLDSA87(seed, msg)
	if err != nil {
		return false, err
	}
	if !aq.VerifyMLDSA87(pub, msg, sig) {
		return false, errors.New("ML-DSA-87 self-test verification failed")
	}
	return true, nil
}