package auronq

import (
    "encoding/json"
    "math/big"
    "net/http"
    "net/http/httptest"
    "os"
    "path/filepath"
    "strings"
    "testing"
)

func buildCheckpointChainFixture(t *testing.T) *Chain {
    t.Helper()
    c,_,_:=consensusTestChain(t,3)
    c.mu.Lock()
    defer c.mu.Unlock()
    work:=WorkForTarget(c.blocks[0].Header.Target)
    for i:=uint64(1);i<=PeerCheckpointInterval+1;i++ {
        prev:=c.blocks[len(c.blocks)-1].Header
        h:=BlockHeader{
            Version:BlockVersion,PowAlgo:PowAlgorithmAQM64,
            Height:i,PrevHash:prev.Hash(),
            Timestamp:prev.Timestamp+TargetBlockSeconds,
            Target:prev.Target,
        }
        c.blocks=append(c.blocks,Block{Header:h})
        work.Add(work,WorkForTarget(h.Target))
    }
    c.state.Height=PeerCheckpointInterval+1
    c.state.Tip=c.blocks[len(c.blocks)-1].Header.Hash()
    c.state.ChainWork=work.Text(16)
    return c
}

func TestP2PSignedCheckpointIdentityPersistsAcrossRestarts(t *testing.T) {
    c:=buildCheckpointChainFixture(t)
    n:=NewNode(c,NodeConfig{})
    first,err:=n.SignedCheckpoint()
    if err!=nil{t.Fatal(err)}
    cp,err:=VerifySignedPeerCheckpoint(first,c.NetworkID(),c.network.Genesis.Hash())
    if err!=nil{t.Fatal(err)}
    if cp.Height!=PeerCheckpointInterval || len(cp.Headers)!=peerCheckpointHistory {
        t.Fatalf("unexpected checkpoint height/history %d/%d",cp.Height,len(cp.Headers))
    }
    confirmed,err:=n.persistOwnCheckpoint()
    if err!=nil || !confirmed {t.Fatalf("checkpoint persistence: %v %v",confirmed,err)}
    if _,err:=os.Stat(filepath.Join(c.dir,"peer-checkpoints","latest.json"));err!=nil{t.Fatal(err)}

    restart:=NewNode(c,NodeConfig{})
    next,err:=restart.SignedCheckpoint()
    if err!=nil{t.Fatal(err)}
    if first.PublicKey!=next.PublicKey {
        t.Fatal("node identity changed after restart")
    }
    if first.Signature!=next.Signature {
        t.Fatal("identical checkpoint was not signed deterministically")
    }
    again,err:=restart.persistOwnCheckpoint()
    if err!=nil || again {t.Fatalf("unexpected rewrite of identical checkpoint %v %v",again,err)}
    if work,_:=new(big.Int).SetString(cp.ChainWork,16);work.Sign()<=0{
        t.Fatal("checkpoint work invalid")
    }
}

func TestP2PSignedCheckpointRejectsTamperAndWrongNetwork(t *testing.T) {
    c:=buildCheckpointChainFixture(t)
    signed,err:=NewNode(c,NodeConfig{}).SignedCheckpoint()
    if err!=nil {t.Fatal(err)}
    tampered:=signed
    tampered.Payload=append(json.RawMessage(nil),signed.Payload...)
    tampered.Payload[len(tampered.Payload)-1]^=1
    if _,err:=VerifySignedPeerCheckpoint(tampered,c.NetworkID(),c.network.Genesis.Hash());err==nil{
        t.Fatal("forged payload accepted")
    }
    tampered=signed
    tampered.Signature=strings.Repeat("A",len(signed.Signature))
    if _,err:=VerifySignedPeerCheckpoint(tampered,c.NetworkID(),c.network.Genesis.Hash());err==nil{
        t.Fatal("tampered signature accepted")
    }
    var bad Hash
    if _,err:=VerifySignedPeerCheckpoint(signed,bad,c.network.Genesis.Hash());err==nil{
        t.Fatal("wrong network identity accepted")
    }
}

func TestP2PCheckpointHTTPAvailabilityAndNoConsensusSideEffects(t *testing.T) {
    c:=buildCheckpointChainFixture(t)
    before:=c.State()
    n:=NewNode(c,NodeConfig{EnablePeerCheckpoints:true})
    req:=httptest.NewRequest(http.MethodGet,"/p2p/checkpoint",nil)
    req.RemoteAddr="8.8.8.8:44444"
    rec:=httptest.NewRecorder()
    n.handler().ServeHTTP(rec,req)
    if rec.Code!=http.StatusOK {
        t.Fatalf("checkpoint HTTP %d body %s",rec.Code,rec.Body.String())
    }
    var att SignedPeerCheckpoint
    if err:=json.Unmarshal(rec.Body.Bytes(),&att);err!=nil{t.Fatal(err)}
    if _,err:=VerifySignedPeerCheckpoint(att,c.NetworkID(),c.network.Genesis.Hash());err!=nil{
        t.Fatal(err)
    }
    after:=c.State()
    if before.Height!=after.Height || before.Tip!=after.Tip ||
        before.ChainWork!=after.ChainWork || before.Issued!=after.Issued {
        t.Fatal("checkpoint endpoint modified consensus state")
    }
}

func TestP2PCheckpointCanaryDisabledByDefault(t *testing.T) {
    t.Setenv("AURONQ_ENABLE_P2P_CHECKPOINTS","")
    c:=buildCheckpointChainFixture(t)
    n:=NewNode(c,NodeConfig{})
    if n.cfg.EnablePeerCheckpoints {t.Fatal("pilot checkpoint service unexpectedly enabled")}
    request:=httptest.NewRequest(http.MethodGet,"/p2p/checkpoint",nil)
    request.RemoteAddr="8.8.8.8:44444"
    rec:=httptest.NewRecorder()
    n.handler().ServeHTTP(rec,request)
    if rec.Code!=http.StatusNotFound {
        t.Fatalf("checkpoint endpoint should be disabled: HTTP %d",rec.Code)
    }
    if _,err:=os.Stat(filepath.Join(c.dir,"checkpoint-node-identity.seed"));!os.IsNotExist(err){
        t.Fatalf("disabled checkpoint service must not create a signing key: %v",err)
    }
}

func TestP2PCheckpointRateLimitDoesNotBlockHeaderSynchronization(t *testing.T){
    c:=buildCheckpointChainFixture(t)
    n:=NewNode(c,NodeConfig{EnablePeerCheckpoints:true})
    endpoint:=n.handler()
    for i:=0;i<maxCheckpointRequestsPerIPWindow+1;i++ {
        req:=httptest.NewRequest(http.MethodGet,"/p2p/checkpoint",nil)
        req.RemoteAddr="8.8.8.8:44444"
        rec:=httptest.NewRecorder()
        endpoint.ServeHTTP(rec,req)
        expected:=http.StatusOK
        if i==maxCheckpointRequestsPerIPWindow {expected=http.StatusTooManyRequests}
        if rec.Code!=expected {
            t.Fatalf("unexpected checkpoint response at request %d: %d, expected %d",i,rec.Code,expected)
        }
    }
    req:=httptest.NewRequest(http.MethodGet,"/p2p/headers?start=0&limit=1",nil)
    req.RemoteAddr="8.8.8.8:44444"
    rec:=httptest.NewRecorder()
    endpoint.ServeHTTP(rec,req)
    if rec.Code!=http.StatusOK {
        t.Fatalf("checkpoint rate limit interfered with header/block endpoint: HTTP %d",rec.Code)
    }
}
