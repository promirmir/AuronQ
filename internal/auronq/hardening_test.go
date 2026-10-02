package auronq

import (
	"context"
	"math/big"
	"testing"
	"time"
)

func easyPowForTest(t *testing.T) {
	t.Helper()
	old := PowLimit
	PowLimit = TargetFromBig(new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 512), big.NewInt(1)))
	t.Cleanup(func() { PowLimit = old })
}

func TestDiscoveredPeerNetgroupDiversityCap(t *testing.T) {
	_, c := testPeerNetwork(t)
	n := NewNode(c, NodeConfig{})
	for _, p := range []string{
		"http://8.8.1.1:18444",
		"http://8.8.2.2:18444",
		"http://8.8.3.3:18444",
		"http://8.8.4.4:18444",
		"http://8.8.5.5:18444",
	} {
		n.addDiscoveredPeer(p)
	}
	if got := len(n.peerList()); got != maxPeersPerNetgroup {
		t.Fatalf("same-netgroup peers=%d want=%d", got, maxPeersPerNetgroup)
	}
	n.addDiscoveredPeer("http://9.9.9.9:18444")
	if got := len(n.peerList()); got != maxPeersPerNetgroup+1 {
		t.Fatalf("diverse peer was not admitted, peers=%d", got)
	}
}

func TestReorgRequeuesDisconnectedTransaction(t *testing.T) {
	easyPowForTest(t)
	founderSeed, founderPub, err := GenerateMLDSA87()
	if err != nil {
		t.Fatal(err)
	}
	_, recipientPub, err := GenerateMLDSA87()
	if err != nil {
		t.Fatal(err)
	}
	_, minerAPub, err := GenerateMLDSA87()
	if err != nil {
		t.Fatal(err)
	}
	_, minerBPub, err := GenerateMLDSA87()
	if err != nil {
		t.Fatal(err)
	}
	founder := AddressFromPub(founderPub, SchemeMLDSA87, TestnetNetworkByte)
	recipient := AddressFromPub(recipientPub, SchemeMLDSA87, TestnetNetworkByte)
	minerA := AddressFromPub(minerAPub, SchemeMLDSA87, TestnetNetworkByte)
	minerB := AddressFromPub(minerBPub, SchemeMLDSA87, TestnetNetworkByte)
	g, err := CreateGenesis(founder, TestnetNetworkByte, time.Now().Unix()-20)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	mr, err := MineParallel(ctx, g, 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	g = mr.Block
	netCfg := &NetworkConfig{Name: "reorg-test", ProtocolVersion: 1, NetworkByte: TestnetNetworkByte, CoinbaseMaturity: 10, FounderAddress: founder, Genesis: g}
	mainChain, err := OpenChain(t.TempDir(), netCfg)
	if err != nil {
		t.Fatal(err)
	}
	forkChain, err := OpenChain(t.TempDir(), netCfg)
	if err != nil {
		t.Fatal(err)
	}

	us, err := mainChain.UTXOsForAddress(founder, false)
	if err != nil {
		t.Fatal(err)
	}
	w := &Wallet{File: WalletFile{NetworkByte: TestnetNetworkByte, Scheme: SchemeMLDSA87, Address: founder}, seed: founderSeed, pub: founderPub}
	tx, _, err := w.BuildTransaction(us, recipient, Coin, TestnetNetworkByte)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := mainChain.AddMempool(tx); err != nil {
		t.Fatal(err)
	}
	tpl, err := mainChain.BuildTemplate(minerA)
	if err != nil {
		t.Fatal(err)
	}
	mr, err = MineParallel(ctx, tpl, 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := mainChain.AddBlock(&mr.Block); err != nil {
		t.Fatal(err)
	}
	if mainChain.MempoolSize() != 0 {
		t.Fatal("confirmed transaction remained in mempool")
	}

	for i := 0; i < 2; i++ {
		tpl, err = forkChain.BuildTemplate(minerB)
		if err != nil {
			t.Fatal(err)
		}
		mr, err = MineParallel(ctx, tpl, 1, nil)
		if err != nil {
			t.Fatal(err)
		}
		if err := forkChain.AddBlock(&mr.Block); err != nil {
			t.Fatal(err)
		}
	}
	if err := mainChain.ReplaceCanonical(forkChain.BlocksThrough(2)); err != nil {
		t.Fatal(err)
	}
	if got := mainChain.MempoolSize(); got != 1 {
		t.Fatalf("reorg mempool size=%d want=1", got)
	}
	pending := mainChain.MempoolTransactions(MaxBlockBytes)
	if len(pending) != 1 || pending[0].ID() != tx.ID() {
		t.Fatal("disconnected transaction was not restored to mempool")
	}
}

func TestMempoolByteLimit(t *testing.T) {
	easyPowForTest(t)
	founderSeed, founderPub, err := GenerateMLDSA87()
	if err != nil {
		t.Fatal(err)
	}
	_, recipientPub, err := GenerateMLDSA87()
	if err != nil {
		t.Fatal(err)
	}
	founder := AddressFromPub(founderPub, SchemeMLDSA87, TestnetNetworkByte)
	recipient := AddressFromPub(recipientPub, SchemeMLDSA87, TestnetNetworkByte)
	g, err := CreateGenesis(founder, TestnetNetworkByte, time.Now().Unix()-20)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	mr, err := MineParallel(ctx, g, 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	g = mr.Block
	netCfg := &NetworkConfig{Name: "mempool-limit-test", ProtocolVersion: 1, NetworkByte: TestnetNetworkByte, CoinbaseMaturity: 10, FounderAddress: founder, Genesis: g}
	c, err := OpenChain(t.TempDir(), netCfg)
	if err != nil {
		t.Fatal(err)
	}
	us, err := c.UTXOsForAddress(founder, false)
	if err != nil {
		t.Fatal(err)
	}
	w := &Wallet{File: WalletFile{NetworkByte: TestnetNetworkByte, Scheme: SchemeMLDSA87, Address: founder}, seed: founderSeed, pub: founderPub}
	tx, _, err := w.BuildTransaction(us, recipient, Coin, TestnetNetworkByte)
	if err != nil {
		t.Fatal(err)
	}
	c.mu.Lock()
	c.mempoolBytes = MaxMempoolBytes
	_, err = c.addMempoolLocked(tx)
	c.mu.Unlock()
	if err == nil {
		t.Fatal("transaction accepted beyond mempool byte limit")
	}
}

func TestCandidateEnvelopeRejectsFabricatedTargetBeforeReorg(t *testing.T) {
	_, c := testPeerNetwork(t)
	tpl, err := c.BuildTemplate(c.network.FounderAddress)
	if err != nil {
		t.Fatal(err)
	}
	// Make the block claim more work than consensus allows. The fork-sync
	// preflight must reject it before the value can contribute to branch work.
	x := tpl.Header.Target.Big()
	x.Sub(x, big.NewInt(1))
	tpl.Header.Target = TargetFromBig(x)
	history := c.RecentBlocksThrough(0, DifficultyWindow+1)
	if err := ValidateCandidateEnvelope(&tpl, history, time.Now().Unix()); err == nil {
		t.Fatal("candidate with fabricated target was accepted")
	}
}