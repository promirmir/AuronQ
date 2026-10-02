package auronq

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"
)

func mineTestBlocks(t *testing.T, c *Chain, miner string, count int) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for i := 0; i < count; i++ {
		tpl, err := c.BuildTemplate(miner)
		if err != nil {
			t.Fatalf("build template %d: %v", i, err)
		}
		mr, err := MineParallel(ctx, tpl, 1, nil)
		if err != nil {
			t.Fatalf("mine block %d: %v", i, err)
		}
		if err := c.AddBlock(&mr.Block); err != nil {
			t.Fatalf("add block %d: %v", i, err)
		}
	}
}

// This is an in-process distributed-systems regression test. Five independent
// chain directories are split into two partitions that mine competing forks.
// After reconnecting, every node must converge to the higher-work branch and a
// restarted node must reconstruct exactly the same canonical state from disk.
func TestFiveNodePartitionReorgAndRestartConvergence(t *testing.T) {
	easyPowForTest(t)

	_, founderPub, err := GenerateMLDSA87()
	if err != nil {
		t.Fatal(err)
	}
	founder := AddressFromPub(founderPub, SchemeMLDSA87, TestnetNetworkByte)

	g, err := CreateGenesis(founder, TestnetNetworkByte, time.Now().Unix()-20)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	minedGenesis, err := MineParallel(ctx, g, 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	g = minedGenesis.Block

	netCfg := &NetworkConfig{
		Name:             "five-node-resilience",
		ProtocolVersion:  1,
		NetworkByte:      TestnetNetworkByte,
		CoinbaseMaturity: TestnetCoinbaseMaturity,
		FounderAddress:   founder,
		Genesis:          g,
	}

	dirs := make([]string, 5)
	chains := make([]*Chain, 5)
	for i := range chains {
		dirs[i] = t.TempDir()
		chains[i], err = OpenChain(dirs[i], netCfg)
		if err != nil {
			t.Fatalf("open node %d: %v", i, err)
		}
	}

	// Partition A: nodes 0,1,2 see a two-block branch.
	mineTestBlocks(t, chains[0], founder, 2)
	nodeA := NewNode(chains[0], NodeConfig{})
	srvA := httptest.NewServer(nodeA.handler())
	defer srvA.Close()
	for _, i := range []int{1, 2} {
		n := NewNode(chains[i], NodeConfig{})
		if err := n.syncPeer(srvA.URL); err != nil {
			t.Fatalf("partition A sync node %d: %v", i, err)
		}
	}

	// Partition B: nodes 3,4 independently create a longer three-block fork.
	mineTestBlocks(t, chains[3], founder, 3)
	nodeB := NewNode(chains[3], NodeConfig{})
	srvB := httptest.NewServer(nodeB.handler())
	defer srvB.Close()
	n4 := NewNode(chains[4], NodeConfig{})
	if err := n4.syncPeer(srvB.URL); err != nil {
		t.Fatalf("partition B sync: %v", err)
	}

	if chains[0].State().Tip == chains[3].State().Tip {
		t.Fatal("partitions unexpectedly produced identical tips")
	}

	// Heal the partition. Nodes on the shorter branch must reorganize to B.
	for _, i := range []int{0, 1, 2} {
		n := NewNode(chains[i], NodeConfig{})
		if err := n.syncPeer(srvB.URL); err != nil {
			t.Fatalf("heal sync node %d: %v", i, err)
		}
	}

	want := chains[3].State()
	for i, c := range chains {
		got := c.State()
		if got.Height != want.Height || got.Tip != want.Tip || got.ChainWork != want.ChainWork || got.Issued != want.Issued {
			t.Fatalf("node %d did not converge: got=%+v want=%+v", i, got, want)
		}
	}

	// Simulate a process restart. Canonical blocks, not cached snapshots, are
	// replayed from disk and must yield the same state.
	restarted, err := OpenChain(dirs[0], netCfg)
	if err != nil {
		t.Fatalf("restart node 0: %v", err)
	}
	got := restarted.State()
	if got.Height != want.Height || got.Tip != want.Tip || got.ChainWork != want.ChainWork || got.Issued != want.Issued {
		t.Fatalf("restarted node state mismatch: got=%+v want=%+v", got, want)
	}
}

func TestRestartWithoutBootstrapKeepsPersistedPublicPeers(t *testing.T) {
	_, c := testPeerNetwork(t)
	store := t.TempDir() + "/public-peers.json"

	first := NewNode(c, NodeConfig{PeerStorePath: store})
	for _, p := range []string{
		"http://8.8.8.8:18444",
		"http://9.9.9.9:18444",
		"http://1.1.1.1:18444",
	} {
		first.addDiscoveredPeer(p)
	}
	first.savePeerStore()

	// No seed/manual/bootstrap peers are supplied to the restarted node.
	restarted := NewNode(c, NodeConfig{PeerStorePath: store})
	got := restarted.peerList()
	if len(got) != 3 {
		t.Fatalf("persisted peers after bootstrap loss=%v", got)
	}
}

func TestMalformedPeerHelloCannotChangeChain(t *testing.T) {
	_, c := testPeerNetwork(t)
	before := c.State()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/p2p/hello" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"protocol_version":1,"network_id":"not-a-hash","height":999999999,"tip":"bad","chain_work":"ffffffff"}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	n := NewNode(c, NodeConfig{})
	if err := n.syncPeer(srv.URL); err == nil {
		t.Fatal("malformed hello unexpectedly accepted")
	}
	after := c.State()
	if after != before {
		t.Fatalf("malformed peer changed chain state: before=%+v after=%+v", before, after)
	}
}
