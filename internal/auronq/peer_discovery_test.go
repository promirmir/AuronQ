package auronq

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func testPeerNetwork(t *testing.T) (*NetworkConfig, *Chain) {
	t.Helper()
	_, pub, err := GenerateMLDSA87()
	if err != nil {
		t.Fatal(err)
	}
	founder := AddressFromPub(pub, SchemeMLDSA87, TestnetNetworkByte)
	g, err := CreateGenesis(founder, TestnetNetworkByte, time.Now().Unix()-20)
	if err != nil {
		t.Fatal(err)
	}
	// The test only exercises peer discovery/hello, so the genesis need not be mined.
	netCfg := &NetworkConfig{Name: "peer-test", ProtocolVersion: 1, NetworkByte: TestnetNetworkByte, CoinbaseMaturity: 10, FounderAddress: founder, Genesis: g}
	c, err := OpenChain(t.TempDir(), netCfg)
	if err != nil {
		t.Fatal(err)
	}
	return netCfg, c
}

func TestHelloAdvertisesOnlyPublicPeers(t *testing.T) {
	_, c := testPeerNetwork(t)
	n := NewNode(c, NodeConfig{Peers: []string{"http://127.0.0.1:18444", "http://localhost:18445", "http://[::1]:18446", "http://100.64.0.10:18445", "http://192.168.1.10:18444", "http://1.1.1.1:18445"}})
	req := httptest.NewRequest("GET", "/p2p/hello", nil)
	rec := httptest.NewRecorder()
	n.handler().ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("status=%d", rec.Code)
	}
	var h Hello
	if err := json.Unmarshal(rec.Body.Bytes(), &h); err != nil {
		t.Fatal(err)
	}
	if len(h.Peers) != 1 || h.Peers[0] != "http://1.1.1.1:18445" {
		t.Fatalf("advertised peers=%v", h.Peers)
	}
}

func TestDiscoveredNonPublicPeerIsIgnored(t *testing.T) {
	_, c := testPeerNetwork(t)
	n := NewNode(c, NodeConfig{Peers: []string{"http://1.1.1.1:18445"}})
	n.addDiscoveredPeer("http://127.0.0.1:18444")
	n.addDiscoveredPeer("http://localhost:18444")
	n.addDiscoveredPeer("http://[::1]:18444")
	n.addDiscoveredPeer("http://100.64.0.10:18444")
	n.addDiscoveredPeer("http://192.168.1.2:18444")
	got := n.peerList()
	if len(got) != 1 || got[0] != "http://1.1.1.1:18445" {
		t.Fatalf("peer list=%v", got)
	}
}

func TestPublicPeerStorePersistsLearnedPeer(t *testing.T) {
	_, c := testPeerNetwork(t)
	store := t.TempDir() + "/peers.json"
	n := NewNode(c, NodeConfig{PeerStorePath: store})
	n.addDiscoveredPeer("http://1.1.1.1:18444")
	got := n.peerList()
	if len(got) != 1 || got[0] != "http://1.1.1.1:18444" {
		t.Fatalf("peer list=%v", got)
	}
	n2 := NewNode(c, NodeConfig{PeerStorePath: store})
	got = n2.peerList()
	if len(got) != 1 || got[0] != "http://1.1.1.1:18444" {
		t.Fatalf("reloaded peer list=%v", got)
	}
}

func TestPrivatePeerAnnouncementRejected(t *testing.T) {
	_, c := testPeerNetwork(t)
	n := NewNode(c, NodeConfig{})
	body := `{"protocol_version":1,"network_id":"` + c.NetworkID().String() + `","advertise":"http://192.168.1.20:18444"}`
	req := httptest.NewRequest("POST", "/p2p/announce", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	n.handler().ServeHTTP(rec, req)
	if rec.Code != 400 {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestDNSSeedDiscoveryAddsOnlyPublicAddresses(t *testing.T) {
	netCfg, c := testPeerNetwork(t)
	netCfg.DNSSeeds = []string{"seed.example"}
	n := NewNode(c, NodeConfig{LookupHost: func(ctx context.Context, host string) ([]string, error) {
		if host != "seed.example" {
			t.Fatalf("unexpected host %q", host)
		}
		return []string{"1.1.1.1", "192.168.1.20", "100.64.0.10"}, nil
	}})
	n.refreshDNSSeeds(context.Background())
	got := n.peerList()
	if len(got) != 1 || got[0] != "http://1.1.1.1:18444" {
		t.Fatalf("peer list=%v", got)
	}
}

func TestObservedAnnouncementPeer(t *testing.T) {
	if got := observedAnnouncementPeer("8.8.8.8:54321", 18444); got != "http://8.8.8.8:18444" {
		t.Fatalf("got %q", got)
	}
	if got := observedAnnouncementPeer("192.168.1.10:54321", 18444); got != "" {
		t.Fatalf("private address accepted: %q", got)
	}
	if got := observedAnnouncementPeer("8.8.8.8:54321", 0); got != "" {
		t.Fatalf("zero port accepted: %q", got)
	}
}

func TestAnnounceSelfSendsListenPortWithoutManualAdvertise(t *testing.T) {
	_, c := testPeerNetwork(t)
	var got PeerAnnounce
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/p2p/announce" {
			t.Fatalf("path=%s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatal(err)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()
	n := NewNode(c, NodeConfig{Listen: "0.0.0.0:18444"})
	n.announceSelf(srv.URL)
	if got.ListenPort != 18444 {
		t.Fatalf("listen_port=%d", got.ListenPort)
	}
	if got.Advertise != "" {
		t.Fatalf("unexpected manual advertise=%q", got.Advertise)
	}
}

func TestDeadLearnedPeerIsPrunedButBootstrapIsKept(t *testing.T) {
	_, c := testPeerNetwork(t)
	n := NewNode(c, NodeConfig{Peers: []string{"http://1.1.1.1:18444"}})
	n.addDiscoveredPeer("http://8.8.8.8:18444")
	for i := 0; i < peerFailureDrop; i++ {
		n.recordPeerFailure("http://8.8.8.8:18444")
		n.recordPeerFailure("http://1.1.1.1:18444")
	}
	got := n.peerList()
	if len(got) != 1 || got[0] != "http://1.1.1.1:18444" {
		t.Fatalf("peer list=%v", got)
	}
}

func TestBootstrapMetadataDoesNotChangeNetworkID(t *testing.T) {
	netCfg, _ := testPeerNetwork(t)
	before := netCfg.NetworkID()
	netCfg.SeedPeers = []string{"http://1.1.1.1:18444"}
	netCfg.DNSSeeds = []string{"seed.example"}
	if got := netCfg.NetworkID(); got != before {
		t.Fatalf("bootstrap metadata changed network id: before=%s after=%s", before, got)
	}
}

func TestNetworkIDCommitsProtocolVersion(t *testing.T) {
	netCfg, _ := testPeerNetwork(t)
	before := netCfg.NetworkID()
	netCfg.ProtocolVersion = 2
	if got := netCfg.NetworkID(); got == before {
		t.Fatalf("protocol version did not change network id: %s", got)
	}
}

func TestPerIPBlockRateLimit(t *testing.T) {
	_, c := testPeerNetwork(t)
	n := NewNode(c, NodeConfig{})
	for i := 0; i < maxBlocksPerIPWindow; i++ {
		r := httptest.NewRequest("POST", "/v1/block", nil)
		r.RemoteAddr = "8.8.8.8:40000"
		if !n.allowBlockRequest(r) {
			t.Fatalf("request %d rejected before configured limit", i+1)
		}
	}
	r := httptest.NewRequest("POST", "/v1/block", nil)
	r.RemoteAddr = "8.8.8.8:40001"
	if n.allowBlockRequest(r) {
		t.Fatal("block request accepted beyond per-IP window limit")
	}

	// A different source IP gets an independent budget.
	r2 := httptest.NewRequest("POST", "/v1/block", nil)
	r2.RemoteAddr = "9.9.9.9:40000"
	if !n.allowBlockRequest(r2) {
		t.Fatal("independent source IP was incorrectly rate limited")
	}
}

func TestBootstrapManifestAddsMatchingNetworkPeers(t *testing.T) {
	netCfg, c := testPeerNetwork(t)
	var srv *httptest.Server
	srv = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bootstrap.json" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(BootstrapManifest{
			NetworkID: netCfg.NetworkID(),
			Peers:     []string{"https://seed-one.example", "http://1.1.1.1:18444"},
			ExpiresAt: time.Now().Add(time.Hour).Unix(),
		})
	}))
	defer srv.Close()
	netCfg.BootstrapManifests = []string{srv.URL + "/bootstrap.json"}
	n := NewNode(c, NodeConfig{})
	n.client = srv.Client()
	n.refreshBootstrapManifests(context.Background())
	got := n.peerList()
	want := map[string]bool{"https://seed-one.example": true, "http://1.1.1.1:18444": true}
	if len(got) != len(want) {
		t.Fatalf("peer list=%v", got)
	}
	for _, p := range got {
		if !want[p] {
			t.Fatalf("unexpected peer %q in %v", p, got)
		}
	}
}

func TestBootstrapManifestRejectsWrongNetworkAndExpired(t *testing.T) {
	netCfg, c := testPeerNetwork(t)
	wrong := netCfg.NetworkID()
	wrong[0] ^= 0xff
	cases := []BootstrapManifest{
		{NetworkID: wrong, Peers: []string{"https://wrong.example"}},
		{NetworkID: netCfg.NetworkID(), Peers: []string{"https://expired.example"}, ExpiresAt: time.Now().Add(-time.Minute).Unix()},
	}
	for i, manifest := range cases {
		srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(manifest)
		}))
		netCfg.BootstrapManifests = []string{srv.URL}
		n := NewNode(c, NodeConfig{})
		n.client = srv.Client()
		n.refreshBootstrapManifests(context.Background())
		if got := n.peerList(); len(got) != 0 {
			t.Fatalf("case %d accepted peers: %v", i, got)
		}
		srv.Close()
	}
}

func TestBootstrapManifestMetadataDoesNotChangeNetworkID(t *testing.T) {
	netCfg, _ := testPeerNetwork(t)
	before := netCfg.NetworkID()
	netCfg.BootstrapManifests = []string{"https://example.com/bootstrap.json"}
	if got := netCfg.NetworkID(); got != before {
		t.Fatalf("bootstrap manifest metadata changed network id: before=%s after=%s", before, got)
	}
}