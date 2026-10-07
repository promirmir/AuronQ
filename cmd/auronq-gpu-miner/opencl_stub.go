//go:build !windows && !linux

package main

import "fmt"

func openOpenCLBackend(path string, device int) (gpuBackend, error) {
	return nil, fmt.Errorf("OpenCL GPU acceleration is not available in this package; use --backend cpu or a supported Windows/Linux GPU build")
}
