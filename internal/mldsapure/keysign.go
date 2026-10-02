package mldsapure

import (
	sha3 "auronq/internal/sha3compat"
	"crypto/rand"
	"errors"
	"io"
)

type Key87 struct {
	rho  [32]byte
	key  [32]byte
	tr   [64]byte
	s1   [L87]RingElement
	s2   [K87]RingElement
	t0   [K87]RingElement
	a    [K87 * L87]NttElement
	t1   [K87]RingElement
	seed [32]byte
}

func GenerateKey87() (*Key87, error) {
	var seed [32]byte
	if _, e := io.ReadFull(rand.Reader, seed[:]); e != nil {
		return nil, e
	}
	return NewKey87(seed[:])
}
func NewKey87(seed []byte) (*Key87, error) {
	if len(seed) != 32 {
		return nil, errors.New("mldsa: invalid seed length")
	}
	k := &Key87{}
	copy(k.seed[:], seed)
	k.generate()
	return k, nil
}
func (k *Key87) Seed() []byte { b := make([]byte, 32); copy(b, k.seed[:]); return b }
func (k *Key87) PublicBytes() []byte {
	b := make([]byte, PublicKeySize87)
	copy(b[:32], k.rho[:])
	off := 32
	for i := 0; i < K87; i++ {
		p := packT1(k.t1[i])
		copy(b[off:], p)
		off += EncodingSize10
	}
	return b
}
func (k *Key87) generate() {
	h := sha3.NewSHAKE256()
	h.Write(k.seed[:])
	h.Write([]byte{K87, L87})
	var expanded [128]byte
	h.Read(expanded[:])
	copy(k.rho[:], expanded[:32])
	rho1 := expanded[32:96]
	copy(k.key[:], expanded[96:])
	for i := 0; i < L87; i++ {
		k.s1[i] = sampleBoundedPoly(rho1, uint16(i))
	}
	for i := 0; i < K87; i++ {
		k.s2[i] = sampleBoundedPoly(rho1, uint16(L87+i))
	}
	for i := 0; i < K87; i++ {
		for j := 0; j < L87; j++ {
			k.a[i*L87+j] = sampleNTTPoly(k.rho[:], byte(j), byte(i))
		}
	}
	var s1n [L87]NttElement
	for i := 0; i < L87; i++ {
		s1n[i] = NTT(k.s1[i])
	}
	for i := 0; i < K87; i++ {
		var acc NttElement
		for j := 0; j < L87; j++ {
			acc = polyAdd(acc, nttMul(k.a[i*L87+j], s1n[j]))
		}
		t := ringAdd(InvNTT(acc), k.s2[i])
		for j := 0; j < N; j++ {
			k.t1[i][j], k.t0[i][j] = power2Round(t[j])
		}
	}
	h.Reset()
	h.Write(k.PublicBytes())
	h.Read(k.tr[:])
}
func (k *Key87) Sign(message []byte) ([]byte, error) {
	if len(message) == 0 {
		return nil, errors.New("mldsa: refusing empty message")
	}
	var rnd [32]byte
	if _, e := io.ReadFull(rand.Reader, rnd[:]); e != nil {
		return nil, e
	}
	m := make([]byte, 2+len(message))
	copy(m[2:], message)
	return k.signInternal(rnd[:], m)
}
func (k *Key87) signInternal(rnd, m []byte) ([]byte, error) {
	h := sha3.NewSHAKE256()
	h.Write(k.tr[:])
	h.Write(m)
	var mu [64]byte
	h.Read(mu[:])
	h.Reset()
	h.Write(k.key[:])
	h.Write(rnd)
	h.Write(mu[:])
	var rhoP [64]byte
	h.Read(rhoP[:])
	var s1n [L87]NttElement
	var s2n [K87]NttElement
	var t0n [K87]NttElement
	for i := 0; i < L87; i++ {
		s1n[i] = NTT(k.s1[i])
	}
	for i := 0; i < K87; i++ {
		s2n[i] = NTT(k.s2[i])
		t0n[i] = NTT(k.t0[i])
	}
	var seedBuf [66]byte
	copy(seedBuf[:64], rhoP[:])
	for kappa := uint16(0); ; kappa += L87 {
		var y [L87]RingElement
		for i := 0; i < L87; i++ {
			seedBuf[64] = byte(kappa + uint16(i))
			seedBuf[65] = byte((kappa + uint16(i)) >> 8)
			y[i] = expandMask19(seedBuf[:])
		}
		var yn [L87]NttElement
		for i := 0; i < L87; i++ {
			yn[i] = NTT(y[i])
		}
		var w [K87]RingElement
		var w1 [K87]RingElement
		for i := 0; i < K87; i++ {
			var acc NttElement
			for j := 0; j < L87; j++ {
				acc = polyAdd(acc, nttMul(k.a[i*L87+j], yn[j]))
			}
			w[i] = InvNTT(acc)
			for j := 0; j < N; j++ {
				w1[i][j] = FieldElement(highBits(w[i][j], Gamma2QMinus1Div32))
			}
		}
		h.Reset()
		h.Write(mu[:])
		for i := 0; i < K87; i++ {
			h.Write(packW1_4(w1[i]))
		}
		var cTilde [Lambda256 / 4]byte
		h.Read(cTilde[:])
		c := sampleChallenge(cTilde[:], Tau60)
		cn := NTT(c)
		var z [L87]RingElement
		for i := 0; i < L87; i++ {
			cs1 := InvNTT(nttMul(cn, s1n[i]))
			z[i] = ringAdd(y[i], cs1)
		}
		if vectorInfinityNorm(z[:]) >= Gamma1Pow19-Beta87 {
			continue
		}
		var r0 [K87][N]int32
		for i := 0; i < K87; i++ {
			cs2 := InvNTT(nttMul(cn, s2n[i]))
			for j := 0; j < N; j++ {
				_, r0[i][j] = decompose(fieldSub(w[i][j], cs2[j]), Gamma2QMinus1Div32)
			}
		}
		if vectorInfinityNormSigned(r0[:]) >= int32(Gamma2QMinus1Div32-Beta87) {
			continue
		}
		var ct0 [K87]RingElement
		for i := 0; i < K87; i++ {
			ct0[i] = InvNTT(nttMul(cn, t0n[i]))
		}
		if vectorInfinityNorm(ct0[:]) >= Gamma2QMinus1Div32 {
			continue
		}
		var hints [K87]RingElement
		for i := 0; i < K87; i++ {
			cs2 := InvNTT(nttMul(cn, s2n[i]))
			for j := 0; j < N; j++ {
				r := fieldSub(w[i][j], cs2[j])
				hints[i][j] = makeHint(ct0[i][j], r, Gamma2QMinus1Div32)
			}
		}
		if countOnes(hints[:]) > Omega75 {
			continue
		}
		sig := make([]byte, SignatureSize87)
		copy(sig, cTilde[:])
		off := len(cTilde)
		for i := 0; i < L87; i++ {
			p := packZ19(z[i])
			copy(sig[off:], p)
			off += EncodingSize20
		}
		copy(sig[off:], packHint(&hints))
		return sig, nil
	}
}