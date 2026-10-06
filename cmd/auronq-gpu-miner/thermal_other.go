//go:build !windows

package main

type thermalController struct{}

func newThermalController(device, targetC, limitC, maxBatch int) *thermalController { return nil }
func (t *thermalController) Target() int { return 0 }
func (t *thermalController) Limit() int { return 0 }
func (t *thermalController) Adjust(batch int) (int, int, string, error) {
	return batch, -1, "", nil
}
