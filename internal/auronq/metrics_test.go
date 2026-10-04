package auronq

import (
	"math"
	"math/big"
	"testing"
)

func TestEstimatedNetworkHashrateUsesRecentBlockWork(t *testing.T) {
	work := WorkForTarget(PowLimit)
	blocks := make([]Block, 4)
	for i := range blocks {
		blocks[i].Header.Height = uint64(i)
		blocks[i].Header.Timestamp = int64(i) * TargetBlockSeconds
		blocks[i].Header.Target = PowLimit
	}
	c := &Chain{blocks: blocks}
	got := c.EstimatedNetworkHashrate(3)
	want, _ := new(big.Float).Quo(
		new(big.Float).SetInt(work),
		big.NewFloat(float64(TargetBlockSeconds)),
	).Float64()
	if math.Abs(got-want) > math.Max(1e-12, want*1e-9) {
		t.Fatalf("hashrate=%g want %g", got, want)
	}
}

func TestEstimatedNetworkHashrateNeedsTimeSpan(t *testing.T) {
	c := &Chain{blocks: []Block{{Header: BlockHeader{Timestamp: 100, Target: PowLimit}}}}
	if got := c.EstimatedNetworkHashrate(60); got != 0 {
		t.Fatalf("hashrate=%g want 0", got)
	}
}
