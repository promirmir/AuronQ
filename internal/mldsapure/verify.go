package mldsapure

import (
	sha3 "auronq/internal/sha3compat"
	"errors"
)

type PublicKey87 struct {
	rho [32]byte
	t1  [K87]RingElement
	tr  [64]byte
	a   [K87 * L87]NttElement
}

func NewPublicKey87(b []byte) (*PublicKey87, error) {
	if len(b) != PublicKeySize87 {
		return nil, errors.New("mldsa: invalid public key length")
	}
	pk := &PublicKey87{}
	copy(pk.rho[:], b[:32])
	off := 32
	for i := 0; i < K87; i++ {
		pk.t1[i] = unpackT1(b[off : off+EncodingSize10])
		off += EncodingSize10
	}
	for i := 0; i < K87; i++ {
		for j := 0; j < L87; j++ {
			pk.a[i*L87+j] = sampleNTTPoly(pk.rho[:], byte(j), byte(i))
		}
	}
	h := sha3.NewSHAKE256()
	h.Write(b)
	h.Read(pk.tr[:])
	return pk, nil
}
func Verify87(pub, sig, message []byte) bool {
	pk, e := NewPublicKey87(pub)
	if e != nil {
		return false
	}
	return pk.Verify(sig, message, nil)
}
func (pk *PublicKey87) Verify(sig, message, context []byte) bool {
	if len(sig) != SignatureSize87 || len(context) > 255 {
		return false
	}
	m := make([]byte, 2+len(context)+len(message))
	m[0] = 0
	m[1] = byte(len(context))
	copy(m[2:], context)
	copy(m[2+len(context):], message)
	return pk.verifyInternal(sig, m)
}
func (pk *PublicKey87) verifyInternal(sig, mPrime []byte) bool {
	h := sha3.NewSHAKE256()
	h.Write(pk.tr[:])
	h.Write(mPrime)
	var mu [64]byte
	h.Read(mu[:])
	cTilde := sig[:Lambda256/4]
	off := Lambda256 / 4
	var z [L87]RingElement
	for i := 0; i < L87; i++ {
		z[i] = unpackZ19(sig[off : off+EncodingSize20])
		off += EncodingSize20
	}
	if vectorInfinityNorm(z[:]) >= Gamma1Pow19-Beta87 {
		return false
	}
	var hints [K87]RingElement
	if !unpackHint(sig[off:], &hints) {
		return false
	}
	c := sampleChallenge(cTilde, Tau60)
	cNTT := NTT(c)
	var zNTT [L87]NttElement
	for i := 0; i < L87; i++ {
		zNTT[i] = NTT(z[i])
	}
	var t1NTT [K87]NttElement
	for i := 0; i < K87; i++ {
		var t RingElement
		for j := 0; j < N; j++ {
			t[j] = pk.t1[i][j] << D
		}
		t1NTT[i] = NTT(t)
	}
	h.Reset()
	h.Write(mu[:])
	for i := 0; i < K87; i++ {
		var acc NttElement
		for j := 0; j < L87; j++ {
			acc = polyAdd(acc, nttMul(pk.a[i*L87+j], zNTT[j]))
		}
		ct1 := nttMul(cNTT, t1NTT[i])
		acc = polySub(acc, ct1)
		w := InvNTT(acc)
		var w1 RingElement
		for j := 0; j < N; j++ {
			w1[j] = useHint(hints[i][j], w[j], Gamma2QMinus1Div32)
		}
		h.Write(packW1_4(w1))
	}
	var check [Lambda256 / 4]byte
	h.Read(check[:])
	var diff byte
	for i := range cTilde {
		diff |= cTilde[i] ^ check[i]
	}
	return diff == 0
}