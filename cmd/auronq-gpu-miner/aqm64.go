package main

import (
	"bytes"
	"encoding/binary"
	"fmt"

	aq "auronq/internal/auronq"
	"auronq/internal/argon2pure"
	sha3 "auronq/internal/sha3compat"
)

const (
	preDomain   = "AURONQ_AQM64_PRE_V1\x00"
	saltDomain  = "AURONQ_AQM64_SALT_V1\x00"
	finalDomain = "AURONQ_AQM64_FINAL_V1\x00"
)

type preparedCandidate struct {
	nonce   uint64
	pre     [64]byte
	initial []uint64
}

func prepareCandidate(header aq.BlockHeader) (preparedCandidate, error) {
	var out preparedCandidate
	out.nonce = header.Nonce

	hb := header.Bytes()
	preInput := make([]byte, 0, len(preDomain)+len(hb))
	preInput = append(preInput, preDomain...)
	preInput = append(preInput, hb...)
	sha3.ShakeSum256(out.pre[:], preInput)

	saltInput := make([]byte, 0, len(saltDomain)+len(header.PrevHash)+10)
	saltInput = append(saltInput, saltDomain...)
	saltInput = append(saltInput, header.PrevHash[:]...)
	var x [10]byte
	binary.BigEndian.PutUint64(x[:8], header.Height)
	binary.BigEndian.PutUint16(x[8:], header.PowAlgo)
	saltInput = append(saltInput, x[:]...)
	var salt [32]byte
	sha3.ShakeSum256(salt[:], saltInput)

	initial, err := argon2pure.PrepareSingleLaneBlocks(
		out.pre[:],
		salt[:],
		aq.AQM64TimeCost,
		aq.AQM64MemoryKiB,
		64,
	)
	if err != nil {
		return preparedCandidate{}, err
	}
	out.initial = initial
	return out, nil
}

func finishCandidate(pre [64]byte, finalBlock []uint64) (aq.Hash, error) {
	mid, err := argon2pure.ExtractSingleLaneKey(finalBlock, 64)
	if err != nil {
		return aq.Hash{}, err
	}
	in := make([]byte, 0, len(finalDomain)+64+len(mid))
	in = append(in, finalDomain...)
	in = append(in, pre[:]...)
	in = append(in, mid...)
	var out aq.Hash
	sha3.ShakeSum256(out[:], in)
	return out, nil
}

func hashMeetsTarget(hash aq.Hash, target aq.Target) bool {
	return bytes.Compare(hash[:], target[:]) <= 0
}

func buildBatch(template aq.Block, startNonce uint64, count int) ([]preparedCandidate, []uint64, error) {
	if count < 1 {
		return nil, nil, fmt.Errorf("invalid batch size %d", count)
	}
	prepared := make([]preparedCandidate, count)
	initial := make([]uint64, count*256)
	for i := 0; i < count; i++ {
		h := template.Header
		h.Nonce = startNonce + uint64(i)
		p, err := prepareCandidate(h)
		if err != nil {
			return nil, nil, err
		}
		prepared[i] = p
		copy(initial[i*256:(i+1)*256], p.initial)
	}
	return prepared, initial, nil
}
