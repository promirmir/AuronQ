//go:build windows

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"unsafe"
)

const (
	className = "AuronQGPUMinerGUI"

	WM_DESTROY = 0x0002
	WM_COMMAND = 0x0111
	WM_TIMER   = 0x0113
	WM_CLOSE   = 0x0010
	WM_SETFONT = 0x0030

	WM_USER = 0x0400

	WS_OVERLAPPED   = 0x00000000
	WS_CAPTION      = 0x00C00000
	WS_SYSMENU      = 0x00080000
	WS_MINIMIZEBOX  = 0x00020000
	WS_CHILD        = 0x40000000
	WS_VISIBLE      = 0x10000000
	WS_BORDER       = 0x00800000
	WS_VSCROLL      = 0x00200000
	WS_TABSTOP      = 0x00010000

	ES_LEFT        = 0x0000
	ES_MULTILINE   = 0x0004
	ES_AUTOVSCROLL = 0x0040
	ES_READONLY    = 0x0800
	ES_NUMBER      = 0x2000

	BS_PUSHBUTTON    = 0x00000000
	BS_DEFPUSHBUTTON = 0x00000001
	BS_AUTOCHECKBOX  = 0x00000003

	BM_GETCHECK = 0x00F0
	BM_SETCHECK = 0x00F1
	BST_CHECKED = 1

	EM_SETSEL     = 0x00B1
	EM_REPLACESEL = 0x00C2

	SW_SHOW = 5

	COLOR_WINDOW = 5
	IDC_ARROW     = 32512

	ID_NODE      = 1001
	ID_ADDRESS   = 1002
	ID_DEVICE    = 1003
	ID_BATCH     = 1004
	ID_SELFTEST  = 1005
	ID_START     = 1010
	ID_STOP      = 1011
	ID_TEST      = 1012
	ID_BENCH     = 1013
	ID_STATUS    = 1020
	ID_LOG       = 1021
	timerID      = 1
)

var (
	user32                  = syscall.NewLazyDLL("user32.dll")
	kernel32                = syscall.NewLazyDLL("kernel32.dll")
	gdi32                   = syscall.NewLazyDLL("gdi32.dll")
	procRegisterClassExW    = user32.NewProc("RegisterClassExW")
	procCreateWindowExW     = user32.NewProc("CreateWindowExW")
	procDefWindowProcW      = user32.NewProc("DefWindowProcW")
	procGetMessageW         = user32.NewProc("GetMessageW")
	procTranslateMessage    = user32.NewProc("TranslateMessage")
	procDispatchMessageW    = user32.NewProc("DispatchMessageW")
	procPostQuitMessage     = user32.NewProc("PostQuitMessage")
	procShowWindow          = user32.NewProc("ShowWindow")
	procUpdateWindow        = user32.NewProc("UpdateWindow")
	procSetWindowTextW      = user32.NewProc("SetWindowTextW")
	procGetWindowTextW      = user32.NewProc("GetWindowTextW")
	procGetWindowTextLength = user32.NewProc("GetWindowTextLengthW")
	procSendMessageW        = user32.NewProc("SendMessageW")
	procEnableWindow        = user32.NewProc("EnableWindow")
	procLoadCursorW         = user32.NewProc("LoadCursorW")
	procSetTimer            = user32.NewProc("SetTimer")
	procKillTimer           = user32.NewProc("KillTimer")
	procMessageBoxW         = user32.NewProc("MessageBoxW")
	procSetProcessDPIAware  = user32.NewProc("SetProcessDPIAware")
	procGetModuleHandleW    = kernel32.NewProc("GetModuleHandleW")
	procGetStockObject      = gdi32.NewProc("GetStockObject")
)

type wndClassEx struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     uintptr
	HIcon         uintptr
	HCursor       uintptr
	HbrBackground uintptr
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       uintptr
}

type point struct {
	X int32
	Y int32
}

type msg struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      point
}

type settings struct {
	Node     string `json:"node"`
	Address  string `json:"address"`
	Device   int    `json:"device"`
	Batch    int    `json:"batch"`
	SelfTest bool   `json:"self_test"`
}

type appState struct {
	hwnd     uintptr
	node     uintptr
	address  uintptr
	device   uintptr
	batch    uintptr
	selftest uintptr
	start    uintptr
	stop     uintptr
	test     uintptr
	bench    uintptr
	status   uintptr
	log      uintptr

	mu          sync.Mutex
	pendingLogs []string
	nextStatus  string
	running     bool
	cmd         *exec.Cmd
}

var app appState

func utf16Ptr(s string) *uint16 {
	p, _ := syscall.UTF16PtrFromString(s)
	return p
}

func loword(v uintptr) uint16 { return uint16(v & 0xffff) }

func setText(hwnd uintptr, s string) {
	procSetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(utf16Ptr(s))))
}

func getText(hwnd uintptr) string {
	n, _, _ := procGetWindowTextLength.Call(hwnd)
	buf := make([]uint16, int(n)+1)
	procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	return syscall.UTF16ToString(buf)
}

func send(hwnd uintptr, msg uint32, wp, lp uintptr) uintptr {
	r, _, _ := procSendMessageW.Call(hwnd, uintptr(msg), wp, lp)
	return r
}

func appendLogText(hwnd uintptr, s string) {
	if s == "" {
		return
	}
	s = strings.ReplaceAll(s, "\n", "\r\n")
	send(hwnd, EM_SETSEL, ^uintptr(0), ^uintptr(0))
	send(hwnd, EM_REPLACESEL, 0, uintptr(unsafe.Pointer(utf16Ptr(s))))
}

func createControl(class, text string, style uint32, x, y, w, h int32, parent uintptr, id int) uintptr {
	hwnd, _, _ := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(utf16Ptr(class))),
		uintptr(unsafe.Pointer(utf16Ptr(text))),
		uintptr(style),
		uintptr(x), uintptr(y), uintptr(w), uintptr(h),
		parent, uintptr(id), 0, 0,
	)
	font, _, _ := procGetStockObject.Call(17) // DEFAULT_GUI_FONT
	send(hwnd, WM_SETFONT, font, 1)
	return hwnd
}

func configPath() string {
	base, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(base, "AuronQ", "gpu-miner.json")
}

func loadSettings() settings {
	s := settings{Node: "http://127.0.0.1:18444", Device: 0, Batch: 60, SelfTest: true}
	p := configPath()
	if p == "" {
		return s
	}
	b, err := os.ReadFile(p)
	if err == nil {
		_ = json.Unmarshal(b, &s)
	}
	if s.Node == "" {
		s.Node = "http://127.0.0.1:18444"
	}
	if s.Batch < 1 || s.Batch > 64 {
		s.Batch = 60
	}
	return s
}

func saveSettings(s settings) {
	p := configPath()
	if p == "" {
		return
	}
	_ = os.MkdirAll(filepath.Dir(p), 0o755)
	b, err := json.MarshalIndent(s, "", "  ")
	if err == nil {
		_ = os.WriteFile(p, b, 0o600)
	}
}

func currentSettings() (settings, error) {
	device, err := strconv.Atoi(strings.TrimSpace(getText(app.device)))
	if err != nil || device < 0 {
		return settings{}, fmt.Errorf("Device must be 0 or a positive integer")
	}
	batch, err := strconv.Atoi(strings.TrimSpace(getText(app.batch)))
	if err != nil || batch < 1 || batch > 64 {
		return settings{}, fmt.Errorf("Batch must be between 1 and 64")
	}
	s := settings{
		Node:     strings.TrimSpace(getText(app.node)),
		Address:  strings.TrimSpace(getText(app.address)),
		Device:   device,
		Batch:    batch,
		SelfTest: send(app.selftest, BM_GETCHECK, 0, 0) == BST_CHECKED,
	}
	if s.Node == "" {
		return settings{}, fmt.Errorf("Node URL is required")
	}
	return s, nil
}

func (a *appState) queueLog(s string) {
	a.mu.Lock()
	a.pendingLogs = append(a.pendingLogs, s)
	a.mu.Unlock()
}

func (a *appState) queueStatus(s string) {
	a.mu.Lock()
	a.nextStatus = s
	a.mu.Unlock()
}

func (a *appState) drainUI() {
	a.mu.Lock()
	logs := append([]string(nil), a.pendingLogs...)
	a.pendingLogs = a.pendingLogs[:0]
	status := a.nextStatus
	a.nextStatus = ""
	a.mu.Unlock()

	for _, line := range logs {
		appendLogText(a.log, line+"\n")
	}
	if status != "" {
		setText(a.status, status)
	}
}

func minerExe() string {
	exe, err := os.Executable()
	if err != nil {
		return "auronq-gpu-miner.exe"
	}
	return filepath.Join(filepath.Dir(exe), "auronq-gpu-miner.exe")
}

func (a *appState) setRunning(running bool) {
	a.mu.Lock()
	a.running = running
	a.mu.Unlock()
	if running {
		procEnableWindow.Call(a.start, 0)
		procEnableWindow.Call(a.test, 0)
		procEnableWindow.Call(a.bench, 0)
		procEnableWindow.Call(a.stop, 1)
	} else {
		procEnableWindow.Call(a.start, 1)
		procEnableWindow.Call(a.test, 1)
		procEnableWindow.Call(a.bench, 1)
		procEnableWindow.Call(a.stop, 0)
	}
}

func (a *appState) isRunning() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.running
}

func (a *appState) run(mode string) {
	if a.isRunning() {
		return
	}
	s, err := currentSettings()
	if err != nil {
		messageBox("AuronQ GPU Miner", err.Error())
		return
	}
	if mode == "mine" && s.Address == "" {
		messageBox("AuronQ GPU Miner", "Enter the AURQ reward address before starting mining.")
		return
	}
	saveSettings(s)

	args := []string{"--device", strconv.Itoa(s.Device), "--batch", strconv.Itoa(s.Batch)}
	switch mode {
	case "selftest":
		args = append(args, "--self-test")
	case "benchmark":
		args = append(args, "--benchmark", "--benchmark-seconds", "15")
	case "mine":
		args = append(args, "--node", s.Node, "--address", s.Address)
		if s.SelfTest {
			args = append(args, "--self-test")
		}
	}

	exe := minerExe()
	cmd := exec.Command(exe, args...)
	cmd.Dir = filepath.Dir(exe)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		messageBox("AuronQ GPU Miner", err.Error())
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		messageBox("AuronQ GPU Miner", err.Error())
		return
	}
	if err := cmd.Start(); err != nil {
		messageBox("AuronQ GPU Miner", "Cannot start miner:\n"+err.Error())
		return
	}

	a.mu.Lock()
	a.cmd = cmd
	a.mu.Unlock()
	a.setRunning(true)
	a.queueLog("")
	a.queueLog("=== "+strings.ToUpper(mode)+" ===")
	a.queueStatus("Running "+mode+"...")

	scan := func(prefix string, r *bufio.Scanner) {
		for r.Scan() {
			line := r.Text()
			a.queueLog(prefix + line)
			switch {
			case strings.Contains(line, "BLOCK FOUND"):
				a.queueStatus("BLOCK FOUND - mining continues")
			case strings.Contains(line, "SELF-TEST OK"):
				a.queueStatus("SELF-TEST OK")
			case strings.Contains(line, "BENCHMARK OK"):
				a.queueStatus("Benchmark complete")
			case strings.Contains(line, "Mining height"):
				a.queueStatus(line)
			}
		}
	}
	outScanner := bufio.NewScanner(stdout)
	errScanner := bufio.NewScanner(stderr)
	outScanner.Buffer(make([]byte, 4096), 1024*1024)
	errScanner.Buffer(make([]byte, 4096), 1024*1024)
	go scan("", outScanner)
	go scan("ERROR: ", errScanner)

	go func() {
		err := cmd.Wait()
		a.mu.Lock()
		a.cmd = nil
		a.mu.Unlock()
		a.setRunning(false)
		if err != nil {
			a.queueLog("Process stopped: " + err.Error())
			a.queueStatus("Stopped with error")
		} else {
			a.queueStatus("Stopped")
		}
	}()
}

func (a *appState) stopMiner() {
	a.mu.Lock()
	cmd := a.cmd
	a.mu.Unlock()
	if cmd != nil && cmd.Process != nil {
		_ = cmd.Process.Kill()
		a.queueStatus("Stopping...")
	}
}

func messageBox(title, text string) {
	procMessageBoxW.Call(0,
		uintptr(unsafe.Pointer(utf16Ptr(text))),
		uintptr(unsafe.Pointer(utf16Ptr(title))),
		0x00000040)
}

func wndProc(hwnd uintptr, message uint32, wParam, lParam uintptr) uintptr {
	switch message {
	case WM_COMMAND:
		switch loword(wParam) {
		case ID_START:
			app.run("mine")
		case ID_STOP:
			app.stopMiner()
		case ID_TEST:
			app.run("selftest")
		case ID_BENCH:
			app.run("benchmark")
		}
		return 0

	case WM_TIMER:
		if wParam == timerID {
			app.drainUI()
		}
		return 0

	case WM_CLOSE:
		app.stopMiner()
		procKillTimer.Call(hwnd, timerID)
		procPostQuitMessage.Call(0)
		return 0

	case WM_DESTROY:
		app.stopMiner()
		procPostQuitMessage.Call(0)
		return 0
	}
	r, _, _ := procDefWindowProcW.Call(hwnd, uintptr(message), wParam, lParam)
	return r
}

func buildUI(hwnd uintptr) {
	s := loadSettings()

	createControl("STATIC", "AuronQ GPU Miner", WS_CHILD|WS_VISIBLE, 20, 15, 280, 28, hwnd, 0)
	createControl("STATIC", "CUDA Mainnet miner - AQM64", WS_CHILD|WS_VISIBLE, 20, 42, 300, 20, hwnd, 0)

	createControl("STATIC", "Full node URL:", WS_CHILD|WS_VISIBLE, 20, 78, 120, 22, hwnd, 0)
	app.node = createControl("EDIT", s.Node, WS_CHILD|WS_VISIBLE|WS_BORDER|WS_TABSTOP|ES_LEFT, 145, 75, 665, 26, hwnd, ID_NODE)

	createControl("STATIC", "Reward address:", WS_CHILD|WS_VISIBLE, 20, 113, 120, 22, hwnd, 0)
	app.address = createControl("EDIT", s.Address, WS_CHILD|WS_VISIBLE|WS_BORDER|WS_TABSTOP|ES_LEFT, 145, 110, 665, 26, hwnd, ID_ADDRESS)

	createControl("STATIC", "CUDA device:", WS_CHILD|WS_VISIBLE, 20, 148, 105, 22, hwnd, 0)
	app.device = createControl("EDIT", strconv.Itoa(s.Device), WS_CHILD|WS_VISIBLE|WS_BORDER|WS_TABSTOP|ES_NUMBER, 125, 145, 55, 26, hwnd, ID_DEVICE)

	createControl("STATIC", "Batch:", WS_CHILD|WS_VISIBLE, 205, 148, 50, 22, hwnd, 0)
	app.batch = createControl("EDIT", strconv.Itoa(s.Batch), WS_CHILD|WS_VISIBLE|WS_BORDER|WS_TABSTOP|ES_NUMBER, 255, 145, 55, 26, hwnd, ID_BATCH)

	app.selftest = createControl("BUTTON", "Run GPU/CPU self-test before mining", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_AUTOCHECKBOX, 340, 145, 280, 26, hwnd, ID_SELFTEST)
	if s.SelfTest {
		send(app.selftest, BM_SETCHECK, BST_CHECKED, 0)
	}

	app.start = createControl("BUTTON", "Start mining", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_DEFPUSHBUTTON, 20, 190, 145, 34, hwnd, ID_START)
	app.stop = createControl("BUTTON", "Stop", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON, 175, 190, 100, 34, hwnd, ID_STOP)
	app.test = createControl("BUTTON", "Self-test", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON, 300, 190, 120, 34, hwnd, ID_TEST)
	app.bench = createControl("BUTTON", "15 s benchmark", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON, 430, 190, 150, 34, hwnd, ID_BENCH)
	procEnableWindow.Call(app.stop, 0)

	createControl("STATIC", "Status:", WS_CHILD|WS_VISIBLE, 20, 241, 55, 20, hwnd, 0)
	app.status = createControl("STATIC", "Ready", WS_CHILD|WS_VISIBLE, 78, 241, 730, 20, hwnd, ID_STATUS)

	createControl("STATIC", "Miner log:", WS_CHILD|WS_VISIBLE, 20, 270, 100, 20, hwnd, 0)
	app.log = createControl("EDIT", "", WS_CHILD|WS_VISIBLE|WS_BORDER|WS_VSCROLL|ES_LEFT|ES_MULTILINE|ES_AUTOVSCROLL|ES_READONLY, 20, 292, 790, 300, hwnd, ID_LOG)

	createControl("STATIC", "Tip: keep AuronQ Desktop/full node synchronized while mining. Batch 60 is validated on RTX 4050 Laptop GPU.", WS_CHILD|WS_VISIBLE, 20, 605, 790, 20, hwnd, 0)
}

func main() {
	procSetProcessDPIAware.Call()
	hInst, _, _ := procGetModuleHandleW.Call(0)
	cursor, _, _ := procLoadCursorW.Call(0, IDC_ARROW)

	wc := wndClassEx{
		CbSize:        uint32(unsafe.Sizeof(wndClassEx{})),
		LpfnWndProc:   syscall.NewCallback(wndProc),
		HInstance:     hInst,
		HCursor:       cursor,
		HbrBackground: COLOR_WINDOW + 1,
		LpszClassName: utf16Ptr(className),
	}
	if r, _, _ := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc))); r == 0 {
		messageBox("AuronQ GPU Miner", "Could not register Windows window class.")
		return
	}

	style := uint32(WS_OVERLAPPED | WS_CAPTION | WS_SYSMENU | WS_MINIMIZEBOX)
	hwnd, _, _ := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(utf16Ptr(className))),
		uintptr(unsafe.Pointer(utf16Ptr("AuronQ GPU Miner"))),
		uintptr(style),
		200, 100, 850, 680,
		0, 0, hInst, 0,
	)
	if hwnd == 0 {
		messageBox("AuronQ GPU Miner", "Could not create the main window.")
		return
	}
	app.hwnd = hwnd
	buildUI(hwnd)
	procSetTimer.Call(hwnd, timerID, 150, 0)
	procShowWindow.Call(hwnd, SW_SHOW)
	procUpdateWindow.Call(hwnd)

	var m msg
	for {
		r, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(r) <= 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}
}
