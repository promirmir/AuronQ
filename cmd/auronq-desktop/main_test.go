package main

import (
	"reflect"
	"testing"

	aq "auronq/internal/auronq"
)

func TestMergeBootstrapMetadataSameNetwork(t *testing.T) {
	existing := &aq.NetworkConfig{
		Name:               "AuronQ Mainnet",
		ProtocolVersion:    1,
		NetworkByte:        81,
		FounderAddress:     "keep-me",
		CoinbaseMaturity:   100,
		SeedPeers:          []string{"https://old-seed.example"},
		BootstrapManifests: nil,
	}
	bundled := *existing
	bundled.FounderAddress = "must-not-replace-existing"
	bundled.SeedPeers = []string{"https://old-seed.example"}
	bundled.DNSSeeds = []string{"seed.auronq.example"}
	bundled.BootstrapManifests = []string{"https://raw.githubusercontent.com/promirmir/AuronQ/main/bootstrap.json"}

	if !mergeBootstrapMetadata(existing, &bundled) {
		t.Fatal("expected bootstrap metadata merge to report a change")
	}
	if existing.FounderAddress != "keep-me" {
		t.Fatalf("consensus/non-bootstrap field was modified: %q", existing.FounderAddress)
	}
	if !reflect.DeepEqual(existing.SeedPeers, []string{"https://old-seed.example"}) {
		t.Fatalf("unexpected seed peers: %#v", existing.SeedPeers)
	}
	if !reflect.DeepEqual(existing.DNSSeeds, []string{"seed.auronq.example"}) {
		t.Fatalf("unexpected DNS seeds: %#v", existing.DNSSeeds)
	}
	if !reflect.DeepEqual(existing.BootstrapManifests, []string{"https://raw.githubusercontent.com/promirmir/AuronQ/main/bootstrap.json"}) {
		t.Fatalf("unexpected bootstrap manifests: %#v", existing.BootstrapManifests)
	}
	if mergeBootstrapMetadata(existing, &bundled) {
		t.Fatal("second merge should be idempotent")
	}
}

func TestMergeBootstrapMetadataRejectsDifferentNetwork(t *testing.T) {
	existing := &aq.NetworkConfig{ProtocolVersion: 1, NetworkByte: 81, CoinbaseMaturity: 100}
	bundled := &aq.NetworkConfig{
		ProtocolVersion:    1,
		NetworkByte:        82,
		CoinbaseMaturity:   100,
		BootstrapManifests: []string{"https://example.invalid/bootstrap.json"},
	}
	if mergeBootstrapMetadata(existing, bundled) {
		t.Fatal("different Network ID must not merge bootstrap metadata")
	}
	if len(existing.BootstrapManifests) != 0 {
		t.Fatalf("different network modified existing config: %#v", existing.BootstrapManifests)
	}
}
