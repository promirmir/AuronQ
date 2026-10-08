package main

import (
    "fmt"
    "time"
)

// adaptiveMiningAgent performs bounded, local, deterministic batch tuning.
// The independent thermal controller remains authoritative: this adviser
// never raises its temperature target, hard limit or bypasses sensor checks.
type adaptiveMiningAgent struct {
    lastAdjustment time.Time
    lastEvaluation time.Time
    referenceRate float64
    windowHashes uint64
    windowTime time.Duration
    observations int
    failedProbes int
    previousBatch int
    probing bool
}

func newAdaptiveMiningAgent() *adaptiveMiningAgent { return &adaptiveMiningAgent{} }

// Observe returns a *suggested* batch. The caller MUST run the thermal
// controller before the next compute invocation. Temperature headroom and
// no duty-cycle pause are prerequisites for upward exploration.
func (a *adaptiveMiningAgent) Observe(now time.Time, batch int, hashes int, elapsed time.Duration, temperature, target, maxBatch int, thermalPause time.Duration) (int, string) {
    if a == nil || elapsed <= 0 || hashes <= 0 || batch < 1 {
        return batch, ""
    }
    a.windowHashes += uint64(hashes)
    a.windowTime += elapsed
    a.observations++
    if a.lastEvaluation.IsZero() { a.lastEvaluation = now; return batch, "" }
    if now.Sub(a.lastEvaluation) < 20*time.Second || a.windowTime < 5*time.Second {
        return batch, ""
    }
    rate := float64(a.windowHashes) / a.windowTime.Seconds()
    a.windowHashes, a.windowTime, a.observations = 0, 0, 0
    a.lastEvaluation = now
    if temperature < 0 || temperature > target-4 || thermalPause > 0 {
        a.probing = false
        a.referenceRate = rate
        return batch, ""
    }
    if a.probing {
        a.probing = false
        if a.referenceRate > 0 && rate < a.referenceRate*0.97 {
            a.failedProbes++
            if a.previousBatch >= 1 && a.previousBatch < batch {
                a.lastAdjustment = now
                return a.previousBatch, fmt.Sprintf("AI_AGENT rollback batch=%d->%d H/s=%.2f baseline=%.2f", batch, a.previousBatch, rate, a.referenceRate)
            }
        } else {
            a.failedProbes = 0
        }
        a.referenceRate = rate
        return batch, ""
    }
    a.referenceRate = rate
    if maxBatch < 1 { maxBatch = 1 }
    if batch >= maxBatch || a.failedProbes >= 3 || (!a.lastAdjustment.IsZero() && now.Sub(a.lastAdjustment) < 90*time.Second) {
        return batch, ""
    }
    a.previousBatch = batch
    a.lastAdjustment = now
    a.probing = true
    return batch+1, fmt.Sprintf("AI_AGENT probe batch=%d->%d baseline=%.2f H/s", batch, batch+1, rate)
}
