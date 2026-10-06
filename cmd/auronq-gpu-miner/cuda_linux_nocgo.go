//go:build linux && !cgo

package main

import "fmt"

func openCUDABackend(path string, device int) (gpuBackend, error) {
	return nil, fmt.Errorf("Linux CUDA miner requires a CGO-enabled build; use the official Linux amd64 release")
}
