//go:build !windows && !linux

package main

import "time"

type thermalController struct{}

func newThermalController(device, targetC, limitC, maxBatch int) *thermalController { return nil }
func (t *thermalController) Target() int { return 0 }
func (t *thermalController) Limit() int { return 0 }
func (t *thermalController) Adjust(batch int) (int, int, time.Duration, string, error) {
	return batch, -1, 0, "", nil
}
