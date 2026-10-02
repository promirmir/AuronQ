package auronq

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"
)

// Regression: an outbound-only node that mined while its public seed was
// temporarily unavailable must repair that seed on a later sync round. rc3
// only pulled stronger remote chains and therefore left the seed behind after
// a missed one-shot block broadcast.
func TestSyncPeerPushesMissingExtensionToBehindPeer(t *testing.T) {
	_, founderPub, err := GenerateMLDSA87()
	if err != nil {
		t.Fatal(err)
	}
	founder := AddressFromPub(founderPub, SchemeMLDSA87, TestnetNetworkByte)
	g, err := CreateGenesis(founder, TestnetNetworkByte, time.Now().Unix()-20)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	minedGenesis, err := MineParallel(ctx, g, 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	g = minedGenesis.Block
	netCfg := &NetworkConfig{Name: "push-sync-test", ProtocolVersion: 1, NetworkByte: TestnetNetworkByte, CoinbaseMaturity: 10, FounderAddress: founder, Genesis: g}

	strong, err := OpenChain(t.TempDir(), netCfg)
	if err != nil {
		t.Fatal(err)
	}
	tpl, err := strong.BuildTemplate(founder)
	if err != nil {
		t.Fatal(err)
	}
	mined, err := MineParallel(ctx, tpl, 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := strong.AddBlock(&mined.Block); err != nil {
		t.Fatal(err)
	}

	behind, err := OpenChain(t.TempDir(), netCfg)
	if err != nil {
		t.Fatal(err)
	}
	remoteNode := NewNode(behind, NodeConfig{})
	srv := httptest.NewServer(remoteNode.handler())
	defer srv.Close()

	outboundNode := NewNode(strong, NodeConfig{})
	if err := outboundNode.syncPeer(srv.URL); err != nil {
		t.Fatal(err)
	}
	if got := behind.State().Height; got != 1 {
		t.Fatalf("behind peer height=%d want=1", got)
	}
	if strong.State().Tip != behind.State().Tip {
		t.Fatalf("tips differ after push repair")
	}
}
