//go:build !windows && !linux

package main

import "fmt"

func openCUDABackend(path string, device int) (gpuBackend, error) {
	return nil, fmt.Errorf("AuronQ GPU Miner CUDA v0.1 currently supports Windows x64 only")
}
