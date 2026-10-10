package bridge

import (
    "crypto/ed25519"
    "crypto/rand"
    "encoding/base64"
    "encoding/json"
    "strings"
    "testing"
    "time"
)

func TestSignedCheckpointRejectsTamperWrongKeyAndExpiredManifest(t *testing.T) {
    public, private, err := ed25519.GenerateKey(rand.Reader)
    if err!=nil {t.Fatal(err)}
    c,err:=reviewedCheckpointCache()
    if err!=nil {t.Fatal(err)}
    now:=time.Now()
    p:=checkpointPayload{
        NetworkID:mainnetNetworkID,GenesisHash:mainnetGenesisHash,
        FixedAnchorHeight:verifiedCheckpointHeight,FixedAnchorTip:verifiedCheckpointTip,
        IssuedAt:now.Unix(),ExpiresAt:now.Add(48*time.Hour).Unix(),Cache:c,
    }
    payload,err:=json.Marshal(p)
    if err!=nil {t.Fatal(err)}
    signed:=checkpointEnvelope{
        Version:1,Payload:payload,
        Signature:base64.StdEncoding.EncodeToString(ed25519.Sign(private,append([]byte(checkpointSignatureDomain),payload...))),
    }
    data,err:=json.Marshal(signed)
    if err!=nil {t.Fatal(err)}
    if _,err:=verifyCheckpointCertificate(data,public,now);err!=nil {t.Fatal(err)}
    if _,err:=verifyCheckpointCertificate(data,public,now.Add(72*time.Hour));err==nil {
        t.Fatal("expired checkpoint was accepted")
    }
    otherPub,_,err:=ed25519.GenerateKey(rand.Reader)
    if err!=nil {t.Fatal(err)}
    if _,err:=verifyCheckpointCertificate(data,otherPub,now);err==nil {
        t.Fatal("checkpoint signature from another key was accepted")
    }
    forged:=strings.Replace(string(data),mainnetNetworkID,strings.Repeat("f",len(mainnetNetworkID)),1)
    if _,err:=verifyCheckpointCertificate([]byte(forged),public,now);err==nil {
        t.Fatal("tampered certificate was accepted")
    }
    // Valid signature under a key not pinned into the app cannot authorize updates.
    // The signer entrypoint also refuses a mismatched seed.
    seed:=private.Seed()
    if _,err:=SignReviewedCheckpoint([]byte("{}"),base64.StdEncoding.EncodeToString(seed),now);err==nil {
        t.Fatal("untrusted signer key was accepted")
    }
}

func TestSignedCheckpointNeverRollsBackLocallyVerifiedHeaders(t *testing.T) {
    c,err:=reviewedCheckpointCache()
    if err!=nil {t.Fatal(err)}
    newer:=c
    newer.VerifiedHeight=c.VerifiedHeight+3
    if err:=checkpointCompatibleWithLocal(newer,c);err==nil {
        t.Fatal("downgrade was allowed")
    }
    wrong:=c
    wrong.VerifiedTip=strings.Repeat("a",128)
    if err:=checkpointCompatibleWithLocal(c,wrong);err==nil {
        t.Fatal("conflicting same-height tip was accepted")
    }
}

func TestCheckpointNetworkGroupsDoNotDoubleCountSameDNSOperator(t *testing.T) {
    a:=checkpointNetworkGroup("https://mir.taild63f46.ts.net")
    b:=checkpointNetworkGroup("https://desktop-4nifg1j.taild63f46.ts.net")
    if a!=b {t.Fatalf("same DNS domain counted twice: %s %s",a,b)}
    if checkpointNetworkGroup("http://45.88.201.77:18444")==checkpointNetworkGroup("http://54.38.81.30:18444") {
        t.Fatal("separate IP ranges collapsed into one group")
    }
}

func TestCheckpointPayloadMustMatchPinnedGenesisAndWork(t *testing.T) {
    c,err:=reviewedCheckpointCache()
    if err!=nil {t.Fatal(err)}
    now:=time.Now()
    p:=checkpointPayload{
        NetworkID:mainnetNetworkID,GenesisHash:mainnetGenesisHash,
        FixedAnchorHeight:verifiedCheckpointHeight,FixedAnchorTip:verifiedCheckpointTip,
        IssuedAt:now.Unix(),ExpiresAt:now.Add(24*time.Hour).Unix(),Cache:c,
    }
    if err:=checkpointPayloadValid(p,now);err!=nil {t.Fatal(err)}
    p.GenesisHash=strings.Repeat("a",128)
    if err:=checkpointPayloadValid(p,now);err==nil {t.Fatal("wrong genesis accepted")}
    p.GenesisHash=mainnetGenesisHash
    p.Cache.ChainWork="1"
    if err:=checkpointPayloadValid(p,now);err==nil {t.Fatal("fake weak chainwork accepted")}
}
