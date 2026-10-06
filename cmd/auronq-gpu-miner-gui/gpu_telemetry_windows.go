//go:build windows

package main

import (
		"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type gpuInfo struct {
	Index          int     `json:"index"`
	Name           string  `json:"name"`
	TemperatureC   int     `json:"temperature_c"`
	FanPercent     int     `json:"fan_percent"`
	UtilPercent    int     `json:"util_percent"`
	MemoryUsedMiB  int     `json:"memory_used_mib"`
	MemoryTotalMiB int     `json:"memory_total_mib"`
	PowerW         float64 `json:"power_w"`
	PowerLimitW    float64 `json:"power_limit_w"`
	CoreClockMHz   int     `json:"core_clock_mhz"`
	MemoryClockMHz int     `json:"memory_clock_mhz"`
	PState         string  `json:"pstate,omitempty"`
	TelemetrySource string  `json:"telemetry_source,omitempty"`
	SampleUnixMS    int64   `json:"sample_unix_ms,omitempty"`
}

func queryNVIDIAGPUs() ([]gpuInfo, error) {
	// Safety-critical telemetry is read directly from the local NVIDIA driver
	// through NVML. Pool/miner stdout, websites and remote services are never
	// used for temperature control.
	return queryDirectNVIDIAGPUs()
}

func parseSMIInt(s string) int {
	s = strings.TrimSpace(s)
	if s == "" || strings.EqualFold(s, "N/A") || strings.HasPrefix(s, "[") {
		return -1
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return -1
	}
	return int(v + 0.5)
}

func parseSMIString(s string) string {
	s = strings.TrimSpace(s)
	if s == "" || strings.EqualFold(s, "N/A") || strings.HasPrefix(s, "[") {
		return ""
	}
	return s
}

func parseSMIFloat(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" || strings.EqualFold(s, "N/A") || strings.HasPrefix(s, "[") {
		return -1
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return -1
	}
	return v
}

func selectedDeviceIndices(s settings, gpus []gpuInfo) ([]int, error) {
	available := make(map[int]bool, len(gpus))
	for _, g := range gpus {
		available[g.Index] = true
	}

	var requested []int
	if s.UseAllDevices {
		for _, g := range gpus {
			requested = append(requested, g.Index)
		}
	} else if len(s.Devices) > 0 {
		requested = append(requested, s.Devices...)
	} else {
		requested = []int{s.Device}
	}

	seen := map[int]bool{}
	out := make([]int, 0, len(requested))
	for _, d := range requested {
		if d < 0 {
			return nil, fmt.Errorf("CUDA device must be 0 or greater")
		}
		if !available[d] {
			return nil, fmt.Errorf("CUDA device %d is not available", d)
		}
		if !seen[d] {
			seen[d] = true
			out = append(out, d)
		}
	}
	sort.Ints(out)
	if len(out) == 0 {
		return nil, errors.New("select at least one NVIDIA GPU")
	}
	return out, nil
}

func deviceCSV(devices []int) string {
	parts := make([]string, 0, len(devices))
	for _, d := range devices {
		parts = append(parts, strconv.Itoa(d))
	}
	return strings.Join(parts, ",")
}

func normalizePoolEndpoint(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	lower := strings.ToLower(s)
	if strings.HasPrefix(lower, "stratum+ssl://") || strings.HasPrefix(lower, "ssl://") {
		return "", errors.New("TLS pool URLs are not translated by the pool bridge; use an endpoint supported directly by the external miner")
	}
	for _, prefix := range []string{"stratum+tcp://", "tcp://"} {
		if strings.HasPrefix(lower, prefix) {
			s = s[len(prefix):]
			break
		}
	}
	s = strings.TrimRight(strings.TrimSpace(s), "/")
	if s == "" || !strings.Contains(s, ":") || strings.ContainsAny(s, " \t\r\n") {
		return "", errors.New("pool endpoint must be host:port or stratum+tcp://host:port")
	}
	return s, nil
}

func resolvePoolMinerPath(configured string) (string, error) {
	if p := strings.TrimSpace(configured); p != "" {
		p = strings.Trim(p, "\"")
		if !filepath.IsAbs(p) {
			if exeDir, err := executableDir(); err == nil {
				p = filepath.Join(exeDir, p)
			}
		}
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p, nil
		}
		return "", fmt.Errorf("pool miner executable not found: %s", p)
	}

	exeDir, err := executableDir()
	if err != nil {
		return "", err
	}
	names := []string{"meshpool-miner.exe", "meshminer.exe"}
	for _, name := range names {
		p := filepath.Join(exeDir, name)
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p, nil
		}
	}

	// MeshMiner's Windows ZIP is commonly extracted as a folder next to the
	// AuronQ miner. Search one directory level down without scanning the disk.
	entries, _ := os.ReadDir(exeDir)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		lower := strings.ToLower(entry.Name())
		if !strings.Contains(lower, "meshminer") && !strings.Contains(lower, "meshpool") {
			continue
		}
		for _, name := range names {
			p := filepath.Join(exeDir, entry.Name(), name)
			if st, err := os.Stat(p); err == nil && !st.IsDir() {
				return p, nil
			}
		}
	}
	return "", errors.New("MeshMiner 0.8.35+ not found; extract its Windows ZIP next to AuronQ-GPU-Miner.exe or enter the path to meshpool-miner.exe")
}


func (a *App) monitorGPUs() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		a.refreshGPUInfo()
		<-ticker.C
	}
}

func (a *App) hardwareTelemetrySnapshot(maxAge time.Duration) ([]gpuInfo, error) {
	a.mu.RLock()
	gpus := append([]gpuInfo(nil), a.gpus...)
	at := a.gpuTelemetryAt
	errText := a.gpuTelemetryErr
	a.mu.RUnlock()

	if len(gpus) == 0 {
		if errText != "" {
			return nil, errors.New(errText)
		}
		return nil, errors.New("no direct NVIDIA hardware telemetry sample available")
	}
	if at.IsZero() || time.Since(at) > maxAge {
		return nil, fmt.Errorf("direct NVIDIA hardware telemetry is stale (%s)", time.Since(at).Round(100*time.Millisecond))
	}
	for _, g := range gpus {
		if g.TelemetrySource != "NVML_DIRECT" || g.TemperatureC < 0 {
			return nil, errors.New("direct NVML temperature sample unavailable")
		}
	}
	return gpus, nil
}

func (a *App) refreshGPUInfo() {
	gpus, err := queryNVIDIAGPUs()
	if err != nil {
		a.mu.Lock()
		a.gpuTelemetryErr = err.Error()
		a.mu.Unlock()
		return
	}

	a.mu.Lock()
	a.gpus = append([]gpuInfo(nil), gpus...)
	a.gpuTelemetryErr = ""
	a.gpuTelemetryAt = time.Now()
	cfg := a.cfg
	running := a.miner.Running
	mode := a.miner.Mode
	backend := a.miner.Backend
	stopping := a.minerStopRequested
	a.mu.Unlock()

	if !running || stopping || cfg.ThermalStopC <= 0 || backend == "cpu" {
		return
	}
	devices, err := selectedDeviceIndices(cfg, gpus)
	if err != nil {
		return
	}
	selected := make(map[int]bool, len(devices))
	for _, d := range devices {
		selected[d] = true
	}
	for _, g := range gpus {
		if !selected[g.Index] || g.TemperatureC < 0 {
			continue
		}

		// Pool mining has its own 2 Hz autonomous governor. At the configured
		// limit it first suspends the external miner, cools the GPU and resumes
		// automatically. Keep this slower monitor as an independent catastrophic
		// fail-safe only, so it does not race the emergency cooldown path.
		stopAt := cfg.ThermalStopC
		if mode == "pool" {
			stopAt += 1
		}
		if g.TemperatureC >= stopAt {
			if mode == "pool" {
				a.addLog(fmt.Sprintf("THERMAL FAILSAFE: GPU %d reached %d C despite direct-NVML pool governor (limit %d C); stopping miner", g.Index, g.TemperatureC, cfg.ThermalStopC))
			} else {
				a.addLog(fmt.Sprintf("THERMAL SAFETY: GPU %d reached %d C (limit %d C); stopping miner", g.Index, g.TemperatureC, cfg.ThermalStopC))
			}
			a.stopWorker()
			return
		}
	}
}
