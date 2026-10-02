package sha3

func newShake128() *state { return newShake128Generic() }
func newShake256() *state { return newShake256Generic() }

// FIPS 204 naming used by Go 1.24 crypto/sha3 and the vendored ML-DSA code.
func NewSHAKE128() ShakeHash { return NewShake128() }
func NewSHAKE256() ShakeHash { return NewShake256() }