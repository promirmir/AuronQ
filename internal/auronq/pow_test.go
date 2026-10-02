package auronq

import (
	"testing"

	"auronq/internal/argon2pure"
)

func TestAQM64DeterministicVector(t *testing.T) {
	var prev Hash
	for i := range prev {
		prev[i] = byte(i)
	}
	var merkle Hash
	for i := range merkle {
		merkle[i] = byte(255 - i)
	}
	h := BlockHeader{
		Version: BlockVersion, PowAlgo: PowAlgorithmAQM64, Height: 42,
		PrevHash: prev, MerkleRoot: merkle, Timestamp: 1700000000,
		Target: PowLimit, Nonce: 123456789,
	}
	got, err := PowHash(h)
	if err != nil {
		t.Fatal(err)
	}
	const want = "868bc0843225cc911d568b26b05df65565dc40677f682b2bc9ace18c5f27ad683ca3b6772c47c08a8326280d7396fa06fd6748bcd1cfb1c7317f6c8e4dbc9353"
	if got.String() != want {
		t.Fatalf("AQM64 vector mismatch: got %s want %s", got, want)
	}
}

func TestAQM64ProductionParameterVector(t *testing.T) {
	var prev Hash
	for i := range prev {
		prev[i] = byte(i)
	}
	var merkle Hash
	for i := range merkle {
		merkle[i] = byte(255 - i)
	}
	header := BlockHeader{
		Version: BlockVersion, PowAlgo: PowAlgorithmAQM64, Height: 42,
		PrevHash: prev, MerkleRoot: merkle, Timestamp: 1700000000,
		Target: PowLimit, Nonce: 123456789,
	}
	h := &aqm64Hasher{
		memory: AQM64MemoryKiB,
		time:   AQM64TimeCost,
		ws:     argon2pure.NewWorkspace(AQM64MemoryKiB, AQM64Parallelism),
	}
	got := h.hash(header)
	const want = "d2d994ed295c897aeaa74467517dfbcaa40b4f91f31c0c54e0e3b642cd879e3cf7f0d6d07560a4fee2ea73d48ea845c6e6ad1793a2d3471dbb50cd3211ab1ea2"
	if got.String() != want {
		t.Fatalf("AQM64 production vector mismatch: got %s want %s", got, want)
	}
}

func TestAQM64RejectsUnknownAlgorithm(t *testing.T) {
	_, err := PowHash(BlockHeader{PowAlgo: 999})
	if err == nil {
		t.Fatal("expected unsupported PoW algorithm error")
	}
}
