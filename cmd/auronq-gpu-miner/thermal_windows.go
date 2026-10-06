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
	"strconv"
	"strings"
	"syscall"
	"time"
)

type thermalController struct {
	device      int
	targetC     int
	limitC      int
	minBatch    int
	maxBatch    int
	lastCheck   time.Time
	lastTemp    int
	coolSamples int
}

func newThermalController(device, targetC, limitC, maxBatch int) *thermalController {
	if maxBatch < 1 {
		maxBatch = 1
	}
	return &thermalController{
		device:   device,
		targetC:  targetC,
		limitC:   limitC,
		minBatch: 1,
		maxBatch: maxBatch,
		lastTemp: -1,
	}
}

func (t *thermalController) Target() int { return t.targetC }
func (t *thermalController) Limit() int  { return t.limitC }

func (t *thermalController) Adjust(batch int) (int, int, string, error) {
	if t == nil {
		return batch, -1, "", nil
	}
	now := time.Now()
	if !t.lastCheck.IsZero() && now.Sub(t.lastCheck) < 2*time.Second {
		return batch, t.lastTemp, "", nil
	}
	t.lastCheck = now

	temp, err := queryNVIDIATemperature(t.device)
	if err != nil {
		return batch, -1, "", nil
	}
	t.lastTemp = temp

	if temp >= t.limitC {
		return batch, temp, "stop", fmt.Errorf("THERMAL LIMIT: GPU %d reached %d C (limit %d C)", t.device, temp, t.limitC)
	}

	newBatch := batch
	action := ""

	switch {
	case temp >= t.limitC-1:
		newBatch = maxInt(t.minBatch, batch/2)
		action = "critical-reduce"
		t.coolSamples = 0
	case temp >= t.targetC+2:
		step := maxInt(2, batch/4)
		newBatch = maxInt(t.minBatch, batch-step)
		action = "reduce"
		t.coolSamples = 0
	case temp >= t.targetC:
		step := maxInt(1, batch/8)
		newBatch = maxInt(t.minBatch, batch-step)
		action = "trim"
		t.coolSamples = 0
	case temp <= t.targetC-4:
		t.coolSamples++
		if t.coolSamples >= 3 && batch < t.maxBatch {
			step := maxInt(1, t.maxBatch/12)
			newBatch = minInt(t.maxBatch, batch+step)
			action = "increase"
			t.coolSamples = 0
		}
	default:
		t.coolSamples = 0
	}

	if newBatch == batch {
		action = ""
	}
	return newBatch, temp, action, nil
}

func queryNVIDIATemperature(device int) (int, error) {
	smi, err := findWorkerNvidiaSMI()
	if err != nil {
		return -1, err
	}
	cmd := exec.Command(smi,
		"--query-gpu=index,temperature.gpu",
		"--format=csv,noheader,nounits",
	)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	out, err := cmd.Output()
	if err != nil {
		return -1, fmt.Errorf("nvidia-smi temperature: %w", err)
	}
	r := csv.NewReader(bytes.NewReader(out))
	r.TrimLeadingSpace = true
	records, err := r.ReadAll()
	if err != nil {
		return -1, err
	}
	for _, rec := range records {
		if len(rec) < 2 {
			continue
		}
		idx, err := strconv.Atoi(strings.TrimSpace(rec[0]))
		if err != nil || idx != device {
			continue
		}
		temp, err := strconv.Atoi(strings.TrimSpace(rec[1]))
		if err != nil {
			return -1, err
		}
		return temp, nil
	}
	return -1, fmt.Errorf("GPU %d temperature not reported", device)
}

func findWorkerNvidiaSMI() (string, error) {
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
	return "", errors.New("nvidia-smi not found")
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
