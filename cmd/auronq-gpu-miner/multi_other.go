//go:build !windows

package main

import "errors"

type multiGPUOptions struct {
	Node             string
	Address          string
	Batch            int
	DLLPath          string
	SelfTest         bool
	Benchmark        bool
	BenchmarkSeconds int
	AutoTune         bool
	AutoTuneSeconds  int
	ThermalAuto      bool
	ThermalLimit     int
	ThermalTarget    int
	NoncePrefix      uint64
}

func resolveCUDADevices(string) ([]int, error) {
	return nil, errors.New("multi-GPU CUDA orchestration currently targets Windows x64")
}

func runMultiGPU([]int, multiGPUOptions) error {
	return errors.New("multi-GPU CUDA orchestration currently targets Windows x64")
}
