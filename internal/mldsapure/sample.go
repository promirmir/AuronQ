package mldsapure

import sha3 "auronq/internal/sha3compat"

func sampleNTTPoly(rho []byte, s, r byte) NttElement {
	h := sha3.NewSHAKE128()
	h.Write(rho)
	h.Write([]byte{s, r})
	var buf [168]byte
	var a NttElement
	j := 0
	for {
		h.Read(buf[:])
		for i := 0; i < len(buf) && j < N; i += 3 {
			v := uint32(buf[i]) | uint32(buf[i+1])<<8 | (uint32(buf[i+2])&0x7f)<<16
			if v < Q {
				a[j] = FieldElement(v)
				j++
			}
		}
		if j >= N {
			return a
		}
	}
}
func sampleChallenge(seed []byte, tau int) RingElement {
	h := sha3.NewSHAKE256()
	h.Write(seed)
	var buf [136]byte
	h.Read(buf[:])
	var signs uint64
	for i := 0; i < 8; i++ {
		signs |= uint64(buf[i]) << (8 * i)
	}
	off := 8
	var c RingElement
	for i := N - tau; i < N; i++ {
		var j byte
		for {
			if off >= len(buf) {
				h.Read(buf[:])
				off = 0
			}
			j = buf[off]
			off++
			if int(j) <= i {
				break
			}
		}
		c[i] = c[j]
		if signs&1 == 0 {
			c[j] = 1
		} else {
			c[j] = Q - 1
		}
		signs >>= 1
	}
	return c
}

func sampleBoundedPoly(seed []byte, nonce uint16) RingElement {
	h := sha3.NewSHAKE256()
	h.Write(seed)
	h.Write([]byte{byte(nonce), byte(nonce >> 8)})
	var buf [136]byte
	var a RingElement
	j, off := 0, 0
	h.Read(buf[:])
	for j < N {
		if off >= len(buf) {
			h.Read(buf[:])
			off = 0
		}
		z0 := buf[off] & 0x0f
		z1 := buf[off] >> 4
		off++
		if z0 < 15 {
			z0 = z0 - (z0/5)*5
			a[j] = fieldSub(2, FieldElement(z0))
			j++
		}
		if j < N && z1 < 15 {
			z1 = z1 - (z1/5)*5
			a[j] = fieldSub(2, FieldElement(z1))
			j++
		}
	}
	return a
}
func expandMask19(seed []byte) RingElement {
	h := sha3.NewSHAKE256()
	h.Write(seed)
	buf := make([]byte, 640)
	h.Read(buf)
	return unpackZ19(buf)
}
