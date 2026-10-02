package auronq

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base32"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/bits"
	"strings"
)

const (
	SchemeMLDSA87        uint16 = 1
	MLDSA87PublicKeySize        = 2592
	MLDSA87SignatureSize        = 4627
	MLDSASeedSize               = 32
)

var b32 = base32.NewEncoding("abcdefghijklmnopqrstuvwxyz234567").WithPadding(base32.NoPadding)

func Hash512(data []byte) Hash    { return Hash(sha512.Sum512(data)) }
func Hash256(data []byte) KeyHash { v := sha512.Sum512_256(data); return KeyHash(v) }

// ScryptKey is a self-contained RFC 7914 scrypt implementation. Keeping it in
// the consensus/wallet code avoids platform crypto-library differences and
// preserves the auronq-wallet-v1 format on Linux and Windows.
func ScryptKey(password string, salt []byte, N, r, p uint64, outLen int) ([]byte, error) {
	if password == "" {
		return nil, errors.New("empty wallet password")
	}
	if len(salt) < 16 {
		return nil, errors.New("salt too short")
	}
	if N < 2 || N&(N-1) != 0 {
		return nil, errors.New("scrypt N must be a power of two > 1")
	}
	if r == 0 || p == 0 || outLen <= 0 {
		return nil, errors.New("invalid scrypt parameters")
	}
	if r > uint64(^uint(0)>>1)/128 || p > uint64(^uint(0)>>1)/(128*r) {
		return nil, errors.New("scrypt parameters too large")
	}
	blockLen := int(128 * r)
	if N > uint64((512*1024*1024)/blockLen) {
		return nil, errors.New("scrypt memory requirement exceeds 512 MiB")
	}
	B := pbkdf2SHA256([]byte(password), salt, 1, int(p)*blockLen)
	for i := 0; i < int(p); i++ {
		if err := scryptROMix(B[i*blockLen:(i+1)*blockLen], int(r), int(N)); err != nil {
			return nil, err
		}
	}
	out := pbkdf2SHA256([]byte(password), B, 1, outLen)
	for i := range B {
		B[i] = 0
	}
	return out, nil
}

func pbkdf2SHA256(password, salt []byte, iter, keyLen int) []byte {
	out := make([]byte, 0, keyLen)
	var ctr [4]byte
	for block := 1; len(out) < keyLen; block++ {
		binary.BigEndian.PutUint32(ctr[:], uint32(block))
		mac := hmac.New(sha256.New, password)
		mac.Write(salt)
		mac.Write(ctr[:])
		u := mac.Sum(nil)
		t := append([]byte(nil), u...)
		for j := 1; j < iter; j++ {
			mac = hmac.New(sha256.New, password)
			mac.Write(u)
			u = mac.Sum(nil)
			for k := range t {
				t[k] ^= u[k]
			}
		}
		out = append(out, t...)
	}
	return out[:keyLen]
}

func scryptROMix(b []byte, r, N int) error {
	if len(b) != 128*r {
		return errors.New("invalid scrypt block length")
	}
	X := append([]byte(nil), b...)
	V := make([]byte, len(X)*N)
	Y := make([]byte, len(X))
	for i := 0; i < N; i++ {
		copy(V[i*len(X):(i+1)*len(X)], X)
		scryptBlockMix(X, Y, r)
		X, Y = Y, X
	}
	for i := 0; i < N; i++ {
		j := int(binary.LittleEndian.Uint64(X[(2*r-1)*64:]) & uint64(N-1))
		v := V[j*len(X) : (j+1)*len(X)]
		for k := range X {
			X[k] ^= v[k]
		}
		scryptBlockMix(X, Y, r)
		X, Y = Y, X
	}
	copy(b, X)
	for i := range X {
		X[i] = 0
	}
	for i := range Y {
		Y[i] = 0
	}
	for i := range V {
		V[i] = 0
	}
	return nil
}

func scryptBlockMix(B, Y []byte, r int) {
	var x [16]uint32
	last := B[(2*r-1)*64 : 2*r*64]
	for i := 0; i < 16; i++ {
		x[i] = binary.LittleEndian.Uint32(last[i*4:])
	}
	tmp := make([]byte, 64)
	for i := 0; i < 2*r; i++ {
		bi := B[i*64 : (i+1)*64]
		for j := 0; j < 16; j++ {
			x[j] ^= binary.LittleEndian.Uint32(bi[j*4:])
		}
		salsa208(&x)
		for j := 0; j < 16; j++ {
			binary.LittleEndian.PutUint32(tmp[j*4:], x[j])
		}
		dest := i / 2
		if i&1 == 1 {
			dest = r + i/2
		}
		copy(Y[dest*64:(dest+1)*64], tmp)
	}
}

func salsa208(x *[16]uint32) {
	z := *x
	for i := 0; i < 8; i += 2 {
		z[4] ^= bits.RotateLeft32(z[0]+z[12], 7)
		z[8] ^= bits.RotateLeft32(z[4]+z[0], 9)
		z[12] ^= bits.RotateLeft32(z[8]+z[4], 13)
		z[0] ^= bits.RotateLeft32(z[12]+z[8], 18)
		z[9] ^= bits.RotateLeft32(z[5]+z[1], 7)
		z[13] ^= bits.RotateLeft32(z[9]+z[5], 9)
		z[1] ^= bits.RotateLeft32(z[13]+z[9], 13)
		z[5] ^= bits.RotateLeft32(z[1]+z[13], 18)
		z[14] ^= bits.RotateLeft32(z[10]+z[6], 7)
		z[2] ^= bits.RotateLeft32(z[14]+z[10], 9)
		z[6] ^= bits.RotateLeft32(z[2]+z[14], 13)
		z[10] ^= bits.RotateLeft32(z[6]+z[2], 18)
		z[3] ^= bits.RotateLeft32(z[15]+z[11], 7)
		z[7] ^= bits.RotateLeft32(z[3]+z[15], 9)
		z[11] ^= bits.RotateLeft32(z[7]+z[3], 13)
		z[15] ^= bits.RotateLeft32(z[11]+z[7], 18)
		z[1] ^= bits.RotateLeft32(z[0]+z[3], 7)
		z[2] ^= bits.RotateLeft32(z[1]+z[0], 9)
		z[3] ^= bits.RotateLeft32(z[2]+z[1], 13)
		z[0] ^= bits.RotateLeft32(z[3]+z[2], 18)
		z[6] ^= bits.RotateLeft32(z[5]+z[4], 7)
		z[7] ^= bits.RotateLeft32(z[6]+z[5], 9)
		z[4] ^= bits.RotateLeft32(z[7]+z[6], 13)
		z[5] ^= bits.RotateLeft32(z[4]+z[7], 18)
		z[11] ^= bits.RotateLeft32(z[10]+z[9], 7)
		z[8] ^= bits.RotateLeft32(z[11]+z[10], 9)
		z[9] ^= bits.RotateLeft32(z[8]+z[11], 13)
		z[10] ^= bits.RotateLeft32(z[9]+z[8], 18)
		z[12] ^= bits.RotateLeft32(z[15]+z[14], 7)
		z[13] ^= bits.RotateLeft32(z[12]+z[15], 9)
		z[14] ^= bits.RotateLeft32(z[13]+z[12], 13)
		z[15] ^= bits.RotateLeft32(z[14]+z[13], 18)
	}
	for i := 0; i < 16; i++ {
		x[i] += z[i]
	}
}

func EncryptSeed(seed []byte, password string) (salt, nonce, ciphertext []byte, err error) {
	salt = make([]byte, 16)
	if _, err = rand.Read(salt); err != nil {
		return
	}
	key, err := ScryptKey(password, salt, 1<<17, 8, 1, 32)
	if err != nil {
		return nil, nil, nil, err
	}
	defer func() {
		for i := range key {
			key[i] = 0
		}
	}()
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, nil, err
	}
	nonce = make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return nil, nil, nil, err
	}
	ciphertext = gcm.Seal(nil, nonce, seed, []byte("AURONQ_WALLET_V1"))
	return
}

func DecryptSeed(salt, nonce, ciphertext []byte, password string) ([]byte, error) {
	key, err := ScryptKey(password, salt, 1<<17, 8, 1, 32)
	if err != nil {
		return nil, err
	}
	defer func() {
		for i := range key {
			key[i] = 0
		}
	}()
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	seed, err := gcm.Open(nil, nonce, ciphertext, []byte("AURONQ_WALLET_V1"))
	if err != nil {
		return nil, errors.New("invalid wallet password or corrupted wallet")
	}
	if len(seed) != MLDSASeedSize {
		return nil, errors.New("corrupt wallet seed")
	}
	return seed, nil
}

func AddressFromPub(pub []byte, scheme uint16, network byte) string {
	kh := Hash256(pub)
	payload := make([]byte, 0, 41)
	payload = append(payload, network, byte(scheme>>8), byte(scheme))
	payload = append(payload, kh[:]...)
	check := Hash256(append([]byte("AURONQ_ADDR_V1"), payload...))
	payload = append(payload, check[:6]...)
	return "aurq1" + strings.ToLower(b32.EncodeToString(payload))
}

func DecodeAddress(addr string) (network byte, scheme uint16, kh KeyHash, err error) {
	lower := strings.ToLower(addr)
	if !strings.HasPrefix(lower, "aurq1") {
		err = errors.New("invalid address prefix")
		return
	}
	body := strings.TrimPrefix(lower, "aurq1")
	raw, e := b32.DecodeString(body)
	if e != nil {
		err = errors.New("invalid address encoding")
		return
	}
	if b32.EncodeToString(raw) != body {
		err = errors.New("non-canonical address encoding")
		return
	}
	if len(raw) != 41 {
		err = fmt.Errorf("invalid address length: %d", len(raw))
		return
	}
	payload := raw[:35]
	check := Hash256(append([]byte("AURONQ_ADDR_V1"), payload...))
	if !constantEqual(raw[35:], check[:6]) {
		err = errors.New("address checksum mismatch")
		return
	}
	network = raw[0]
	scheme = uint16(raw[1])<<8 | uint16(raw[2])
	copy(kh[:], raw[3:35])
	return
}
func constantEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var x byte
	for i := range a {
		x |= a[i] ^ b[i]
	}
	return x == 0
}
func Hex(b []byte) string { return hex.EncodeToString(b) }
