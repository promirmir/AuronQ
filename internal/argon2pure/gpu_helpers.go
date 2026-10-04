package argon2pure

import (
	"encoding/binary"
	"errors"
)

// PrepareSingleLaneBlocks computes the two Argon2 v1.3 initial 1 KiB blocks
// for a single-lane Argon2id instance. It exists so an external accelerator
// can process the memory-hard block graph while AuronQ keeps the exact
// initialization/finalization logic in one reviewed implementation.
//
// The returned slice contains block 0 followed by block 1, 128 uint64 words
// each, using the same little-endian word representation as Workspace.
func PrepareSingleLaneBlocks(password, salt []byte, timeCost, memory, keyLen uint32) ([]uint64, error) {
	if timeCost < 1 {
		return nil, errors.New("argon2: number of rounds too small")
	}
	memory = normalizedMemory(memory, 1)
	if memory < 8 {
		return nil, errors.New("argon2: memory too small")
	}
	h0 := initHash(password, salt, timeCost, memory, 1, keyLen)
	first := make([]block, 2)
	initBlocksInto(first, &h0, memory, 1)
	out := make([]uint64, 2*blockLength)
	copy(out[:blockLength], first[0][:])
	copy(out[blockLength:], first[1][:])
	return out, nil
}

// ExtractSingleLaneKey applies the standard Argon2 variable-length final hash
// to the final 1 KiB block returned by a single-lane accelerator.
func ExtractSingleLaneKey(finalBlock []uint64, keyLen uint32) ([]byte, error) {
	if len(finalBlock) != blockLength {
		return nil, errors.New("argon2: final block must contain 128 uint64 words")
	}
	if keyLen == 0 {
		return nil, errors.New("argon2: key length must be positive")
	}
	raw := make([]byte, 1024)
	for i, v := range finalBlock {
		binary.LittleEndian.PutUint64(raw[i*8:], v)
	}
	return blake2bLong(keyLen, raw), nil
}
