package main

import "testing"

func TestSafeCPUMiningThreads(t *testing.T) {
	got, err := safeCPUMiningThreads(0)
	if err != nil {
		t.Fatal(err)
	}
	if got < 1 || got > 2 {
		t.Fatalf("safe automatic CPU mining threads = %d, want 1..2", got)
	}
	if _, err := safeCPUMiningThreads(-1); err == nil {
		t.Fatal("expected negative thread count to fail")
	}
	if _, err := safeCPUMiningThreads(17); err == nil {
		t.Fatal("expected >16 threads to fail")
	}
	got, err = safeCPUMiningThreads(1)
	if err != nil || got != 1 {
		t.Fatalf("explicit 1 thread: got %d err=%v", got, err)
	}
}
