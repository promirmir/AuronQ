package auronq

import (
	"encoding/binary"
	"errors"
	"math/big"
	"sync"

	"auronq/internal/argon2pure"
	sha3 "auronq/internal/sha3compat"
)

const (
	aqm64PreDomain   = "AURONQ_AQM64_PRE_V1\x00"
	aqm64SaltDomain  = "AURONQ_AQM64_SALT_V1\x00"
	aqm64FinalDomain = "AURONQ_AQM64_FINAL_V1\x00"
)

type aqm64Hasher struct {
	memory uint32
	time   uint32
	ws     *argon2pure.Workspace
}

func newAQM64Hasher() *aqm64Hasher {
	return &aqm64Hasher{
		memory: aqm64MemoryKiB,
		time:   aqm64TimeCost,
		ws:     argon2pure.NewWorkspace(aqm64MemoryKiB, AQM64Parallelism),
	}
}

func (a *aqm64Hasher) hash(header BlockHeader) Hash {
	// AQM64 deliberately composes standardized primitives instead of inventing
	// a new cryptographic primitive. SHAKE256 domain-separates the header,
	// Argon2id v1.3 imposes the 64 MiB memory cost, and SHAKE256 produces the
	// 512-bit proof value used by AuronQ's existing target/work arithmetic.
	hb := header.Bytes()

	preInput := make([]byte, 0, len(aqm64PreDomain)+len(hb))
	preInput = append(preInput, aqm64PreDomain...)
	preInput = append(preInput, hb...)
	pre := make([]byte, 64)
	sha3.ShakeSum256(pre, preInput)

	saltInput := make([]byte, 0, len(aqm64SaltDomain)+len(header.PrevHash)+8+2)
	saltInput = append(saltInput, aqm64SaltDomain...)
	saltInput = append(saltInput, header.PrevHash[:]...)
	var x [10]byte
	binary.BigEndian.PutUint64(x[:8], header.Height)
	binary.BigEndian.PutUint16(x[8:], header.PowAlgo)
	saltInput = append(saltInput, x[:]...)
	salt := make([]byte, 32)
	sha3.ShakeSum256(salt, saltInput)

	mid := a.ws.IDKey(pre, salt, a.time, 64)

	finalInput := make([]byte, 0, len(aqm64FinalDomain)+len(pre)+len(mid))
	finalInput = append(finalInput, aqm64FinalDomain...)
	finalInput = append(finalInput, pre...)
	finalInput = append(finalInput, mid...)
	var out Hash
	sha3.ShakeSum256(out[:], finalInput)

	for i := range pre {
		pre[i] = 0
	}
	for i := range mid {
		mid[i] = 0
	}
	for i := range salt {
		salt[i] = 0
	}
	return out
}

var aqm64Pool sync.Pool

func acquireAQM64Hasher() *aqm64Hasher {
	if v := aqm64Pool.Get(); v != nil {
		h := v.(*aqm64Hasher)
		if h.memory == aqm64MemoryKiB && h.time == aqm64TimeCost {
			return h
		}
	}
	return newAQM64Hasher()
}

func releaseAQM64Hasher(h *aqm64Hasher) {
	if h != nil && h.memory == aqm64MemoryKiB && h.time == aqm64TimeCost {
		aqm64Pool.Put(h)
	}
}

func PowHash(header BlockHeader) (Hash, error) {
	if header.PowAlgo != PowAlgorithmAQM64 {
		return Hash{}, errors.New("unsupported proof-of-work algorithm")
	}
	h := acquireAQM64Hasher()
	defer releaseAQM64Hasher(h)
	return h.hash(header), nil
}

func ValidProofOfWork(header BlockHeader) bool {
	h, err := PowHash(header)
	if err != nil {
		return false
	}
	return h.Big().Cmp(header.Target.Big()) <= 0
}

func (h Hash) Big() *big.Int {
	return new(big.Int).SetBytes(h[:])
}