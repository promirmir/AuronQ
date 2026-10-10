package bridge

import (
 "strings"
 "testing"
)

func TestQuickNetgroupClassification(t *testing.T) {
 good:=map[string]string{
  "http://45.88.201.77:18444":"ipv4:45.88",
  "http://54.38.81.30:18444":"ipv4:54.38",
  "http://69.173.206.211:18444":"ipv4:69.173",
 }
 for addr,want:=range good {
  if got:=publicPeerNetgroup(addr);got!=want{t.Fatalf("%s: got %s wanted %s",addr,got,want)}
 }
 for _,addr:=range []string{"https://node.example.org","http://127.0.0.1:18444","http://192.168.1.1:18444","http://203.0.113.1:18444"}{
  if g:=publicPeerNetgroup(addr);g!=""{t.Fatalf("unverifiable public netgroup accepted: %s",addr)}
 }
}

func TestQuickRequiresThreeDistinctGroupsOnSameTipAndWork(t *testing.T){
 tip:=strings.Repeat("a",128)
 peers:=[]mobileNodeObservation{
   {Node:"http://45.88.201.77:18444",Height:10000,Tip:tip,ChainWork:"abc"},
   {Node:"http://45.88.222.3:18444",Height:10000,Tip:tip,ChainWork:"abc"},
   {Node:"http://54.38.81.30:18444",Height:10000,Tip:tip,ChainWork:"abc"},
 }
 if x:=matchingThreeNetgroups(peers);len(x)!=0{t.Fatal("two /16 netgroups misreported as three")}
 peers=append(peers,mobileNodeObservation{Node:"http://69.173.206.211:18444",Height:10000,Tip:tip,ChainWork:"abc"})
 if x:=matchingThreeNetgroups(peers);len(x)!=3{t.Fatalf("three agreeing /16 groups rejected: %+v",x)}
 peers[3].Tip=strings.Repeat("b",128)
 if x:=matchingThreeNetgroups(peers);len(x)!=0{t.Fatal("same-height fork accepted as consensus")}
 peers[3].Tip=tip
 peers[3].ChainWork="abd"
 if x:=matchingThreeNetgroups(peers);len(x)!=0{t.Fatal("peer with higher work conflicting tip must block readiness")}
}

func TestQuickInvalidOrGenesisTipRejected(t *testing.T){
 if quickStateValid(0,strings.Repeat("a",128),"123"){t.Fatal("genesis tip cannot qualify for spend readiness")}
 if quickStateValid(10,"junk","123"){t.Fatal("invalid hash accepted")}
 if quickStateValid(10,strings.Repeat("a",128),"not-a-hex-work"){t.Fatal("invalid chainwork accepted")}
 if !quickStateValid(10000,strings.Repeat("a",128),"123abc"){t.Fatal("valid reported tip rejected")}
}
