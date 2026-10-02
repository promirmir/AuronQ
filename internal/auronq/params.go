package auronq

import "math/big"

const (
	Coin                    = uint64(100_000_000)
	MaxSupplyCoins          = uint64(21_000_000)
	MaxSupplyAtoms          = MaxSupplyCoins * Coin
	FounderCoins            = uint64(210_000)
	FounderAtoms            = FounderCoins * Coin
	InitialSubsidyAtoms     = uint64(49)*Coin + Coin/2 // 49.5 AURQ; 99% of Bitcoin-style 50 because 1% is allocated in genesis
	HalvingInterval         = uint64(210_000)
	MainnetCoinbaseMaturity = uint64(100)
	TestnetCoinbaseMaturity = uint64(10)
	TargetBlockSeconds      = int64(600)
	DifficultyWindow        = 60
	MedianTimeWindow        = 11
	MaxFutureSeconds        = int64(2 * 60 * 60)
	MaxBlockBytes           = 4 * 1024 * 1024
	MaxTxBytes              = 128 * 1024
	MaxPubKeyBytes          = 4096
	MaxSignatureBytes       = 8192
	MaxMempoolTx            = 10_000
	MaxMempoolBytes         = 64 * 1024 * 1024
	MinRelayFeePerKB        = uint64(1_000)
	DefaultP2PPort          = 18444
	MainnetNetworkByte      = byte(0x51)
	TestnetNetworkByte      = byte(0x71)

	// AuronQ-PoW v1 (AQM64): Argon2id with 64 MiB per mining lane,
	// followed by SHAKE256. The primitives are standard; the composition is
	// AuronQ-specific. Mainnet RC1 freezes these parameters for launch compatibility; independent review is still strongly recommended.
	PowAlgorithmAQM64 uint16 = 1
	AQM64MemoryKiB    uint32 = 64 * 1024
	AQM64TimeCost     uint32 = 2
	AQM64Parallelism  uint8  = 1
)

var PowLimit = func() Target {
	// AQM64 is intentionally expensive per attempt. Mainnet RC1 freezes a 10-bit
	// bootstrap PoW limit so a new network can begin on ordinary CPUs while LWMA
	// converges toward observed network hashrate. This is a launch calibration,
	// not a claim of optimal economic/security calibration.
	x := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 502), big.NewInt(1))
	return TargetFromBig(x)
}()

func BlockSubsidy(height uint64, issued uint64) uint64 {
	if height == 0 || issued >= MaxSupplyAtoms {
		return 0
	}
	halvings := (height - 1) / HalvingInterval
	if halvings >= 64 {
		return 0
	}
	subsidy := InitialSubsidyAtoms >> halvings
	rem := MaxSupplyAtoms - issued
	if subsidy > rem {
		subsidy = rem
	}
	return subsidy
}

// Internal test hooks. Production code never changes these values. Tests use
// a smaller scratchpad so consensus validation suites remain fast.
var (
	aqm64MemoryKiB = AQM64MemoryKiB
	aqm64TimeCost  = AQM64TimeCost
)
