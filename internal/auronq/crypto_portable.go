//go:build !legacycrypto

package auronq

import (
	"crypto/rand"
	"errors"
	"fmt"

	"auronq/internal/mldsapure"
)

// GenerateMLDSA87 creates a FIPS 204 ML-DSA-87 key from a 32-byte random seed.
// This implementation is portable Go and does not depend on Windows CNG or OpenSSL.
func GenerateMLDSA87() (seed, pub []byte, err error) {
	seed = make([]byte, MLDSASeedSize)
	if _, err = rand.Read(seed); err != nil {
		return nil, nil, err
	}
	pub, err = PublicFromSeed(seed)
	if err != nil {
		for i := range seed {
			seed[i] = 0
		}
		return nil, nil, err
	}
	return seed, pub, nil
}

func PublicFromSeed(seed []byte) ([]byte, error) {
	if len(seed) != MLDSASeedSize {
		return nil, fmt.Errorf("invalid seed length: %d", len(seed))
	}
	k, err := mldsapure.NewKey87(seed)
	if err != nil {
		return nil, err
	}
	return k.PublicBytes(), nil
}

func SignMLDSA87(seed, msg []byte) ([]byte, error) {
	if len(seed) != MLDSASeedSize {
		return nil, errors.New("invalid ML-DSA seed")
	}
	if len(msg) == 0 {
		return nil, errors.New("refusing to sign empty message")
	}
	k, err := mldsapure.NewKey87(seed)
	if err != nil {
		return nil, err
	}
	return k.Sign(msg)
}

func VerifyMLDSA87(pub, msg, sig []byte) bool {
	if len(pub) != MLDSA87PublicKeySize || len(sig) != MLDSA87SignatureSize || len(msg) == 0 {
		return false
	}
	return mldsapure.Verify87(pub, sig, msg)
}
