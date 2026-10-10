package bridge

import (
  "crypto/sha256"
  "encoding/json"
  "errors"
  "fmt"
  "io"
  "net/http"
  "os"
  "path/filepath"
  "strings"
  "time"

  "github.com/sigstore/sigstore-go/pkg/bundle"
  "github.com/sigstore/sigstore-go/pkg/root"
  "github.com/sigstore/sigstore-go/pkg/tuf"
  "github.com/sigstore/sigstore-go/pkg/verify"
)

const (
  keylessCheckpointManifestURL = "https://raw.githubusercontent.com/promirmir/AuronQ/automation/mobile-checkpoints/latest.json"
  keylessCheckpointBundleURL = "https://raw.githubusercontent.com/promirmir/AuronQ/automation/mobile-checkpoints/latest.sigstore.json"
  keylessOIDCIssuer = "https://token.actions.githubusercontent.com"
  // Only this reviewed workflow on main may issue accepted checkpoint bundles.
  keylessWorkflowIdentity = "^https://github\\.com/promirmir/AuronQ/\\.github/workflows/mobile-keyless-checkpoints\\.yml@refs/heads/main$"
)

func readCheckpointArtifact(url string, maxBytes int64)([]byte,error){
  request,err:=http.NewRequest(http.MethodGet,url,nil)
  if err!=nil{return nil,err}
  request.Header.Set("User-Agent","AuronQ-Mobile/keyless-checkpoint-v1")
  client:=mobileHTTP(8*time.Second)
  res,err:=client.Do(request)
  if err!=nil {return nil,err}
  defer res.Body.Close()
  if res.StatusCode!=http.StatusOK {return nil,fmt.Errorf("checkpoint HTTP %d",res.StatusCode)}
  data,err:=io.ReadAll(io.LimitReader(res.Body,maxBytes+1))
  if err!=nil{return nil,err}
  if int64(len(data))>maxBytes{return nil,errors.New("checkpoint response oversized")}
  return data,nil
}

// Signature and artifact digest are authenticated with Sigstore's Fulcio CA
// and Rekor transparency log, not with a hand-configured signing key.
// The certificate MUST attest our exact GitHub Actions workflow on main.
// Never fall back to unsigned GitHub manifests on verification failures.
func verifyKeylessArtifact(checkpointJSON, bundleJSON []byte)error{
  if len(checkpointJSON)==0||len(bundleJSON)==0{return errors.New("unsigned checkpoint")}
  path,err:=os.CreateTemp("","auronq-sigstore-bundle-*.json")
  if err!=nil{return err}
  defer os.Remove(path.Name())
  if _,err:=path.Write(bundleJSON);err!=nil {path.Close();return err}
  if err:=path.Close();err!=nil{return err}
  signed,err:=bundle.LoadJSONFromPath(path.Name())
  if err!=nil{return fmt.Errorf("invalid Sigstore bundle: %w",err)}
  tufClient,err:=tuf.New(tuf.DefaultOptions())
  if err!=nil{return fmt.Errorf("Sigstore TUF unavailable: %w",err)}
  trusted,err:=root.GetTrustedRoot(tufClient)
  if err!=nil{return fmt.Errorf("cannot load Sigstore trust roots: %w",err)}
  verifier,err:=verify.NewVerifier(trusted,
     verify.WithSignedCertificateTimestamps(1),
     verify.WithTransparencyLog(1),
     verify.WithObserverTimestamps(1),
  )
  if err!=nil{return err}
  id,err:=verify.NewShortCertificateIdentity(
    keylessOIDCIssuer,"","",keylessWorkflowIdentity)
  if err!=nil{return err}
  checksum:=sha256.Sum256(checkpointJSON)
  _,err=verifier.Verify(signed,verify.NewPolicy(
    verify.WithArtifactDigest("sha256",checksum[:]),
    verify.WithCertificateIdentity(id),
  ))
  if err!=nil{return fmt.Errorf("invalid checkpoint provenance, digest or signature: %w",err)}
  return nil
}

// Read-only, authenticated manifest update. Skips when the old signed
// checkpoint is ahead of this candidate, and never bypasses the local
// AQM64 header verifier or changes wallet files, keys or spending rules.
func UpdateKeylessCheckpoint(knownNodesJSON, cachePath string)(string,error){
  if strings.TrimSpace(cachePath)=="" {return "",errors.New("header cache path is empty")}
  payload,err:=readCheckpointArtifact(keylessCheckpointManifestURL,256<<10)
  if err!=nil{return "",err}
  proof,err:=readCheckpointArtifact(keylessCheckpointBundleURL,256<<10)
  if err!=nil{return "",err}
  if err:=verifyKeylessArtifact(payload,proof);err!=nil{return "",err}
  var p checkpointPayload
  if err:=json.Unmarshal(payload,&p);err!=nil{return "",err}
  if err:=checkpointPayloadValid(p,time.Now());err!=nil{return "",err}
  previous,err:=loadHeaderCache(cachePath)
  if err!=nil{return "",err}
  next:=p.Cache
  if next.VerifiedHeight<previous.VerifiedHeight{return "already-ahead",nil}
  if err:=checkpointCompatibleWithLocal(previous,next);err!=nil{return "",err}
  if next.VerifiedHeight==previous.VerifiedHeight{return "already-current",nil}
  if err:=checkCheckpointWitnesses(next,previous,knownNodesJSON);err!=nil{return "",err}
  again,err:=loadHeaderCache(cachePath)
  if err!=nil{return "",err}
  if again.VerifiedHeight!=previous.VerifiedHeight || again.VerifiedTip!=previous.VerifiedTip {
    return "changed-during-update",nil
  }
  if err:=os.MkdirAll(filepath.Dir(cachePath),0700);err!=nil{return "",err}
  if err:=saveHeaderCache(cachePath,next);err!=nil{return "",err}
  return fmt.Sprintf("installed Sigstore-keyless checkpoint at %d",next.VerifiedHeight),nil
}
