package argon2pure

import (
	"encoding/hex"
	"testing"
)

func TestBlake2bVectors(t *testing.T) {
	got := hex.EncodeToString(blake2bSum(nil, 64))
	want := "786a02f742015903c6c6fd852552d272912f4740e15847618a86e217f71f5419d25e1031afee585313896444934eb04b903a685b1448b755d56f701afe9be2ce"
	if got != want {
		t.Fatalf("blake2b empty: %s", got)
	}
}

func TestArgon2idKnownVectors(t *testing.T) {
	// Independently generated with argon2-cffi low_level.hash_secret_raw,
	// Argon2 version 19 (v1.3), type=ID. Vectors intentionally exercise
	// different memory/time/parallelism/output sizes.
	vectors := []struct {
		passwordHex string
		saltHex     string
		time        uint32
		memory      uint32
		parallelism uint8
		outLen      uint32
		wantHex     string
	}{
		{"70617373776f7264", "736f6d6573616c74", 2, 32, 1, 32, "31111cc053ba0a799c0884148fd7ec9dc3631f3e8cf476cca9521d4ccc5136e8"},
		{"6175726f6e71", "3132333435363738", 1, 64, 1, 64, "5aa56a3d68a9b7a5967f8a3b514c76f8e4077c6c35385281fe6514e9bf13ecf798a84ba4e0dcc7179388e9a2bad44f6cac71d25ba2b7a2385699a3dafaf33474"},
		{"000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f", "30313233343536373839616263646566", 3, 128, 2, 48, "313e557bc7cda2e772aca33d9e2d8cd924453f97018453a773c8b0fcaedaed0010af3b94e2f719a768edb4673529b893"},
	}
	for i, v := range vectors {
		password, _ := hex.DecodeString(v.passwordHex)
		salt, _ := hex.DecodeString(v.saltHex)
		want, _ := hex.DecodeString(v.wantHex)
		got := IDKey(password, salt, v.time, v.memory, v.parallelism, v.outLen)
		if string(got) != string(want) {
			t.Fatalf("vector %d: got %x want %x", i, got, want)
		}
	}
}

func TestSingleLaneAcceleratorBoundaryMatchesIDKey(t *testing.T) {
	password := []byte("auronq-gpu-boundary-test")
	salt := []byte("0123456789abcdef0123456789abcdef")
	const (
		timeCost = uint32(2)
		memory   = uint32(64)
		keyLen   = uint32(64)
	)
	initial, err := PrepareSingleLaneBlocks(password, salt, timeCost, memory, keyLen)
	if err != nil {
		t.Fatal(err)
	}
	m := normalizedMemory(memory, 1)
	B := make([]block, m)
	copy(B[0][:], initial[:blockLength])
	copy(B[1][:], initial[blockLength:])
	processBlocks(B, timeCost, m, 1)

	final := make([]uint64, blockLength)
	copy(final, B[m-1][:])
	got, err := ExtractSingleLaneKey(final, keyLen)
	if err != nil {
		t.Fatal(err)
	}
	want := IDKey(password, salt, timeCost, memory, 1, keyLen)
	if string(got) != string(want) {
		t.Fatalf("accelerator boundary mismatch: got %x want %x", got, want)
	}
}
