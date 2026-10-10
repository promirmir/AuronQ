package bridge

import (
    "encoding/hex"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "net/http"
    "strings"
    "time"
)

// This is presentation-only explorer metadata, never a source of consensus
// trust, wallet balance, transaction authorization or chain validation.
// A block header alone cannot reveal its number of transactions.
type recentBlockTxCount struct {
    Height uint64 `json:"height"`
    Hash string `json:"hash"`
    Transactions int `json:"transactions"`
}

func parseRecentBlockTxCounts(body []byte, limit int) (string,error) {
    if limit<1 || limit>12 {return "",errors.New("invalid recent-block limit")}
    if len(body)==0 || len(body)>48<<10 {return "",errors.New("explorer response exceeds maximum size")}
    var parsed struct {
        Blocks []recentBlockTxCount `json:"blocks"`
    }
    if err:=json.Unmarshal(body,&parsed);err!=nil{return "",err}
    if len(parsed.Blocks)>limit {return "",errors.New("explorer returned too many blocks")}
    seen:=map[string]bool{}
    for _,item:=range parsed.Blocks {
        if len(item.Hash)!=128 || item.Transactions<0 {
            return "",errors.New("invalid explorer block metadata")
        }
        if _,err:=hex.DecodeString(item.Hash);err!=nil{return "",errors.New("invalid explorer block hash")}
        key:=fmt.Sprintf("%d:%s",item.Height,strings.ToLower(item.Hash))
        if seen[key] {return "",errors.New("duplicate explorer block")}
        seen[key]=true
    }
    canonical,err:=json.Marshal(parsed)
    if err!=nil{return "",err}
    return string(canonical),nil
}

// RecentBlockTransactionCounts gets one tiny JSON summary from the selected
// verified-header peer. Unlike fetching full blocks, it does not parse bulky
// ML-DSA transaction bodies. UI compares height AND hash to the header snapshot
// before showing counts; on timeout it retains "— tx" rather than inventing 0.
func RecentBlockTransactionCounts(nodeURL string, limit int)(string,error) {
    nodeURL=normalizeMobileNode(nodeURL)
    if nodeURL=="" {return "",errors.New("invalid public AuronQ peer URL")}
    if limit<1 || limit>12 {return "",errors.New("invalid recent-block limit")}
    req,err:=http.NewRequest(http.MethodGet,
        fmt.Sprintf("%s/v1/explorer/blocks?limit=%d",nodeURL,limit),nil)
    if err!=nil{return "",err}
    req.Header.Set("User-Agent","AuronQ-Mobile/explorer-block-summary")
    resp,err:=mobileHTTP(2500*time.Millisecond).Do(req)
    if err!=nil{return "",err}
    defer resp.Body.Close()
    if resp.StatusCode!=http.StatusOK {
        return "",fmt.Errorf("explorer block summaries HTTP %d",resp.StatusCode)
    }
    raw,err:=io.ReadAll(io.LimitReader(resp.Body,(48<<10)+1))
    if err!=nil{return "",err}
    return parseRecentBlockTxCounts(raw,limit)
}
