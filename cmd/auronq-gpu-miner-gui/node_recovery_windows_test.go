//go:build windows

package main

import "testing"

func TestNodeRecoveryClearsStaleHashrate(t *testing.T) {
    a := &App{}
    a.miner.Hashrate = 250
    a.parseWorkerLine("NODE_STATUS_BACKOFF wait=5s attempt=1")
    if a.miner.Hashrate != 0 {
        t.Fatalf("expected 0 H/s while node is unavailable, got %v", a.miner.Hashrate)
    }
    if a.miner.LastError == "" {
        t.Fatal("missing recovery reason in GUI state")
    }
    a.parseWorkerLine("Mining height 1040 target=abcdef nonce_base=0")
    a.parseWorkerLine("hashes=300 rate=180.00 H/s avg=200.00 H/s current_height=1040 batch=25")
    if a.miner.Hashrate != 180 {
        t.Fatalf("expected fresh recovered rate 180 H/s, got %v", a.miner.Hashrate)
    }
    if a.miner.LastError != "" {
        t.Fatalf("stale recovery warning: %q", a.miner.LastError)
    }
}

func TestNodeStartupPause(t *testing.T) {
    a := &App{}
    a.miner.Hashrate = 100
    a.parseWorkerLine("NODE_STARTUP_RETRY wait=5s error=connection refused")
    if a.miner.Hashrate != 0 {
        t.Fatalf("expected zero hashrate while node starts, got %v", a.miner.Hashrate)
    }
}
