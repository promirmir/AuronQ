package main

import (
    "testing"
    "time"
)

func TestNodeRecoveryDelayBounded(t *testing.T) {
    cases := []struct{ attempt int; want time.Duration }{
        {-1, 5*time.Second}, {0, 5*time.Second},
        {1, 10*time.Second}, {11, 60*time.Second},
        {12, 60*time.Second}, {100000, 60*time.Second},
    }
    for _, tc := range cases {
        if got := nodeRecoveryDelay(tc.attempt); got != tc.want {
            t.Errorf("attempt %d: got %v want %v", tc.attempt, got, tc.want)
        }
    }
}
