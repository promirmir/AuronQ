package argon2pure

import (
	"encoding/binary"
	"math/bits"
)

const Version = 0x13
const blockLength = 128
const syncPoints = 4
const argon2id = 2

type block [blockLength]uint64

// Workspace owns Argon2 memory and can be reused across hashes. Reuse is
// important for proof-of-work mining because allocating 64 MiB on every nonce
// would otherwise dominate the runtime and garbage collector.
type Workspace struct {
	blocks  []block
	memory  uint32
	threads uint32
}

func normalizedMemory(memory, threads uint32) uint32 {
	memory = memory / (syncPoints * threads) * (syncPoints * threads)
	if memory < 2*syncPoints*threads {
		memory = 2 * syncPoints * threads
	}
	return memory
}

func NewWorkspace(memory uint32, threads uint8) *Workspace {
	if threads < 1 {
		panic("argon2: parallelism degree too low")
	}
	m := normalizedMemory(memory, uint32(threads))
	return &Workspace{blocks: make([]block, m), memory: m, threads: uint32(threads)}
}

// IDKey implements Argon2id version 1.3. memory is in KiB.
func IDKey(password, salt []byte, timeCost, memory uint32, threads uint8, keyLen uint32) []byte {
	return NewWorkspace(memory, threads).IDKey(password, salt, timeCost, keyLen)
}

// IDKey derives a key using the workspace's fixed memory/parallelism settings.
func (w *Workspace) IDKey(password, salt []byte, timeCost, keyLen uint32) []byte {
	if timeCost < 1 {
		panic("argon2: number of rounds too small")
	}
	if w == nil || w.threads < 1 || uint32(len(w.blocks)) != w.memory {
		panic("argon2: invalid workspace")
	}
	h0 := initHash(password, salt, timeCost, w.memory, w.threads, keyLen)
	initBlocksInto(w.blocks, &h0, w.memory, w.threads)
	processBlocks(w.blocks, timeCost, w.memory, w.threads)
	return extractKey(w.blocks, w.memory, w.threads, keyLen)
}

func initHash(password, salt []byte, timeCost, memory, threads, keyLen uint32) [72]byte {
	var params [24]byte
	binary.LittleEndian.PutUint32(params[0:4], threads)
	binary.LittleEndian.PutUint32(params[4:8], keyLen)
	binary.LittleEndian.PutUint32(params[8:12], memory)
	binary.LittleEndian.PutUint32(params[12:16], timeCost)
	binary.LittleEndian.PutUint32(params[16:20], uint32(Version))
	binary.LittleEndian.PutUint32(params[20:24], uint32(argon2id))
	buf := make([]byte, 0, 24+4+len(password)+4+len(salt)+8)
	buf = append(buf, params[:]...)
	var tmp [4]byte
	binary.LittleEndian.PutUint32(tmp[:], uint32(len(password)))
	buf = append(buf, tmp[:]...)
	buf = append(buf, password...)
	binary.LittleEndian.PutUint32(tmp[:], uint32(len(salt)))
	buf = append(buf, tmp[:]...)
	buf = append(buf, salt...)
	binary.LittleEndian.PutUint32(tmp[:], 0)
	buf = append(buf, tmp[:]...) // secret length
	binary.LittleEndian.PutUint32(tmp[:], 0)
	buf = append(buf, tmp[:]...) // associated data length
	sum := blake2bSum(buf, 64)
	var h0 [72]byte
	copy(h0[:64], sum)
	return h0
}

func initBlocksInto(B []block, h0 *[72]byte, memory, threads uint32) {
	var in [72]byte
	copy(in[:], h0[:])
	for lane := uint32(0); lane < threads; lane++ {
		j := lane * (memory / threads)
		binary.LittleEndian.PutUint32(in[68:72], lane)
		binary.LittleEndian.PutUint32(in[64:68], 0)
		block0 := blake2bLong(1024, in[:])
		for i := range B[j] {
			B[j][i] = binary.LittleEndian.Uint64(block0[i*8:])
		}
		binary.LittleEndian.PutUint32(in[64:68], 1)
		block1 := blake2bLong(1024, in[:])
		for i := range B[j+1] {
			B[j+1][i] = binary.LittleEndian.Uint64(block1[i*8:])
		}
	}
}

func processBlocks(B []block, timeCost, memory, threads uint32) {
	lanes := memory / threads
	segments := lanes / syncPoints
	for pass := uint32(0); pass < timeCost; pass++ {
		for slice := uint32(0); slice < syncPoints; slice++ {
			// Consensus uses threads=1 for AQM64. Keep this generic and deterministic;
			// lanes are processed in order to avoid scheduler-dependent behavior.
			for lane := uint32(0); lane < threads; lane++ {
				processSegment(B, pass, slice, lane, memory, timeCost, threads, lanes, segments)
			}
		}
	}
}

func processSegment(B []block, pass, slice, lane, memory, timeCost, threads, lanes, segments uint32) {
	var addresses, in, zero block
	independent := pass == 0 && slice < syncPoints/2
	if independent {
		in[0] = uint64(pass)
		in[1] = uint64(lane)
		in[2] = uint64(slice)
		in[3] = uint64(memory)
		in[4] = uint64(timeCost)
		in[5] = uint64(argon2id)
	}
	index := uint32(0)
	if pass == 0 && slice == 0 {
		index = 2
		in[6]++
		processBlock(&addresses, &in, &zero)
		processBlock(&addresses, &addresses, &zero)
	}
	offset := lane*lanes + slice*segments + index
	for index < segments {
		prev := offset - 1
		if index == 0 && slice == 0 {
			prev += lanes
		}
		var random uint64
		if independent {
			if index%blockLength == 0 {
				in[6]++
				processBlock(&addresses, &in, &zero)
				processBlock(&addresses, &addresses, &zero)
			}
			random = addresses[index%blockLength]
		} else {
			random = B[prev][0]
		}
		newOffset := indexAlpha(random, lanes, segments, threads, pass, slice, lane, index)
		if pass == 0 {
			processBlock(&B[offset], &B[prev], &B[newOffset])
		} else {
			processBlockXOR(&B[offset], &B[prev], &B[newOffset])
		}
		index++
		offset++
	}
}

func extractKey(B []block, memory, threads, keyLen uint32) []byte {
	lanes := memory / threads
	last := B[memory-1]
	for lane := uint32(0); lane < threads-1; lane++ {
		x := B[lane*lanes+lanes-1]
		for i := range last {
			last[i] ^= x[i]
		}
	}
	raw := make([]byte, 1024)
	for i, v := range last {
		binary.LittleEndian.PutUint64(raw[i*8:], v)
	}
	return blake2bLong(keyLen, raw)
}

func indexAlpha(rand uint64, lanes, segments, threads, pass, slice, lane, index uint32) uint32 {
	refLane := uint32(rand>>32) % threads
	if pass == 0 && slice == 0 {
		refLane = lane
	}
	m, s := 3*segments, ((slice+1)%syncPoints)*segments
	if lane == refLane {
		m += index
	}
	if pass == 0 {
		m, s = slice*segments, 0
		if slice == 0 || lane == refLane {
			m += index
		}
	}
	if index == 0 || lane == refLane {
		m--
	}
	return phi(rand, uint64(m), uint64(s), refLane, lanes)
}

func phi(rand, m, s uint64, lane, lanes uint32) uint32 {
	p := rand & 0xffffffff
	p = (p * p) >> 32
	p = (p * m) >> 32
	return lane*lanes + uint32((s+m-(p+1))%uint64(lanes))
}

func blamka(x, y uint64) uint64 {
	return x + y + 2*uint64(uint32(x))*uint64(uint32(y))
}

func g(v *[16]uint64, a, b, c, d int) {
	v[a] = blamka(v[a], v[b])
	v[d] = bits.RotateLeft64(v[d]^v[a], -32)
	v[c] = blamka(v[c], v[d])
	v[b] = bits.RotateLeft64(v[b]^v[c], -24)
	v[a] = blamka(v[a], v[b])
	v[d] = bits.RotateLeft64(v[d]^v[a], -16)
	v[c] = blamka(v[c], v[d])
	v[b] = bits.RotateLeft64(v[b]^v[c], -63)
}

func round(v *[16]uint64) {
	g(v, 0, 4, 8, 12)
	g(v, 1, 5, 9, 13)
	g(v, 2, 6, 10, 14)
	g(v, 3, 7, 11, 15)
	g(v, 0, 5, 10, 15)
	g(v, 1, 6, 11, 12)
	g(v, 2, 7, 8, 13)
	g(v, 3, 4, 9, 14)
}

func processBlock(out, in1, in2 *block) {
	var r, z block
	for i := 0; i < blockLength; i++ {
		r[i] = in1[i] ^ in2[i]
		z[i] = r[i]
	}
	for i := 0; i < 8; i++ {
		var v [16]uint64
		copy(v[:], z[i*16:(i+1)*16])
		round(&v)
		copy(z[i*16:(i+1)*16], v[:])
	}
	for i := 0; i < 8; i++ {
		var v [16]uint64
		for j := 0; j < 8; j++ {
			v[2*j] = z[16*j+2*i]
			v[2*j+1] = z[16*j+2*i+1]
		}
		round(&v)
		for j := 0; j < 8; j++ {
			z[16*j+2*i] = v[2*j]
			z[16*j+2*i+1] = v[2*j+1]
		}
	}
	for i := 0; i < blockLength; i++ {
		out[i] = r[i] ^ z[i]
	}
}

func processBlockXOR(out, in1, in2 *block) {
	var r, z block
	for i := 0; i < blockLength; i++ {
		r[i] = in1[i] ^ in2[i]
		z[i] = r[i]
	}
	for i := 0; i < 8; i++ {
		var v [16]uint64
		copy(v[:], z[i*16:(i+1)*16])
		round(&v)
		copy(z[i*16:(i+1)*16], v[:])
	}
	for i := 0; i < 8; i++ {
		var v [16]uint64
		for j := 0; j < 8; j++ {
			v[2*j] = z[16*j+2*i]
			v[2*j+1] = z[16*j+2*i+1]
		}
		round(&v)
		for j := 0; j < 8; j++ {
			z[16*j+2*i] = v[2*j]
			z[16*j+2*i+1] = v[2*j+1]
		}
	}
	for i := 0; i < blockLength; i++ {
		out[i] ^= r[i] ^ z[i]
	}
}