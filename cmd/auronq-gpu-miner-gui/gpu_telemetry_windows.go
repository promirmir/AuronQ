//go:build windows

package main

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"os/exec"
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
}

func findNvidiaSMI() (string, error) {
	if p, err := exec.LookPath("nvidia-smi"); err == nil {
		return p, nil
	}
	candidates := []string{
		filepath.Join(os.Getenv("ProgramW6432"), "NVIDIA Corporation", "NVSMI", "nvidia-smi.exe"),
		filepath.Join(os.Getenv("ProgramFiles"), "NVIDIA Corporation", "NVSMI", "nvidia-smi.exe"),
	}
	for _, p := range candidates {
		if strings.TrimSpace(p) == "" {
			continue
		}
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p, nil
		}
	}
	return "", errors.New("nvidia-smi not found; install/update the NVIDIA driver")
}

func queryNVIDIAGPUs() ([]gpuInfo, error) {
	smi, err := findNvidiaSMI()
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(smi,
		"--query-gpu=index,name,temperature.gpu,fan.speed,utilization.gpu,memory.used,memory.total,power.draw,power.limit",
		"--format=csv,noheader,nounits",
	)
	cmd.SysProcAttr = &syscallSysProcAttr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("nvidia-smi: %w", err)
	}
	r := csv.NewReader(bytes.NewReader(out))
	r.TrimLeadingSpace = true
	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("parse nvidia-smi output: %w", err)
	}
	gpus := make([]gpuInfo, 0, len(records))
	for _, rec := range records {
		if len(rec) < 9 {
			continue
		}
		idx, err := strconv.Atoi(strings.TrimSpace(rec[0]))
		if err != nil {
			continue
		}
		gpus = append(gpus, gpuInfo{
			Index:          idx,
			Name:           strings.TrimSpace(rec[1]),
			TemperatureC:   parseSMIInt(rec[2]),
			FanPercent:     parseSMIInt(rec[3]),
			UtilPercent:    parseSMIInt(rec[4]),
			MemoryUsedMiB:  parseSMIInt(rec[5]),
			MemoryTotalMiB: parseSMIInt(rec[6]),
			PowerW:         parseSMIFloat(rec[7]),
			PowerLimitW:    parseSMIFloat(rec[8]),
		})
	}
	if len(gpus) == 0 {
		return nil, errors.New("no NVIDIA CUDA GPUs reported by nvidia-smi")
	}
	sort.Slice(gpus, func(i, j int) bool { return gpus[i].Index < gpus[j].Index })
	return gpus, nil
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
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p, nil
		}
		return "", fmt.Errorf("pool miner executable not found: %s", p)
	}
	exeDir, err := executableDir()
	if err != nil {
		return "", err
	}
	for _, name := range []string{"meshpool-miner.exe", "meshminer.exe"} {
		p := filepath.Join(exeDir, name)
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p, nil
		}
	}
	return "", errors.New("pool mode needs a compatible external pool miner (for example MeshMiner 0.8.35+); place meshpool-miner.exe next to AuronQ-GPU-Miner.exe or set its path")
}


func (a *App) monitorGPUs() {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	for {
		a.refreshGPUInfo()
		<-ticker.C
	}
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
	cfg := a.cfg
	running := a.miner.Running
	stopping := a.minerStopRequested
	a.mu.Unlock()

	if !running || stopping || cfg.ThermalStopC <= 0 {
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
		if g.TemperatureC >= cfg.ThermalStopC {
			a.addLog(fmt.Sprintf("THERMAL SAFETY: GPU %d reached %d C (limit %d C); stopping miner", g.Index, g.TemperatureC, cfg.ThermalStopC))
			a.stopWorker()
			return
		}
	}
}
