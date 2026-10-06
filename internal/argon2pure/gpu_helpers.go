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


// SingleLaneWorkspace reuses the 64 MiB Argon2 memory region for prepared
// AQM64 CPU work. It lets the universal miner provide a CPU fallback without
// allocating a fresh scratchpad for every nonce.
type SingleLaneWorkspace struct {
	blocks []block
	memory uint32
}

// NewSingleLaneWorkspace creates a reusable single-lane workspace. memory is
// expressed in KiB and is normalized exactly like the canonical Argon2 code.
func NewSingleLaneWorkspace(memory uint32) *SingleLaneWorkspace {
	m := normalizedMemory(memory, 1)
	return &SingleLaneWorkspace{
		blocks: make([]block, m),
		memory: m,
	}
}

// ProcessPreparedSingleLane continues Argon2id from the two initial 1 KiB
// blocks produced by PrepareSingleLaneBlocks and returns the final 1 KiB
// block. The workspace is not safe for concurrent use.
func (w *SingleLaneWorkspace) ProcessPreparedSingleLane(initial []uint64, timeCost uint32) ([]uint64, error) {
	if w == nil || len(w.blocks) == 0 || w.memory < 8 {
		return nil, errors.New("argon2: invalid single-lane workspace")
	}
	if timeCost < 1 {
		return nil, errors.New("argon2: number of rounds too small")
	}
	if len(initial) != 2*blockLength {
		return nil, errors.New("argon2: prepared input must contain exactly two blocks")
	}

	// processBlocks overwrites the rest of the memory graph deterministically,
	// so only the two initial blocks need to be replaced for each nonce.
	copy(w.blocks[0][:], initial[:blockLength])
	copy(w.blocks[1][:], initial[blockLength:2*blockLength])
	processBlocks(w.blocks, timeCost, w.memory, 1)

	out := make([]uint64, blockLength)
	copy(out, w.blocks[w.memory-1][:])
	return out, nil
}
