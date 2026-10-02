package auronq

import "fmt"

// SignatureSchemeID is encoded in every address/UTXO. The wire format is
// deliberately algorithm-agnostic: adding a future post-quantum signature
// scheme does not require a new transaction or address format.
type SignatureSchemeID uint16

type SignatureSchemeInfo struct {
	ID            uint16
	Name          string
	PublicKeySize int
	SignatureSize int
	SecurityClass string
}

var signatureSchemes = map[uint16]SignatureSchemeInfo{
	SchemeMLDSA87: {
		ID:            SchemeMLDSA87,
		Name:          "ML-DSA-87",
		PublicKeySize: MLDSA87PublicKeySize,
		SignatureSize: MLDSA87SignatureSize,
		SecurityClass: "NIST FIPS 204 / category 5",
	},
}

func SignatureScheme(scheme uint16) (SignatureSchemeInfo, bool) {
	i, ok := signatureSchemes[scheme]
	return i, ok
}

func IsConsensusScheme(scheme uint16) bool {
	_, ok := SignatureScheme(scheme)
	return ok
}

func VerifySignature(scheme uint16, pub, msg, sig []byte) bool {
	info, ok := SignatureScheme(scheme)
	if !ok || len(pub) != info.PublicKeySize || len(sig) != info.SignatureSize {
		return false
	}
	switch scheme {
	case SchemeMLDSA87:
		return VerifyMLDSA87(pub, msg, sig)
	default:
		return false
	}
}

func SignWithScheme(scheme uint16, secret, msg []byte) ([]byte, error) {
	if !IsConsensusScheme(scheme) {
		return nil, fmt.Errorf("unsupported signature scheme %d", scheme)
	}
	switch scheme {
	case SchemeMLDSA87:
		return SignMLDSA87(secret, msg)
	default:
		return nil, fmt.Errorf("signature scheme %d has no signer", scheme)
	}
}

// MaxWitnessSizes returns consensus-safe maximums derived from all currently
// supported schemes. The transaction format can therefore accept a future
// scheme after a consensus upgrade without changing the serialization layout.
func MaxWitnessSizes() (pubKey, signature int) {
	for _, s := range signatureSchemes {
		if s.PublicKeySize > pubKey {
			pubKey = s.PublicKeySize
		}
		if s.SignatureSize > signature {
			signature = s.SignatureSize
		}
	}
	return
}