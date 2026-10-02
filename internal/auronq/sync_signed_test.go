package auronq

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"
)

// Regression: a fresh node must be able to download and validate a chain that
// contains an ML-DSA-signed non-coinbase transaction. This is the path that
// failed on Windows versions without native CNG ML-DSA support.
func TestSyncChainContainingSignedMLDSATransaction(t *testing.T) {
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
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	mr, err := MineParallel(ctx, g, 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	g = mr.Block
	netCfg := &NetworkConfig{Name: "sync-signed-test", ProtocolVersion: 1, NetworkByte: TestnetNetworkByte, CoinbaseMaturity: 10, FounderAddress: founder, Genesis: g}
	c1, err := OpenChain(t.TempDir(), netCfg)
	if err != nil {
		t.Fatal(err)
	}
	us, err := c1.UTXOsForAddress(founder, false)
	if err != nil {
		t.Fatal(err)
	}
	w := &Wallet{File: WalletFile{NetworkByte: TestnetNetworkByte, Scheme: SchemeMLDSA87, Address: founder}, seed: founderSeed, pub: founderPub}
	amount := uint64(10) * Coin
	tx, _, err := w.BuildTransaction(us, recipient, amount, TestnetNetworkByte)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = c1.AddMempool(tx); err != nil {
		t.Fatal(err)
	}
	tpl, err := c1.BuildTemplate(founder)
	if err != nil {
		t.Fatal(err)
	}
	mr, err = MineParallel(ctx, tpl, 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err = c1.AddBlock(&mr.Block); err != nil {
		t.Fatal(err)
	}

	n1 := NewNode(c1, NodeConfig{})
	srv := httptest.NewServer(n1.handler())
	defer srv.Close()
	c2, err := OpenChain(t.TempDir(), netCfg)
	if err != nil {
		t.Fatal(err)
	}
	n2 := NewNode(c2, NodeConfig{})
	if err = n2.syncPeer(srv.URL); err != nil {
		t.Fatal(err)
	}
	if got := c2.State().Height; got != 1 {
		t.Fatalf("synced height=%d want=1", got)
	}
	sp, total, err := c2.Balance(recipient)
	if err != nil {
		t.Fatal(err)
	}
	if sp != amount || total != amount {
		t.Fatalf("synced recipient balance=%s/%s", FormatAmount(sp), FormatAmount(total))
	}
}
