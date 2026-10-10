package bridge

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "sync"
    "time"

    aq "auronq/internal/auronq"
)

type peerCheckpointReport struct {
    Checked int `json:"checked"`
    VerifiedSignatures int `json:"verified_signatures"`
    Netgroups int `json:"netgroups"`
    MatchingHeight uint64 `json:"matching_height"`
    MatchingNetgroups int `json:"matching_netgroups"`
    HintOnly bool `json:"hint_only"`
    HeaderCacheTrustedFromPeers bool `json:"header_cache_trusted_from_peers"`
}

type peerCheckpointObservation struct {
    node string
    group string
    checkpoint aq.PeerCheckpoint
    identity string
    err error
}

func mobileMainnetHash(text string) (aq.Hash,error){
    var hash aq.Hash
    err:=json.Unmarshal([]byte(fmt.Sprintf("%q",text)),&hash)
    return hash,err
}

func fetchSignedCheckpointFromPeer(node string,networkID,genesis aq.Hash)(aq.PeerCheckpoint,string,error){
    var empty aq.PeerCheckpoint
    client:=mobileHTTP(5*time.Second)
    request,err:=http.NewRequest(http.MethodGet,node+"/p2p/checkpoint",nil)
    if err!=nil{return empty,"",err}
    request.Header.Set("User-Agent","AuronQ-Mobile/p2p-advisory-checkpoint-v1")
    response,err:=client.Do(request)
    if err!=nil{return empty,"",err}
    defer response.Body.Close()
    if response.StatusCode!=http.StatusOK {
        return empty,"",fmt.Errorf("peer checkpoint HTTP %d",response.StatusCode)
    }
    var att aq.SignedPeerCheckpoint
    decoder:=json.NewDecoder(io.LimitReader(response.Body,96<<10))
    if err:=decoder.Decode(&att);err!=nil{return empty,"",err}
    cp,err:=aq.VerifySignedPeerCheckpoint(att,networkID,genesis)
    if err!=nil{return empty,"",err}
    return cp,att.PublicKey,nil
}

// PeerCheckpointSummary collects only UNTRUSTED P2P hints authenticated to
// each peer's self-generated identity. It never writes header cache, never
// reads wallets, and cannot authorize balance/history/spending decisions.
func PeerCheckpointSummary(knownNodesJSON string)(string,error){
    candidates:=mobileCandidates(knownNodesJSON)
    if len(candidates)>8{candidates=candidates[:8]}
    netID,err:=mobileMainnetHash(mainnetNetworkID)
    if err!=nil{return "",err}
    genesis,err:=mobileMainnetHash(mainnetGenesisHash)
    if err!=nil{return "",err}
    observed:=make(chan peerCheckpointObservation,len(candidates))
    var wg sync.WaitGroup
    for _,candidate:=range candidates {
        node:=candidate
        wg.Add(1)
        go func(){
            defer wg.Done()
            cp,id,err:=fetchSignedCheckpointFromPeer(node,netID,genesis)
            observed<-peerCheckpointObservation{
                node:node,group:checkpointNetworkGroup(node),
                checkpoint:cp,identity:id,err:err,
            }
        }()
    }
    wg.Wait()
    close(observed)
    // Count by network group AND public node identity, never by number of
    // hostnames. Multiple hosts can be controlled by the same operator.
    signed:=0
    groups:=map[string]bool{}
    snapshots:=map[string]map[string]bool{}
    heights:=map[string]uint64{}
    seenKey:=map[string]bool{}
    for o:=range observed {
        if o.err!=nil || o.group=="" || o.identity=="" {continue}
        signed++
        if seenKey[o.identity]{continue}
        seenKey[o.identity]=true
        groups[o.group]=true
        k:=fmt.Sprintf("%d:%s:%s",o.checkpoint.Height,o.checkpoint.Tip.String(),o.checkpoint.ChainWork)
        if snapshots[k]==nil {snapshots[k]=map[string]bool{};heights[k]=o.checkpoint.Height}
        snapshots[k][o.group]=true
    }
    var bestHeight uint64
    bestGroups:=0
    for key,g:=range snapshots {
        count:=len(g)
        if count>=2 && (heights[key]>bestHeight || (heights[key]==bestHeight && count>bestGroups)) {
            bestHeight=heights[key];bestGroups=count
        }
    }
    out:=peerCheckpointReport{
        Checked:len(candidates),VerifiedSignatures:signed,
        Netgroups:len(groups),MatchingHeight:bestHeight,
        MatchingNetgroups:bestGroups,
        HintOnly:true,HeaderCacheTrustedFromPeers:false,
    }
    data,err:=json.Marshal(out)
    if err!=nil{return "",err}
    return string(data),nil
}
