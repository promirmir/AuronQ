package auronq

import (
	"context"
	"net/http"
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


func TestTwentyNodeMultiPartitionConvergence(t *testing.T) {
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
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	minedGenesis, err := MineParallel(ctx, g, 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	g = minedGenesis.Block

	netCfg := &NetworkConfig{
		Name:             "twenty-node-chaos",
		ProtocolVersion:  1,
		NetworkByte:      TestnetNetworkByte,
		CoinbaseMaturity: TestnetCoinbaseMaturity,
		FounderAddress:   founder,
		Genesis:          g,
	}

	const nodeCount = 20
	dirs := make([]string, nodeCount)
	chains := make([]*Chain, nodeCount)
	for i := range chains {
		dirs[i] = t.TempDir()
		chains[i], err = OpenChain(dirs[i], netCfg)
		if err != nil {
			t.Fatalf("open node %d: %v", i, err)
		}
	}

	leaders := []int{0, 5, 10, 15}
	miners := make([]string, len(leaders))
	for i := range miners {
		_, pub, err := GenerateMLDSA87()
		if err != nil {
			t.Fatal(err)
		}
		miners[i] = AddressFromPub(pub, SchemeMLDSA87, TestnetNetworkByte)
	}

	servers := make([]*httptest.Server, len(leaders))
	for part, leader := range leaders {
		mineTestBlocks(t, chains[leader], miners[part], part+1)
		n := NewNode(chains[leader], NodeConfig{})
		servers[part] = httptest.NewServer(n.handler())
		defer servers[part].Close()

		for i := leader + 1; i < leader+5; i++ {
			follower := NewNode(chains[i], NodeConfig{})
			if err := follower.syncPeer(servers[part].URL); err != nil {
				t.Fatalf("partition %d sync node %d: %v", part, i, err)
			}
		}
	}

	// Four independently mined partitions must not all share the same tip.
	distinct := map[Hash]struct{}{}
	for _, leader := range leaders {
		distinct[chains[leader].Tip()] = struct{}{}
	}
	if len(distinct) < 4 {
		t.Fatalf("expected four distinct partition tips, got %d", len(distinct))
	}

	// Partition 3 has the most validated cumulative work (four blocks). After
	// healing, every node must independently reorganize/converge to it.
	winner := leaders[len(leaders)-1]
	for i := 0; i < nodeCount; i++ {
		if i >= winner && i < winner+5 {
			continue
		}
		n := NewNode(chains[i], NodeConfig{})
		if err := n.syncPeer(servers[len(servers)-1].URL); err != nil {
			t.Fatalf("heal sync node %d: %v", i, err)
		}
	}

	want := chains[winner].State()
	for i, chain := range chains {
		got := chain.State()
		if got.Height != want.Height || got.Tip != want.Tip || got.ChainWork != want.ChainWork || got.Issued != want.Issued {
			t.Fatalf("node %d did not converge: got=%+v want=%+v", i, got, want)
		}
	}

	// Restart nodes from different former partitions. Canonical state must be
	// reconstructed from block files without any bootstrap peer.
	for _, i := range []int{0, 7, 13, 19} {
		restarted, err := OpenChain(dirs[i], netCfg)
		if err != nil {
			t.Fatalf("restart node %d: %v", i, err)
		}
		got := restarted.State()
		if got.Height != want.Height || got.Tip != want.Tip || got.ChainWork != want.ChainWork || got.Issued != want.Issued {
			t.Fatalf("restarted node %d mismatch: got=%+v want=%+v", i, got, want)
		}
	}
}
