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


func TestListenPort(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"0.0.0.0:18444", 18444},
		{"[::]:18444", 18444},
		{"127.0.0.1:12345", 12345},
	}
	for _, tc := range cases {
		got, err := listenPort(tc.in)
		if err != nil {
			t.Fatalf("listenPort(%q): %v", tc.in, err)
		}
		if got != tc.want {
			t.Fatalf("listenPort(%q) = %d, want %d", tc.in, got, tc.want)
		}
	}
	for _, bad := range []string{"18444", "0.0.0.0:0", "0.0.0.0:70000"} {
		if _, err := listenPort(bad); err == nil {
			t.Fatalf("listenPort(%q) unexpectedly succeeded", bad)
		}
	}
}
