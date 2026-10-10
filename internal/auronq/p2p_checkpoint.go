package auronq

import (
    "context"
    "crypto/ed25519"
    "crypto/rand"
    "encoding/base64"
    "encoding/hex"
    "encoding/json"
    "errors"
    "fmt"
    "math/big"
    "os"
    "path/filepath"
    "strings"
    "time"
)

const (
    // Advisory only: this is NOT a protocol upgrade or consensus checkpoint.
    PeerCheckpointInterval uint64 = 256
    peerCheckpointHistory = 64 // > DifficultyWindow+1 and MedianTimeWindow
    peerCheckpointDomain = "AURONQ_P2P_ADVISORY_CHECKPOINT_V1\x00"
)

type PeerCheckpoint struct {
    Version uint32 `json:"version"`
    NetworkID Hash `json:"network_id"`
    GenesisHash Hash `json:"genesis_hash"`
    Height uint64 `json:"height"`
    Tip Hash `json:"tip"`
    ChainWork string `json:"chain_work"`
    Headers []BlockHeader `json:"headers"`
}

type SignedPeerCheckpoint struct {
    Version uint32 `json:"version"`
    Payload json.RawMessage `json:"payload"`
    PublicKey string `json:"public_key"`
    Signature string `json:"signature"`
}

// The checkpoint is ALWAYS derived from this node's already FULLY VALIDATED
// canonical blocks and never imported from peers. Full validation continues
// from genesis on full-node restarts; these hints cannot bypass that rule.
// Read lock produces a consistent historic tip even if mining proceeds.
func (c *Chain) LatestPeerCheckpoint() (PeerCheckpoint,error) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    const v=1
    h:=c.state.Height / PeerCheckpointInterval * PeerCheckpointInterval
    if h==0 || h>=uint64(len(c.blocks)) {
        return PeerCheckpoint{},errors.New("not enough validated blocks for first advisory checkpoint")
    }
    work,ok:=new(big.Int).SetString(strings.TrimSpace(c.state.ChainWork),16)
    if !ok || work.Sign()<=0 {
        return PeerCheckpoint{},errors.New("invalid full-node cumulative work")
    }
    // At most 255 later blocks to subtract, never full chain replay here.
    for i:=h+1;i<=c.state.Height;i++ {
        work.Sub(work,WorkForTarget(c.blocks[i].Header.Target))
    }
    if work.Sign()<=0 {return PeerCheckpoint{},errors.New("invalid checkpoint work")}
    start:=uint64(0)
    if h+1>peerCheckpointHistory {start=h+1-peerCheckpointHistory}
    headers:=make([]BlockHeader,0,h-start+1)
    for i:=start;i<=h;i++ {headers=append(headers,c.blocks[i].Header)}
    return PeerCheckpoint{
        Version:v,NetworkID:c.network.NetworkID(),
        GenesisHash:c.network.Genesis.Hash(),Height:h,
        Tip:c.blocks[h].Header.Hash(),ChainWork:work.Text(16),
        Headers:headers,
    },nil
}

// Node-specific identity is generated once locally and kept with node data.
// It belongs to the NODE, not to the operator's wallet or Mainnet consensus.
func (n *Node) privateCheckpointKey() (ed25519.PrivateKey,error) {
    n.checkpointMu.Lock()
    defer n.checkpointMu.Unlock()
    if len(n.checkpointPrivateKey)==ed25519.PrivateKeySize {
        return append(ed25519.PrivateKey(nil),n.checkpointPrivateKey...),nil
    }
    path:=filepath.Join(n.Chain.dir,"checkpoint-node-identity.seed")
    var seed []byte
    existing,err:=os.ReadFile(path)
    switch {
    case err==nil:
        seed=existing
    case errors.Is(err,os.ErrNotExist):
        if err:=os.MkdirAll(filepath.Dir(path),0700);err!=nil{return nil,err}
        seed=make([]byte,ed25519.SeedSize)
        if _,err:=rand.Read(seed);err!=nil{return nil,err}
        // O_EXCL prevents silent overwrite of an existing peer identity.
        f,err:=os.OpenFile(path,os.O_WRONLY|os.O_CREATE|os.O_EXCL,0600)
        if errors.Is(err,os.ErrExist) {
            seed,err=os.ReadFile(path)
            if err!=nil{return nil,err}
        }else if err!=nil{return nil,err
        }else{
            _,writeErr:=f.Write(seed)
            closeErr:=f.Close()
            if writeErr!=nil{return nil,writeErr}
            if closeErr!=nil{return nil,closeErr}
        }
    default:
        return nil,err
    }
    if len(seed)!=ed25519.SeedSize {
        return nil,errors.New("invalid local checkpoint identity seed length")
    }
    private:=ed25519.NewKeyFromSeed(seed)
    n.checkpointPrivateKey=append([]byte(nil),private...)
    return private,nil
}

func (n *Node) SignedCheckpoint()(SignedPeerCheckpoint,error){
    cp,err:=n.Chain.LatestPeerCheckpoint()
    if err!=nil {return SignedPeerCheckpoint{},err}
    priv,err:=n.privateCheckpointKey()
    if err!=nil {return SignedPeerCheckpoint{},fmt.Errorf("checkpoint local signer not available: %w",err)}
    b,err:=json.Marshal(cp)
    if err!=nil{return SignedPeerCheckpoint{},err}
    msg:=append([]byte(peerCheckpointDomain),b...)
    return SignedPeerCheckpoint{
        Version:1,Payload:b,
        PublicKey:hex.EncodeToString(priv.Public().(ed25519.PublicKey)),
        Signature:base64.StdEncoding.EncodeToString(ed25519.Sign(priv,msg)),
    },nil
}

// Checks only signature, local network identity and internal consistency;
// a self-issued signature CANNOT prove that this peer is honest and must
// NEVER be promoted into a wallet's verified headers/UTXO state on its own.
func VerifySignedPeerCheckpoint(att SignedPeerCheckpoint,networkID,genesis Hash)(PeerCheckpoint,error){
    if att.Version!=1 || len(att.Payload)==0 || len(att.Payload)>64<<10 {
        return PeerCheckpoint{},errors.New("invalid signed checkpoint framing")
    }
    key,err:=hex.DecodeString(att.PublicKey)
    if err!=nil || len(key)!=ed25519.PublicKeySize {
        return PeerCheckpoint{},errors.New("invalid peer identity public key")
    }
    sig,err:=base64.StdEncoding.DecodeString(att.Signature)
    if err!=nil || len(sig)!=ed25519.SignatureSize {
        return PeerCheckpoint{},errors.New("invalid peer checkpoint signature encoding")
    }
    if !ed25519.Verify(ed25519.PublicKey(key),append([]byte(peerCheckpointDomain),att.Payload...),sig) {
        return PeerCheckpoint{},errors.New("peer checkpoint signature invalid")
    }
    var cp PeerCheckpoint
    if err:=json.Unmarshal(att.Payload,&cp);err!=nil{return cp,err}
    if cp.Version!=1 || cp.NetworkID!=networkID || cp.GenesisHash!=genesis ||
        cp.Height==0 || cp.Height%PeerCheckpointInterval!=0 {
        return PeerCheckpoint{},errors.New("peer checkpoint has wrong network identity or height")
    }
    n:=len(cp.Headers)
    if n<2 || n>peerCheckpointHistory {
        return PeerCheckpoint{},errors.New("invalid peer checkpoint header window")
    }
    for i,h:=range cp.Headers {
        if i>0 {
            if h.Height!=cp.Headers[i-1].Height+1 || h.PrevHash!=cp.Headers[i-1].Hash() {
                return PeerCheckpoint{},errors.New("broken peer checkpoint header continuity")
            }
        }
    }
    if cp.Headers[n-1].Height!=cp.Height || cp.Headers[n-1].Hash()!=cp.Tip {
        return PeerCheckpoint{},errors.New("peer checkpoint tip differs from included header")
    }
    work,ok:=new(big.Int).SetString(cp.ChainWork,16)
    if !ok || work.Sign()<=0 {
        return PeerCheckpoint{},errors.New("invalid claimed cumulative work")
    }
    return cp,nil
}

func (n *Node) persistOwnCheckpoint()(bool,error){
    att,err:=n.SignedCheckpoint()
    if err!=nil {
        // Before block 256 there is no checkpoint; not an operational error.
        if n.Chain.Height()<PeerCheckpointInterval {return false,nil}
        return false,err
    }
    dir:=filepath.Join(n.Chain.dir,"peer-checkpoints")
    dest:=filepath.Join(dir,"latest.json")
    // Check height to avoid writing unchanged checkpoint every minute.
    if raw,readErr:=os.ReadFile(dest);readErr==nil {
        var old SignedPeerCheckpoint
        if json.Unmarshal(raw,&old)==nil {
            var previous PeerCheckpoint
            if json.Unmarshal(old.Payload,&previous)==nil {
                var current PeerCheckpoint
                if json.Unmarshal(att.Payload,&current)==nil &&
                    previous.Height==current.Height && previous.Tip==current.Tip {
                    return false,nil
                }
            }
        }
    }
    b,err:=json.MarshalIndent(att,"","  ")
    if err!=nil{return false,err}
    if err:=atomicWrite(dest,append(b,'\n'),0600);err!=nil{return false,err}
    return true,nil
}

// This loop is purely auxiliary: failures must not stop mining, block
// validation, peer sync, wallet service, or normal node startup.
func (n *Node) checkpointLoop(ctx context.Context){
    check:=func(){
        _,_ = n.persistOwnCheckpoint()
    }
    check()
    t:=time.NewTicker(60*time.Second)
    defer t.Stop()
    for {
        select {
        case <-ctx.Done(): return
        case <-t.C: check()
        }
    }
}
