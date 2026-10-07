//go:build windows

package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	gt "auronq/internal/gputelemetry"
)

type multiGPUOptions struct {
	Backend          string
	Node             string
	Address          string
	Batch            int
	DLLPath          string
	LegacyDLLPath    string
	KeplerDLLPath    string
	OpenCLPath       string
	SelfTest         bool
	Benchmark        bool
	BenchmarkSeconds int
	AutoTune         bool
	AutoTuneSeconds  int
	ThermalAuto      bool
	ThermalLimit     int
	ThermalTarget    int
	NoncePrefix      uint64
}

type multiGPUState struct {
	mu      sync.Mutex
	rates   map[int]float64
	heights map[int]uint64
}

type multiGPUChild struct {
	device int
	cmd    *exec.Cmd
}

func resolveCUDADevices(raw string) ([]int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, errors.New("empty CUDA device selection")
	}
	if strings.EqualFold(raw, "all") {
		gpus, err := gt.QueryNVIDIA()
		if err != nil {
			return nil, fmt.Errorf("detect CUDA devices from direct local NVIDIA telemetry: %w", err)
		}
		parts := make([]string, 0, len(gpus))
		for _, gpu := range gpus {
			parts = append(parts, strconv.Itoa(gpu.Index))
		}
		raw = strings.Join(parts, ",")
	}

	seen := map[int]bool{}
	var devices []int
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		v, err := strconv.Atoi(part)
		if err != nil || v < 0 {
			return nil, fmt.Errorf("invalid CUDA device %q", part)
		}
		if !seen[v] {
			seen[v] = true
			devices = append(devices, v)
		}
	}
	if len(devices) == 0 {
		return nil, errors.New("no CUDA devices selected")
	}
	sort.Ints(devices)
	return devices, nil
}

func runMultiGPU(devices []int, opt multiGPUOptions) error {
	if len(devices) < 2 {
		return errors.New("multi-GPU runner needs at least two devices")
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	state := &multiGPUState{
		rates:   make(map[int]float64, len(devices)),
		heights: make(map[int]uint64, len(devices)),
	}
	children := make([]multiGPUChild, 0, len(devices))
	done := make(chan error, len(devices))

	for slot, device := range devices {
		prefix := opt.NoncePrefix + (uint64(slot) << 56)
		args := []string{
			"--multi-child",
			"--backend", opt.Backend,
			"--device", strconv.Itoa(device),
			"--batch", strconv.Itoa(opt.Batch),
			"--cuda-dll", opt.DLLPath,
			"--cuda-legacy-dll", opt.LegacyDLLPath,
			"--cuda-kepler-dll", opt.KeplerDLLPath,
			"--opencl-dll", opt.OpenCLPath,
			"--nonce-prefix", strconv.FormatUint(prefix, 10),
		}
		if opt.SelfTest {
			args = append(args, "--self-test")
		}
		if opt.AutoTune {
			args = append(args, "--auto-tune", "--auto-tune-seconds", strconv.Itoa(opt.AutoTuneSeconds))
		}
		if opt.ThermalAuto && !opt.Benchmark {
			args = append(args, "--thermal-auto", "--thermal-limit", strconv.Itoa(opt.ThermalLimit))
			if opt.ThermalTarget > 0 {
				args = append(args, "--thermal-target", strconv.Itoa(opt.ThermalTarget))
			}
		}
		if opt.Benchmark {
			args = append(args, "--benchmark", "--benchmark-seconds", strconv.Itoa(opt.BenchmarkSeconds))
		} else {
			args = append(args, "--node", opt.Node)
			if strings.TrimSpace(opt.Address) != "" {
				args = append(args, "--address", opt.Address)
			}
		}

		cmd := exec.Command(exe, args...)
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			killChildren(children)
			return err
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			killChildren(children)
			return err
		}
		if err := cmd.Start(); err != nil {
			killChildren(children)
			return fmt.Errorf("start GPU %d worker: %w", device, err)
		}
		children = append(children, multiGPUChild{device: device, cmd: cmd})
		fmt.Printf("MULTI GPU %d worker started pid=%d nonce_base=%d\n", device, cmd.Process.Pid, prefix)

		go scanMultiStream(stdout, device, state, false)
		go scanMultiStream(stderr, device, state, true)
		go func(device int, cmd *exec.Cmd) {
			err := cmd.Wait()
			if err != nil {
				done <- fmt.Errorf("GPU %d worker stopped: %w", device, err)
				return
			}
			done <- nil
		}(device, cmd)
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	remaining := len(children)
	var firstErr error
	for remaining > 0 {
		select {
		case err := <-done:
			remaining--
			if err != nil && firstErr == nil {
				firstErr = err
				killChildren(children)
			}
		case <-ticker.C:
			rate, height := state.snapshot()
			fmt.Printf("hashes=0 rate=%.2f H/s avg=%.2f H/s current_height=%d multi_gpu=%d\n", rate, rate, height, len(devices))
		}
	}

	rate, _ := state.snapshot()
	if opt.Benchmark {
		fmt.Printf("BENCHMARK OK hashes=0 elapsed=multi avg=%.3f H/s batch=per-device\n", rate)
	}
	if firstErr != nil {
		return firstErr
	}
	return nil
}

func killChildren(children []multiGPUChild) {
	for _, child := range children {
		if child.cmd != nil && child.cmd.Process != nil {
			_ = child.cmd.Process.Kill()
		}
	}
}

func scanMultiStream(r io.Reader, device int, state *multiGPUState, stderr bool) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 4096), 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "hashes=") || strings.HasPrefix(line, "BENCHMARK OK") {
			rate, ok := parseMultiNumberAfter(line, "rate=")
			if !ok {
				rate, ok = parseMultiNumberAfter(line, "avg=")
			}
			if ok {
				state.mu.Lock()
				state.rates[device] = rate
				state.mu.Unlock()
			}
		}
		if h, ok := parseMultiUintAfter(line, "current_height="); ok {
			state.mu.Lock()
			state.heights[device] = h
			state.mu.Unlock()
		}
		if strings.HasPrefix(line, "Mining height ") {
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				if h, err := strconv.ParseUint(fields[2], 10, 64); err == nil {
					state.mu.Lock()
					state.heights[device] = h
					state.mu.Unlock()
				}
			}
		}

		prefix := fmt.Sprintf("[GPU %d] ", device)
		if stderr {
			prefix += "ERROR: "
		}
		fmt.Println(prefix + line)
		if strings.HasPrefix(line, "BLOCK FOUND ") {
			fmt.Printf("%s gpu=%d\n", line, device)
		}
	}
}

func (s *multiGPUState) snapshot() (float64, uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	total := 0.0
	var height uint64
	for _, rate := range s.rates {
		total += rate
	}
	for _, h := range s.heights {
		if h > height {
			height = h
		}
	}
	return total, height
}

func parseMultiNumberAfter(line, key string) (float64, bool) {
	i := strings.Index(line, key)
	if i < 0 {
		return 0, false
	}
	s := line[i+len(key):]
	if j := strings.IndexAny(s, " \t"); j >= 0 {
		s = s[:j]
	}
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return v, err == nil
}

func parseMultiUintAfter(line, key string) (uint64, bool) {
	i := strings.Index(line, key)
	if i < 0 {
		return 0, false
	}
	s := line[i+len(key):]
	if j := strings.IndexAny(s, " \t"); j >= 0 {
		s = s[:j]
	}
	v, err := strconv.ParseUint(strings.TrimSpace(s), 10, 64)
	return v, err == nil
}
