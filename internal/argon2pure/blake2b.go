package argon2pure

import (
	"encoding/binary"
	"math/bits"
)

var blake2bIV = [8]uint64{
	0x6a09e667f3bcc908, 0xbb67ae8584caa73b, 0x3c6ef372fe94f82b, 0xa54ff53a5f1d36f1,
	0x510e527fade682d1, 0x9b05688c2b3e6c1f, 0x1f83d9abfb41bd6b, 0x5be0cd19137e2179,
}

var blake2bSigma = [12][16]uint8{
	{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
	{14, 10, 4, 8, 9, 15, 13, 6, 1, 12, 0, 2, 11, 7, 5, 3},
	{11, 8, 12, 0, 5, 2, 15, 13, 10, 14, 3, 6, 7, 1, 9, 4},
	{7, 9, 3, 1, 13, 12, 11, 14, 2, 6, 5, 10, 4, 0, 15, 8},
	{9, 0, 5, 7, 2, 4, 10, 15, 14, 1, 11, 12, 6, 8, 3, 13},
	{2, 12, 6, 10, 0, 11, 8, 3, 4, 13, 7, 5, 15, 14, 1, 9},
	{12, 5, 1, 15, 14, 13, 4, 10, 0, 7, 6, 3, 9, 2, 8, 11},
	{13, 11, 7, 14, 12, 1, 3, 9, 5, 0, 15, 4, 8, 6, 2, 10},
	{6, 15, 14, 9, 11, 3, 0, 8, 12, 2, 13, 7, 1, 4, 10, 5},
	{10, 2, 8, 4, 7, 6, 1, 5, 15, 11, 9, 14, 3, 12, 13, 0},
	{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
	{14, 10, 4, 8, 9, 15, 13, 6, 1, 12, 0, 2, 11, 7, 5, 3},
}

func blake2bG(v *[16]uint64, a, b, c, d int, x, y uint64) {
	v[a] = v[a] + v[b] + x
	v[d] = bits.RotateLeft64(v[d]^v[a], -32)
	v[c] += v[d]
	v[b] = bits.RotateLeft64(v[b]^v[c], -24)
	v[a] = v[a] + v[b] + y
	v[d] = bits.RotateLeft64(v[d]^v[a], -16)
	v[c] += v[d]
	v[b] = bits.RotateLeft64(v[b]^v[c], -63)
}

func blake2bCompress(h *[8]uint64, block []byte, t0, t1 uint64, final bool) {
	var m [16]uint64
	for i := 0; i < 16; i++ {
		m[i] = binary.LittleEndian.Uint64(block[i*8:])
	}
	var v [16]uint64
	copy(v[:8], h[:])
	copy(v[8:], blake2bIV[:])
	v[12] ^= t0
	v[13] ^= t1
	if final {
		v[14] = ^v[14]
	}
	for r := 0; r < 12; r++ {
		s := blake2bSigma[r]
		blake2bG(&v, 0, 4, 8, 12, m[s[0]], m[s[1]])
		blake2bG(&v, 1, 5, 9, 13, m[s[2]], m[s[3]])
		blake2bG(&v, 2, 6, 10, 14, m[s[4]], m[s[5]])
		blake2bG(&v, 3, 7, 11, 15, m[s[6]], m[s[7]])
		blake2bG(&v, 0, 5, 10, 15, m[s[8]], m[s[9]])
		blake2bG(&v, 1, 6, 11, 12, m[s[10]], m[s[11]])
		blake2bG(&v, 2, 7, 8, 13, m[s[12]], m[s[13]])
		blake2bG(&v, 3, 4, 9, 14, m[s[14]], m[s[15]])
	}
	for i := 0; i < 8; i++ {
		h[i] ^= v[i] ^ v[i+8]
	}
}

// blake2bSum computes an unkeyed BLAKE2b digest of size 1..64 bytes.
func blake2bSum(data []byte, outLen int) []byte {
	if outLen < 1 || outLen > 64 {
		panic("blake2b: invalid output length")
	}
	h := blake2bIV
	h[0] ^= 0x01010000 ^ uint64(outLen)
	var t0, t1 uint64
	for len(data) > 128 {
		old := t0
		t0 += 128
		if t0 < old {
			t1++
		}
		blake2bCompress(&h, data[:128], t0, t1, false)
		data = data[128:]
	}
	var block [128]byte
	copy(block[:], data)
	old := t0
	t0 += uint64(len(data))
	if t0 < old {
		t1++
	}
	blake2bCompress(&h, block[:], t0, t1, true)
	full := make([]byte, 64)
	for i := 0; i < 8; i++ {
		binary.LittleEndian.PutUint64(full[i*8:], h[i])
	}
	return append([]byte(nil), full[:outLen]...)
}

func blake2bLong(outLen uint32, input []byte) []byte {
	prefix := make([]byte, 4+len(input))
	binary.LittleEndian.PutUint32(prefix[:4], outLen)
	copy(prefix[4:], input)
	if outLen <= 64 {
		return blake2bSum(prefix, int(outLen))
	}
	out := make([]byte, outLen)
	block := blake2bSum(prefix, 64)
	copy(out[:32], block[:32])
	pos := 32
	for int(outLen)-pos > 64 {
		block = blake2bSum(block, 64)
		copy(out[pos:pos+32], block[:32])
		pos += 32
	}
	block = blake2bSum(block, int(outLen)-pos)
	copy(out[pos:], block)
	return out
}
