//go:build windows

package main

import "testing"

func TestFiletimeValue(t *testing.T) {
	v := winFiletime{LowDateTime: 0x89abcdef, HighDateTime: 0x01234567}
	if got, want := filetimeValue(v), uint64(0x0123456789abcdef); got != want {
		t.Fatalf("filetimeValue=%x want=%x", got, want)
	}
}

func TestCPUSamplerMemoryAccounting(t *testing.T) {
	s := &cpuSampler{name: "Test CPU", mhz: 3200}
	got := s.Sample(8, 2)
	if got.Name != "Test CPU" || got.NominalClockMHz != 3200 {
		t.Fatalf("identity mismatch: %+v", got)
	}
	if got.LogicalProcessors != 8 || got.MiningThreads != 2 {
		t.Fatalf("thread telemetry mismatch: %+v", got)
	}
	if got.AQM64MemoryMiB != 128 {
		t.Fatalf("AQM64 memory=%d want=128", got.AQM64MemoryMiB)
	}
	if got.TemperatureC != -1 {
		t.Fatalf("CPU temperature must remain unavailable without a trusted sensor, got=%v", got.TemperatureC)
	}
}
