//go:build windows

package main

import (
	"strings"
	"testing"
)

func TestUniversalMinerWebControlsPresent(t *testing.T) {
	b, err := webFS.ReadFile("web/index.html")
	if err != nil {
		t.Fatal(err)
	}
	html := string(b)
	for _, needle := range []string{
		`id="soloBackend"`,
		`value="auto"`,
		`value="opencl"`,
		`backend_opencl`,
		`id="cpuThreads"`,
		`id="poolBackend"`,
		`AuronQ Universal Miner`,
		`cpu_fallback_ready`,
		`cpu_usage`,
		`cpu_threads_live`,
		`aqm64_memory`,
		`cpu_temp_na`,
	} {
		if !strings.Contains(html, needle) {
			t.Fatalf("universal miner UI missing %q", needle)
		}
	}
	if strings.Contains(html, `id="poolThreads" type="number" min="0" max="256"`) {
		t.Fatal("unsafe legacy pool CPU thread limit is still present")
	}
}

func TestNormalizeUniversalSettings(t *testing.T) {
	s := settings{
		Device: 0, Batch: 60, AutoTuneSeconds: 1, ThermalStopC: 81,
		MiningMode: "solo", SoloBackend: "unknown", CPUThreads: 99,
		PoolBackend: "unknown", PoolThreads: 99, Language: "en",
	}
	normalizeSettings(&s)
	if s.SoloBackend != "auto" {
		t.Fatalf("SoloBackend = %q, want auto", s.SoloBackend)
	}
	if s.PoolBackend != "auto" {
		t.Fatalf("PoolBackend = %q, want auto", s.PoolBackend)
	}
	if s.CPUThreads != 0 || s.PoolThreads != 0 {
		t.Fatalf("unsafe thread settings were not normalized: cpu=%d pool=%d", s.CPUThreads, s.PoolThreads)
	}
}

func TestSafeGUIThreads(t *testing.T) {
	got := safeGUIThreads(0)
	if got < 1 || got > 2 {
		t.Fatalf("safe automatic GUI CPU threads = %d, want 1..2", got)
	}
	if got := safeGUIThreads(999); got < 1 || got > 16 {
		t.Fatalf("explicit GUI CPU cap = %d, want 1..16", got)
	}
}


func TestActiveCPUThreadsUsesPoolOverride(t *testing.T) {
	cfg := settings{CPUThreads: 1, PoolThreads: 3}
	miner := minerState{Running: true, Mode: "pool", Backend: "cpu"}
	got := activeCPUThreads(cfg, miner)
	if got < 1 || got > 3 {
		t.Fatalf("active pool CPU threads = %d, want 1..3 depending on logical CPU count", got)
	}
}

func TestCompactWorkerErrorUsesLastDiagnosticLine(t *testing.T) {
	got := compactWorkerError([]byte("line one\nCUDA batch failed: no kernel image is available\n"))
	if got != "CUDA batch failed: no kernel image is available" {
		t.Fatalf("compact error = %q", got)
	}
}


func TestNormalizeOpenCLSoloBackend(t *testing.T) {
	s := settings{
		Device: 0, Batch: 60, AutoTuneSeconds: 1, ThermalStopC: 81,
		MiningMode: "solo", SoloBackend: "opencl", CPUThreads: 0,
		PoolBackend: "auto", Language: "en",
	}
	normalizeSettings(&s)
	if s.SoloBackend != "opencl" {
		t.Fatalf("SoloBackend = %q, want opencl", s.SoloBackend)
	}
}


func TestPoolDoesNotUseNativeCUDAEligibilityGate(t *testing.T) {
	pool := settings{MiningMode: "pool", PoolBackend: "auto"}
	if shouldValidateNativeCUDA("mining", pool) {
		t.Fatal("Pool mode must not reject a GPU based on AuronQ native CUDA compatibility")
	}

	solo := settings{MiningMode: "solo", SoloBackend: "auto"}
	if !shouldValidateNativeCUDA("mining", solo) {
		t.Fatal("Solo mode must validate AuronQ native CUDA before using it")
	}
}
