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
			var cudaErr error
			if b, err := openCUDABackend(cudaPath, device); err == nil {
				if testErr := runSelfTest(b); testErr == nil {
					return b, "cuda", "", nil
				} else {
					_ = b.Close()
					cudaErr = fmt.Errorf("CUDA backend failed canonical AQM64 validation: %w", testErr)
				}
			} else {
				cudaErr = err
			}
			cpu, cpuErr := openCPUBackend(cpuThreads)
			if cpuErr != nil {
				return nil, "", "", fmt.Errorf("CUDA unavailable (%v) and CPU fallback failed: %w", cudaErr, cpuErr)
			}
			return cpu, "cpu", cudaErr.Error(), nil
		}
		cpu, err := openCPUBackend(cpuThreads)
		if err != nil {
			return nil, "", "", err
		}
		return cpu, "cpu", fmt.Sprintf("CUDA is not supported on %s", runtime.GOOS), nil
	}
}
