package bridge

import (
    "errors"
    "sync/atomic"
    "testing"
    "time"

    aq "auronq/internal/auronq"
)

func fakeSequentialHeaders(count int) ([]aq.BlockHeader, []aq.BlockHeader) {
    base:=make([]aq.BlockHeader,128)
    base[0]=aq.BlockHeader{Height:0,Timestamp:1}
    for i:=1;i<128;i++ {
        base[i]=aq.BlockHeader{Height:uint64(i),Timestamp:int64(i+1),PrevHash:base[i-1].Hash()}
    }
    headers:=make([]aq.BlockHeader,count)
    prev:=base[len(base)-1]
    for i:=range headers {
        headers[i]=aq.BlockHeader{Height:prev.Height+1,Timestamp:prev.Timestamp+1,PrevHash:prev.Hash()}
        prev=headers[i]
    }
    return base,headers
}

func TestFast053ParallelBatchPreservesExactOriginalConsensusHistory(t *testing.T) {
    base,headers:=fakeSequentialHeaders(7)
    seen:=make([]int32,len(headers))
    for _,workerCount:=range []int{1,2} {
        for i:=range seen {atomic.StoreInt32(&seen[i],0)}
        at,err:=validateHeaderBatch053With(base,headers,workerCount,
            func(h aq.BlockHeader,history []aq.BlockHeader,now int64)error{
                idx:=int(h.Height-base[len(base)-1].Height-1)
                if idx<0||idx>=len(headers){return errors.New("wrong index")}
                expected:=make([]aq.BlockHeader,0,len(base)+idx)
                expected=append(expected,base...)
                expected=append(expected,headers[:idx]...)
                if len(expected)>headerCacheKeep{expected=expected[len(expected)-headerCacheKeep:]}
                if len(history)!=len(expected){return errors.New("wrong consensus history length")}
                for n:=range expected{
                    if history[n].Height!=expected[n].Height||history[n].Hash()!=expected[n].Hash(){
                        return errors.New("unverified/missing predecessor in difficulty history")
                    }
                }
                atomic.AddInt32(&seen[idx],1)
                return nil
            })
        if err!=nil||at!=-1{t.Fatalf("workers=%d: failed at=%d: %v",workerCount,at,err)}
        for i:=range seen{if atomic.LoadInt32(&seen[i])!=1{t.Fatalf("workers=%d: header %d not checked exactly once",workerCount,i)}}
    }
}

func TestFast053ParallelRespectsOriginalEarliestFailure(t *testing.T) {
    base,headers:=fakeSequentialHeaders(6)
    fail:=errors.New("invalid 0.5.3 proof")
    for _,n:=range []int{1,2,4}{
        idx,err:=validateHeaderBatch053With(base,headers,n,func(h aq.BlockHeader,_ []aq.BlockHeader,_ int64)error {
            if h.Height==headers[2].Height{return fail}
            if h.Height==headers[3].Height{return errors.New("later failure")}
            return nil
        })
        if idx!=2 || err!=fail{t.Fatalf("workers=%d: returned index %d, err %v",n,idx,err)}
    }
}

func TestFast053ParallelRejectsBrokenLinkBeforeExpensivePow(t *testing.T){
    base,headers:=fakeSequentialHeaders(2)
    headers[1].PrevHash=aq.Hash{}
    var called int32
    idx,err:=validateHeaderBatch053With(base,headers,2,func(_ aq.BlockHeader,_ []aq.BlockHeader,_ int64)error {
        atomic.AddInt32(&called,1)
        return nil
    })
    if idx!=1||err==nil||called!=0{t.Fatalf("invalid link accepted: idx=%d, err=%v, called=%d",idx,err,called)}
}

func TestFast053ParallelTrulyBoundedAtTwo(t *testing.T){
    base,headers:=fakeSequentialHeaders(6)
    var active,peak int32
    idx,err:=validateHeaderBatch053With(base,headers,2,func(_ aq.BlockHeader,_ []aq.BlockHeader,_ int64)error{
        a:=atomic.AddInt32(&active,1)
        for {
            p:=atomic.LoadInt32(&peak)
            if a<=p || atomic.CompareAndSwapInt32(&peak,p,a) {break}
        }
        time.Sleep(20*time.Millisecond)
        atomic.AddInt32(&active,-1)
        return nil
    })
    if err!=nil||idx!=-1{t.Fatalf("headers rejected: %d %v",idx,err)}
    if peak!=2{t.Fatalf("expected two concurrently validating workers, max=%d",peak)}
}

func TestFast053NoUncheckedInvalidEnvelopeCanPass(t *testing.T) {
    base,headers:=fakeSequentialHeaders(2)
    // Invalid version is rejected by EXACTLY the old consensus verifier.
    for _,workerCount:=range []int{1,2} {
        idx,err:=validateHeaderBatch053With(base,headers,workerCount,aq.ValidateHeaderEnvelope)
        if idx!=0||err==nil{t.Fatalf("invalid PoW/version accepted: workers=%d, idx=%d, err=%v",workerCount,idx,err)}
    }
}

func TestFast053RuntimeVerificationWorkerCap(t *testing.T){
    defer SetHeaderValidationWorkers(1)
    for _,n:=range []int{-1,0,1,3,100}{
        SetHeaderValidationWorkers(n)
        if got:=mobileHeaderWorkers.Load();got!=1{t.Fatalf("unsafe worker setting: got %d",got)}
    }
    SetHeaderValidationWorkers(2)
    if mobileHeaderWorkers.Load()!=2{t.Fatal("bounded two-worker mode not enabled")}
}
