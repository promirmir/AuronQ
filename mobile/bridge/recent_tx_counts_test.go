package bridge

import (
    "encoding/json"
    "fmt"
    "strings"
    "testing"
)

func TestParseRecentBlockTxCountsAcceptsRealZeroAndNonzero(t *testing.T) {
    h1:=strings.Repeat("a",128)
    h2:=strings.Repeat("b",128)
    body:=fmt.Sprintf(`{"blocks":[{"height":1331,"hash":"%s","transactions":1,"prev_hash":"irrelevant"},{"height":1330,"hash":"%s","transactions":0}]}`,h1,h2)
    raw,err:=parseRecentBlockTxCounts([]byte(body),6)
    if err!=nil {t.Fatal(err)}
    var response struct {Blocks []recentBlockTxCount `json:"blocks"`}
    if err:=json.Unmarshal([]byte(raw),&response);err!=nil {t.Fatal(err)}
    if len(response.Blocks)!=2 ||
        response.Blocks[0].Transactions!=1 || response.Blocks[1].Transactions!=0 {
        t.Fatalf("incorrect counts: %+v",response.Blocks)
    }
}

func TestParseRecentBlockTxCountsRejectsFabricatedMissingMalformedData(t *testing.T) {
    goodHash:=strings.Repeat("a",128)
    testcases:=[]string{
        `{"blocks":[{"height":3,"hash":"abc","transactions":4}]}`,
        fmt.Sprintf(`{"blocks":[{"height":3,"hash":"%s","transactions":-1}]}`,goodHash),
        fmt.Sprintf(`{"blocks":[{"height":3,"hash":"%s"}]}`,goodHash),
        fmt.Sprintf(`{"blocks":[{"height":3,"hash":"%s","transactions":null}]}`,goodHash),
        fmt.Sprintf(`{"blocks":[{"height":3,"hash":"%s","transactions":"0"}]}`,goodHash),
        fmt.Sprintf(`{"blocks":[{"height":3,"hash":"%s","transactions":0.5}]}`,goodHash),
        fmt.Sprintf(`{"blocks":[{"height":3,"hash":"%s","transactions":1},{"height":3,"hash":"%s","transactions":1}]}`,goodHash,goodHash),
        fmt.Sprintf(`{"blocks":[{"height":3,"hash":"%s","transactions":0},{"height":2,"hash":"%s","transactions":1},{"height":1,"hash":"%s","transactions":2}]}`,goodHash,goodHash,goodHash),
        `{"blocks":[{}`, // invalid JSON
    }
    for i,input:=range testcases{
        limit:=1
        if strings.Contains(input,`"height":1`){limit=2}
        if strings.Contains(input,`"height":3,"hash":"`+goodHash+`","transactions":1},{"height":3`){limit=6}
        if _,err:=parseRecentBlockTxCounts([]byte(input),limit);err==nil {
            t.Fatalf("bad explorer metadata accepted, case %d",i)
        }
    }
    if _,err:=parseRecentBlockTxCounts(make([]byte,(48<<10)+1),6);err==nil {
        t.Fatal("oversized explorer response accepted")
    }
}
