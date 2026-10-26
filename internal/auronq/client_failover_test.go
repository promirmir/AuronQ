package auronq

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func testClientNode(t *testing.T, networkID Hash, height uint64, tip Hash, peers []string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/status":
			_ = json.NewEncoder(w).Encode(Status{
				Network:   "AuronQ Mainnet",
				NetworkID: networkID,
				Height:    height,
				Tip:       tip,
				Peers:     len(peers),
			})
		case "/p2p/hello":
			_ = json.NewEncoder(w).Encode(Hello{
				ProtocolVersion: 1,
				NetworkID:       networkID,
				Height:          height,
				Tip:             tip,
				Peers:           peers,
			})
		default:
			http.NotFound(w, r)
		}
	}))
}

func TestResolveVerifiedClientFallsBackFromDeadPreferred(t *testing.T) {
	var networkID Hash
	networkID[0] = 0x44
	var tip Hash
	tip[0] = 0x99
	seed := testClientNode(t, networkID, 100, tip, nil)
	defer seed.Close()

	store := filepath.Join(t.TempDir(), "client-peers.json")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	got, err := ResolveVerifiedClient(ctx, VerifiedClientConfig{
		ExpectedNetworkID: networkID,
		Preferred:         "http://127.0.0.1:1",
		Seeds:             []string{seed.URL},
		StorePath:         store,
		Timeout:           300 * time.Millisecond,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Endpoint != seed.URL {
		t.Fatalf("endpoint=%q want %q", got.Endpoint, seed.URL)
	}
	if got.Status.Height != 100 || got.Status.Tip != tip {
		t.Fatalf("unexpected status: %+v", got.Status)
	}
}

func TestResolveVerifiedClientFailsClosedOnPreferredNetworkMismatch(t *testing.T) {
	var expected, wrong Hash
	expected[0] = 1
	wrong[0] = 2
	var tip Hash
	tip[0] = 3
	preferred := testClientNode(t, wrong, 10, tip, nil)
	defer preferred.Close()
	seed := testClientNode(t, expected, 10, tip, nil)
	defer seed.Close()

	_, err := ResolveVerifiedClient(context.Background(), VerifiedClientConfig{
		ExpectedNetworkID: expected,
		Preferred:         preferred.URL,
		Seeds:             []string{seed.URL},
		Timeout:           time.Second,
	})
	if !errors.Is(err, ErrClientNetworkMismatch) {
		t.Fatalf("expected network mismatch, got %v", err)
	}
}

func TestResolveVerifiedClientFailsClosedOnSameHeightConflict(t *testing.T) {
	var networkID Hash
	networkID[0] = 7
	var tipA, tipB Hash
	tipA[0] = 1
	tipB[0] = 2
	a := testClientNode(t, networkID, 77, tipA, nil)
	defer a.Close()
	b := testClientNode(t, networkID, 77, tipB, nil)
	defer b.Close()

	_, err := ResolveVerifiedClient(context.Background(), VerifiedClientConfig{
		ExpectedNetworkID: networkID,
		Seeds:             []string{a.URL, b.URL},
		Timeout:           time.Second,
	})
	if !errors.Is(err, ErrClientChainConflict) {
		t.Fatalf("expected chain conflict, got %v", err)
	}
}

func TestResolveVerifiedClientPersistsOnlyVerifiedGossipPeer(t *testing.T) {
	var networkID Hash
	networkID[0] = 0x42
	var tip Hash
	tip[0] = 0x55

	learned := testClientNode(t, networkID, 55, tip, nil)
	defer learned.Close()
	unverified := "http://127.0.0.1:1"
	preferred := testClientNode(t, networkID, 55, tip, []string{learned.URL, unverified})
	defer preferred.Close()

	store := filepath.Join(t.TempDir(), "client-peers.json")
	_, err := ResolveVerifiedClient(context.Background(), VerifiedClientConfig{
		ExpectedNetworkID: networkID,
		Preferred:         preferred.URL,
		StorePath:         store,
		Timeout:           300 * time.Millisecond,
	})
	if err != nil {
		t.Fatal(err)
	}
	records := loadClientPeerStore(store)
	if _, ok := records[preferred.URL]; !ok {
		t.Fatalf("preferred verified peer not stored: %v", records)
	}
	if _, ok := records[learned.URL]; !ok {
		t.Fatalf("independently verified gossip peer not stored: %v", records)
	}
	if _, ok := records[unverified]; ok {
		t.Fatalf("unreachable gossip peer was persisted: %v", records)
	}
}

func TestClientPeerStoreQuarantinesRepeatedFailure(t *testing.T) {
	var networkID Hash
	networkID[0] = 9
	var tip Hash
	tip[0] = 8
	healthy := testClientNode(t, networkID, 20, tip, nil)
	defer healthy.Close()

	store := filepath.Join(t.TempDir(), "client-peers.json")
	dead := "http://127.0.0.1:1"
	initial := clientPeerStore{Peers: []clientPeerRecord{{URL: dead, Failures: 2}}}
	b, err := json.Marshal(initial)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(store, b, 0600); err != nil {
		t.Fatal(err)
	}

	_, err = ResolveVerifiedClient(context.Background(), VerifiedClientConfig{
		ExpectedNetworkID: networkID,
		Seeds:             []string{healthy.URL},
		StorePath:         store,
		Timeout:           300 * time.Millisecond,
	})
	if err != nil {
		t.Fatal(err)
	}
	records := loadClientPeerStore(store)
	rec, ok := records[dead]
	if !ok {
		t.Fatalf("dead peer record missing: %v", records)
	}
	if rec.Failures < 3 || rec.QuarantineUntil <= time.Now().Unix() {
		t.Fatalf("dead peer not quarantined: %+v", rec)
	}
}
