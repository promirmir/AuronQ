package main

import (
	"net"
	"testing"
)

func TestLAN24CandidatesPrivateIPv4(t *testing.T) {
	got := lan24Candidates(net.ParseIP("192.168.1.6"), 18444)
	if len(got) != 253 {
		t.Fatalf("count=%d", len(got))
	}
	for _, p := range got {
		if p == "http://192.168.1.6:18444" {
			t.Fatal("own address included")
		}
	}
	if got[0] != "http://192.168.1.1:18444" {
		t.Fatalf("first=%q", got[0])
	}
}

func TestLAN24CandidatesRejectsPublicIPv4(t *testing.T) {
	if got := lan24Candidates(net.ParseIP("8.8.8.8"), 18444); len(got) != 0 {
		t.Fatalf("public IP produced %d candidates", len(got))
	}
}
