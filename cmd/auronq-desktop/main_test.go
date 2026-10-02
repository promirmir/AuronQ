package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
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

func TestWalletDeleteRemovesOnlySelectedWallet(t *testing.T) {
	dir := t.TempDir()
	wallets := filepath.Join(dir, "wallets")
	if err := os.MkdirAll(wallets, 0700); err != nil { t.Fatal(err) }
	target := filepath.Join(wallets, "miner.wallet")
	other := filepath.Join(wallets, "keep.wallet")
	if err := os.WriteFile(target, []byte("{}"), 0600); err != nil { t.Fatal(err) }
	if err := os.WriteFile(other, []byte("{}"), 0600); err != nil { t.Fatal(err) }
	a := &App{walletsDir: wallets}
	r := httptest.NewRequest(http.MethodPost, "/api/wallet/delete", strings.NewReader(`{"Name":"miner","Confirm":"miner"}`))
	w := httptest.NewRecorder()
	a.handleWalletDelete(w, r)
	if w.Code != http.StatusOK { t.Fatalf("delete status = %d, body=%s", w.Code, w.Body.String()) }
	if _, err := os.Stat(target); !os.IsNotExist(err) { t.Fatalf("target wallet still exists or unexpected stat error: %v", err) }
	if _, err := os.Stat(other); err != nil { t.Fatalf("unrelated wallet was touched: %v", err) }
}

func TestWalletDeleteRejectsTraversalAndWrongConfirmation(t *testing.T) {
	dir := t.TempDir()
	wallets := filepath.Join(dir, "wallets")
	if err := os.MkdirAll(wallets, 0700); err != nil { t.Fatal(err) }
	outside := filepath.Join(dir, "victim.wallet")
	if err := os.WriteFile(outside, []byte("keep"), 0600); err != nil { t.Fatal(err) }
	a := &App{walletsDir: wallets}
	for _, body := range []string{
		`{"Name":"../victim","Confirm":"../victim"}`,
		`{"Name":"miner","Confirm":"wrong"}`,
	} {
		r := httptest.NewRequest(http.MethodPost, "/api/wallet/delete", strings.NewReader(body))
		w := httptest.NewRecorder()
		a.handleWalletDelete(w, r)
		if w.Code == http.StatusOK { t.Fatalf("unsafe delete unexpectedly succeeded for %s", body) }
	}
	if _, err := os.Stat(outside); err != nil { t.Fatalf("outside file was touched: %v", err) }
}

func TestWalletDeleteRejectsActiveMiningWallet(t *testing.T) {
	dir := t.TempDir()
	wallets := filepath.Join(dir, "wallets")
	if err := os.MkdirAll(wallets, 0700); err != nil { t.Fatal(err) }
	p := filepath.Join(wallets, "miner.wallet")
	if err := os.WriteFile(p, []byte("{}"), 0600); err != nil { t.Fatal(err) }
	a := &App{walletsDir: wallets, miner: minerState{Running: true, Wallet: "miner"}}
	r := httptest.NewRequest(http.MethodPost, "/api/wallet/delete", strings.NewReader(`{"Name":"miner","Confirm":"miner"}`))
	w := httptest.NewRecorder()
	a.handleWalletDelete(w, r)
	if w.Code != http.StatusConflict { t.Fatalf("status = %d, want %d; body=%s", w.Code, http.StatusConflict, w.Body.String()) }
	if _, err := os.Stat(p); err != nil { t.Fatalf("active mining wallet was removed: %v", err) }
}
