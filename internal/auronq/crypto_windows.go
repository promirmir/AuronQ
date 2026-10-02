//go:build windows && legacycrypto

package auronq

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"runtime"
	"syscall"
	"unsafe"
)

var (
	bcrypt      = syscall.NewLazyDLL("bcrypt.dll")
	procOpen    = bcrypt.NewProc("BCryptOpenAlgorithmProvider")
	procClose   = bcrypt.NewProc("BCryptCloseAlgorithmProvider")
	procImport  = bcrypt.NewProc("BCryptImportKeyPair")
	procExport  = bcrypt.NewProc("BCryptExportKey")
	procDestroy = bcrypt.NewProc("BCryptDestroyKey")
	procSign    = bcrypt.NewProc("BCryptSignHash")
	procVerify  = bcrypt.NewProc("BCryptVerifySignature")
)

const (
	mldsaPublicMagic = uint32(0x4B505344)
	mldsaSeedMagic   = uint32(0x53535344)
)

func wstr(s string) *uint16 { p, _ := syscall.UTF16PtrFromString(s); return p }
func ntOK(r uintptr) bool   { return uint32(r) == 0 }

func openMLDSA() (uintptr, error) {
	var h uintptr
	r, _, _ := procOpen.Call(uintptr(unsafe.Pointer(&h)), uintptr(unsafe.Pointer(wstr("ML-DSA"))), 0, 0)
	if !ntOK(r) {
		return 0, fmt.Errorf("Windows CNG ML-DSA unavailable (NTSTATUS 0x%08x); Windows 11 24H2+ is required", uint32(r))
	}
	return h, nil
}
func closeAlg(h uintptr) {
	if h != 0 {
		procClose.Call(h, 0)
	}
}
func destroyKey(h uintptr) {
	if h != 0 {
		procDestroy.Call(h)
	}
}

func pqBlob(magic uint32, key []byte) []byte {
	// BCRYPT_PQDSA_KEY_BLOB (3x uint32 LE), UTF-16LE L"87\0" (6 bytes), key material.
	out := make([]byte, 12+6+len(key))
	binary.LittleEndian.PutUint32(out[0:4], magic)
	binary.LittleEndian.PutUint32(out[4:8], 6)
	binary.LittleEndian.PutUint32(out[8:12], uint32(len(key)))
	copy(out[12:18], []byte{'8', 0, '7', 0, 0, 0})
	copy(out[18:], key)
	return out
}
func parsePQBlob(blob []byte, wantMagic uint32, wantLen int) ([]byte, error) {
	if len(blob) < 18 {
		return nil, errors.New("short CNG PQDSA blob")
	}
	if binary.LittleEndian.Uint32(blob[:4]) != wantMagic {
		return nil, errors.New("unexpected CNG PQDSA magic")
	}
	ps := int(binary.LittleEndian.Uint32(blob[4:8]))
	kl := int(binary.LittleEndian.Uint32(blob[8:12]))
	if ps < 2 || 12+ps+kl != len(blob) {
		return nil, errors.New("invalid CNG PQDSA blob")
	}
	if wantLen > 0 && kl != wantLen {
		return nil, fmt.Errorf("unexpected CNG key size %d", kl)
	}
	return append([]byte(nil), blob[12+ps:]...), nil
}
func importBlob(alg uintptr, typ string, blob []byte) (uintptr, error) {
	var key uintptr
	var ptr uintptr
	if len(blob) > 0 {
		ptr = uintptr(unsafe.Pointer(&blob[0]))
	}
	r, _, _ := procImport.Call(alg, 0, uintptr(unsafe.Pointer(wstr(typ))), uintptr(unsafe.Pointer(&key)), ptr, uintptr(uint32(len(blob))), 0)
	runtime.KeepAlive(blob)
	if !ntOK(r) {
		return 0, fmt.Errorf("BCryptImportKeyPair failed: 0x%08x", uint32(r))
	}
	return key, nil
}
func exportBlob(key uintptr, typ string) ([]byte, error) {
	var n uint32
	r, _, _ := procExport.Call(key, 0, uintptr(unsafe.Pointer(wstr(typ))), 0, 0, uintptr(unsafe.Pointer(&n)), 0)
	if !ntOK(r) {
		return nil, fmt.Errorf("BCryptExportKey(size) failed: 0x%08x", uint32(r))
	}
	if n == 0 || n > 1<<20 {
		return nil, errors.New("invalid CNG export size")
	}
	out := make([]byte, n)
	r, _, _ = procExport.Call(key, 0, uintptr(unsafe.Pointer(wstr(typ))), uintptr(unsafe.Pointer(&out[0])), uintptr(n), uintptr(unsafe.Pointer(&n)), 0)
	runtime.KeepAlive(out)
	if !ntOK(r) {
		return nil, fmt.Errorf("BCryptExportKey failed: 0x%08x", uint32(r))
	}
	return out[:n], nil
}
func keyFromSeed(seed []byte) (uintptr, uintptr, error) {
	if len(seed) != MLDSASeedSize {
		return 0, 0, fmt.Errorf("invalid seed length: %d", len(seed))
	}
	alg, err := openMLDSA()
	if err != nil {
		return 0, 0, err
	}
	key, err := importBlob(alg, "PQDSAPRIVATESEEDBLOB", pqBlob(mldsaSeedMagic, seed))
	if err != nil {
		closeAlg(alg)
		return 0, 0, err
	}
	return alg, key, nil
}

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
	alg, key, err := keyFromSeed(seed)
	if err != nil {
		return nil, err
	}
	defer closeAlg(alg)
	defer destroyKey(key)
	blob, err := exportBlob(key, "PQDSAPUBLICBLOB")
	if err != nil {
		return nil, err
	}
	return parsePQBlob(blob, mldsaPublicMagic, MLDSA87PublicKeySize)
}
func SignMLDSA87(seed, msg []byte) ([]byte, error) {
	if len(msg) == 0 {
		return nil, errors.New("refusing to sign empty message")
	}
	alg, key, err := keyFromSeed(seed)
	if err != nil {
		return nil, err
	}
	defer closeAlg(alg)
	defer destroyKey(key)
	var n uint32
	r, _, _ := procSign.Call(key, 0, uintptr(unsafe.Pointer(&msg[0])), uintptr(uint32(len(msg))), 0, 0, uintptr(unsafe.Pointer(&n)), 0)
	runtime.KeepAlive(msg)
	if !ntOK(r) {
		return nil, fmt.Errorf("BCryptSignHash(size) failed: 0x%08x", uint32(r))
	}
	if n != MLDSA87SignatureSize {
		return nil, fmt.Errorf("unexpected ML-DSA signature size %d", n)
	}
	sig := make([]byte, n)
	r, _, _ = procSign.Call(key, 0, uintptr(unsafe.Pointer(&msg[0])), uintptr(uint32(len(msg))), uintptr(unsafe.Pointer(&sig[0])), uintptr(n), uintptr(unsafe.Pointer(&n)), 0)
	runtime.KeepAlive(msg)
	runtime.KeepAlive(sig)
	if !ntOK(r) {
		return nil, fmt.Errorf("BCryptSignHash failed: 0x%08x", uint32(r))
	}
	return sig[:n], nil
}
func VerifyMLDSA87(pub, msg, sig []byte) bool {
	if len(pub) != MLDSA87PublicKeySize || len(sig) != MLDSA87SignatureSize || len(msg) == 0 {
		return false
	}
	alg, err := openMLDSA()
	if err != nil {
		return false
	}
	defer closeAlg(alg)
	key, err := importBlob(alg, "PQDSAPUBLICBLOB", pqBlob(mldsaPublicMagic, pub))
	if err != nil {
		return false
	}
	defer destroyKey(key)
	r, _, _ := procVerify.Call(key, 0, uintptr(unsafe.Pointer(&msg[0])), uintptr(uint32(len(msg))), uintptr(unsafe.Pointer(&sig[0])), uintptr(uint32(len(sig))), 0)
	runtime.KeepAlive(msg)
	runtime.KeepAlive(sig)
	return ntOK(r)
}