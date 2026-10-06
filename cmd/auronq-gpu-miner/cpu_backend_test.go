package main

import (
	"testing"

	aq "auronq/internal/auronq"
)

func TestCPUBackendMatchesCanonicalAQM64(t *testing.T) {
	backend, err := openCPUBackend(1)
	if err != nil {
		t.Fatal(err)
	}
	defer backend.Close()

	header := aq.BlockHeader{
		Version:   aq.BlockVersion,
		PowAlgo:   aq.PowAlgorithmAQM64,
		Height:    4242,
		Timestamp: 1790951480,
		Target:    aq.PowLimit,
		Nonce:     0x1234567890abcdef,
	}
	for i := range header.PrevHash {
		header.PrevHash[i] = byte(i*7 + 1)
	}
	for i := range header.MerkleRoot {
		header.MerkleRoot[i] = byte(i*11 + 3)
	}

	prepared, err := prepareCandidate(header)
	if err != nil {
		t.Fatal(err)
	}
	finals, err := backend.Run(prepared.initial, 1)
	if err != nil {
		t.Fatal(err)
	}
	got, err := finishCandidate(prepared.pre, finals[:128])
	if err != nil {
		t.Fatal(err)
	}
	want, err := aq.PowHash(header)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("CPU fallback hash mismatch\ngot  %s\nwant %s", got.String(), want.String())
	}
}

func TestSafeCPUThreads(t *testing.T) {
	got := safeCPUThreads(0)
	if got < 1 || got > 4 {
		t.Fatalf("automatic safe CPU threads = %d, want 1..4", got)
	}
	if got := safeCPUThreads(999); got < 1 || got > 16 {
		t.Fatalf("explicit CPU thread cap = %d, want 1..16", got)
	}
}
