//go:build windows

package gputelemetry

import "testing"

func TestParseGPUUUID(t *testing.T) {
	got, err := parseGPUUUID("GPU-00112233-4455-6677-8899-aabbccddeeff")
	if err != nil {
		t.Fatal(err)
	}
	want := [16]byte{0x00,0x11,0x22,0x33,0x44,0x55,0x66,0x77,0x88,0x99,0xaa,0xbb,0xcc,0xdd,0xee,0xff}
	if got != want {
		t.Fatalf("got %x want %x", got, want)
	}
	if _, err := parseGPUUUID("bad"); err == nil {
		t.Fatal("expected invalid UUID error")
	}
}
