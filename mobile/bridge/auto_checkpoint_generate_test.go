package bridge

import (
    "os"
    "path/filepath"
    "sort"
    "strings"
    "testing"
    "time"
    "math/big"
    "encoding/json"

    aq "auronq/internal/auronq"
)

// Invoked only by the scheduled checkpoint publisher, not normal unit tests.
// The runner performs exactly the same PoW+difficulty verification as a phone,
// but advances from the previous Ed25519-verified checkpoint, if available.
// It never signs only a claim made by a remote node.
func TestGenerateScheduledSignedCheckpoint(t *testing.T) {
    if os.Getenv("AURONQ_GENERATE_SIGNED_CHECKPOINT")!="1" {
        t.Skip("requires the scheduled checkpoint publisher")
    }
    signingSeed:=os.Getenv("AURONQ_CP_ED25519_PRIVATE_KEY")
    if strings.TrimSpace(signingSeed)=="" {
        t.Fatal("publisher signing key secret has not been provisioned")
    }

    prior,err:=reviewedCheckpointCache()
    if err!=nil {t.Fatal(err)}
    path:=os.Getenv("AURONQ_PREVIOUS_SIGNED_MANIFEST")
    if path!="" {
        if raw,readErr:=os.ReadFile(path);readErr==nil {
            c,verifyErr:=ReadSignedCheckpointForCI(raw)
            if verifyErr!=nil {
                t.Fatalf("previous signed certificate invalid; refusing publication: %v",verifyErr)
            }
            if c.VerifiedHeight>=prior.VerifiedHeight {prior=c}
        } else if !os.IsNotExist(readErr) {
            t.Fatalf("could not read existing signed manifest: %v",readErr)
        }
    }
    observations,err:=collectNodeObservations("[]")
    if err!=nil {t.Fatal(err)}
    if len(observations)<2 {t.Fatal("need two publicly reachable independent peer observations")}
    sort.Slice(observations,func(i,j int)bool{
        return observations[i].Height>observations[j].Height
    })
    // Stay twelve blocks behind the second-highest reachable peer.
    // Do not sign speculative tip history.
    anchor:=observations[1].Height
    if anchor<=12 {t.Skip("chain not ready")}
    anchor-=12
    if anchor<=prior.VerifiedHeight+24 {
        t.Skipf("already sufficiently fresh: verified=%d target=%d",prior.VerifiedHeight,anchor)
    }
    var next headerCache
    var validated bool
    for _,o:=range observations {
        if o.Height<anchor {continue}
        cachePath:=filepath.Join(t.TempDir(),"verified.json")
        if err:=saveHeaderCache(cachePath,prior);err!=nil {t.Fatal(err)}
        _,err:=verifyHeaderChain(o.Node,cachePath)
        if err!=nil {
            t.Logf("peer rejected for independent header validation: %s: %v",o.Node,err)
            continue
        }
        verified,err:=loadHeaderCache(cachePath)
        if err!=nil {t.Fatal(err)}
        // We retain at most 128 past headers; never invent skipped work.
        var tip *aq.BlockHeader
        for i:=range verified.History {
            if verified.History[i].Height==anchor {
                tip=&verified.History[i]
                break
            }
        }
        if tip==nil || verified.VerifiedHeight<anchor {continue}
        var history []aq.BlockHeader
        work,ok:=new(big.Int).SetString(verified.ChainWork,16)
        if !ok {t.Fatal("bad accumulated verified work")}
        for _,h:=range verified.History {
            if h.Height>anchor {
                work.Sub(work,aq.WorkForTarget(h.Target))
            } else {
                history=append(history,h)
            }
        }
        if len(history)<62 || work.Sign()<=0 {continue}
        next=headerCache{
            Version:headerCacheVersion,NetworkID:mainnetNetworkID,
            VerifiedHeight:anchor,VerifiedTip:tip.Hash().String(),
            ChainWork:work.Text(16),History:history,
        }
        validated=true
        break
    }
    if !validated {t.Fatal("no peer provided the complete, consensus-verified interval")}
    if err:=checkpointCompatibleWithLocal(prior,next);err!=nil {
        t.Fatalf("checkpoint would downgrade or conflict with previous verified state: %v",err)
    }
    if err:=checkCheckpointWitnesses(next,prior,"[]");err!=nil {
        t.Fatalf("new checkpoint not corroborated by independent peers: %v",err)
    }
    payload,err:=json.Marshal(next)
    if err!=nil {t.Fatal(err)}
    signed,err:=SignReviewedCheckpoint(payload,signingSeed,time.Now())
    if err!=nil {t.Fatal(err)}
    // Strong self-check using the public key pinned into all APKs.
    roundtrip,err:=ReadSignedCheckpointForCI(signed)
    if err!=nil || roundtrip.VerifiedTip!=next.VerifiedTip {
        t.Fatalf("signed checkpoint failed self-verification: %v",err)
    }
    destination:=os.Getenv("AURONQ_CHECKPOINT_OUTPUT_PATH")
    if destination=="" {t.Fatal("no output path")}
    if err:=os.WriteFile(destination,append(signed,'\n'),0600);err!=nil {t.Fatal(err)}
    t.Logf("SIGNED_AURONQ_CHECKPOINT verified_height=%d signed_date=%s witnessed_in_two_netgroups=true",
        next.VerifiedHeight,time.Now().UTC().Format(time.RFC3339))
}
