package bridge

import "testing"

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
