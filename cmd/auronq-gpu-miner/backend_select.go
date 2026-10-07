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
	case "opencl", "ocl", "amd", "intel":
		return "opencl"
	default:
		return "auto"
	}
}

func validateAcceleratorBackend(kind string, b gpuBackend) error {
	if b == nil {
		return fmt.Errorf("%s backend is nil", kind)
	}
	if err := runSelfTest(b); err != nil {
		return fmt.Errorf("%s backend failed canonical AQM64 validation: %w", kind, err)
	}
	return nil
}

func openSelectedBackend(mode, cudaPath, openclPath string, device, cpuThreads int) (gpuBackend, string, string, error) {
	mode = normalizeComputeBackend(mode)

	openValidatedCUDA := func() (gpuBackend, error) {
		if runtime.GOOS != "windows" && runtime.GOOS != "linux" {
			return nil, fmt.Errorf("CUDA backend is not available on %s", runtime.GOOS)
		}
		b, err := openCUDABackend(cudaPath, device)
		if err != nil {
			return nil, err
		}
		if err := validateAcceleratorBackend("CUDA", b); err != nil {
			_ = b.Close()
			return nil, err
		}
		return b, nil
	}

	openValidatedOpenCL := func() (gpuBackend, error) {
		if runtime.GOOS != "windows" && runtime.GOOS != "linux" {
			return nil, fmt.Errorf("OpenCL GPU backend is not available in this %s package", runtime.GOOS)
		}
		b, err := openOpenCLBackend(openclPath, device)
		if err != nil {
			return nil, err
		}
		if err := validateAcceleratorBackend("OpenCL", b); err != nil {
			_ = b.Close()
			return nil, err
		}
		return b, nil
	}

	switch mode {
	case "cpu":
		b, err := openCPUBackend(cpuThreads)
		return b, "cpu", "", err
	case "cuda":
		b, err := openValidatedCUDA()
		if err != nil {
			return nil, "", "", err
		}
		return b, "cuda", "", nil
	case "opencl":
		b, err := openValidatedOpenCL()
		if err != nil {
			return nil, "", "", err
		}
		return b, "opencl", "", nil
	default:
		var reasons []string

		if b, err := openValidatedCUDA(); err == nil {
			return b, "cuda", "", nil
		} else {
			reasons = append(reasons, "CUDA: "+err.Error())
		}

		if b, err := openValidatedOpenCL(); err == nil {
			return b, "opencl", strings.Join(reasons, " | "), nil
		} else {
			reasons = append(reasons, "OpenCL: "+err.Error())
		}

		cpu, cpuErr := openCPUBackend(cpuThreads)
		if cpuErr != nil {
			return nil, "", "", fmt.Errorf("accelerator backends unavailable (%s) and CPU fallback failed: %w",
				strings.Join(reasons, " | "), cpuErr)
		}
		return cpu, "cpu", strings.Join(reasons, " | "), nil
	}
}
