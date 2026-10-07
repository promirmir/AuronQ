package auronq

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func survivabilityTestNode(peerCount int) *Node {
	n := &Node{peers: map[string]struct{}{}}
	for i := 0; i < peerCount; i++ {
		// 11/8 is globally routable and gives us many independent-looking
		// netgroups for the diversity ordering used by publicPeerList.
		n.peers[fmt.Sprintf("http://11.%d.%d.%d:18444", (i%200)+1, ((i/200)%200)+1, (i%250)+1)] = struct{}{}
	}
	return n
}

func TestPersistentPeerListIsNotLimitedByGossipWindow(t *testing.T) {
	n := survivabilityTestNode(80)
	if got := len(n.advertisedPeerList()); got != 32 {
		t.Fatalf("gossip list length = %d, want 32", got)
	}
	if got := len(n.persistentPeerList()); got != 80 {
		t.Fatalf("persistent peer list length = %d, want 80", got)
	}
}

func TestAdvertisedPeerListRotatesAcrossKnownPeers(t *testing.T) {
	n := survivabilityTestNode(80)
	first := n.advertisedPeerList()
	second := n.advertisedPeerList()
	if reflect.DeepEqual(first, second) {
		t.Fatal("successive gossip windows are identical; later peers would be starved")
	}
	seen := map[string]bool{}
	for _, p := range first {
		seen[p] = true
	}
	for _, p := range second {
		seen[p] = true
	}
	if len(seen) <= 32 {
		t.Fatalf("rotation exposed only %d unique peers, want more than 32", len(seen))
	}
}

func TestSavePeerStoreRetainsMoreThanHelloLimit(t *testing.T) {
	n := survivabilityTestNode(80)
	n.cfg.PeerStorePath = filepath.Join(t.TempDir(), "peers.json")
	n.savePeerStore()

	b, err := os.ReadFile(n.cfg.PeerStorePath)
	if err != nil {
		t.Fatal(err)
	}
	var peers []string
	if err := json.Unmarshal(b, &peers); err != nil {
		t.Fatal(err)
	}
	if got := len(peers); got != 80 {
		t.Fatalf("saved peer count = %d, want 80", got)
	}
}
