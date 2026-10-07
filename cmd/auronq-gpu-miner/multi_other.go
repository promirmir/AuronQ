//go:build !windows && !linux

package main

import "errors"

type multiGPUOptions struct {
	Backend          string
	Node             string
	Address          string
	Batch            int
	DLLPath          string
	LegacyDLLPath    string
	KeplerDLLPath    string
	OpenCLPath       string
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
	return nil, errors.New("multi-GPU accelerator orchestration currently targets Windows/Linux")
}

func runMultiGPU([]int, multiGPUOptions) error {
	return errors.New("multi-GPU accelerator orchestration currently targets Windows/Linux")
}
