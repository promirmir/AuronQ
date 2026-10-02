// Package mldsapure contains a portable ML-DSA-87 verifier used as a
// compatibility fallback when the host OS does not expose ML-DSA via CNG.
// Core arithmetic is adapted from KarpelesLab/mldsa (MIT), FIPS 204.
package mldsapure

const (
	N                  = 256
	Q                  = 8380417
	D                  = 13
	K87                = 8
	L87                = 7
	Beta87             = 2 * 60
	Gamma2QMinus1Div32 = (Q - 1) / 32
	Gamma1Pow19        = 1 << 19
	Tau60              = 60
	Omega75            = 75
	Lambda256          = 256
	PublicKeySize87    = 32 + K87*N*10/8
	SignatureSize87    = Lambda256/4 + L87*N*20/8 + Omega75 + K87
	EncodingSize10     = N * 10 / 8
	EncodingSize20     = N * 20 / 8
	EncodingSize4      = N * 4 / 8
	QMinus1Div2        = (Q - 1) / 2
)

type FieldElement uint32
type RingElement [N]FieldElement
type NttElement [N]FieldElement

const (
	qNegInv = 4236238847
	invN    = 41978
)

func fieldReduceOnce(a uint32) FieldElement   { x := a - Q; x += (x >> 31) * Q; return FieldElement(x) }
func fieldAdd(a, b FieldElement) FieldElement { return fieldReduceOnce(uint32(a) + uint32(b)) }
func fieldSub(a, b FieldElement) FieldElement { return fieldReduceOnce(uint32(a) - uint32(b) + Q) }
func fieldReduce(a uint64) FieldElement {
	t := uint32(a) * qNegInv
	return fieldReduceOnce(uint32((a + uint64(t)*Q) >> 32))
}
func fieldMul(a, b FieldElement) FieldElement { return fieldReduce(uint64(a) * uint64(b)) }
func polyAdd(a, b NttElement) (c NttElement) {
	for i := range c {
		c[i] = fieldAdd(a[i], b[i])
	}
	return
}
func polySub(a, b NttElement) (c NttElement) {
	for i := range c {
		c[i] = fieldSub(a[i], b[i])
	}
	return
}

func ringAdd(a, b RingElement) (c RingElement) {
	for i := range c {
		c[i] = fieldAdd(a[i], b[i])
	}
	return
}