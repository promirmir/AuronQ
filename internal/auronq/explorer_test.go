package auronq

import (
	"strings"
	"testing"
)

func TestAddressFromKeyHashRoundTrip(t *testing.T) {
	kh := KeyHash{1, 2, 3, 4, 5}
	addr := AddressFromKeyHash(kh, SchemeMLDSA87, TestnetNetworkByte)
	net, scheme, got, err := DecodeAddress(addr)
	if err != nil {
		t.Fatal(err)
	}
	if net != TestnetNetworkByte || scheme != SchemeMLDSA87 || got != kh {
		t.Fatalf("address roundtrip mismatch net=%d scheme=%d kh=%x", net, scheme, got)
	}
}

func TestExplorerViewsResolveBlocksAndTransactions(t *testing.T) {
	kh1 := KeyHash{1}
	kh2 := KeyHash{2}
	tx0 := Transaction{
		Version:        TxVersion,
		Coinbase:       true,
		CoinbaseHeight: 0,
		Outputs: []TxOutput{{
			Value:   10 * Coin,
			Scheme:  SchemeMLDSA87,
			KeyHash: kh1,
		}},
	}
	tx1 := Transaction{
		Version: TxVersion,
		Inputs: []TxInput{{
			PrevTx:    tx0.ID(),
			PrevIndex: 0,
			Sequence:  SequenceFinal,
		}},
		Outputs: []TxOutput{{
			Value:   9 * Coin,
			Scheme:  SchemeMLDSA87,
			KeyHash: kh2,
		}},
	}
	b0 := Block{Header: BlockHeader{Version: BlockVersion, PowAlgo: PowAlgorithmAQM64, Height: 0, Timestamp: 100, Target: PowLimit}, Transactions: []Transaction{tx0}}
	b1 := Block{Header: BlockHeader{Version: BlockVersion, PowAlgo: PowAlgorithmAQM64, Height: 1, PrevHash: b0.Hash(), Timestamp: 200, Target: PowLimit}, Transactions: []Transaction{tx1}}
	c := &Chain{
		network: &NetworkConfig{Name: "explorer-test", NetworkByte: TestnetNetworkByte},
		blocks:  []Block{b0, b1},
		state:   ChainState{Height: 1, Tip: b1.Hash()},
		mempool: map[string]MempoolEntry{},
	}

	recent := c.ExplorerRecentBlocks(20)
	if len(recent) != 2 || recent[0].Height != 1 || recent[1].Height != 0 {
		t.Fatalf("unexpected recent blocks: %+v", recent)
	}

	h := uint64(1)
	block, ok, err := c.ExplorerBlock(&h, "")
	if err != nil || !ok {
		t.Fatalf("block lookup failed ok=%v err=%v", ok, err)
	}
	if block.Summary.Hash != b1.Hash().String() || len(block.Transactions) != 1 || block.Transactions[0] != tx1.ID().String() {
		t.Fatalf("unexpected block view: %+v", block)
	}

	byHash, ok, err := c.ExplorerBlock(nil, b0.Hash().String())
	if err != nil || !ok || byHash.Summary.Height != 0 {
		t.Fatalf("hash lookup failed ok=%v err=%v block=%+v", ok, err, byHash)
	}

	view, ok, err := c.ExplorerTransaction(tx1.ID().String())
	if err != nil || !ok {
		t.Fatalf("tx lookup failed ok=%v err=%v", ok, err)
	}
	if view.Height == nil || *view.Height != 1 || view.Confirmations != 1 {
		t.Fatalf("unexpected tx location: %+v", view)
	}
	if view.FeeAtoms != Coin {
		t.Fatalf("fee=%s want 1 AURQ", view.Fee)
	}
	if len(view.Inputs) != 1 || view.Inputs[0].Address != AddressFromKeyHash(kh1, SchemeMLDSA87, TestnetNetworkByte) {
		t.Fatalf("input not resolved: %+v", view.Inputs)
	}
	if len(view.Outputs) != 1 || view.Outputs[0].Address != AddressFromKeyHash(kh2, SchemeMLDSA87, TestnetNetworkByte) {
		t.Fatalf("output not resolved: %+v", view.Outputs)
	}
}

func TestExplorerRejectsInvalidIdentifiers(t *testing.T) {
	c := &Chain{}
	if _, _, err := c.ExplorerTransaction("not-a-txid"); err == nil {
		t.Fatal("invalid txid accepted")
	}
	if _, _, err := c.ExplorerBlock(nil, "not-a-hash"); err == nil {
		t.Fatal("invalid block hash accepted")
	}
}

func TestExplorerHTMLWiresLiveHeight(t *testing.T) {
	if !strings.Contains(explorerIndexHTML, `id="height"`) {
		t.Fatal("explorer height element missing")
	}
	if !strings.Contains(explorerIndexHTML, `$('height')`) {
		t.Fatal("explorer height script binding missing")
	}
	if !strings.Contains(explorerIndexHTML, "/v1/explorer/blocks") {
		t.Fatal("explorer block API binding missing")
	}
}
