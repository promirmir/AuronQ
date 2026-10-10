package bridge

import (
    "crypto/ed25519"
    "encoding/hex"
    "encoding/base64"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "math/big"
    "net"
    "net/http"
    "net/url"
    "os"
    "path/filepath"
    "strings"
    "time"

)

// The release-pinned key is independent of both GitHub TLS and the Mainnet's
// consensus identity. The private signing seed MUST NOT be committed to GitHub.
const checkpointSignerPublicHex = "488a16b26e8643f2f450fd90a98969f1a3ca7eb428ae661332f11a4b1f208806"
const checkpointManifestURL = "https://raw.githubusercontent.com/promirmir/AuronQ/automation/mobile-checkpoints/latest.json"
const checkpointSignatureDomain = "AURONQ_MOBILE_CHECKPOINT_V1\x00"
const checkpointMaxAge = 35 * 24 * time.Hour

type checkpointPayload struct {
    NetworkID string `json:"network_id"`
    GenesisHash string `json:"genesis_hash"`
    FixedAnchorHeight uint64 `json:"fixed_anchor_height"`
    FixedAnchorTip string `json:"fixed_anchor_tip"`
    IssuedAt int64 `json:"issued_at"`
    ExpiresAt int64 `json:"expires_at"`
    Cache headerCache `json:"checkpoint"`
}

type checkpointEnvelope struct {
    Version int `json:"version"`
    Payload json.RawMessage `json:"payload"`
    Signature string `json:"signature"`
}

func checkpointPayloadValid(p checkpointPayload, now time.Time) error {
    c:=p.Cache
    if p.NetworkID!=mainnetNetworkID || p.GenesisHash!=mainnetGenesisHash ||
        p.FixedAnchorHeight!=verifiedCheckpointHeight || p.FixedAnchorTip!=verifiedCheckpointTip {
        return errors.New("checkpoint certificate is not anchored to this AuronQ Mainnet release")
    }
    if p.IssuedAt<=0 || p.ExpiresAt<=p.IssuedAt || p.IssuedAt>now.Unix()+3600 ||
        p.ExpiresAt<now.Unix() || p.ExpiresAt-p.IssuedAt>int64(checkpointMaxAge.Seconds()) {
        return errors.New("checkpoint certificate is expired, future-dated or too long-lived")
    }
    if c.Version!=headerCacheVersion || c.NetworkID!=mainnetNetworkID ||
        c.VerifiedHeight<verifiedCheckpointHeight || c.VerifiedTip=="" {
        return errors.New("checkpoint header cache has incompatible identity")
    }
    if len(c.History)<62 || len(c.History)>headerCacheKeep {
        return errors.New("checkpoint lacks bounded consensus difficulty history")
    }
    if c.History[len(c.History)-1].Height!=c.VerifiedHeight ||
        c.History[len(c.History)-1].Hash().String()!=c.VerifiedTip {
        return errors.New("checkpoint tip does not match its final header")
    }
    for i:=1;i<len(c.History);i++ {
        if c.History[i].Height!=c.History[i-1].Height+1 ||
            c.History[i].PrevHash!=c.History[i-1].Hash() {
            return errors.New("checkpoint historical headers are not contiguous")
        }
    }
    work,ok:=new(big.Int).SetString(strings.TrimSpace(c.ChainWork),16)
    if !ok || work.Sign()<=0 || chainWorkCmp(c.ChainWork,verifiedCheckpointWork)<0 {
        return errors.New("checkpoint cumulative work is invalid or below the fixed anchor")
    }
    if c.VerifiedHeight==verifiedCheckpointHeight &&
       (c.VerifiedTip!=verifiedCheckpointTip || !strings.EqualFold(c.ChainWork,verifiedCheckpointWork)) {
        return errors.New("checkpoint attempts to replace the immutable release anchor")
    }
    // For a recent anchored history the fixed-height header must match exactly.
    for _,h:=range c.History {
        if h.Height==verifiedCheckpointHeight && h.Hash().String()!=verifiedCheckpointTip {
            return errors.New("checkpoint conflicts with immutable release history")
        }
    }
    return nil
}

func verifyCheckpointCertificate(raw []byte, publicKey ed25519.PublicKey, now time.Time) (checkpointPayload,error) {
    var e checkpointEnvelope
    if len(raw)==0 || len(raw)>256<<10 {
        return checkpointPayload{},errors.New("checkpoint manifest exceeds size limit")
    }
    if err:=json.Unmarshal(raw,&e);err!=nil{return checkpointPayload{},err}
    if e.Version!=1 || len(e.Payload)==0 || len(publicKey)!=ed25519.PublicKeySize {
        return checkpointPayload{},errors.New("unsupported checkpoint certificate version")
    }
    sig,err:=base64.StdEncoding.DecodeString(e.Signature)
    if err!=nil || len(sig)!=ed25519.SignatureSize {
        return checkpointPayload{},errors.New("invalid checkpoint signature encoding")
    }
    // Verify the exact payload bytes and a domain prefix before parsing any
    // untrusted network-supplied checkpoint fields.
    message:=append([]byte(checkpointSignatureDomain),e.Payload...)
    if !ed25519.Verify(publicKey,message,sig) {
        return checkpointPayload{},errors.New("checkpoint Ed25519 signature verification failed")
    }
    var payload checkpointPayload
    if err:=json.Unmarshal(e.Payload,&payload);err!=nil {return checkpointPayload{},err}
    if err:=checkpointPayloadValid(payload,now);err!=nil{return checkpointPayload{},err}
    return payload,nil
}

// SignReviewedCheckpoint is used only on the publishing runner. Secret material
// is read from the runner environment and never stored in source, app or logs.
func SignReviewedCheckpoint(cacheJSON []byte, base64PrivateSeed string, now time.Time) ([]byte,error) {
    seed,err:=base64.StdEncoding.DecodeString(strings.TrimSpace(base64PrivateSeed))
    if err!=nil || len(seed)!=ed25519.SeedSize {
        return nil,errors.New("signing seed is missing or malformed (32-byte base64)")
    }
    key:=ed25519.NewKeyFromSeed(seed)
    expected,err:=hex.DecodeString(checkpointSignerPublicHex)
    if err!=nil {return nil,err}
    if hex.EncodeToString(key.Public().(ed25519.PublicKey)) != hex.EncodeToString(expected) {
        return nil,errors.New("signing seed does not match the public key pinned into AuronQ Mobile")
    }
    var cache headerCache
    if err:=json.Unmarshal(cacheJSON,&cache);err!=nil {return nil,err}
    p:=checkpointPayload{
        NetworkID:mainnetNetworkID, GenesisHash:mainnetGenesisHash,
        FixedAnchorHeight:verifiedCheckpointHeight, FixedAnchorTip:verifiedCheckpointTip,
        IssuedAt:now.Unix(),ExpiresAt:now.Add(28*24*time.Hour).Unix(),Cache:cache,
    }
    if err:=checkpointPayloadValid(p,now);err!=nil {return nil,err}
    canonical,err:=json.Marshal(p)
    if err!=nil {return nil,err}
    sig:=ed25519.Sign(key,append([]byte(checkpointSignatureDomain),canonical...))
    return json.MarshalIndent(checkpointEnvelope{
        Version:1,Payload:canonical,Signature:base64.StdEncoding.EncodeToString(sig),
    },"","  ")
}

func fetchSignedCheckpoint() ([]byte,error) {
    request,err:=http.NewRequest(http.MethodGet,checkpointManifestURL,nil)
    if err!=nil{return nil,err}
    request.Header.Set("User-Agent","AuronQ-Mobile/checkpoint-v1")
    cl:=mobileHTTP(7*time.Second) // redirects disabled
    response,err:=cl.Do(request)
    if err!=nil{return nil,err}
    defer response.Body.Close()
    if response.StatusCode!=http.StatusOK {
        return nil,fmt.Errorf("signed checkpoint HTTP status: %d",response.StatusCode)
    }
    return io.ReadAll(io.LimitReader(response.Body,(256<<10)+1))
}

func checkpointNetworkGroup(node string) string {
    u,err:=url.Parse(node)
    if err!=nil{return ""}
    host:=strings.ToLower(u.Hostname())
    if ip:=net.ParseIP(host);ip!=nil {
        if v:=ip.To4();v!=nil{return fmt.Sprintf("ip:%d.%d",v[0],v[1])}
        return "ip6:"+ip.String()
    }
    parts:=strings.Split(host,".")
    if len(parts)>=3 {return "dns:"+strings.Join(parts[len(parts)-3:],".")}
    return "dns:"+host
}

func checkCheckpointWitnesses(cache, previous headerCache, knownJSON string) error {
    candidates:=mobileCandidates(knownJSON)
    type result struct { node string; group string; ok bool }
    observations:=make(chan result,len(candidates))
    for _,candidate:=range candidates {
        node:=candidate
        go func(){
            st,err:=statusFromNode(node)
            if err!=nil || st.Height<cache.VerifiedHeight ||
                chainWorkCmp(st.ChainWork,cache.ChainWork)<0 {
                observations<-result{};return
            }
            headers,err:=fetchHeaderBatch(node,cache.VerifiedHeight,1)
            if err!=nil || len(headers)!=1 || headers[0].Hash().String()!=cache.VerifiedTip {
                observations<-result{};return
            }
            // Prevent silent replacement of a previously verified local chain
            // when the old tip is no longer in the downloaded history window.
            if previous.VerifiedHeight>verifiedCheckpointHeight && previous.VerifiedHeight<cache.VerifiedHeight {
                old,err:=fetchHeaderBatch(node,previous.VerifiedHeight,1)
                if err!=nil || len(old)!=1 || old[0].Hash().String()!=previous.VerifiedTip {
                    observations<-result{};return
                }
            }
            observations<-result{node:node,group:checkpointNetworkGroup(node),ok:true}
        }()
    }
    deadline:=time.NewTimer(11*time.Second)
    defer deadline.Stop()
    groups:=map[string]bool{}
    for i:=0;i<len(candidates);i++ {
        select {
        case result:=<-observations:
            if result.ok && result.group!="" {
                groups[result.group]=true
                if len(groups)>=2 {return nil}
            }
        case <-deadline.C:
            return errors.New("new signed checkpoint not confirmed by two reachable, distinct peer groups")
        }
    }
    return errors.New("new signed checkpoint not confirmed by two reachable, distinct peer groups")
}

func checkpointCompatibleWithLocal(previous, candidate headerCache) error {
    if candidate.VerifiedHeight<previous.VerifiedHeight {
        return errors.New("checkpoint rollback rejected")
    }
    if candidate.VerifiedHeight==previous.VerifiedHeight {
        if candidate.VerifiedTip!=previous.VerifiedTip {
            return errors.New("same-height signed checkpoint conflicts with local verified tip")
        }
        return nil
    }
    for _,h:=range candidate.History {
        if h.Height==previous.VerifiedHeight && h.Hash().String()!=previous.VerifiedTip {
            return errors.New("signed checkpoint conflicts with locally verified history")
        }
    }
    if chainWorkCmp(candidate.ChainWork,previous.ChainWork)<=0 {
        return errors.New("signed checkpoint did not increase cumulative work")
    }
    return nil
}

// UpdateSignedCheckpoint is a network-only, read/verify/atomic-write operation.
// It NEVER reads wallet files, passwords or keys, and does not perform spending.
// The manifest MUST have a valid signature, monotonic chainwork, stable network
// identity, and 2 externally reachable witness netgroups before replacing cache.
func UpdateSignedCheckpoint(knownNodesJSON,cachePath string) (string,error) {
    if strings.TrimSpace(cachePath)=="" {
        return "",errors.New("header cache path is empty")
    }
    data,err:=fetchSignedCheckpoint()
    if err!=nil{return "",err}
    publicKeyBytes,err:=hex.DecodeString(checkpointSignerPublicHex)
    if err!=nil{return "",err}
    manifest,err:=verifyCheckpointCertificate(data,ed25519.PublicKey(publicKeyBytes),time.Now())
    if err!=nil{return "",err}
    previous,err:=loadHeaderCache(cachePath)
    if err!=nil{return "",err}
    candidate:=manifest.Cache
    if err:=checkpointCompatibleWithLocal(previous,candidate);err!=nil {
        if candidate.VerifiedHeight<previous.VerifiedHeight {return "already-ahead",nil}
        return "",err
    }
    if candidate.VerifiedHeight==previous.VerifiedHeight{return "already-current",nil}
    if err:=checkCheckpointWitnesses(candidate,previous,knownNodesJSON);err!=nil {
        return "",err
    }
    // loadHeaderCache never downgrades a newer locally verified chain. Recheck
    // immediately before writing to avoid stale candidate overwrites.
    newer,err:=loadHeaderCache(cachePath)
    if err!=nil{return "",err}
    if newer.VerifiedHeight!=previous.VerifiedHeight || newer.VerifiedTip!=previous.VerifiedTip {
        return "changed-during-update",nil
    }
    if err:=os.MkdirAll(filepath.Dir(cachePath),0700);err!=nil{return "",err}
    if err:=saveHeaderCache(cachePath,candidate);err!=nil{return "",err}
    return fmt.Sprintf("installed signed checkpoint at height %d",candidate.VerifiedHeight),nil
}

// Certificate contents can be used by CI to resume verified work without
// any dependency on wallet material.
func ReadSignedCheckpointForCI(raw []byte) (headerCache,error) {
    k,err:=hex.DecodeString(checkpointSignerPublicHex)
    if err!=nil{return headerCache{},err}
    p,err:=verifyCheckpointCertificate(raw,ed25519.PublicKey(k),time.Now())
    if err!=nil{return headerCache{},err}
    return p.Cache,nil
}
