package bridge

import (
    "errors"
    "sync"
    "sync/atomic"
    "time"

    aq "auronq/internal/auronq"
)

// Fresh 0.5.3 wallets must validate EVERY header from the immutable genesis.
// This is a bounded-performance optimization, not a trusted checkpoint.
// AQM64 takes 64 MiB per worker, so the caller may opt in to at most two
// workers on a sufficiently well-equipped device. Default remains one.
var mobileHeaderWorkers atomic.Int32

// SetHeaderValidationWorkers is called by the Android app after checking
// device memory. No remote peer can choose the worker count.
func SetHeaderValidationWorkers(workers int) {
    if workers != 2 { workers = 1 }
    mobileHeaderWorkers.Store(int32(workers))
}

type headerValidator func(aq.BlockHeader, []aq.BlockHeader, int64) error

func validateHeaderBatch053(base []aq.BlockHeader, headers []aq.BlockHeader) (int, error) {
    n:=int(mobileHeaderWorkers.Load())
    if n!=2 { n=1 }
    return validateHeaderBatch053With(base,headers,n,aq.ValidateHeaderEnvelope)
}

// The exact original consensus validator checks PoW, retargets, timestamps,
// height, and previous-hash links. No header is trusted or silently skipped.
// For each batch candidate we pass the same previous 128 headers that the
// original sequential 0.5.3 loop would have used, including earlier batch
// candidates. ONLY AFTER every check passes do we commit the ordered batch.
// No unverified candidate can persist to wallet state or the header cache.
func validateHeaderBatch053With(base []aq.BlockHeader, headers []aq.BlockHeader,
    workers int, validate headerValidator) (int, error) {

    if len(base)==0 { return 0,errors.New("header history is empty") }
    if workers!=2 { workers=1 }
    if validate==nil { return 0,errors.New("header validator is missing") }

    // Adjacent-link precheck is deliberately cheap and prevents a malicious
    // peer from forcing expensive 64 MiB AQM64 work on obviously bad batches.
    last:=base[len(base)-1]
    for i,h:=range headers {
        if h.Height!=last.Height+1 { return i,errors.New("nonconsecutive header height") }
        if h.PrevHash!=last.Hash() { return i,errors.New("header previous hash mismatch") }
        last=h
    }

    for start:=0;start<len(headers);start+=workers {
        end:=start+workers
        if end>len(headers){end=len(headers)}
        errorsAt:=make([]error,end-start)
        var wg sync.WaitGroup
        for i:=start;i<end;i++ {
            // Snapshot the exact rolling history that the previous version
            // would supply at the same index; never share mutable history.
            history:=make([]aq.BlockHeader,0,len(base)+i)
            history=append(history,base...)
            history=append(history,headers[:i]...)
            if len(history)>headerCacheKeep {history=history[len(history)-headerCacheKeep:]}
            h:=headers[i]
            idx:=i-start
            if workers==1 {
                errorsAt[idx]=validate(h,history,time.Now().Unix())
            } else {
                wg.Add(1)
                go func(){
                    defer wg.Done()
                    errorsAt[idx]=validate(h,history,time.Now().Unix())
                }()
            }
        }
        wg.Wait()
        // Preserve original earliest-error behavior, don't commit a partially
        // or differently validated chain even if two workers complete out of order.
        for j,err:=range errorsAt {
            if err!=nil {return start+j,err}
        }
    }
    return -1,nil
}
