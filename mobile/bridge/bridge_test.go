package bridge

import (
	"path/filepath"
	"testing"
)

func TestChooseNodeQuorumPrefersPeerAgreementOverLoneHigherClaim(t *testing.T) {
	obs := []mobileNodeObservation{
		{Node: "https://a.example", Height: 100, Tip: "aa", ChainWork: "1000"},
		{Node: "https://b.example", Height: 100, Tip: "aa", ChainWork: "1000"},
		{Node: "https://evil.example", Height: 999999, Tip: "ff", ChainWork: "999999999999"},
	}
	selected, agreeing, err := chooseNodeQuorum(obs)
	if err != nil {
		t.Fatal(err)
	}
	if len(agreeing) != 2 || selected.Height != 100 || selected.Tip != "aa" {
		t.Fatalf("quorum selection=%+v agreeing=%v", selected, agreeing)
	}
}

func TestChooseNodeQuorumTieUsesGreaterReportedWorkButIsNotAgreement(t *testing.T) {
	obs := []mobileNodeObservation{
		{Node: "https://a.example", Height: 100, Tip: "aa", ChainWork: "1000"},
		{Node: "https://b.example", Height: 101, Tip: "bb", ChainWork: "1001"},
	}
	selected, agreeing, err := chooseNodeQuorum(obs)
	if err != nil {
		t.Fatal(err)
	}
	if selected.Node != "https://b.example" || len(agreeing) != 1 {
		t.Fatalf("tie selection=%+v agreeing=%v", selected, agreeing)
	}
}

func TestMobileCandidatesPreferLearnedPeersButKeepBundledFallbacks(t *testing.T) {
	got := mobileCandidates(`["https://learned-a.example","https://learned-b.example","https://learned-a.example","http://unsafe.example"]`)
	if len(got) < 4 {
		t.Fatalf("candidates=%v", got)
	}
	if got[0] != "https://learned-a.example" || got[1] != "https://learned-b.example" {
		t.Fatalf("learned peers not preferred: %v", got)
	}
	seen := map[string]bool{}
	for _, p := range got {
		if seen[p] {
			t.Fatalf("duplicate candidate %q in %v", p, got)
		}
		seen[p] = true
	}
	for _, p := range bundledBootstrapPeers {
		if !seen[p] {
			t.Fatalf("bundled fallback %q missing from %v", p, got)
		}
	}
}


func TestChainWorkComparisonUsesHexEncoding(t *testing.T) {
	// Hex 0x10 (16) must outrank hex 0x0f (15). Parsing as decimal would
	// reject "f" and fall back to lexicographic ordering.
	if got := chainWorkCmp("0f", "10"); got >= 0 {
		t.Fatalf("hex chain work comparison returned %d", got)
	}
	if got := chainWorkCmp("ff", "100"); got >= 0 {
		t.Fatalf("hex chain work comparison returned %d", got)
	}
}

func TestFreshHeaderCacheAnchorsExactMainnetGenesis(t *testing.T) {
	cache, err := freshHeaderCache()
	if err != nil {
		t.Fatal(err)
	}
	if cache.VerifiedHeight != 0 || cache.VerifiedTip != mainnetGenesisHash || len(cache.History) != 1 {
		t.Fatalf("unexpected genesis cache: %+v", cache)
	}
	if cache.History[0].Hash().String() != mainnetGenesisHash {
		t.Fatalf("embedded genesis hash=%s", cache.History[0].Hash().String())
	}
}

func TestHeaderCacheRoundTripPreservesVerifiedAnchor(t *testing.T) {
	cache, err := freshHeaderCache()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "headers", "mainnet.json")
	if err := saveHeaderCache(path, cache); err != nil {
		t.Fatal(err)
	}
	got, err := loadHeaderCache(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.VerifiedTip != cache.VerifiedTip || got.ChainWork != cache.ChainWork || got.NetworkID != mainnetNetworkID {
		t.Fatalf("cache mismatch: got=%+v want=%+v", got, cache)
	}
}
