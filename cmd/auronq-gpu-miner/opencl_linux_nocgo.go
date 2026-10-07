//go:build linux && !cgo

package main

import "fmt"

func openOpenCLBackend(path string, device int) (gpuBackend, error) {
	return nil, fmt.Errorf("Linux OpenCL miner requires a CGO-enabled build; use the official Linux amd64 release")
}
