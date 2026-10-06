//go:build windows

package main

import "testing"

func TestResolveCUDADevicesExplicit(t *testing.T) {
	got, err := resolveCUDADevices("2,0,2,1")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 || got[0] != 0 || got[1] != 1 || got[2] != 2 {
		t.Fatalf("unexpected devices: %v", got)
	}
	if _, err := resolveCUDADevices("-1"); err == nil {
		t.Fatal("expected invalid device error")
	}
}
