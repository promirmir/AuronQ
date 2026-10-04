//go:build windows

package main

import (
	"bufio"
	"bytes"
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
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	aq "auronq/internal/auronq"
)

const (
	guiVersion       = "0.2.0-alpha"
	guiListen        = "127.0.0.1:18446"
	localNodeURL     = "http://127.0.0.1:18444"
	mainnetNetworkID = "44e62c2ace002a6660c14e252173c1aa303529c68e40c998e92da2b453f44f30b1e58c94d533587e2186004593fb856c433fcdb5418ed430ec8617e29529365c"
)

//go:embed web/*
var webFS embed.FS

type settings struct {
	Address    string `json:"address"`
	Device     int    `json:"device"`
	Batch      int    `json:"batch"`
	AutoPublic bool   `json:"auto_public"`
	SelfTest   bool   `json:"self_test"`
	Language   string `json:"language"`
}

type minerState struct {
	Running     bool    `json:"running"`
	Mode        string  `json:"mode,omitempty"`
	Hashrate    float64 `json:"hashrate"`
	Height      uint64  `json:"height"`
	BlocksFound uint64  `json:"blocks_found"`
	LastBlock   string  `json:"last_block,omitempty"`
	LastError   string  `json:"last_error,omitempty"`
	StartedAt   int64   `json:"started_at,omitempty"`
}

type appState struct {
	Version         string     `json:"version"`
	NetworkName     string     `json:"network_name"`
	NetworkID       string     `json:"network_id"`
	NodeRunning     bool       `json:"node_running"`
	NodeOwned       bool       `json:"node_owned"`
	NodeError       string     `json:"node_error,omitempty"`
	Height          uint64     `json:"height"`
	Tip             string     `json:"tip,omitempty"`
	Peers           int        `json:"peers"`
	NetworkHashrate float64    `json:"network_hashrate"`
	PublicEndpoint  string     `json:"public_endpoint,omitempty"`
	SyncTarget      uint64     `json:"sync_target"`
	Synchronized    bool       `json:"synchronized"`
	GPUName         string     `json:"gpu_name,omitempty"`
	Miner           minerState `json:"miner"`
	Settings        settings   `json:"settings"`
	Logs            []string   `json:"logs"`
}

type App struct {
	mu sync.RWMutex

	baseDir     string
	nodeDir     string
	settingsPath string
	network     *aq.NetworkConfig
	bootstrap   []string

	node       *aq.Node
	chain      *aq.Chain
	nodeCancel context.CancelFunc
	nodeRun    bool
	nodeOwned  bool
	nodeErr    string

	portMapping *aq.PortMapping
	publicCancel context.CancelFunc

	minerCmd           *exec.Cmd
	miner              minerState
	minerStopRequested bool
	gpuName            string

	cfg        settings
	syncTarget uint64
	logs       []string
	token      string
	exit       chan struct{}
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
	ln, err := net.Listen("tcp", guiListen)
	if err != nil {
		if reopenExisting("http://" + guiListen) {
			return
		}
		log.Fatal(err)
	}

	srv := &http.Server{
		Handler:           securityHeaders(mux),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	go func() {
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			app.addLog("GUI HTTP: " + err.Error())
		}
	}()

	go func() {
		if err := app.ensureNode(); err != nil {
			app.addLog("Node: " + err.Error())
		}
		app.monitorNetwork()
	}()

	if err := openDesktopWindow("http://" + guiListen + "/"); err != nil {
		app.addLog("UI: " + err.Error())
	}

	<-app.exit
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}

func newApp() (*App, error) {
	cfgRoot, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	base := filepath.Join(cfgRoot, "AuronQ", "gpu-miner")
	if err := os.MkdirAll(base, 0700); err != nil {
		return nil, err
	}

	a := &App{
		baseDir:      base,
		nodeDir:      filepath.Join(base, "node"),
		settingsPath: filepath.Join(base, "settings.json"),
		exit:         make(chan struct{}, 1),
	}
	if err := os.MkdirAll(a.nodeDir, 0700); err != nil {
		return nil, err
	}
	a.cfg = settings{Device: 0, Batch: 60, AutoPublic: true, SelfTest: true, Language: "pl"}
	a.loadSettings()

	exeDir, err := executableDir()
	if err != nil {
		return nil, err
	}
	networkPath := filepath.Join(exeDir, "network.json")
	n, err := aq.LoadNetwork(networkPath)
	if err != nil {
		return nil, fmt.Errorf("network.json: %w", err)
	}
	if n.NetworkID().String() != mainnetNetworkID {
		return nil, errors.New("bundled network.json is not AuronQ Mainnet")
	}
	a.network = n
	a.bootstrap = loadBootstrap(filepath.Join(exeDir, "bootstrap.json"), n)
	a.bootstrap = mergePeers(a.bootstrap, n.SeedPeers)

	var tokenBytes [32]byte
	if _, err := rand.Read(tokenBytes[:]); err != nil {
		return nil, err
	}
	a.token = hex.EncodeToString(tokenBytes[:])
	a.addLog("AuronQ GPU Miner " + guiVersion)
	a.addLog(fmt.Sprintf("Mainnet loaded: %s", n.NetworkID().String()))
	a.addLog(fmt.Sprintf("Bootstrap peers: %d", len(a.bootstrap)))
	return a, nil
}

func executableDir() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(exe), nil
}

func mergePeers(a, b []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(a)+len(b))
	for _, list := range [][]string{a, b} {
		for _, raw := range list {
			p := strings.TrimSpace(strings.TrimRight(raw, "/"))
			if p == "" || seen[p] {
				continue
			}
			seen[p] = true
			out = append(out, p)
		}
	}
	return out
}

func loadBootstrap(path string, n *aq.NetworkConfig) []string {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var m aq.BootstrapManifest
	if json.Unmarshal(b, &m) != nil || m.NetworkID != n.NetworkID() {
		return nil
	}
	if m.ExpiresAt != 0 && time.Now().Unix() > m.ExpiresAt {
		return nil
	}
	return mergePeers(nil, m.Peers)
}

func (a *App) loadSettings() {
	b, err := os.ReadFile(a.settingsPath)
	if err != nil {
		return
	}
	var s settings
	if json.Unmarshal(b, &s) != nil {
		return
	}
	if s.Device < 0 {
		s.Device = 0
	}
	if s.Batch < 1 || s.Batch > 64 {
		s.Batch = 60
	}
	if s.Language != "en" {
		s.Language = "pl"
	}
	a.cfg = s
}

func (a *App) saveSettings(s settings) error {
	if s.Device < 0 {
		return errors.New("CUDA device must be 0 or greater")
	}
	if s.Batch < 1 || s.Batch > 64 {
		return errors.New("batch must be between 1 and 64")
	}
	if s.Language != "en" {
		s.Language = "pl"
	}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	tmp := a.settingsPath + ".tmp"
	if err := os.WriteFile(tmp, b, 0600); err != nil {
		return err
	}
	if err := os.Rename(tmp, a.settingsPath); err != nil {
		return err
	}
	a.mu.Lock()
	a.cfg = s
	a.mu.Unlock()
	return nil
}

func (a *App) addLog(s string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.logs = append(a.logs, time.Now().Format("15:04:05")+"  "+s)
	if len(a.logs) > 400 {
		a.logs = append([]string(nil), a.logs[len(a.logs)-400:]...)
	}
}

func (a *App) routes(mux *http.ServeMux) {
	mux.HandleFunc("/", a.handleIndex)
	mux.HandleFunc("/api/state", a.handleState)
	mux.HandleFunc("/api/settings", a.guard(a.handleSettings))
	mux.HandleFunc("/api/miner/start", a.guard(a.handleMinerStart))
	mux.HandleFunc("/api/miner/stop", a.guard(a.handleMinerStop))
	mux.HandleFunc("/api/self-test", a.guard(a.handleSelfTest))
	mux.HandleFunc("/api/benchmark", a.guard(a.handleBenchmark))
	mux.HandleFunc("/api/node/reconnect", a.guard(a.handleReconnect))
	mux.HandleFunc("/api/exit", a.guard(a.handleExit))
}

func (a *App) guard(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method", http.StatusMethodNotAllowed)
			return
		}
		if r.Header.Get("X-AuronQ-Token") != a.token {
			http.Error(w, "forbidden", http.StatusForbidden)
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
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline'; connect-src 'self'; img-src 'self' data:; frame-ancestors 'none'")
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
	_, _ = io.WriteString(w, s)
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
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return err
	}
	return nil
}

func (a *App) handleState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, a.state())
}

func (a *App) state() appState {
	a.mu.RLock()
	cfg := a.cfg
	miner := a.miner
	nodeRun := a.nodeRun
	nodeOwned := a.nodeOwned
	nodeErr := a.nodeErr
	syncTarget := a.syncTarget
	gpu := a.gpuName
	logs := append([]string(nil), a.logs...)
	mapping := a.portMapping
	a.mu.RUnlock()

	st := appState{
		Version:     guiVersion,
		NetworkName: a.network.Name,
		NetworkID:   a.network.NetworkID().String(),
		NodeRunning: nodeRun,
		NodeOwned:   nodeOwned,
		NodeError:   nodeErr,
		SyncTarget:  syncTarget,
		GPUName:     gpu,
		Miner:       miner,
		Settings:    cfg,
		Logs:        logs,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 900*time.Millisecond)
	defer cancel()
	if ns, err := aq.NewClient(localNodeURL).StatusContext(ctx); err == nil && ns.NetworkID == a.network.NetworkID() {
		st.NodeRunning = true
		st.Height = ns.Height
		st.Tip = ns.Tip.String()
		st.Peers = ns.Peers
		st.NetworkHashrate = ns.NetworkHashrate
		st.PublicEndpoint = ns.PublicAdvertise
		if st.SyncTarget < ns.Height {
			st.SyncTarget = ns.Height
		}
	} else if !nodeOwned {
		st.NodeRunning = false
	}
	if st.PublicEndpoint == "" && mapping != nil {
		st.PublicEndpoint = mapping.Advertise
	}
	if st.SyncTarget == 0 {
		st.SyncTarget = st.Height
	}
	st.Synchronized = st.NodeRunning && st.Height >= st.SyncTarget
	return st
}

func (a *App) handleSettings(w http.ResponseWriter, r *http.Request) {
	var s settings
	if err := readBody(r, &s); err != nil {
		apiError(w, 400, err)
		return
	}
	if err := a.saveSettings(s); err != nil {
		apiError(w, 400, err)
		return
	}
	writeJSON(w, map[string]any{"ok": true})
}

func (a *App) ensureNode() error {
	a.mu.RLock()
	trackedRun := a.nodeRun
	trackedOwned := a.nodeOwned
	a.mu.RUnlock()
	if trackedRun && trackedOwned {
		return nil
	}
	if trackedRun && !trackedOwned {
		ctx, cancel := context.WithTimeout(context.Background(), 900*time.Millisecond)
		st, err := aq.NewClient(localNodeURL).StatusContext(ctx)
		cancel()
		if err == nil && st.NetworkID == a.network.NetworkID() {
			return nil
		}
		a.mu.Lock()
		a.nodeRun = false
		a.nodeOwned = false
		a.mu.Unlock()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1200*time.Millisecond)
	st, err := aq.NewClient(localNodeURL).StatusContext(ctx)
	cancel()
	if err == nil {
		if st.NetworkID != a.network.NetworkID() {
			return errors.New("port 18444 is used by a different network")
		}
		a.mu.Lock()
		a.nodeRun = true
		a.nodeOwned = false
		a.nodeErr = ""
		a.mu.Unlock()
		a.addLog(fmt.Sprintf("Using existing local AuronQ full node at height %d", st.Height))
		return nil
	}

	networkNodeDir := filepath.Join(a.nodeDir, a.network.NetworkID().String())
	if err := os.MkdirAll(networkNodeDir, 0700); err != nil {
		return err
	}
	chain, err := aq.OpenChain(networkNodeDir, a.network)
	if err != nil {
		return err
	}
	node := aq.NewNode(chain, aq.NodeConfig{
		Listen:        "0.0.0.0:18444",
		Peers:         append([]string(nil), a.bootstrap...),
		PeerStorePath: filepath.Join(networkNodeDir, "public-peers.json"),
	})
	nodeCtx, nodeCancel := context.WithCancel(context.Background())

	a.mu.Lock()
	a.node = node
	a.chain = chain
	a.nodeCancel = nodeCancel
	a.nodeRun = true
	a.nodeOwned = true
	a.nodeErr = ""
	a.mu.Unlock()
	a.addLog("Starting embedded full node on TCP/18444")

	go func() {
		err := node.Run(nodeCtx)
		a.mu.Lock()
		if a.node == node {
			a.node = nil
			a.chain = nil
			a.nodeCancel = nil
			a.nodeRun = false
			a.nodeOwned = false
			if err != nil && nodeCtx.Err() == nil {
				a.nodeErr = err.Error()
			}
		}
		a.mu.Unlock()
		if err != nil && nodeCtx.Err() == nil {
			a.addLog("Embedded node stopped: " + err.Error())
		}
	}()

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		time.Sleep(120 * time.Millisecond)
		cctx, ccancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		ns, e := aq.NewClient(localNodeURL).StatusContext(cctx)
		ccancel()
		if e == nil && ns.NetworkID == a.network.NetworkID() {
			a.addLog(fmt.Sprintf("Embedded full node ready at height %d", ns.Height))
			return nil
		}
		a.mu.RLock()
		nodeErr := a.nodeErr
		running := a.nodeRun
		a.mu.RUnlock()
		if !running && nodeErr != "" {
			return errors.New(nodeErr)
		}
	}
	return errors.New("full node did not become ready on TCP/18444")
}

func (a *App) stopOwnedNode() {
	a.mu.Lock()
	cancel := a.nodeCancel
	owned := a.nodeOwned
	a.nodeCancel = nil
	if owned {
		a.nodeRun = false
		a.nodeOwned = false
	}
	a.mu.Unlock()
	if owned && cancel != nil {
		cancel()
	}
}

func (a *App) handleReconnect(w http.ResponseWriter, r *http.Request) {
	if a.isWorkerRunning() {
		apiError(w, 409, errors.New("stop mining/benchmark/self-test before reconnecting the node"))
		return
	}
	a.closePublicMapping()
	a.stopOwnedNode()
	time.Sleep(200 * time.Millisecond)
	a.mu.Lock()
	// External-node tracking is only a cached observation. Clear it so ensureNode
	// performs a fresh Network-ID-checked probe instead of trusting stale state.
	if !a.nodeOwned {
		a.nodeRun = false
	}
	a.nodeErr = ""
	a.mu.Unlock()
	if err := a.ensureNode(); err != nil {
		apiError(w, 400, err)
		return
	}
	a.refreshSyncTarget()
	writeJSON(w, map[string]any{"ok": true})
}

func (a *App) monitorNetwork() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		a.refreshSyncTarget()
		select {
		case <-a.exit:
			return
		case <-ticker.C:
		}
	}
}

func (a *App) refreshSyncTarget() {
	localHeight := uint64(0)
	ctx, cancel := context.WithTimeout(context.Background(), 900*time.Millisecond)
	if st, err := aq.NewClient(localNodeURL).StatusContext(ctx); err == nil && st.NetworkID == a.network.NetworkID() {
		localHeight = st.Height
	}
	cancel()

	type result struct {
		height uint64
	}
	ch := make(chan result, len(a.bootstrap))
	for _, peer := range a.bootstrap {
		peer := peer
		go func() {
			c := &http.Client{Timeout: 1800 * time.Millisecond}
			resp, err := c.Get(strings.TrimRight(peer, "/") + "/p2p/hello")
			if err != nil {
				ch <- result{}
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				ch <- result{}
				return
			}
			var h aq.Hello
			if json.NewDecoder(io.LimitReader(resp.Body, 64<<10)).Decode(&h) != nil ||
				h.ProtocolVersion != 1 || h.NetworkID != a.network.NetworkID() {
				ch <- result{}
				return
			}
			ch <- result{height: h.Height}
		}()
	}
	target := localHeight
	for range a.bootstrap {
		r := <-ch
		if r.height > target {
			target = r.height
		}
	}
	a.mu.Lock()
	a.syncTarget = target
	a.mu.Unlock()
}

func (a *App) validateRewardAddress(addr string) error {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return errors.New("reward address is required")
	}
	netByte, _, _, err := aq.DecodeAddress(addr)
	if err != nil || netByte != aq.MainnetNetworkByte {
		return errors.New("reward address must be a valid AuronQ Mainnet address")
	}
	return nil
}

func workerPath() string {
	exe, err := os.Executable()
	if err != nil {
		return "auronq-gpu-miner.exe"
	}
	return filepath.Join(filepath.Dir(exe), "auronq-gpu-miner.exe")
}

func (a *App) isWorkerRunning() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.miner.Running
}

func (a *App) startWorker(mode string, s settings) error {
	a.mu.Lock()
	if a.miner.Running {
		a.mu.Unlock()
		return errors.New("a miner job is already running")
	}
	a.mu.Unlock()

	if s.Device < 0 || s.Batch < 1 || s.Batch > 64 {
		return errors.New("invalid CUDA device or batch")
	}
	if mode == "mining" {
		if err := a.validateRewardAddress(s.Address); err != nil {
			return err
		}
		if err := a.ensureNode(); err != nil {
			return fmt.Errorf("full node: %w", err)
		}
		a.refreshSyncTarget()
		state := a.state()
		if state.SyncTarget > 0 && state.Height < state.SyncTarget {
			return fmt.Errorf("full node is synchronizing: local height %d, peer height %d", state.Height, state.SyncTarget)
		}
	}
	if err := a.saveSettings(s); err != nil {
		return err
	}

	args := []string{"--device", strconv.Itoa(s.Device), "--batch", strconv.Itoa(s.Batch)}
	switch mode {
	case "selftest":
		args = append(args, "--self-test")
	case "benchmark":
		args = append(args, "--benchmark", "--benchmark-seconds", "15")
	case "mining":
		args = append(args, "--node", localNodeURL, "--address", strings.TrimSpace(s.Address))
		if s.SelfTest {
			args = append(args, "--self-test")
		}
	default:
		return errors.New("unknown worker mode")
	}

	cmd := exec.Command(workerPath(), args...)
	cmd.Dir = filepath.Dir(workerPath())
	cmd.SysProcAttr = &syscallSysProcAttr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	a.mu.Lock()
	a.minerCmd = cmd
	a.minerStopRequested = false
	a.miner = minerState{
		Running:     true,
		Mode:        mode,
		BlocksFound: a.miner.BlocksFound,
		LastBlock:   a.miner.LastBlock,
		StartedAt:   time.Now().Unix(),
	}
	a.mu.Unlock()
	a.addLog(strings.ToUpper(mode) + " started")

	if mode == "mining" && s.AutoPublic {
		go a.ensurePublicPeer()
	}

	go a.scanWorker(stdout, "")
	go a.scanWorker(stderr, "ERROR: ")
	go func() {
		err := cmd.Wait()
		a.mu.Lock()
		stopped := a.minerStopRequested
		if a.minerCmd == cmd {
			a.minerCmd = nil
		}
		finishedMode := a.miner.Mode
		a.miner.Running = false
		a.miner.Mode = ""
		if finishedMode == "mining" {
			a.miner.Hashrate = 0
		}
		if err != nil && !stopped {
			a.miner.LastError = err.Error()
		}
		a.mu.Unlock()
		if finishedMode == "mining" {
			a.closePublicMapping()
		}
		if err != nil && !stopped {
			a.addLog("Worker stopped: " + err.Error())
		} else {
			a.addLog("Worker stopped")
		}
	}()
	return nil
}

var syscallSysProcAttr = syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}

func (a *App) scanWorker(r io.Reader, prefix string) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 4096), 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		a.addLog(prefix + line)
		a.parseWorkerLine(line)
	}
}

func (a *App) parseWorkerLine(line string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if strings.HasPrefix(line, "GPU: ") {
		a.gpuName = strings.TrimSpace(strings.TrimPrefix(line, "GPU: "))
	}
	if strings.HasPrefix(line, "Mining height ") {
		fields := strings.Fields(line)
		if len(fields) >= 3 {
			if h, err := strconv.ParseUint(fields[2], 10, 64); err == nil {
				a.miner.Height = h
			}
		}
	}
	if strings.HasPrefix(line, "hashes=") {
		if v, ok := numberAfter(line, "avg="); ok {
			a.miner.Hashrate = v
		}
		if v, ok := uintAfter(line, "current_height="); ok {
			a.miner.Height = v
		}
	}
	if strings.HasPrefix(line, "BENCHMARK OK") {
		if v, ok := numberAfter(line, "avg="); ok {
			a.miner.Hashrate = v
		}
	}
	if strings.HasPrefix(line, "BLOCK FOUND ") {
		a.miner.BlocksFound++
		if v, ok := uintAfter(line, "height="); ok {
			a.miner.Height = v
		}
		if h, ok := wordAfter(line, "hash="); ok {
			a.miner.LastBlock = h
		}
	}
	if strings.Contains(line, "SELF-TEST FAILED") || strings.Contains(line, "BENCHMARK FAILED") {
		a.miner.LastError = line
	}
}

func numberAfter(line, key string) (float64, bool) {
	i := strings.Index(line, key)
	if i < 0 {
		return 0, false
	}
	s := line[i+len(key):]
	if j := strings.IndexAny(s, " 	"); j >= 0 {
		s = s[:j]
	}
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return v, err == nil
}

func uintAfter(line, key string) (uint64, bool) {
	i := strings.Index(line, key)
	if i < 0 {
		return 0, false
	}
	s := line[i+len(key):]
	if j := strings.IndexAny(s, " 	"); j >= 0 {
		s = s[:j]
	}
	v, err := strconv.ParseUint(strings.TrimSpace(s), 10, 64)
	return v, err == nil
}

func wordAfter(line, key string) (string, bool) {
	i := strings.Index(line, key)
	if i < 0 {
		return "", false
	}
	s := line[i+len(key):]
	if j := strings.IndexAny(s, " 	"); j >= 0 {
		s = s[:j]
	}
	s = strings.TrimSpace(s)
	return s, s != ""
}

func (a *App) stopWorker() {
	a.mu.Lock()
	cmd := a.minerCmd
	a.minerStopRequested = true
	a.mu.Unlock()
	if cmd != nil && cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
}

func (a *App) handleMinerStart(w http.ResponseWriter, r *http.Request) {
	var s settings
	if err := readBody(r, &s); err != nil {
		apiError(w, 400, err)
		return
	}
	if err := a.startWorker("mining", s); err != nil {
		apiError(w, 400, err)
		return
	}
	writeJSON(w, map[string]any{"ok": true})
}

func (a *App) handleMinerStop(w http.ResponseWriter, r *http.Request) {
	a.stopWorker()
	writeJSON(w, map[string]any{"ok": true})
}

func (a *App) handleSelfTest(w http.ResponseWriter, r *http.Request) {
	var s settings
	if err := readBody(r, &s); err != nil {
		apiError(w, 400, err)
		return
	}
	if err := a.startWorker("selftest", s); err != nil {
		apiError(w, 400, err)
		return
	}
	writeJSON(w, map[string]any{"ok": true})
}

func (a *App) handleBenchmark(w http.ResponseWriter, r *http.Request) {
	var s settings
	if err := readBody(r, &s); err != nil {
		apiError(w, 400, err)
		return
	}
	if err := a.startWorker("benchmark", s); err != nil {
		apiError(w, 400, err)
		return
	}
	writeJSON(w, map[string]any{"ok": true})
}

func (a *App) ensurePublicPeer() {
	a.mu.RLock()
	if a.portMapping != nil {
		a.mu.RUnlock()
		return
	}
	ownedNode := a.node
	a.mu.RUnlock()

	// If the local node is already directly public (or AuronQ Desktop already
	// configured a public endpoint), do not create a second router mapping.
	statusCtx, statusCancel := context.WithTimeout(context.Background(), 900*time.Millisecond)
	if st, err := aq.NewClient(localNodeURL).StatusContext(statusCtx); err == nil &&
		st.NetworkID == a.network.NetworkID() && strings.TrimSpace(st.PublicAdvertise) != "" {
		statusCancel()
		a.addLog("Public node already active: " + st.PublicAdvertise)
		return
	}
	statusCancel()

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	mapping, err := aq.TryUPnPPortMapping(ctx, 18444)
	cancel()
	if err != nil {
		a.addLog("Public node: UPnP unavailable (" + err.Error() + "); continuing outbound-only")
		return
	}

	a.mu.Lock()
	if a.portMapping != nil {
		a.mu.Unlock()
		mapping.Close()
		return
	}
	a.portMapping = mapping
	a.mu.Unlock()

	if ownedNode != nil {
		if !ownedNode.SetPublicAdvertise(mapping.Advertise) {
			a.addLog("Public node: endpoint rejected by local node")
			a.closePublicMapping()
			return
		}
	}
	a.addLog("Public node: TCP/18444 mapped to " + mapping.Advertise)
	a.announcePublic(mapping.Advertise)

	ctxLoop, cancelLoop := context.WithCancel(context.Background())
	a.mu.Lock()
	if a.publicCancel != nil {
		a.publicCancel()
	}
	a.publicCancel = cancelLoop
	a.mu.Unlock()
	go func() {
		t := time.NewTicker(2 * time.Minute)
		defer t.Stop()
		for {
			select {
			case <-ctxLoop.Done():
				return
			case <-t.C:
				a.announcePublic(mapping.Advertise)
			}
		}
	}()
}

func (a *App) announcePublic(endpoint string) {
	payload := aq.PeerAnnounce{
		ProtocolVersion: 1,
		NetworkID:       a.network.NetworkID(),
		Advertise:       endpoint,
		ListenPort:      18444,
	}
	body, _ := json.Marshal(payload)
	success := 0
	var wg sync.WaitGroup
	var smu sync.Mutex
	for _, peer := range a.bootstrap {
		peer := peer
		wg.Add(1)
		go func() {
			defer wg.Done()
			req, err := http.NewRequest(http.MethodPost, strings.TrimRight(peer, "/")+"/p2p/announce", bytes.NewReader(body))
			if err != nil {
				return
			}
			req.Header.Set("Content-Type", "application/json")
			c := &http.Client{Timeout: 7 * time.Second}
			resp, err := c.Do(req)
			if err != nil {
				return
			}
			io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))
			resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				smu.Lock()
				success++
				smu.Unlock()
			}
		}()
	}
	wg.Wait()
	a.addLog(fmt.Sprintf("Public node: announcement accepted by %d/%d bootstrap peers; remote peers callback-verify the endpoint", success, len(a.bootstrap)))
}

func (a *App) closePublicMapping() {
	a.mu.Lock()
	mapping := a.portMapping
	cancel := a.publicCancel
	ownedNode := a.node
	a.portMapping = nil
	a.publicCancel = nil
	a.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	if mapping != nil {
		if ownedNode != nil {
			ownedNode.ClearPublicAdvertise(mapping.Advertise)
		}
		mapping.Close()
		a.addLog("Public node: TCP/18444 mapping closed")
	}
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

func (a *App) shutdown() {
	a.stopWorker()
	a.closePublicMapping()
	a.stopOwnedNode()
}

func reopenExisting(base string) bool {
	c := &http.Client{Timeout: 750 * time.Millisecond}
	resp, err := c.Get(base + "/api/state")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false
	}
	var st appState
	if json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&st) != nil || st.Version == "" {
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
