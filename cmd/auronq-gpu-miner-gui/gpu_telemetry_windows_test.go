//go:build windows

package main

import "testing"

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
