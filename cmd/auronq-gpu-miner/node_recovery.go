package main

import "time"

// nodeRecoveryDelay caps reconnect backoff without returning to mining on
// stale node state. Thermal and backend errors are handled separately.
func nodeRecoveryDelay(attempt int) time.Duration {
    if attempt < 0 { attempt = 0 }
    if attempt >= 12 { return 60 * time.Second }
    d := time.Duration(attempt+1) * 5 * time.Second
    if d > 60*time.Second { return 60*time.Second }
    return d
}
