package bridge

import (
  "encoding/json"
  "net/http"
  "net/http/httptest"
  "strconv"
  "testing"
  "time"

  aq "auronq/internal/auronq"
)

// Regression: recent-block data remain full blocks, ordered most recent first,
// with exact original transaction counts; a slower response must not reorder.
func TestFast053RecentSnapshotPreservesFullBlockHistory(t *testing.T) {
  blocks:=map[uint64]aq.Block{
     11:{Header:aq.BlockHeader{Height:11,Timestamp:300},Transactions:[]aq.Transaction{{},{}}},
     10:{Header:aq.BlockHeader{Height:10,Timestamp:250},Transactions:[]aq.Transaction{{}}},
      9:{Header:aq.BlockHeader{Height:9,Timestamp:200},Transactions:[]aq.Transaction{}},
  }
  srv:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
     if r.URL.Path!="/p2p/getblock" {http.NotFound(w,r);return}
     h,err:=strconv.ParseUint(r.URL.Query().Get("height"),10,64)
     if err!=nil {http.Error(w,"bad height",400);return}
     if h==11 {time.Sleep(45*time.Millisecond)}
     block,ok:=blocks[h]
     if !ok {http.NotFound(w,r);return}
     w.Header().Set("Content-Type","application/json")
     if err:=json.NewEncoder(w).Encode(block);err!=nil{t.Errorf("http block: %v",err)}
  }))
  defer srv.Close()
  out:=fetchRecentSnapshotBlocks(srv.URL,11,3)
  if len(out)!=3 {t.Fatalf("want 3 got %d",len(out))}
  for i,height:=range []uint64{11,10,9} {
    block:=blocks[height]
    if out[i].Height!=height || out[i].Hash!=block.Hash().String() ||
       out[i].Transactions!=len(block.Transactions) {
        t.Fatalf("wrong block at index %d: %+v",i,out[i])
    }
  }
}

func TestFast053RecentSnapshotStopsAtFirstMissingBlock(t *testing.T){
  blocks:=map[uint64]aq.Block{
      7:{Header:aq.BlockHeader{Height:7,Timestamp:300},Transactions:[]aq.Transaction{{}}},
      5:{Header:aq.BlockHeader{Height:5,Timestamp:100}},
  }
  srv:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
     h,err:=strconv.ParseUint(r.URL.Query().Get("height"),10,64)
     if err!=nil {http.Error(w,"bad",400);return}
     b,ok:=blocks[h]
     if !ok {http.Error(w,"temporarily unavailable",503);return}
     _=json.NewEncoder(w).Encode(b)
  }))
  defer srv.Close()
  out:=fetchRecentSnapshotBlocks(srv.URL,7,3)
  if len(out)!=1 || out[0].Height!=7 || out[0].Transactions!=1 {
     t.Fatalf("must preserve old 0.5.3 first failure prefix: %+v",out)
  }
}

func TestFast053StatusPreviewSkipsFullBlockDownloads(t *testing.T){
  calls:=0
  srv:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){calls++;http.Error(w,"unexpected full block request",500)}))
  defer srv.Close()
  if b:=fetchRecentSnapshotBlocks(srv.URL,1234,0);len(b)!=0 || calls!=0 {
     t.Fatalf("preview requested full blocks: got %d, requests %d",len(b),calls)
  }
}
