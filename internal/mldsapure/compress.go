package mldsapure

func highBits(r FieldElement, gamma2 uint32) uint32 {
	r1 := int32((r + 127) >> 7)
	r1 = (r1*1025 + (1 << 21)) >> 22
	return uint32(r1) & 15
}
func decompose(r FieldElement, gamma2 uint32) (uint32, int32) {
	r1 := highBits(r, gamma2)
	r0 := int32(r) - int32(r1)*int32(gamma2)*2
	r0 -= ((int32(QMinus1Div2) - r0) >> 31) & Q
	return r1, r0
}
func useHint(hint, r FieldElement, gamma2 uint32) FieldElement {
	r1, r0 := decompose(r, gamma2)
	if hint == 0 {
		return FieldElement(r1)
	}
	if r0 > 0 {
		return FieldElement((r1 + 1) & 15)
	}
	return FieldElement((r1 - 1) & 15)
}
func infinityNorm(a FieldElement) uint32 {
	if uint32(a) <= QMinus1Div2 {
		return uint32(a)
	}
	return Q - uint32(a)
}
func vectorInfinityNorm(v []RingElement) uint32 {
	var max uint32
	for i := range v {
		for j := range v[i] {
			n := infinityNorm(v[i][j])
			if n > max {
				max = n
			}
		}
	}
	return max
}

func power2Round(r FieldElement) (r1, r0 FieldElement) {
	r1 = r >> D
	r0 = r - r1<<D
	const half = 1 << (D - 1)
	if r0 > half {
		r0 = fieldSub(r0, 1<<D)
		r1++
	}
	return
}
func makeHint(z, r FieldElement, gamma2 uint32) FieldElement {
	if highBits(fieldAdd(r, z), gamma2) != highBits(r, gamma2) {
		return 1
	}
	return 0
}
func countOnes(v []RingElement) int {
	n := 0
	for i := range v {
		for j := range v[i] {
			if v[i][j] != 0 {
				n++
			}
		}
	}
	return n
}
func vectorInfinityNormSigned(v [][N]int32) int32 {
	var max int32
	for i := range v {
		for j := range v[i] {
			x := v[i][j]
			if x < 0 {
				x = -x
			}
			if x > max {
				max = x
			}
		}
	}
	return max
}