package main

import (
	"fmt"
	"runtime"
	"strings"
)

func normalizeComputeBackend(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "cpu":
		return "cpu"
	case "cuda", "nvidia":
		return "cuda"
	default:
		return "auto"
	}
}

func openSelectedBackend(mode, cudaPath string, device, cpuThreads int) (gpuBackend, string, string, error) {
	mode = normalizeComputeBackend(mode)
	switch mode {
	case "cpu":
		b, err := openCPUBackend(cpuThreads)
		return b, "cpu", "", err
	case "cuda":
		if runtime.GOOS != "windows" && runtime.GOOS != "linux" {
			return nil, "", "", fmt.Errorf("CUDA backend is not available on %s; use --backend cpu or --backend auto", runtime.GOOS)
		}
		b, err := openCUDABackend(cudaPath, device)
		if err != nil {
			return nil, "", "", err
		}
		return b, "cuda", "", nil
	default:
		if runtime.GOOS == "windows" || runtime.GOOS == "linux" {
			if b, err := openCUDABackend(cudaPath, device); err == nil {
				return b, "cuda", "", nil
			} else {
				cpu, cpuErr := openCPUBackend(cpuThreads)
				if cpuErr != nil {
					return nil, "", "", fmt.Errorf("CUDA unavailable (%v) and CPU fallback failed: %w", err, cpuErr)
				}
				return cpu, "cpu", err.Error(), nil
			}
		}
		cpu, err := openCPUBackend(cpuThreads)
		if err != nil {
			return nil, "", "", err
		}
		return cpu, "cpu", fmt.Sprintf("CUDA is not supported on %s", runtime.GOOS), nil
	}
}
