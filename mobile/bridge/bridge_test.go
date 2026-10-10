package bridge

import (
    "encoding/json"
    "os"
	"errors"
	"path/filepath"
	"strings"

	aq "auronq/internal/auronq"
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
    cache,err:=reviewedCheckpointCache()
    if err!=nil {t.Fatal(err)}
    path:=filepath.Join(t.TempDir(),"headers","mainnet.json")
    if err:=saveHeaderCache(path,cache);err!=nil {t.Fatal(err)}
    got,err:=loadHeaderCache(path)
    if err!=nil {t.Fatal(err)}
    if got.VerifiedHeight!=cache.VerifiedHeight ||
       got.VerifiedTip!=cache.VerifiedTip ||
       got.ChainWork!=cache.ChainWork || got.NetworkID!=mainnetNetworkID {
        t.Fatalf("checkpoint not preserved got=%+v want=%+v",got,cache)
    }
}
func TestMissingOrOldCacheUsesPinnedCheckpointWithoutFullReplay(t *testing.T) {
    path:=filepath.Join(t.TempDir(),"header-cache.json")
    got,err:=loadHeaderCache(path)
    if err!=nil {t.Fatal(err)}
    if got.VerifiedHeight!=verifiedCheckpointHeight ||
      got.VerifiedTip!=verifiedCheckpointTip ||
      len(got.History)<62 {
        t.Fatalf("cold start must use immutable release checkpoint: %+v",got)
    }
    genesis,err:=freshHeaderCache()
    if err!=nil {t.Fatal(err)}
    if err:=saveHeaderCache(path,genesis);err!=nil {t.Fatal(err)}
    got,err=loadHeaderCache(path)
    if err!=nil {t.Fatal(err)}
    if got.VerifiedHeight!=verifiedCheckpointHeight {
       t.Fatalf("stale genesis cache should be upgraded to checkpoint: %d",got.VerifiedHeight)
    }
}


func TestHistoryFingerprintIgnoresPeerOrdering(t *testing.T) {
	h1 := uint64(10)
	a := []aq.WalletHistoryItem{
		{TXID: "bb", Type: "received", Status: "confirmed", Height: &h1, AmountAtoms: 2},
		{TXID: "aa", Type: "sent", Status: "confirmed", Height: &h1, AmountAtoms: 1},
	}
	b := []aq.WalletHistoryItem{a[1], a[0]}
	if historyFingerprint(a) != historyFingerprint(b) {
		t.Fatalf("equivalent history produced different fingerprints")
	}
}

func TestUTXOFingerprintIgnoresPeerOrdering(t *testing.T) {
	var h1, h2 aq.Hash
	h1[63] = 1
	h2[63] = 2
	a := []aq.UTXORecord{
		{OutPoint: aq.OutPoint{TxID: h2, Index: 1}, UTXO: aq.UTXO{Height: 2}},
		{OutPoint: aq.OutPoint{TxID: h1, Index: 0}, UTXO: aq.UTXO{Height: 1}},
	}
	b := []aq.UTXORecord{a[1], a[0]}
	if utxoFingerprint(a) != utxoFingerprint(b) {
		t.Fatalf("equivalent UTXO sets produced different fingerprints")
	}
}

func TestUTXOFingerprintDetectsStateConflict(t *testing.T) {
	var h aq.Hash
	h[63] = 1
	a := []aq.UTXORecord{{OutPoint: aq.OutPoint{TxID: h, Index: 0}, UTXO: aq.UTXO{Height: 1, Out: aq.TxOutput{Value: 100}}}}
	b := []aq.UTXORecord{{OutPoint: aq.OutPoint{TxID: h, Index: 0}, UTXO: aq.UTXO{Height: 1, Out: aq.TxOutput{Value: 101}}}}
	if utxoFingerprint(a) == utxoFingerprint(b) {
		t.Fatalf("conflicting UTXO sets produced the same fingerprint")
	}
}


func TestHistoryFingerprintIgnoresMempoolPropagationDifference(t *testing.T) {
	h := uint64(10)
	base := []aq.WalletHistoryItem{
		{TXID: "confirmed", Type: "received", Status: "confirmed", Height: &h, AmountAtoms: 5},
	}
	withPending := append([]aq.WalletHistoryItem(nil), base...)
	withPending = append(withPending, aq.WalletHistoryItem{TXID: "pending", Type: "sent", Status: "pending", AmountAtoms: 1})
	if historyFingerprint(base) != historyFingerprint(withPending) {
		t.Fatal("mempool propagation difference changed canonical history fingerprint")
	}
}

func TestMergeHistoryPendingUnionsAcrossAgreeingPeers(t *testing.T) {
	h := uint64(10)
	confirmed := aq.WalletHistoryItem{TXID: "confirmed", Type: "received", Status: "confirmed", Height: &h}
	p1 := aq.WalletHistoryItem{TXID: "p1", Type: "sent", Status: "pending", Timestamp: 10}
	p2 := aq.WalletHistoryItem{TXID: "p2", Type: "received", Status: "pending", Timestamp: 20}
	group := []historyObservation{
		{Node: "a", Items: []aq.WalletHistoryItem{p1, confirmed}},
		{Node: "b", Items: []aq.WalletHistoryItem{p2, confirmed}},
	}
	got := mergeHistoryPending(group)
	if len(got) != 3 || got[0].TXID != "p2" || got[1].TXID != "p1" || got[2].TXID != "confirmed" {
		t.Fatalf("unexpected merged history: %+v", got)
	}
}

func TestCachedTipCheckDoesNotResetForLaggingPeer(t *testing.T) {
	cache, err := freshHeaderCache()
	if err != nil {
		t.Fatal(err)
	}
	cache.VerifiedHeight = 10
	st := aq.Status{Height: 9}
	reset, err := cachedTipCheck(cache, st, nil, nil)
	if err == nil {
		t.Fatal("lagging peer should be rejected")
	}
	if reset {
		t.Fatal("lagging peer must not reset a valid local header cache")
	}
}

func TestCachedTipCheckDoesNotResetOnTransportFailure(t *testing.T) {
	cache, err := freshHeaderCache()
	if err != nil {
		t.Fatal(err)
	}
	st := aq.Status{Height: cache.VerifiedHeight}
	reset, err := cachedTipCheck(cache, st, nil, errors.New("temporary timeout"))
	if err == nil {
		t.Fatal("transport failure should be returned")
	}
	if reset {
		t.Fatal("temporary transport failure must not reset the local header cache")
	}
}

func TestCachedTipCheckResetsOnlyOnConfirmedHistoryMismatch(t *testing.T) {
	cache, err := freshHeaderCache()
	if err != nil {
		t.Fatal(err)
	}
	st := aq.Status{Height: cache.VerifiedHeight}
	other := cache.History[0]
	other.Nonce++
	reset, err := cachedTipCheck(cache, st, []aq.BlockHeader{other}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !reset {
		t.Fatal("confirmed same-height header mismatch must request a cache rebuild")
	}
}

func TestNormalizeMobileNodeAcceptsOnlySafePublicHTTPOrHTTPS(t *testing.T) {
	for _, raw := range []string{
		"https://node.example",
		"https://45.88.201.77:18444",
		"http://45.88.201.77:18444",
		"http://[2606:4700:4700::1111]:18444",
	} {
		if got := normalizeMobileNode(raw); got == "" {
			t.Fatalf("safe public node rejected: %q", raw)
		}
	}
	for _, raw := range []string{
		"http://node.example",
		"http://192.168.1.10:18444",
		"http://10.0.0.2:18444",
		"http://100.64.0.1:18444",
		"http://127.0.0.1:18444",
		"https://localhost:18444",
		"https://node.local:18444",
		"https://203.0.113.5:18444",
	} {
		if got := normalizeMobileNode(raw); got != "" {
			t.Fatalf("unsafe mobile node accepted: %q -> %q", raw, got)
		}
	}
}

func TestBundledMobilePeersIncludeIndependentPublicIPv4Fallbacks(t *testing.T) {
	publicHTTP := 0
	for _, p := range bundledBootstrapPeers {
		if strings.HasPrefix(p, "http://") && normalizeMobileNode(p) != "" {
			publicHTTP++
		}
	}
	if publicHTTP < 3 {
		t.Fatalf("expected at least three public IPv4 fallback peers, got %d: %v", publicHTTP, bundledBootstrapPeers)
	}
}

func TestHeaderSyncProgressFreshCacheAndPersistedHeight(t *testing.T) {
    path := filepath.Join(t.TempDir(), "mainnet.json")
    raw, err := HeaderSyncProgress(path)
    if err != nil {
        t.Fatal(err)
    }
    var progress struct {
        Height uint64 `json:"verified_height"`
        Tip string `json:"verified_tip"`
        Network string `json:"network_id"`
    }
    if err := json.Unmarshal([]byte(raw), &progress); err != nil {
        t.Fatal(err)
    }
    if progress.Height != 0 || progress.Tip != mainnetGenesisHash || progress.Network != mainnetNetworkID {
        t.Fatalf("unexpected fresh progress %+v", progress)
    }
    cache, err := freshHeaderCache()
    if err != nil {
        t.Fatal(err)
    }
    if err := saveHeaderCache(path, cache); err != nil {
        t.Fatal(err)
    }
    raw, err = HeaderSyncProgress(path)
    if err != nil {
        t.Fatal(err)
    }
    if err := json.Unmarshal([]byte(raw), &progress); err != nil {
        t.Fatal(err)
    }
    if progress.Height != 0 || progress.Tip != mainnetGenesisHash {
        t.Fatalf("unexpected persisted progress %+v", progress)
    }
    if _, err := os.Stat(path); err != nil {
        t.Fatalf("progress read unexpectedly removed the cache: %v", err)
    }
}
