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

func openSelectedBackend(mode, cudaPath, legacyCUDAPath, openclPath string, device, cpuThreads int) (gpuBackend, string, string, error) {
	mode = normalizeComputeBackend(mode)

	openValidatedCUDA := func(path, label string) (gpuBackend, error) {
		if runtime.GOOS != "windows" && runtime.GOOS != "linux" {
			return nil, fmt.Errorf("%s is not available on %s", label, runtime.GOOS)
		}
		b, err := openCUDABackend(path, device)
		if err != nil {
			return nil, err
		}
		if err := validateAcceleratorBackend(label, b); err != nil {
			_ = b.Close()
			return nil, err
		}
		return b, nil
	}

	openAnyCUDA := func() (gpuBackend, string, error) {
		var reasons []string
		if b, err := openValidatedCUDA(cudaPath, "CUDA"); err == nil {
			return b, "", nil
		} else {
			reasons = append(reasons, "CUDA: "+err.Error())
		}

		if strings.TrimSpace(legacyCUDAPath) != "" && legacyCUDAPath != cudaPath {
			if b, err := openValidatedCUDA(legacyCUDAPath, "Legacy CUDA"); err == nil {
				return b, strings.Join(reasons, " | ") + " | selected Legacy CUDA", nil
			} else {
				reasons = append(reasons, "Legacy CUDA: "+err.Error())
			}
		}
		return nil, strings.Join(reasons, " | "), fmt.Errorf("%s", strings.Join(reasons, " | "))
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
		b, reason, err := openAnyCUDA()
		if err != nil {
			return nil, "", "", err
		}
		return b, "cuda", reason, nil
	case "opencl":
		b, err := openValidatedOpenCL()
		if err != nil {
			return nil, "", "", err
		}
		return b, "opencl", "", nil
	default:
		var reasons []string

		if b, reason, err := openAnyCUDA(); err == nil {
			return b, "cuda", reason, nil
		} else if reason != "" {
			reasons = append(reasons, reason)
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
