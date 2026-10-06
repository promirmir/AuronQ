//go:build !windows

package gputelemetry

import "errors"

func QueryNVIDIA() ([]NVIDIADevice, error) {
	return nil, errors.New("direct NVIDIA NVML telemetry is currently implemented on Windows only")
}

func Temperature(device int) (int, error) {
	return -1, errors.New("direct NVIDIA NVML temperature is currently implemented on Windows only")
}
