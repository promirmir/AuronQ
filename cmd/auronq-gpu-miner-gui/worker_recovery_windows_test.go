//go:build windows

package main

import "testing"

func TestWorkerRecoveryOnlyPermitsConnectionFailures(t *testing.T) {
    cases := []struct { line string; recover bool }{
        {"node status: connection refused", true},
        {"miner stopped: node status unavailable after throttled retries", true},
        {"miner stopped: template: temporarily unavailable", true},
        {"miner stopped: temperature sensor missing", false},
        {"miner stopped: thermal hard stop 80C", false},
        {"miner stopped: GPU backend failed", false},
        {"miner stopped: refusing to mine: node Network ID changed during reconnect", false},
        {"SELF-TEST FAILED: AQM64 mismatch", false},
        {"Worker stopped: exit status 1", false},
        {"", false},
    }
    for _, tc := range cases {
        if got := recoverableSoloExit(tc.line); got != tc.recover {
            t.Errorf("%q: got %t want %t", tc.line, got, tc.recover)
        }
    }
}
