package argon2pure

import (
	"bytes"
	"testing"
)

func TestPreparedSingleLaneWorkspaceMatchesIDKey(t *testing.T) {
	password := []byte("auronq-universal-miner-test")
	salt := []byte("0123456789abcdef0123456789abcdef")
	const (
		timeCost = uint32(2)
		memory   = uint32(1024)
		keyLen   = uint32(64)
	)
	initial, err := PrepareSingleLaneBlocks(password, salt, timeCost, memory, keyLen)
	if err != nil {
		t.Fatal(err)
	}
	ws := NewSingleLaneWorkspace(memory)
	finalBlock, err := ws.ProcessPreparedSingleLane(initial, timeCost)
	if err != nil {
		t.Fatal(err)
	}
	got, err := ExtractSingleLaneKey(finalBlock, keyLen)
	if err != nil {
		t.Fatal(err)
	}
	want := IDKey(password, salt, timeCost, memory, 1, keyLen)
	if !bytes.Equal(got, want) {
		t.Fatalf("prepared workspace result does not match canonical Argon2id")
	}

	// Reuse the same workspace for a second input to catch stale-memory bugs.
	password2 := []byte("second-input")
	initial2, err := PrepareSingleLaneBlocks(password2, salt, timeCost, memory, keyLen)
	if err != nil {
		t.Fatal(err)
	}
	finalBlock2, err := ws.ProcessPreparedSingleLane(initial2, timeCost)
	if err != nil {
		t.Fatal(err)
	}
	got2, err := ExtractSingleLaneKey(finalBlock2, keyLen)
	if err != nil {
		t.Fatal(err)
	}
	want2 := IDKey(password2, salt, timeCost, memory, 1, keyLen)
	if !bytes.Equal(got2, want2) {
		t.Fatalf("reused prepared workspace result does not match canonical Argon2id")
	}
}
