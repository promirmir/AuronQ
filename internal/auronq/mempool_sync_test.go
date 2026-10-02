package auronq

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSyncMempoolWhenChainWorkIsEqual(t *testing.T) {
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	mr, err := MineParallel(ctx, g, 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	g = mr.Block
	netCfg := &NetworkConfig{Name: "mempool-sync-test", ProtocolVersion: 1, NetworkByte: TestnetNetworkByte, CoinbaseMaturity: 10, FounderAddress: founder, Genesis: g}
	c1, err := OpenChain(t.TempDir(), netCfg)
	if err != nil {
		t.Fatal(err)
	}
	c2, err := OpenChain(t.TempDir(), netCfg)
	if err != nil {
		t.Fatal(err)
	}

	us, err := c1.UTXOsForAddress(founder, false)
	if err != nil {
		t.Fatal(err)
	}
	w := &Wallet{File: WalletFile{NetworkByte: TestnetNetworkByte, Scheme: SchemeMLDSA87, Address: founder}, seed: founderSeed, pub: founderPub}
	tx, _, err := w.BuildTransaction(us, recipient, Coin, TestnetNetworkByte)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = c1.AddMempool(tx); err != nil {
		t.Fatal(err)
	}

	n1 := NewNode(c1, NodeConfig{})
	srv := httptest.NewServer(n1.handler())
	defer srv.Close()
	n2 := NewNode(c2, NodeConfig{})
	if err = n2.syncPeer(srv.URL); err != nil {
		t.Fatal(err)
	}
	if got := c2.MempoolSize(); got != 1 {
		t.Fatalf("synced mempool size=%d want=1", got)
	}
}
