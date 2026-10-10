package bridge

import (
 "encoding/json"
 "errors"
 "fmt"
 "math/big"
 "net"
 "net/url"
 "sort"
 "strings"
 "sync"
 "time"

 aq "auronq/internal/auronq"
)

// Three distinct publicly routed IPv4 /16 network groups are used as a
// conservative *network-diversity heuristic*. They do NOT prove separate
// owners or that a peer is truthful. No Mainnet consensus rule is changed.
const quickMinGroups = 3

func publicPeerNetgroup(raw string) string {
 norm:=normalizeMobileNode(raw)
 if norm=="" {return ""}
 u,err:=url.Parse(norm)
 if err!=nil{return ""}
 ip:=net.ParseIP(u.Hostname())
 if ip==nil||mobileNonPublicIP(ip){return ""} // DNS ≠ independently owned seed
 if v4:=ip.To4();v4!=nil{return fmt.Sprintf("ipv4:%d.%d",v4[0],v4[1])}
 v6:=ip.To16()
 return fmt.Sprintf("ipv6:%x:%x",uint16(v6[0])<<8|uint16(v6[1]),uint16(v6[2])<<8|uint16(v6[3]))
}

type quickPeerView struct {
 Obs mobileNodeObservation
 Group string
 Balance aq.BalanceResponse
 History []aq.WalletHistoryItem
 UTXOs []aq.UTXORecord
}

func quickStateValid(height uint64,tip,work string)bool{
 if height==0||len(tip)!=128 {return false}
 if _,ok:=new(big.Int).SetString(strings.TrimSpace(work),16);!ok{return false}
 return true
}

// Peer failures return UNKNOWN, never a fabricated 0. This helper intentionally
// leaves independent AQM64 validation out of the fast path, so UI MUST label
// its output "peer-confirmed, not trustlessly locally verified".
func quickMatchingPeers(knownJSON string) ([]mobileNodeObservation,error) {
 candidates:=mobileCandidates(knownJSON)
 first:=quickMatchingFromCandidates(candidates)
 if len(first)>=quickMinGroups{return first,nil}
 // Optional peer manifest is a fallback, never mandatory for new users.
 // This keeps the wallet usable even if GitHub or the bootstrap site goes down.
 if extra,err:=fetchManifest();err==nil{
  seen:=map[string]bool{}
  for _,p:=range candidates{seen[p]=true}
  for _,p:=range extra{addMobileCandidate(&candidates,seen,p)}
  if len(candidates)>maxMobileProbeCandidates{candidates=candidates[:maxMobileProbeCandidates]}
  if found:=quickMatchingFromCandidates(candidates);len(found)>=quickMinGroups{return found,nil}
 }
 return nil,errors.New("fewer than three agreeing publicly routed network groups; balance and payments unavailable")
}

func quickMatchingFromCandidates(candidates []string) []mobileNodeObservation{
 ch:=make(chan mobileNodeObservation,len(candidates))
 pending:=0
 for _,node:=range candidates {
  node:=node
  if publicPeerNetgroup(node)==""{continue}
  pending++
  go func(){
   st,err:=statusFromNode(node)
   if err!=nil||!quickStateValid(st.Height,st.Tip.String(),st.ChainWork){ch<-mobileNodeObservation{};return}
   ch<-mobileNodeObservation{Node:node,Height:st.Height,Tip:st.Tip.String(),ChainWork:st.ChainWork,Peers:st.Peers}
  }()
 }
 results:=make([]mobileNodeObservation,0,pending)
 timer:=time.NewTimer(7*time.Second);defer timer.Stop()
 for pending>0 {
  select{
  case obs:=<-ch:
   pending--
   if obs.Node!="" {results=append(results,obs)}
   if selected:=matchingThreeNetgroups(results);len(selected)>=quickMinGroups{
    return selected
   }
  case <-timer.C:pending=0
  }
 }
 return nil
}

func matchingThreeNetgroups(obs []mobileNodeObservation) []mobileNodeObservation {
 // Group by exact tip AND cumulative work; same-height forks are not merged.
 // Select greatest reported work among groups with >=3 distinct network
 // groups; a greater-work conflicting peer prevents false certainty.
 maxWork:=new(big.Int)
 for _,o:=range obs {
  if _,ok:=new(big.Int).SetString(strings.TrimSpace(o.ChainWork),16);!ok{continue}
  w,_:=new(big.Int).SetString(strings.TrimSpace(o.ChainWork),16)
  if w.Cmp(maxWork)>0 {maxWork=w}
 }
 groups:=map[string][]mobileNodeObservation{}
 for _,o:=range obs {
  if publicPeerNetgroup(o.Node)=="" {continue}
  w,ok:=new(big.Int).SetString(strings.TrimSpace(o.ChainWork),16)
  if !ok||w.Cmp(maxWork)!=0{continue}
  key:=mobileStateKey(o)
  groups[key]=append(groups[key],o)
 }
 for _,group:=range groups {
  sort.Slice(group,func(i,j int)bool{return group[i].Node<group[j].Node})
  used:=map[string]bool{}
  selected:=make([]mobileNodeObservation,0,quickMinGroups)
  for _,o:=range group {
   g:=publicPeerNetgroup(o.Node)
   if used[g]{continue}
   used[g]=true
   selected=append(selected,o)
  }
  if len(selected)>=quickMinGroups{return selected[:quickMinGroups]}
 }
 return nil
}

func quickReadOne(obs mobileNodeObservation,address string,limit int) (quickPeerView,error){
 var v quickPeerView
 v.Obs=obs
 v.Group=publicPeerNetgroup(obs.Node)
 client:=aq.NewClient(obs.Node)
 client.HTTP=mobileHTTP(6*time.Second)
 var err error
 if v.Balance,err=client.Balance(address);err!=nil{return v,err}
 if v.UTXOs,err=client.UTXOs(address);err!=nil{return v,err}
 if v.History,err=client.History(address,limit);err!=nil{return v,err}
 // Avoid constructing a cross-height state from replies spanning a reorg.
 end,err:=client.Status()
 if err!=nil || end.NetworkID.String()!=mainnetNetworkID ||
   end.Height!=obs.Height || end.Tip.String()!=obs.Tip ||
   !strings.EqualFold(end.ChainWork,obs.ChainWork){
   return v,errors.New("chain tip changed during account snapshot")
 }
 return v,nil
}

func quickFetchAccount(knownJSON,address string,limit int)([]quickPeerView,error){
 if n,_,_,e:=aq.DecodeAddress(strings.TrimSpace(address));e!=nil||n!=aq.MainnetNetworkByte{
  return nil,errors.New("invalid mainnet wallet address")
 }
 if limit<1 {limit=50}
 if limit>100{limit=100}
 peers,err:=quickMatchingPeers(knownJSON)
 if err!=nil{return nil,err}
 ch:=make(chan struct{view quickPeerView;err error},len(peers))
 var wg sync.WaitGroup
 for _,o:=range peers {
  o:=o
  wg.Add(1)
  go func(){
   defer wg.Done()
   view,e:=quickReadOne(o,address,limit)
   ch<-struct{view quickPeerView;err error}{view,e}
  }()
 }
 wg.Wait()
 close(ch)
 views:=make([]quickPeerView,0,len(peers))
 for r:=range ch {
  if r.err!=nil{return nil,fmt.Errorf("three-peer read failed: %w",r.err)}
  views=append(views,r.view)
 }
 if len(views)!=quickMinGroups{return nil,errors.New("three peers did not return a consistent account snapshot")}
 ref:=views[0]
 groups:=map[string]bool{}
 for _,v:=range views {
  if groups[v.Group]||v.Group=="" {return nil,errors.New("peers are not network-diverse")}
  groups[v.Group]=true
  if v.Obs.Height!=ref.Obs.Height||v.Obs.Tip!=ref.Obs.Tip||
   !strings.EqualFold(v.Obs.ChainWork,ref.Obs.ChainWork)||
   v.Balance.Spendable!=ref.Balance.Spendable||v.Balance.Total!=ref.Balance.Total||
   utxoFingerprint(v.UTXOs)!=utxoFingerprint(ref.UTXOs)||
   historyFingerprint(v.History)!=historyFingerprint(ref.History){
   return nil,errors.New("three full nodes disagree about wallet state; no balance can be trusted")
  }
 }
 sort.Slice(views,func(i,j int)bool{return views[i].Group<views[j].Group})
 return views,nil
}

// QuickNetworkSnapshot returns only peer-OBSERVED chain metadata. It is not a
// proof that full nodes are honest, and is never used as verified AQM64 chainwork.
func QuickNetworkSnapshot(knownJSON string)(string,error){
 peers,err:=quickMatchingPeers(knownJSON)
 if err!=nil{return "",err}
 raw,err:=NetworkSnapshot(peers[0].Node,0) // no full-block network delay
 if err!=nil{return "",err}
 var out map[string]any
 if err:=json.Unmarshal([]byte(raw),&out);err!=nil{return "",err}
 groups:=make([]string,0,len(peers))
 for _,p:=range peers{groups=append(groups,publicPeerNetgroup(p.Node))}
 out["peer_observed"]=len(peers)
 out["peer_agreement"]=len(peers)
 out["network_groups"]=groups
 out["multi_peer_confirmed"]=true
 out["header_verified"]=false
 out["state_trust"]="three-netgroups-peer-observed-not-independent-chain-proof"
 b,e:=json.Marshal(out)
 return string(b),e
}

// QuickAccountSnapshot is an explicitly PEER-TRUSTED light-client read, not an
// independently proved historical chain or UTXO commitment.
func QuickAccountSnapshot(knownJSON,address string,limit int)(string,error){
 views,err:=quickFetchAccount(knownJSON,address,limit)
 if err!=nil{return "",err}
 v:=views[0]
 nodes:=make([]string,0,len(views))
 ng:=make([]string,0,len(views))
 for _,x:=range views{nodes=append(nodes,x.Obs.Node);ng=append(ng,x.Group)}
 // Pending mempool entries may differ between honest full nodes.
 // Never portray pending observed at just ONE peer as confirmed by THREE.
 items:=make([]aq.WalletHistoryItem,0,len(v.History))
 for _,item:=range v.History{
  if item.Status!="pending"{items=append(items,item)}
 }
 if len(items)>limit && limit>0{items=items[:limit]}
 out:=map[string]any{
  "address":strings.TrimSpace(address),"height":v.Obs.Height,"tip":v.Obs.Tip,
  "chain_work":v.Obs.ChainWork,"spendable":aq.FormatAmount(v.Balance.Spendable),
  "total":aq.FormatAmount(v.Balance.Total),"items":items,
  "peer_agreement":len(views),"peer_observed":len(views),
  "multi_peer_confirmed":true,"netgroups":ng,"agreement_nodes":nodes,
  "state_trust":"three-netgroups-peer-observed-not-independent-chain-proof",
  "header_verified":false,
 }
 raw,e:=json.Marshal(out);return string(raw),e
}

// SendWithQuickQuorum reevaluates the three-node state and UTXO set at send
// time. Locally builds and signs a transaction using original v0.5.3 wallet
// code; the remote nodes never receive a key/password. The response explicitly
// identifies a peer-trusted, rather than cryptographically proven, state.
func SendWithQuickQuorum(knownJSON,walletPath,password,to,amount string)(string,error){
 w,err:=aq.LoadWallet(walletPath,password)
 if err!=nil{return "",err}
 defer w.Close()
 if w.File.NetworkByte!=aq.MainnetNetworkByte {return "",errors.New("not an AuronQ mainnet wallet")}
 amt,err:=aq.ParseAmount(strings.TrimSpace(amount))
 if err!=nil{return "",err}
 views,err:=quickFetchAccount(knownJSON,w.Address(),50)
 if err!=nil{return "",err}
 tx,fee,err:=w.BuildTransaction(views[0].UTXOs,strings.TrimSpace(to),amt,aq.MainnetNetworkByte)
 if err!=nil{return "",err}
 // A consent dialog is also required at Android level.
 accepted:=make([]string,0,3)
 for _,v:=range views{
  cl:=aq.NewClient(v.Obs.Node)
  cl.HTTP=mobileHTTP(10*time.Second)
  if _,e:=cl.SubmitTx(tx);e==nil{accepted=append(accepted,v.Obs.Node)}
 }
 if len(accepted)==0 {return "",errors.New("no peer accepted transaction")}
 out:=map[string]any{
  "txid":tx.ID().String(),"fee":aq.FormatAmount(fee),"fee_atoms":fee,
  "direct_accepted":len(accepted),"broadcast_attempted":len(views),
  "utxo_peer_agreement":3,"utxo_peer_observed":3,
  "state_trust":"three-netgroups-peer-observed-not-independent-chain-proof",
 }
 b,_:=json.Marshal(out)
 return string(b),nil
}
