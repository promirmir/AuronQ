//go:build linux

package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
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

func (t *thermalController) Adjust(batch int) (int, int, time.Duration, string, error) {
	if t == nil {
		return batch, -1, 0, "", nil
	}
	now := time.Now()
	if !t.lastCheck.IsZero() && now.Sub(t.lastCheck) < 500*time.Millisecond {
		return batch, t.lastTemp, t.cooldownFor(t.lastTemp), "", nil
	}
	t.lastCheck = now

	temp, err := queryNVIDIATemperature(t.device)
	if err != nil {
		return batch, -1, 0, "stop", fmt.Errorf("THERMAL TELEMETRY FAILSAFE: local NVIDIA temperature unavailable for GPU %d: %w", t.device, err)
	}
	if temp < 0 || temp > 125 {
		return batch, temp, 0, "stop", fmt.Errorf("THERMAL SENSOR INVALID: GPU %d returned %d C", t.device, temp)
	}
	t.lastTemp = temp
	if temp >= t.limitC {
		return batch, temp, 0, "stop", fmt.Errorf("THERMAL LIMIT: GPU %d reached %d C (limit %d C)", t.device, temp, t.limitC)
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
		// Braiins-style staged recovery: regain throughput cautiously after
		// consecutive cool samples; retain our original hard thermal stop.
		t.coolSamples++
		if t.coolSamples >= 3 && batch < t.maxBatch {
			step := maxInt(1, t.maxBatch/12)
			newBatch = minInt(t.maxBatch, batch+step)
			action = "increase"
			t.coolSamples = 0
		}
	case temp <= t.targetC-2:
		// Recover more slowly in the near-target band; previous logic
		// could leave a throttled GPU stuck indefinitely at target-2C.
		t.coolSamples++
		if t.coolSamples >= 6 && batch < t.maxBatch {
			newBatch = minInt(t.maxBatch, batch+1)
			action = "increase"
			t.coolSamples = 0
		}
	case temp == t.targetC-1:
		// Very slow recovery one degree below target. The next hot sample
		// still immediately trims batch and applies the existing duty pause.
		t.coolSamples++
		if t.coolSamples >= 10 && batch < t.maxBatch {
			newBatch = minInt(t.maxBatch, batch+1)
			action = "increase"
			t.coolSamples = 0
		}
	default:
		t.coolSamples = 0
	}
	if newBatch == batch {
		action = ""
	}
	return newBatch, temp, t.cooldownFor(temp), action, nil
}

func (t *thermalController) cooldownFor(temp int) time.Duration {
	if temp < 0 || temp < t.targetC {
		return 0
	}
	delta := temp - t.targetC
	switch {
	case temp >= t.limitC-1:
		return 900 * time.Millisecond
	case delta >= 3:
		return 500 * time.Millisecond
	case delta == 2:
		return 300 * time.Millisecond
	case delta == 1:
		return 160 * time.Millisecond
	default:
		return 80 * time.Millisecond
	}
}

func queryNVIDIATemperature(device int) (int, error) {
	smi, err := exec.LookPath("nvidia-smi")
	if err != nil {
		return -1, fmt.Errorf("nvidia-smi not found: %w", err)
	}
	cmd := exec.Command(smi, "--query-gpu=index,temperature.gpu", "--format=csv,noheader,nounits")
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

func maxInt(a, b int) int {
	if a > b { return a }
	return b
}
func minInt(a, b int) int {
	if a < b { return a }
	return b
}
