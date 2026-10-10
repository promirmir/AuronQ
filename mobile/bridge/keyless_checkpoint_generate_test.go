package bridge

import (
  "encoding/json"
  "math/big"
  "os"
  "path/filepath"
  "sort"
  "strings"
  "testing"
  "time"

  aq "auronq/internal/auronq"
)

// The publisher runs from a GitHub Actions identity proven by Sigstore,
// without any manually provisioned long-lived private signing key.
// The runner independently validates all headers missing since the last
// authenticated checkpoint, then requires corroboration by separate peers.
func TestGenerateKeylessCheckpointEvery256Blocks(t *testing.T) {
  if os.Getenv("AURONQ_GENERATE_KEYLESS_CHECKPOINT")!="1" {t.Skip("only scheduled publisher")}
  previous,err:=reviewedCheckpointCache()
  if err!=nil {t.Fatal(err)}
  priorFile:=os.Getenv("AURONQ_PREVIOUS_KEYLESS_PAYLOAD")
  if priorFile!="" {
    if raw,readErr:=os.ReadFile(priorFile);readErr==nil {
      var p checkpointPayload
      if err:=json.Unmarshal(raw,&p);err!=nil {t.Fatal(err)}
      // CI has ALREADY authenticated the raw payload with a pinned GitHub
      // Actions OIDC identity using cosign before providing this file.
      p.IssuedAt=time.Now().Unix()
      p.ExpiresAt=time.Now().Add(28*24*time.Hour).Unix()
      if err:=checkpointPayloadValid(p,time.Now());err!=nil {
        t.Fatalf("previous keyless checkpoint has invalid header identity: %v",err)
      }
      if p.Cache.VerifiedHeight>=previous.VerifiedHeight {previous=p.Cache}
    }else if !os.IsNotExist(readErr){t.Fatal(readErr)}
  }
  obs,err:=collectNodeObservations("[]")
  if err!=nil || len(obs)<2 {t.Fatalf("no public peer quorum: %v",err)}
  sort.Slice(obs,func(i,j int)bool{return obs[i].Height>obs[j].Height})
  target:=obs[1].Height
  if target<=12 {t.Skip("chain too short")}
  target-=12
  const blocksPerCheckpoint=256
  if target < previous.VerifiedHeight+blocksPerCheckpoint {
    t.Skipf("new checkpoint only after 256 verified blocks: local=%d stable=%d",previous.VerifiedHeight,target)
  }
  var next headerCache
  ok:=false
  for _,p:=range obs {
    if p.Height<target {continue}
    localPath:=filepath.Join(t.TempDir(),"headers.json")
    if err:=saveHeaderCache(localPath,previous);err!=nil {t.Fatal(err)}
    if _,err:=verifyHeaderChain(p.Node,localPath);err!=nil {
      t.Logf("invalid peer chain %s: %v",p.Node,err)
      continue
    }
    verified,err:=loadHeaderCache(localPath)
    if err!=nil {t.Fatal(err)}
    if verified.VerifiedHeight<target {continue}
    work,valid:=new(big.Int).SetString(verified.ChainWork,16)
    if !valid {t.Fatal("invalid PoW cumulative work")}
    var headers []aq.BlockHeader
    var targetTip string
    for _,h:=range verified.History {
      if h.Height>target {work.Sub(work,aq.WorkForTarget(h.Target))}
      if h.Height<=target {headers=append(headers,h);if h.Height==target{targetTip=h.Hash().String()}}
    }
    if targetTip=="" || len(headers)<62 || work.Sign()<=0 {continue}
    next=headerCache{
      Version:headerCacheVersion,NetworkID:mainnetNetworkID,VerifiedHeight:target,
      VerifiedTip:targetTip,ChainWork:work.Text(16),History:headers,
    }
    ok=true
    break
  }
  if !ok {t.Fatal("no publicly reachable peer supplied a verified AQM64 history interval")}
  if err:=checkpointCompatibleWithLocal(previous,next);err!=nil {t.Fatal(err)}
  if err:=checkCheckpointWitnesses(next,previous,"[]");err!=nil {t.Fatal(err)}
  now:=time.Now().UTC()
  payload:=checkpointPayload{
    NetworkID:mainnetNetworkID,GenesisHash:mainnetGenesisHash,
    FixedAnchorHeight:verifiedCheckpointHeight,FixedAnchorTip:verifiedCheckpointTip,
    IssuedAt:now.Unix(),ExpiresAt:now.Add(28*24*time.Hour).Unix(),Cache:next,
  }
  if err:=checkpointPayloadValid(payload,now);err!=nil {t.Fatal(err)}
  body,err:=json.MarshalIndent(payload,"","  ")
  if err!=nil {t.Fatal(err)}
  dest:=os.Getenv("AURONQ_KEYLESS_PAYLOAD_OUTPUT")
  if dest=="" {t.Fatal("output path required")}
  if err:=os.WriteFile(dest,append(body,'\n'),0600);err!=nil {t.Fatal(err)}
  t.Logf("KEYLESS_CHECKPOINT height=%d every=256 signer=GitHubActionsOIDC",next.VerifiedHeight)
}
