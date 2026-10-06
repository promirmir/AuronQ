//go:build windows

package main

import (
	"strings"
	"testing"
	"time"
)

func TestSelectedDeviceIndices(t *testing.T) {
	gpus := []gpuInfo{{Index: 0, Name: "A"}, {Index: 2, Name: "B"}}

	all, err := selectedDeviceIndices(settings{UseAllDevices: true}, gpus)
	if err != nil || len(all) != 2 || all[0] != 0 || all[1] != 2 {
		t.Fatalf("all devices: %v %v", all, err)
	}

	picked, err := selectedDeviceIndices(settings{Devices: []int{2, 2, 0}}, gpus)
	if err != nil || len(picked) != 2 || picked[0] != 0 || picked[1] != 2 {
		t.Fatalf("picked devices: %v %v", picked, err)
	}

	if _, err := selectedDeviceIndices(settings{Devices: []int{1}}, gpus); err == nil {
		t.Fatal("expected unavailable device error")
	}
}

func TestNormalizePoolEndpoint(t *testing.T) {
	got, err := normalizePoolEndpoint("stratum+tcp://pool.example:3359/")
	if err != nil || got != "pool.example:3359" {
		t.Fatalf("got %q err=%v", got, err)
	}
	if _, err := normalizePoolEndpoint("bad endpoint"); err == nil {
		t.Fatal("expected invalid endpoint")
	}
}

func TestParseSMIValues(t *testing.T) {
	if got := parseSMIInt(" 77 "); got != 77 {
		t.Fatalf("int=%d", got)
	}
	if got := parseSMIInt("N/A"); got != -1 {
		t.Fatalf("N/A int=%d", got)
	}
	if got := parseSMIFloat("123.45"); got < 123.44 || got > 123.46 {
		t.Fatalf("float=%f", got)
	}
}


func TestNormalizePoolEndpointCaseInsensitiveScheme(t *testing.T) {
	got, err := normalizePoolEndpoint("STRATUM+TCP://pool.example:4444/")
	if err != nil || got != "pool.example:4444" {
		t.Fatalf("got %q err=%v", got, err)
	}
}

func TestParseHashrateText(t *testing.T) {
	cases := []struct {
		line string
		want float64
	}{
		{"speed 42.5 H/s", 42.5},
		{"total 1.25 kH/s", 1250},
		{"GPU 0: 2.5 MH/s accepted", 2500000},
	}
	for _, tc := range cases {
		got, ok := parseHashrateText(tc.line)
		if !ok || got != tc.want {
			t.Fatalf("%q => %f %v, want %f", tc.line, got, ok, tc.want)
		}
	}
}


func TestMeshMinerArgsCUDA(t *testing.T) {
	s := settings{
		PoolBackend: "cuda",
		PoolFanAuto: true,
		PoolRetune: true,
	}
	args, err := meshMinerArgs(s, []int{0, 2}, "pool.meshpool.net:3359", "aurq1test.rig")
	if err != nil {
		t.Fatal(err)
	}
	got := strings.Join(args, " ")
	for _, want := range []string{
		"--pool pool.meshpool.net:3359",
		"--user aurq1test.rig",
		"--backend cuda",
		"--algo auronq",
		"--device 0,2",
		"--fan auto",
		"--retune",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("args %q missing %q", got, want)
		}
	}
}

func TestMeshMinerArgsCPU(t *testing.T) {
	s := settings{
		PoolBackend: "cpu",
		PoolThreads: 6,
		PoolFanAuto: true,
	}
	args, err := meshMinerArgs(s, nil, "example:1234", "aurq1test.cpu")
	if err != nil {
		t.Fatal(err)
	}
	got := strings.Join(args, " ")
	if !strings.Contains(got, "--backend cpu") || !strings.Contains(got, "--threads 6") {
		t.Fatalf("unexpected args: %q", got)
	}
	if strings.Contains(got, "--device") || strings.Contains(got, "--fan") {
		t.Fatalf("CPU args should not include GPU flags: %q", got)
	}
}

func TestMeshMinerArgsCUDARequiresGPU(t *testing.T) {
	_, err := meshMinerArgs(settings{PoolBackend: "cuda"}, nil, "example:1234", "aurq1test.rig")
	if err == nil {
		t.Fatal("expected CUDA backend to require a GPU")
	}
}


func TestExternalThermalPausePolicy(t *testing.T) {
	target, limit := 76, 81
	cases := []struct {
		temp int
		want time.Duration
	}{
		{70, 0},
		{75, 0},
		{76, 100 * time.Millisecond},
		{77, 220 * time.Millisecond},
		{78, 400 * time.Millisecond},
		{79, 600 * time.Millisecond},
		{80, 850 * time.Millisecond},
		{81, 0},
	}
	for _, tc := range cases {
		if got := externalThermalPause(tc.temp, target, limit); got != tc.want {
			t.Fatalf("temp %d: got %s want %s", tc.temp, got, tc.want)
		}
	}
}

func TestParseMeshMinerExactHashrateLine(t *testing.T) {
	line := "[17:26:52] Total  :   426.20 H/s [    12|    0|    0]"
	got, ok := parseHashrateText(line)
	if !ok || got != 426.20 {
		t.Fatalf("got %f ok=%v", got, ok)
	}
}
