package mldsapure

func unpackT1(b []byte) RingElement {
	var f RingElement
	for i := 0; i < N; i += 4 {
		x := uint64(b[0]) | uint64(b[1])<<8 | uint64(b[2])<<16 | uint64(b[3])<<24 | uint64(b[4])<<32
		f[i] = FieldElement(x & 0x3ff)
		f[i+1] = FieldElement((x >> 10) & 0x3ff)
		f[i+2] = FieldElement((x >> 20) & 0x3ff)
		f[i+3] = FieldElement((x >> 30) & 0x3ff)
		b = b[5:]
	}
	return f
}
func unpackZ19(b []byte) RingElement {
	var f RingElement
	const g = 1 << 19
	const mask = (1 << 20) - 1
	for i := 0; i < N; i += 4 {
		x1 := uint64(b[0]) | uint64(b[1])<<8 | uint64(b[2])<<16 | uint64(b[3])<<24 | uint64(b[4])<<32 | uint64(b[5])<<40 | uint64(b[6])<<48 | uint64(b[7])<<56
		x2 := uint64(b[8]) | uint64(b[9])<<8
		b = b[10:]
		f[i] = fieldSub(g, FieldElement(x1&mask))
		f[i+1] = fieldSub(g, FieldElement((x1>>20)&mask))
		f[i+2] = fieldSub(g, FieldElement((x1>>40)&mask))
		f[i+3] = fieldSub(g, FieldElement(((x1>>60)|(x2<<4))&mask))
	}
	return f
}
func packW1_4(f RingElement) []byte {
	b := make([]byte, EncodingSize4)
	for i := 0; i < N; i += 2 {
		b[i/2] = byte(f[i]) | byte(f[i+1])<<4
	}
	return b
}
func unpackHint(b []byte, hints *[K87]RingElement) bool {
	idx := 0
	for i := 0; i < K87; i++ {
		limit := int(b[Omega75+i])
		if limit < idx || limit > Omega75 {
			return false
		}
		prev := idx
		for ; idx < limit; idx++ {
			pos := b[idx]
			if idx > prev && b[idx-1] >= pos {
				return false
			}
			hints[i][pos] = 1
		}
	}
	for ; idx < Omega75; idx++ {
		if b[idx] != 0 {
			return false
		}
	}
	return true
}

func packT1(f RingElement) []byte {
	b := make([]byte, EncodingSize10)
	for i := 0; i < N; i += 4 {
		x := uint64(f[i]) | uint64(f[i+1])<<10 | uint64(f[i+2])<<20 | uint64(f[i+3])<<30
		p := i / 4 * 5
		b[p] = byte(x)
		b[p+1] = byte(x >> 8)
		b[p+2] = byte(x >> 16)
		b[p+3] = byte(x >> 24)
		b[p+4] = byte(x >> 32)
	}
	return b
}
func packZ19(f RingElement) []byte {
	b := make([]byte, EncodingSize20)
	const g = 1 << 19
	idx := 0
	for i := 0; i < N; i += 4 {
		var x1, x2 uint64
		x1 = uint64(fieldSub(g, f[i]))
		x1 |= uint64(fieldSub(g, f[i+1])) << 20
		x1 |= uint64(fieldSub(g, f[i+2])) << 40
		x2 = uint64(fieldSub(g, f[i+3]))
		x1 |= x2 << 60
		x2 >>= 4
		for j := 0; j < 8; j++ {
			b[idx+j] = byte(x1 >> (8 * j))
		}
		b[idx+8] = byte(x2)
		b[idx+9] = byte(x2 >> 8)
		idx += 10
	}
	return b
}
func packHint(hints *[K87]RingElement) []byte {
	b := make([]byte, Omega75+K87)
	idx := 0
	for i := 0; i < K87; i++ {
		for j := 0; j < N; j++ {
			if hints[i][j] != 0 {
				b[idx] = byte(j)
				idx++
			}
		}
		b[Omega75+i] = byte(idx)
	}
	return b
}