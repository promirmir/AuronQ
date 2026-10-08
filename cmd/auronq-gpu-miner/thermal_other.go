//go:build !windows && !linux

package main

import (
 "fmt"
 "time"
)

// Keep field-compatible with Windows/Linux so portable CLI builds compile.
// This unsupported platform never activates the thermal controller.
type thermalController struct { lastTemp int; maxBatch int }

func newThermalController(device, targetC, limitC, maxBatch int) *thermalController { return nil }
func (t *thermalController) Target() int { return 0 }
func (t *thermalController) Limit() int { return 0 }
func (t *thermalController) Adjust(batch int) (int, int, time.Duration, string, error) {
	return batch, -1, 0, "", nil
}

// Platforms without a trusted local NVIDIA sensor must fail closed for CUDA thermal autotune.
func queryNVIDIATemperature(device int) (int, error) {
 return -1, fmt.Errorf("NVIDIA temperature telemetry unsupported on this platform (device %d)", device)
}
