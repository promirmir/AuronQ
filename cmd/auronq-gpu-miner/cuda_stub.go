//go:build !windows && !linux

package main

import "fmt"

func openCUDABackend(path string, device int) (gpuBackend, error) {
	return nil, fmt.Errorf("NVIDIA CUDA acceleration is not available on this platform; use --backend auto or --backend cpu")
}
