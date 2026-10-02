package auronq

import (
	"math/big"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Consensus tests exercise validation/reorg logic, not the production
	// memory cost. Keep them fast while preserving exactly the same AQM64
	// code path and algorithm structure.
	aqm64MemoryKiB = 32
	aqm64TimeCost = 1
	PowLimit = TargetFromBig(new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 512), big.NewInt(1)))
	os.Exit(m.Run())
}
