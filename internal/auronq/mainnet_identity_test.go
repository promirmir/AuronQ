package auronq

import (
	"path/filepath"
	"testing"
)

// Mainnet identity is frozen. This test is intentionally exact: discovery,
// UI, Android and test-harness work must never silently create a new network.
func TestMainnetIdentityIsFrozen(t *testing.T) {
	n, err := LoadNetwork(filepath.Join("..", "..", "network.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateGenesis(n); err != nil {
		t.Fatalf("mainnet genesis no longer validates: %v", err)
	}

	const wantNetworkID = "44e62c2ace002a6660c14e252173c1aa303529c68e40c998e92da2b453f44f30b1e58c94d533587e2186004593fb856c433fcdb5418ed430ec8617e29529365c"
	const wantGenesis = "5750a455c04bfe93c9edfef1a12744b05e29ac6da1a9dd5b790566629dea2080c581beb2f0324efba2067c9efb113ed7a29598fd3ffbed965ced31f265d0cec4"
	const wantFounder = "aurq1keaacmqgcvostprfhfejx55noyd746frjzv3qusb62lmutxurypahx7yiqq3mbg2lq"

	if got := n.NetworkID().String(); got != wantNetworkID {
		t.Fatalf("MAINNET NETWORK ID CHANGED: got %s want %s", got, wantNetworkID)
	}
	if got := n.Genesis.Hash().String(); got != wantGenesis {
		t.Fatalf("MAINNET GENESIS CHANGED: got %s want %s", got, wantGenesis)
	}
	if n.FounderAddress != wantFounder {
		t.Fatalf("MAINNET FOUNDER ADDRESS CHANGED: got %s want %s", n.FounderAddress, wantFounder)
	}
	if n.ProtocolVersion != 1 {
		t.Fatalf("mainnet protocol version changed: %d", n.ProtocolVersion)
	}
	if n.NetworkByte != MainnetNetworkByte {
		t.Fatalf("mainnet network byte changed: %d", n.NetworkByte)
	}
	if n.Maturity() != MainnetCoinbaseMaturity {
		t.Fatalf("mainnet coinbase maturity changed: %d", n.Maturity())
	}
	if len(n.Genesis.Transactions) != 1 || len(n.Genesis.Transactions[0].Outputs) != 1 {
		t.Fatal("mainnet genesis allocation structure changed")
	}
	if got := n.Genesis.Transactions[0].Outputs[0].Value; got != FounderAtoms {
		t.Fatalf("mainnet founder allocation changed: %d", got)
	}
}
