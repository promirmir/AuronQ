//go:build windows

package main

import gt "auronq/internal/gputelemetry"

func queryDirectNVIDIAGPUs() ([]gpuInfo, error) {
	devices, err := gt.QueryNVIDIA()
	if err != nil {
		return nil, err
	}
	out := make([]gpuInfo, 0, len(devices))
	for _, d := range devices {
		out = append(out, gpuInfo{
			Index:           d.Index,
			Name:            d.Name,
			TemperatureC:    d.TemperatureC,
			FanPercent:      d.FanPercent,
			UtilPercent:     d.UtilPercent,
			MemoryUsedMiB:   d.MemoryUsedMiB,
			MemoryTotalMiB:  d.MemoryTotalMiB,
			PowerW:          d.PowerW,
			PowerLimitW:     d.PowerLimitW,
			CoreClockMHz:    d.CoreClockMHz,
			MemoryClockMHz:  d.MemoryClockMHz,
			PState:          d.PState,
			TelemetrySource: d.Source,
			SampleUnixMS:    d.SampleUnixMS,
		})
	}
	return out, nil
}
