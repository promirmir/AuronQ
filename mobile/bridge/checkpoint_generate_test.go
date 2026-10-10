package bridge

import (
    "encoding/json"
    "errors"
    "fmt"
    "math/big"
    "os"
    "sort"
    "strings"
    "testing"

    aq "auronq/internal/auronq"
)

// Manual release-engineering task. Disabled for regular Go tests.
// This verifies AQM64/difficulty from genesis once on a CI runner,
// then requires another public node to serve the identical anchor.
// Generated data is reviewed and baked into the application. Runtime
// does not download or silently trust a mutable bootstrap checkpoint.
func TestGenerateReviewedMobileCheckpoint(t *testing.T) {
    if os.Getenv("AURONQ_GENERATE_CHECKPOINT") != "1" {
        t.Skip("manual checkpoint generation only")
    }
    candidates := mobileCandidates("[]")
    var online []mobileNodeObservation
    for _, node := range candidates {
        st, err := statusFromNode(node)
        if err != nil {
            t.Logf("peer unavailable: %s: %v", node, err)
            continue
        }
        online = append(online, mobileNodeObservation{
            Node: node, Height: st.Height, Tip: st.Tip.String(),
            ChainWork: st.ChainWork, Peers: st.Peers,
        })
    }
    if len(online) < 2 {
        t.Fatalf("need at least two public peers for a reviewed checkpoint; got %d", len(online))
    }
    sort.Slice(online, func(i,j int) bool { return online[i].Height > online[j].Height })
    anchorHeight := online[1].Height
    if anchorHeight > 3 {anchorHeight -= 3}
    if anchorHeight < 64 {
        t.Fatal("not enough chain history for meaningful checkpoint")
    }
    // The checkpoint is intentionally conservative and not at an
    // unconfirmed current chain tip.
    var validated headerCache
    var chosen string
    for _, o := range online {
        if o.Height < anchorHeight {continue}
        temp := t.TempDir()+"/headers.json"
        v, err := verifyHeaderChain(o.Node, temp)
        if err != nil {
            t.Logf("full header validation from genesis failed via %s: %v", o.Node, err)
            continue
        }
        cached, err := loadHeaderCache(temp)
        if err != nil {t.Fatal(err)}
        if v.Height != cached.VerifiedHeight {t.Fatal("verified height mismatch")}
        validated = cached
        chosen = o.Node
        break
    }
    if chosen == "" {t.Fatal("no reachable peer provided fully AQM64-validated headers from genesis")}
    if validated.VerifiedHeight < anchorHeight {t.Fatal("validated peer is behind checkpoint")}
    var at *aq.BlockHeader
    for i := range validated.History {
        if validated.History[i].Height == anchorHeight {
            at = &validated.History[i]
            break
        }
    }
    if at == nil {t.Fatal("checkpoint outside verified header history")}
    confirmed := 0
    for _, o := range online {
        if o.Node == chosen || o.Height < anchorHeight {continue}
        hs, err := fetchHeaderBatch(o.Node,anchorHeight,1)
        if err == nil && len(hs)==1 && hs[0].Hash() == at.Hash() {
            confirmed++
            t.Logf("independent peer agreeing: %s",o.Node)
        }
    }
    if confirmed == 0 {
        t.Fatal("no second independent peer agrees on fully verified checkpoint header")
    }

    history := make([]aq.BlockHeader,0,headerCacheKeep)
    for _, h := range validated.History {
        if h.Height <= anchorHeight {history=append(history,h)}
    }
    if len(history)<62 {t.Fatalf("need >=62 preceding headers for LWMA: got %d",len(history))}
    // Verified cumulative work is anchored at the selected height, not
    // the remote node's newer tip.
    work, ok := new(big.Int).SetString(strings.TrimSpace(validated.ChainWork),16)
    if !ok {t.Fatal("bad cumulative work")}
    for _, h := range validated.History {
        if h.Height>anchorHeight {work.Sub(work,aq.WorkForTarget(h.Target))}
    }
    if work.Sign()<=0 {t.Fatal("invalid checkpoint cumulative work")}
    cp := headerCache{
        Version: headerCacheVersion, NetworkID: mainnetNetworkID,
        VerifiedHeight: anchorHeight, VerifiedTip: at.Hash().String(),
        ChainWork: work.Text(16), History: history,
    }
    b, err := json.MarshalIndent(cp,"","  ")
    if err != nil {t.Fatal(err)}
    // Do not allow branch CI to produce an unverified or malformed anchor.
    if cp.History[len(cp.History)-1].Hash().String()!=cp.VerifiedTip {
        t.Fatal(errors.New("malformed checkpoint"))
    }
    if err := os.WriteFile("verified-checkpoint.json",append(b,'\n'),0644);err!=nil {t.Fatal(err)}
    t.Log(fmt.Sprintf("REVIEWED_CHECKPOINT height=%d tip=%s confirmed_by=%d second_nodes anchor=%s",cp.VerifiedHeight,cp.VerifiedTip,confirmed,chosen))
}
