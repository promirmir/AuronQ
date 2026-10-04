package auronq

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type NetworkConfig struct {
	Name               string   `json:"name"`
	ProtocolVersion    uint32   `json:"protocol_version"`
	NetworkByte        byte     `json:"network_byte"`
	FounderAddress     string   `json:"founder_address"`
	Genesis            Block    `json:"genesis"`
	SeedPeers          []string `json:"seed_peers"`
	DNSSeeds           []string `json:"dns_seeds,omitempty"`
	BootstrapManifests []string `json:"bootstrap_manifests,omitempty"`
	CoinbaseMaturity   uint64   `json:"coinbase_maturity"`
}

func (n *NetworkConfig) Maturity() uint64 {
	if n != nil && n.CoinbaseMaturity != 0 {
		return n.CoinbaseMaturity
	}
	if n != nil && n.NetworkByte == TestnetNetworkByte {
		return TestnetCoinbaseMaturity
	}
	return MainnetCoinbaseMaturity
}

func ConsensusFingerprint() Hash {
	var b bytes.Buffer
	b.WriteString("AURONQ_CONSENSUS_V3\x00")
	writeU64(&b, MaxSupplyAtoms)
	writeU64(&b, FounderAtoms)
	writeU64(&b, InitialSubsidyAtoms)
	writeU64(&b, HalvingInterval)
	writeU64(&b, MainnetCoinbaseMaturity)
	writeU64(&b, uint64(TargetBlockSeconds))
	writeU64(&b, uint64(DifficultyWindow))
	writeU64(&b, uint64(MedianTimeWindow))
	writeU64(&b, uint64(MaxFutureSeconds))
	writeU64(&b, uint64(MaxBlockBytes))
	writeU64(&b, uint64(MaxTxBytes))
	writeU64(&b, uint64(MaxPubKeyBytes))
	writeU64(&b, uint64(MaxSignatureBytes))
	writeU16(&b, TxVersion)
	writeU16(&b, BlockVersion)
	b.Write(PowLimit[:])
	writeU16(&b, PowAlgorithmAQM64)
	writeU32(&b, AQM64MemoryKiB)
	writeU32(&b, AQM64TimeCost)
	b.WriteByte(AQM64Parallelism)
	writeU16(&b, 2) // crypto-agility registry version
	for _, scheme := range []uint16{SchemeMLDSA87} {
		info, _ := SignatureScheme(scheme)
		writeU16(&b, scheme)
		writeU32(&b, uint32(info.PublicKeySize))
		writeU32(&b, uint32(info.SignatureSize))
	}
	return Hash512(b.Bytes())
}
func (n *NetworkConfig) NetworkID() Hash {
	g := n.Genesis.Hash()
	f := ConsensusFingerprint()
	b := append([]byte("AURONQ_NETWORK_ID_V1\x00"), g[:]...)
	b = append(b, f[:]...)
	writeU64Buf := func(v uint64) { var x bytes.Buffer; writeU64(&x, v); b = append(b, x.Bytes()...) }
	writeU64Buf(n.Maturity())
	var pv bytes.Buffer
	writeU32(&pv, n.ProtocolVersion)
	b = append(b, pv.Bytes()...)
	b = append(b, n.NetworkByte)
	return Hash512(b)
}

func LoadNetwork(path string) (*NetworkConfig, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var n NetworkConfig
	if err = json.Unmarshal(b, &n); err != nil {
		return nil, err
	}
	if n.CoinbaseMaturity == 0 {
		if n.NetworkByte == TestnetNetworkByte {
			n.CoinbaseMaturity = TestnetCoinbaseMaturity
		} else {
			n.CoinbaseMaturity = MainnetCoinbaseMaturity
		}
	}
	if n.ProtocolVersion != 1 {
		return nil, fmt.Errorf("unsupported protocol version %d", n.ProtocolVersion)
	}
	if err := ValidateGenesis(&n); err != nil {
		return nil, err
	}
	return &n, nil
}
func SaveNetwork(path string, n *NetworkConfig) error {
	b, _ := json.MarshalIndent(n, "", "  ")
	return atomicWrite(path, b, 0644)
}

func ValidateGenesis(n *NetworkConfig) error {
	if n.CoinbaseMaturity == 0 {
		n.CoinbaseMaturity = n.Maturity()
	}
	if n.CoinbaseMaturity < 1 || n.CoinbaseMaturity > 100000 {
		return errors.New("invalid coinbase maturity")
	}
	g := &n.Genesis
	if g.Header.Version != BlockVersion || g.Header.PowAlgo != PowAlgorithmAQM64 || g.Header.Height != 0 || !g.Header.PrevHash.IsZero() {
		return errors.New("invalid genesis header")
	}
	if g.Header.Target != PowLimit {
		return errors.New("genesis target must equal consensus pow limit")
	}
	if g.Header.MerkleRoot != MerkleRoot(g.Transactions) {
		return errors.New("invalid genesis merkle root")
	}
	if len(g.Transactions) != 1 {
		return errors.New("genesis must contain exactly one allocation transaction")
	}
	tx := g.Transactions[0]
	if err := tx.ValidateBasic(); err != nil {
		return fmt.Errorf("invalid genesis transaction: %w", err)
	}
	if !tx.Coinbase || tx.CoinbaseHeight != 0 || tx.LockHeight != 0 || len(tx.Inputs) != 0 || len(tx.Outputs) != 1 || tx.Outputs[0].Value != FounderAtoms {
		return errors.New("invalid founder allocation in genesis")
	}
	if !IsConsensusScheme(tx.Outputs[0].Scheme) {
		return errors.New("invalid genesis signature scheme")
	}
	if n.FounderAddress == "" {
		return errors.New("missing founder address metadata")
	}
	net, scheme, kh, err := DecodeAddress(n.FounderAddress)
	if err != nil || net != n.NetworkByte || scheme != tx.Outputs[0].Scheme || kh != tx.Outputs[0].KeyHash {
		return errors.New("founder address metadata does not match genesis allocation")
	}
	pow, err := PowHash(g.Header)
	if err != nil || pow.Big().Cmp(PowLimit.Big()) > 0 {
		return errors.New("genesis proof of work invalid")
	}
	return nil
}

type MempoolEntry struct {
	Tx    Transaction
	Fee   uint64
	Added int64
}
type UTXORecord struct {
	OutPoint OutPoint `json:"outpoint"`
	UTXO     UTXO     `json:"utxo"`
}

type Chain struct {
	mu           sync.RWMutex
	dir          string
	network      *NetworkConfig
	blocks       []Block
	utxos        map[string]UTXO
	state        ChainState
	mempool      map[string]MempoolEntry
	mempoolBytes int
}

func OpenChain(dir string, network *NetworkConfig) (*Chain, error) {
	c := &Chain{dir: dir, network: network, utxos: map[string]UTXO{}, mempool: map[string]MempoolEntry{}}
	if err := os.MkdirAll(filepath.Join(dir, "blocks"), 0755); err != nil {
		return nil, err
	}
	if err := c.loadOrRebuild(); err != nil {
		return nil, err
	}
	return c, nil
}

// EstimatedNetworkHashrate estimates recent AQM64 work per second from canonical blocks.
// It is an observed rolling estimate, not a direct measurement of every miner.
func (c *Chain) EstimatedNetworkHashrate(window int) float64 {
	if window <= 0 {
		window = DifficultyWindow
	}
	if window > 240 {
		window = 240
	}
	c.mu.RLock()
	count := len(c.blocks)
	if count < 2 {
		c.mu.RUnlock()
		return 0
	}
	intervals := window
	if max := count - 1; intervals > max {
		intervals = max
	}
	start := count - 1 - intervals
	blocks := append([]Block(nil), c.blocks[start:]...)
	c.mu.RUnlock()

	elapsed := blocks[len(blocks)-1].Header.Timestamp - blocks[0].Header.Timestamp
	if elapsed <= 0 {
		return 0
	}
	work := new(big.Int)
	for i := 1; i < len(blocks); i++ {
		work.Add(work, WorkForTarget(blocks[i].Header.Target))
	}
	if work.Sign() <= 0 {
		return 0
	}
	rate := new(big.Float).Quo(new(big.Float).SetInt(work), big.NewFloat(float64(elapsed)))
	v, _ := rate.Float64()
	return v
}

func (c *Chain) loadOrRebuild() error {
	// Consensus state is replayed from canonical block files on every start.
	// This deliberately favors integrity over fast startup: a modified interior
	// block cannot be hidden behind a stale/corrupt UTXO snapshot.
	return c.rebuild()
}

func loadBlockFiles(dir string) ([]Block, error) {
	files, err := filepath.Glob(filepath.Join(dir, "blocks", "*.json"))
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	out := make([]Block, 0, len(files))
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}
		var bl Block
		if err = json.Unmarshal(b, &bl); err != nil {
			return nil, fmt.Errorf("%s: %w", f, err)
		}
		out = append(out, bl)
	}
	return out, nil
}

func (c *Chain) rebuild() error {
	files, _ := filepath.Glob(filepath.Join(c.dir, "blocks", "*.json"))
	sort.Strings(files)
	blocks := make([]Block, 0, len(files)+1)
	if len(files) == 0 {
		blocks = append(blocks, c.network.Genesis)
		if err := saveBlockFile(c.dir, &c.network.Genesis); err != nil {
			return err
		}
	} else {
		for _, f := range files {
			b, err := os.ReadFile(f)
			if err != nil {
				return err
			}
			var bl Block
			if err = json.Unmarshal(b, &bl); err != nil {
				return err
			}
			blocks = append(blocks, bl)
		}
		if len(blocks) == 0 || blocks[0].Hash() != c.network.Genesis.Hash() {
			return errors.New("data directory belongs to a different network")
		}
	}
	ut := map[string]UTXO{}
	st := ChainState{Height: 0, Tip: c.network.Genesis.Hash(), Issued: FounderAtoms, ChainWork: WorkForTarget(c.network.Genesis.Header.Target).Text(16)}
	if err := applyGenesisUTXO(ut, &c.network.Genesis); err != nil {
		return err
	}
	history := []Block{c.network.Genesis}
	for i := 1; i < len(blocks); i++ {
		if err := validateApplyBlock(ut, &st, &blocks[i], history, time.Now().Unix(), c.network.Maturity()); err != nil {
			return fmt.Errorf("invalid stored block %d: %w", i, err)
		}
		history = append(history, blocks[i])
	}
	c.blocks = blocks
	c.utxos = ut
	c.state = st
	return c.saveSnapshotLocked()
}

func applyGenesisUTXO(ut map[string]UTXO, g *Block) error {
	tx := g.Transactions[0]
	id := tx.ID()
	for i, o := range tx.Outputs {
		ut[OutPoint{id, uint32(i)}.Key()] = UTXO{Out: o, Height: 0, Coinbase: false}
	}
	return nil
}

func saveBlockFile(dir string, b *Block) error {
	x, _ := json.Marshal(b)
	return atomicWrite(filepath.Join(dir, "blocks", fmt.Sprintf("%012d.json", b.Header.Height)), x, 0644)
}
func atomicWrite(path string, b []byte, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, mode); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func (c *Chain) saveSnapshotLocked() error {
	s, _ := json.MarshalIndent(c.state, "", "  ")
	u, _ := json.Marshal(c.utxos)
	if err := atomicWrite(filepath.Join(c.dir, "state.json"), s, 0644); err != nil {
		return err
	}
	return atomicWrite(filepath.Join(c.dir, "utxo.json"), u, 0600)
}

func (c *Chain) State() ChainState { c.mu.RLock(); defer c.mu.RUnlock(); return c.state }
func (c *Chain) Height() uint64    { c.mu.RLock(); defer c.mu.RUnlock(); return c.state.Height }
func (c *Chain) Tip() Hash         { c.mu.RLock(); defer c.mu.RUnlock(); return c.state.Tip }
func (c *Chain) Block(height uint64) (Block, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if height >= uint64(len(c.blocks)) {
		return Block{}, false
	}
	return c.blocks[height], true
}

// Headers returns a bounded canonical header range beginning at start.
func (c *Chain) Headers(start uint64, limit int) []BlockHeader {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if limit <= 0 || start >= uint64(len(c.blocks)) {
		return nil
	}
	if limit > 256 {
		limit = 256
	}
	end := start + uint64(limit)
	if end > uint64(len(c.blocks)) {
		end = uint64(len(c.blocks))
	}
	out := make([]BlockHeader, 0, end-start)
	for i := start; i < end; i++ {
		out = append(out, c.blocks[i].Header)
	}
	return out
}
func (c *Chain) NetworkID() Hash         { return c.network.NetworkID() }
func (c *Chain) Network() *NetworkConfig { return c.network }

func expectedTargetHeaders(history []BlockHeader, height uint64) (Target, error) {
	if height == 0 {
		return PowLimit, nil
	}
	if len(history) == 0 {
		return Target{}, errors.New("no history")
	}
	prev := history[len(history)-1]
	if height == 1 || len(history) < 2 {
		return prev.Target, nil
	}

	// LWMA-inspired per-block adjustment. It reacts quickly to a new network's
	// hashrate while smoothing timestamp noise over up to 60 solved blocks.
	n := len(history) - 1
	if n > DifficultyWindow {
		n = DifficultyWindow
	}
	start := len(history) - 1 - n
	var weighted int64
	var weightSum int64
	targetSum := new(big.Int)
	for j := 1; j <= n; j++ {
		i := start + j
		solve := history[i].Timestamp - history[i-1].Timestamp
		if solve < 1 {
			solve = 1
		}
		if solve > TargetBlockSeconds*6 {
			solve = TargetBlockSeconds * 6
		}
		w := int64(j)
		weighted += solve * w
		weightSum += w
		targetSum.Add(targetSum, history[i].Target.Big())
	}
	if weightSum == 0 || n == 0 {
		return prev.Target, nil
	}
	avgSolve := weighted / weightSum
	if avgSolve < 1 {
		avgSolve = 1
	}
	avgTarget := new(big.Int).Div(targetSum, big.NewInt(int64(n)))
	next := new(big.Int).Mul(avgTarget, big.NewInt(avgSolve))
	next.Div(next, big.NewInt(TargetBlockSeconds))

	// Prevent pathological one-block jumps and never exceed the launch pow limit.
	prevBig := prev.Target.Big()
	minT := new(big.Int).Div(new(big.Int).Set(prevBig), big.NewInt(4))
	maxT := new(big.Int).Mul(new(big.Int).Set(prevBig), big.NewInt(4))
	if next.Cmp(minT) < 0 {
		next = minT
	}
	if next.Cmp(maxT) > 0 {
		next = maxT
	}
	if next.Cmp(PowLimit.Big()) > 0 {
		next = PowLimit.Big()
	}
	if next.Sign() <= 0 {
		next = big.NewInt(1)
	}
	return TargetFromBig(next), nil
}

func expectedTarget(history []Block, height uint64) (Target, error) {
	headers := make([]BlockHeader, len(history))
	for i := range history {
		headers[i] = history[i].Header
	}
	return expectedTargetHeaders(headers, height)
}

func medianTimePastHeaders(history []BlockHeader) int64 {
	n := MedianTimeWindow
	if len(history) < n {
		n = len(history)
	}
	ts := make([]int64, n)
	for i := 0; i < n; i++ {
		ts[i] = history[len(history)-n+i].Timestamp
	}
	sort.Slice(ts, func(i, j int) bool { return ts[i] < ts[j] })
	return ts[n/2]
}

func medianTimePast(history []Block) int64 {
	headers := make([]BlockHeader, len(history))
	for i := range history {
		headers[i] = history[i].Header
	}
	return medianTimePastHeaders(headers)
}

// ValidateHeaderEnvelope validates the consensus-critical properties that can be
// checked from headers alone. It is shared with light clients so they do not
// need a second implementation of AuronQ difficulty/timestamp/PoW rules.
//
// This does not validate transactions, UTXO state, coinbase value or Merkle
// contents and therefore is not a substitute for full-node validation.
func ValidateHeaderEnvelope(header BlockHeader, history []BlockHeader, now int64) error {
	if len(history) == 0 {
		return errors.New("header history is empty")
	}
	prev := history[len(history)-1]
	if header.Version != BlockVersion {
		return errors.New("unsupported block version")
	}
	if header.PowAlgo != PowAlgorithmAQM64 {
		return errors.New("unsupported proof-of-work algorithm")
	}
	if header.Height != prev.Height+1 {
		return fmt.Errorf("unexpected height %d", header.Height)
	}
	if header.PrevHash != prev.Hash() {
		return errors.New("previous hash mismatch")
	}
	target, err := expectedTargetHeaders(history, header.Height)
	if err != nil {
		return err
	}
	if header.Target != target {
		return errors.New("incorrect target")
	}
	pow, err := PowHash(header)
	if err != nil || pow.Big().Cmp(header.Target.Big()) > 0 {
		return errors.New("insufficient proof of work")
	}
	if header.Timestamp <= medianTimePastHeaders(history) {
		return errors.New("timestamp not greater than median time past")
	}
	if header.Timestamp > now+MaxFutureSeconds {
		return errors.New("timestamp too far in future")
	}
	return nil
}

func validateApplyBlock(base map[string]UTXO, state *ChainState, b *Block, history []Block, now int64, maturity uint64) error {
	if b.Header.Version != BlockVersion {
		return errors.New("unsupported block version")
	}
	if b.Header.PowAlgo != PowAlgorithmAQM64 {
		return errors.New("unsupported proof-of-work algorithm")
	}
	if b.Header.Height != state.Height+1 {
		return fmt.Errorf("unexpected height %d", b.Header.Height)
	}
	if b.Header.PrevHash != state.Tip {
		return errors.New("previous hash mismatch")
	}
	if b.Size() > MaxBlockBytes {
		return fmt.Errorf("block exceeds %d bytes", MaxBlockBytes)
	}
	if len(b.Transactions) == 0 {
		return errors.New("block has no transactions")
	}
	if b.Transactions[0].Coinbase == false {
		return errors.New("first transaction must be coinbase")
	}
	for i := 1; i < len(b.Transactions); i++ {
		if b.Transactions[i].Coinbase {
			return errors.New("multiple coinbase transactions")
		}
	}
	if b.Header.MerkleRoot != MerkleRoot(b.Transactions) {
		return errors.New("merkle root mismatch")
	}
	expTarget, err := expectedTarget(history, b.Header.Height)
	if err != nil {
		return err
	}
	if b.Header.Target != expTarget {
		return errors.New("incorrect target")
	}
	pow, err := PowHash(b.Header)
	if err != nil || pow.Big().Cmp(b.Header.Target.Big()) > 0 {
		return errors.New("insufficient proof of work")
	}
	mtp := medianTimePast(history)
	if b.Header.Timestamp <= mtp {
		return errors.New("timestamp not greater than median time past")
	}
	if b.Header.Timestamp > now+MaxFutureSeconds {
		return errors.New("timestamp too far in future")
	}
	overlay := map[string]*UTXO{}
	lookup := func(k string) (UTXO, bool) {
		if v, ok := overlay[k]; ok {
			if v == nil {
				return UTXO{}, false
			}
			return *v, true
		}
		v, ok := base[k]
		return v, ok
	}
	var fees uint64
	seenTx := map[string]bool{}
	for ti := 1; ti < len(b.Transactions); ti++ {
		tx := &b.Transactions[ti]
		if err := tx.ValidateBasic(); err != nil {
			return fmt.Errorf("tx %d: %w", ti, err)
		}
		id := tx.ID().String()
		if seenTx[id] {
			return errors.New("duplicate txid in block")
		}
		seenTx[id] = true
		if tx.LockHeight > b.Header.Height {
			return errors.New("transaction lock height not reached")
		}
		var inSum, outSum uint64
		for oi, o := range tx.Outputs {
			if ^uint64(0)-outSum < o.Value {
				return errors.New("output overflow")
			}
			outSum += o.Value
			_ = oi
		}
		for ii, in := range tx.Inputs {
			key := OutPoint{in.PrevTx, in.PrevIndex}.Key()
			u, ok := lookup(key)
			if !ok {
				return fmt.Errorf("missing/spent input %s", key)
			}
			if u.Coinbase && b.Header.Height < u.Height+maturity {
				return errors.New("immature coinbase spend")
			}
			info, ok := SignatureScheme(u.Out.Scheme)
			if !ok {
				return errors.New("unsupported previous output scheme")
			}
			if len(in.PublicKey) != info.PublicKeySize || Hash256(in.PublicKey) != u.Out.KeyHash {
				return errors.New("public key does not satisfy output")
			}
			if len(in.Signature) != info.SignatureSize {
				return errors.New("invalid signature size")
			}
			sigh := tx.SigHash(ii, u.Out)
			if !VerifySignature(u.Out.Scheme, in.PublicKey, sigh[:], in.Signature) {
				return errors.New("invalid signature")
			}
			if ^uint64(0)-inSum < u.Out.Value {
				return errors.New("input overflow")
			}
			inSum += u.Out.Value
			overlay[key] = nil
		}
		if inSum < outSum {
			return errors.New("transaction spends more than inputs")
		}
		fee := inSum - outSum
		if ^uint64(0)-fees < fee {
			return errors.New("fee overflow")
		}
		fees += fee
		txid := tx.ID()
		for oi, o := range tx.Outputs {
			u := UTXO{Out: o, Height: b.Header.Height, Coinbase: false}
			overlay[OutPoint{txid, uint32(oi)}.Key()] = &u
		}
	}
	cb := &b.Transactions[0]
	if err := cb.ValidateBasic(); err != nil {
		return fmt.Errorf("coinbase: %w", err)
	}
	if cb.CoinbaseHeight != b.Header.Height {
		return errors.New("coinbase height mismatch")
	}
	var cbSum uint64
	for _, o := range cb.Outputs {
		if ^uint64(0)-cbSum < o.Value {
			return errors.New("coinbase overflow")
		}
		cbSum += o.Value
	}
	subsidy := BlockSubsidy(b.Header.Height, state.Issued)
	expected := subsidy + fees
	if expected < subsidy {
		return errors.New("reward overflow")
	}
	if cbSum != expected {
		return fmt.Errorf("coinbase amount %d, expected %d", cbSum, expected)
	}
	cbid := cb.ID()
	for oi, o := range cb.Outputs {
		u := UTXO{Out: o, Height: b.Header.Height, Coinbase: true}
		overlay[OutPoint{cbid, uint32(oi)}.Key()] = &u
	}
	for k, v := range overlay {
		if v == nil {
			delete(base, k)
		} else {
			base[k] = *v
		}
	}
	state.Height = b.Header.Height
	state.Tip = b.Hash()
	state.Issued += subsidy
	cw := new(big.Int)
	cw.SetString(state.ChainWork, 16)
	cw.Add(cw, WorkForTarget(b.Header.Target))
	state.ChainWork = cw.Text(16)
	return nil
}

func captureAffected(base map[string]UTXO, b *Block) map[string]*UTXO {
	old := make(map[string]*UTXO)
	capture := func(k string) {
		if _, seen := old[k]; seen {
			return
		}
		if u, ok := base[k]; ok {
			v := u
			old[k] = &v
		} else {
			old[k] = nil
		}
	}
	for _, tx := range b.Transactions {
		for _, in := range tx.Inputs {
			capture(OutPoint{in.PrevTx, in.PrevIndex}.Key())
		}
		id := tx.ID()
		for i := range tx.Outputs {
			capture(OutPoint{id, uint32(i)}.Key())
		}
	}
	return old
}
func restoreAffected(base map[string]UTXO, old map[string]*UTXO) {
	for k, v := range old {
		if v == nil {
			delete(base, k)
		} else {
			base[k] = *v
		}
	}
}

func (c *Chain) AddBlock(b *Block) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	oldState := c.state
	oldUTXO := captureAffected(c.utxos, b)
	if err := validateApplyBlock(c.utxos, &c.state, b, c.blocks, time.Now().Unix(), c.network.Maturity()); err != nil {
		return err
	}
	if err := saveBlockFile(c.dir, b); err != nil {
		restoreAffected(c.utxos, oldUTXO)
		c.state = oldState
		return err
	}
	c.blocks = append(c.blocks, *b)
	// Snapshots are acceleration artifacts only; canonical block files are the
	// source of truth and are fully replayed on restart.
	_ = c.saveSnapshotLocked()
	// Remove confirmed and conflicting transactions.
	spent := map[string]bool{}
	confirmed := map[string]bool{}
	for _, tx := range b.Transactions {
		confirmed[tx.ID().String()] = true
		for _, in := range tx.Inputs {
			spent[OutPoint{in.PrevTx, in.PrevIndex}.Key()] = true
		}
	}
	for id, e := range c.mempool {
		remove := confirmed[id]
		if !remove {
			for _, in := range e.Tx.Inputs {
				if spent[OutPoint{in.PrevTx, in.PrevIndex}.Key()] {
					remove = true
					break
				}
			}
		}
		if remove {
			sz := e.Tx.BaseSize()
			if c.mempoolBytes >= sz {
				c.mempoolBytes -= sz
			} else {
				c.mempoolBytes = 0
			}
			delete(c.mempool, id)
		}
	}
	return nil
}

func (c *Chain) txFeeAgainstUTXO(tx *Transaction) (uint64, error) {
	var inSum, outSum uint64
	for _, o := range tx.Outputs {
		outSum += o.Value
	}
	for ii, in := range tx.Inputs {
		u, ok := c.utxos[OutPoint{in.PrevTx, in.PrevIndex}.Key()]
		if !ok {
			return 0, errors.New("input missing or spent")
		}
		if u.Coinbase && c.state.Height+1 < u.Height+c.network.Maturity() {
			return 0, errors.New("immature coinbase")
		}
		if Hash256(in.PublicKey) != u.Out.KeyHash {
			return 0, errors.New("wrong public key")
		}
		h := tx.SigHash(ii, u.Out)
		if !VerifySignature(u.Out.Scheme, in.PublicKey, h[:], in.Signature) {
			return 0, errors.New("invalid signature")
		}
		inSum += u.Out.Value
	}
	if inSum < outSum {
		return 0, errors.New("overspend")
	}
	return inSum - outSum, nil
}

func (c *Chain) addMempoolLocked(tx Transaction) (uint64, error) {
	if tx.Coinbase {
		return 0, errors.New("coinbase not allowed in mempool")
	}
	if err := tx.ValidateBasic(); err != nil {
		return 0, err
	}
	if tx.LockHeight > c.state.Height+1 {
		return 0, errors.New("transaction is timelocked")
	}
	id := tx.ID().String()
	if _, ok := c.mempool[id]; ok {
		return 0, errors.New("transaction already in mempool")
	}
	used := map[string]bool{}
	for _, e := range c.mempool {
		for _, in := range e.Tx.Inputs {
			used[OutPoint{in.PrevTx, in.PrevIndex}.Key()] = true
		}
	}
	for _, in := range tx.Inputs {
		if used[OutPoint{in.PrevTx, in.PrevIndex}.Key()] {
			return 0, errors.New("mempool double spend")
		}
	}
	fee, err := c.txFeeAgainstUTXO(&tx)
	if err != nil {
		return 0, err
	}
	min := uint64((tx.BaseSize()+999)/1000) * MinRelayFeePerKB
	if fee < min {
		return 0, fmt.Errorf("fee %d below minimum %d", fee, min)
	}
	if len(c.mempool) >= MaxMempoolTx {
		return 0, errors.New("mempool transaction limit reached")
	}
	sz := tx.BaseSize()
	if sz > MaxMempoolBytes || c.mempoolBytes > MaxMempoolBytes-sz {
		return 0, errors.New("mempool byte limit reached")
	}
	c.mempool[id] = MempoolEntry{Tx: tx, Fee: fee, Added: time.Now().Unix()}
	c.mempoolBytes += sz
	return fee, nil
}

func (c *Chain) AddMempool(tx Transaction) (uint64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.addMempoolLocked(tx)
}

func (c *Chain) BuildTemplate(minerAddress string) (Block, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	net, scheme, kh, err := DecodeAddress(minerAddress)
	if err != nil {
		return Block{}, err
	}
	if net != c.network.NetworkByte {
		return Block{}, errors.New("miner address belongs to another network")
	}
	if !IsConsensusScheme(scheme) {
		return Block{}, errors.New("unsupported miner address scheme")
	}
	txs := make([]Transaction, 0, len(c.mempool)+1)
	var fees uint64
	ids := SortedTxIDs(func() map[string]Transaction {
		m := map[string]Transaction{}
		for k, e := range c.mempool {
			m[k] = e.Tx
		}
		return m
	}())
	size := 0
	for _, id := range ids {
		e := c.mempool[id]
		if size+e.Tx.BaseSize() > MaxBlockBytes-64*1024 {
			break
		}
		txs = append(txs, e.Tx)
		fees += e.Fee
		size += e.Tx.BaseSize()
	}
	height := c.state.Height + 1
	sub := BlockSubsidy(height, c.state.Issued)
	cb := Transaction{Version: TxVersion, Coinbase: true, CoinbaseHeight: height, Outputs: []TxOutput{{Value: sub + fees, Scheme: scheme, KeyHash: kh}}, LockHeight: 0}
	txs = append([]Transaction{cb}, txs...)
	target, err := expectedTarget(c.blocks, height)
	if err != nil {
		return Block{}, err
	}
	ts := time.Now().Unix()
	mtp := medianTimePast(c.blocks)
	if ts <= mtp {
		ts = mtp + 1
	}
	b := Block{Header: BlockHeader{Version: BlockVersion, PowAlgo: PowAlgorithmAQM64, Height: height, PrevHash: c.state.Tip, Timestamp: ts, Target: target}, Transactions: txs}
	b.Header.MerkleRoot = MerkleRoot(b.Transactions)
	return b, nil
}

func (c *Chain) UTXOsForAddress(addr string, includeImmature bool) ([]UTXORecord, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	net, scheme, kh, err := DecodeAddress(addr)
	if err != nil {
		return nil, err
	}
	if net != c.network.NetworkByte {
		return nil, errors.New("address network mismatch")
	}
	out := []UTXORecord{}
	for k, u := range c.utxos {
		if u.Out.Scheme != scheme || u.Out.KeyHash != kh {
			continue
		}
		if !includeImmature && u.Coinbase && c.state.Height+1 < u.Height+c.network.Maturity() {
			continue
		}
		parts := strings.Split(k, ":")
		if len(parts) != 2 {
			continue
		}
		hb, _ := hexToHash(parts[0])
		idx64, _ := strconv.ParseUint(parts[1], 16, 32)
		out = append(out, UTXORecord{OutPoint: OutPoint{hb, uint32(idx64)}, UTXO: u})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].UTXO.Height == out[j].UTXO.Height {
			return out[i].OutPoint.Key() < out[j].OutPoint.Key()
		}
		return out[i].UTXO.Height < out[j].UTXO.Height
	})
	return out, nil
}
func hexToHash(s string) (Hash, error) {
	var h Hash
	b, err := fmtHexDecode(s)
	if err != nil {
		return h, err
	}
	if len(b) != 64 {
		return h, errors.New("bad hash")
	}
	copy(h[:], b)
	return h, nil
}
func fmtHexDecode(s string) ([]byte, error) { // local helper to avoid accidental variable-length parsing elsewhere
	if len(s)%2 != 0 {
		return nil, errors.New("odd hex")
	}
	b := make([]byte, len(s)/2)
	for i := 0; i < len(b); i++ {
		v, err := strconv.ParseUint(s[i*2:i*2+2], 16, 8)
		if err != nil {
			return nil, err
		}
		b[i] = byte(v)
	}
	return b, nil
}
func (c *Chain) Balance(addr string) (spendable, total uint64, err error) {
	all, e := c.UTXOsForAddress(addr, true)
	if e != nil {
		return 0, 0, e
	}
	height := c.Height()
	for _, r := range all {
		total += r.UTXO.Out.Value
		if !r.UTXO.Coinbase || height+1 >= r.UTXO.Height+c.network.Maturity() {
			spendable += r.UTXO.Out.Value
		}
	}
	return
}

func (c *Chain) ReplaceCanonical(candidate []Block) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(candidate) == 0 || candidate[0].Hash() != c.network.Genesis.Hash() {
		return errors.New("candidate genesis mismatch")
	}
	ut := map[string]UTXO{}
	applyGenesisUTXO(ut, &candidate[0])
	st := ChainState{Height: 0, Tip: candidate[0].Hash(), Issued: FounderAtoms, ChainWork: WorkForTarget(candidate[0].Header.Target).Text(16)}
	hist := []Block{candidate[0]}
	for i := 1; i < len(candidate); i++ {
		if err := validateApplyBlock(ut, &st, &candidate[i], hist, time.Now().Unix(), c.network.Maturity()); err != nil {
			return fmt.Errorf("candidate block %d invalid: %w", i, err)
		}
		hist = append(hist, candidate[i])
	}
	cur := new(big.Int)
	cur.SetString(c.state.ChainWork, 16)
	nw := new(big.Int)
	nw.SetString(st.ChainWork, 16)
	if nw.Cmp(cur) <= 0 {
		return errors.New("candidate chain does not have more cumulative work")
	}

	// Preserve transactions that may become valid again after a reorganization.
	// We first remember the old mempool and all non-coinbase transactions from
	// disconnected blocks, then revalidate them against the new canonical UTXO set.
	requeue := make([]Transaction, 0, len(c.mempool)+64)
	oldEntries := make([]MempoolEntry, 0, len(c.mempool))
	for _, e := range c.mempool {
		oldEntries = append(oldEntries, e)
	}
	sort.Slice(oldEntries, func(i, j int) bool {
		if oldEntries[i].Added != oldEntries[j].Added {
			return oldEntries[i].Added < oldEntries[j].Added
		}
		return oldEntries[i].Tx.ID().String() < oldEntries[j].Tx.ID().String()
	})
	for _, e := range oldEntries {
		requeue = append(requeue, e.Tx)
	}
	newTx := map[string]struct{}{}
	for _, b := range candidate {
		for i := range b.Transactions {
			newTx[b.Transactions[i].ID().String()] = struct{}{}
		}
	}
	for _, b := range c.blocks {
		for i := 1; i < len(b.Transactions); i++ {
			tx := b.Transactions[i]
			if _, confirmed := newTx[tx.ID().String()]; !confirmed {
				requeue = append(requeue, tx)
			}
		}
	}

	// Rewrite canonical block files only after full validation.
	tmpDir := filepath.Join(c.dir, "blocks.new")
	os.RemoveAll(tmpDir)
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return err
	}
	for _, b := range candidate {
		x, _ := json.Marshal(b)
		if err := os.WriteFile(filepath.Join(tmpDir, fmt.Sprintf("%012d.json", b.Header.Height)), x, 0644); err != nil {
			return err
		}
	}
	old := filepath.Join(c.dir, "blocks.old")
	os.RemoveAll(old)
	if err := os.Rename(filepath.Join(c.dir, "blocks"), old); err != nil {
		return err
	}
	if err := os.Rename(tmpDir, filepath.Join(c.dir, "blocks")); err != nil {
		_ = os.Rename(old, filepath.Join(c.dir, "blocks"))
		return err
	}
	os.RemoveAll(old)
	c.blocks = candidate
	c.utxos = ut
	c.state = st
	c.mempool = map[string]MempoolEntry{}
	c.mempoolBytes = 0
	for _, tx := range requeue {
		_, _ = c.addMempoolLocked(tx)
	}
	return c.saveSnapshotLocked()
}

func ChainWorkBig(st ChainState) *big.Int { x := new(big.Int); x.SetString(st.ChainWork, 16); return x }

func (c *Chain) BlocksThrough(height uint64) []Block {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if height >= uint64(len(c.blocks)) {
		height = uint64(len(c.blocks) - 1)
	}
	out := make([]Block, height+1)
	copy(out, c.blocks[:height+1])
	return out
}

// RecentBlocksThrough returns at most count canonical blocks ending at height.
// It is used to validate a candidate fork header without copying the whole chain.
func (c *Chain) RecentBlocksThrough(height uint64, count int) []Block {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if len(c.blocks) == 0 || count <= 0 {
		return nil
	}
	if height >= uint64(len(c.blocks)) {
		height = uint64(len(c.blocks) - 1)
	}
	start := uint64(0)
	if height+1 > uint64(count) {
		start = height + 1 - uint64(count)
	}
	out := make([]Block, height-start+1)
	copy(out, c.blocks[start:height+1])
	return out
}

// WorkAfter returns cumulative proof-of-work for the canonical suffix strictly
// above height. The shared prefix cancels when comparing a competing fork.
func (c *Chain) WorkAfter(height uint64) *big.Int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := new(big.Int)
	if len(c.blocks) == 0 || height >= uint64(len(c.blocks)-1) {
		return out
	}
	for i := height + 1; i < uint64(len(c.blocks)); i++ {
		out.Add(out, WorkForTarget(c.blocks[i].Header.Target))
	}
	return out
}

// ValidateCandidateEnvelope performs all expensive/header-level consensus checks
// that do not require the fork's UTXO state. A remote peer must therefore supply
// valid proof-of-work before it can trigger a full reorganization replay.
func ValidateCandidateEnvelope(b *Block, history []Block, now int64) error {
	if len(history) == 0 {
		return errors.New("candidate history is empty")
	}
	prev := history[len(history)-1]
	if b.Header.Version != BlockVersion {
		return errors.New("unsupported block version")
	}
	if b.Header.PowAlgo != PowAlgorithmAQM64 {
		return errors.New("unsupported proof-of-work algorithm")
	}
	if b.Header.Height != prev.Header.Height+1 {
		return fmt.Errorf("unexpected height %d", b.Header.Height)
	}
	if b.Header.PrevHash != prev.Hash() {
		return errors.New("previous hash mismatch")
	}
	if b.Size() > MaxBlockBytes {
		return fmt.Errorf("block exceeds %d bytes", MaxBlockBytes)
	}
	if len(b.Transactions) == 0 || !b.Transactions[0].Coinbase {
		return errors.New("first transaction must be coinbase")
	}
	for i := 1; i < len(b.Transactions); i++ {
		if b.Transactions[i].Coinbase {
			return errors.New("multiple coinbase transactions")
		}
	}
	if b.Header.MerkleRoot != MerkleRoot(b.Transactions) {
		return errors.New("merkle root mismatch")
	}
	target, err := expectedTarget(history, b.Header.Height)
	if err != nil {
		return err
	}
	if b.Header.Target != target {
		return errors.New("incorrect target")
	}
	pow, err := PowHash(b.Header)
	if err != nil || pow.Big().Cmp(b.Header.Target.Big()) > 0 {
		return errors.New("insufficient proof of work")
	}
	if b.Header.Timestamp <= medianTimePast(history) {
		return errors.New("timestamp not greater than median time past")
	}
	if b.Header.Timestamp > now+MaxFutureSeconds {
		return errors.New("timestamp too far in future")
	}
	return nil
}

func (c *Chain) MempoolSize() int { c.mu.RLock(); defer c.mu.RUnlock(); return len(c.mempool) }

// MempoolTransactions returns a deterministic snapshot of pending transactions
// capped by their consensus-encoded base size. It is used for peer catch-up so
// a node that joins after a transaction was first broadcast can still learn it.
func (c *Chain) MempoolTransactions(maxBytes int) []Transaction {
	c.mu.RLock()
	defer c.mu.RUnlock()
	type item struct {
		id string
		e  MempoolEntry
	}
	items := make([]item, 0, len(c.mempool))
	for id, e := range c.mempool {
		items = append(items, item{id: id, e: e})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].e.Added != items[j].e.Added {
			return items[i].e.Added < items[j].e.Added
		}
		return items[i].id < items[j].id
	})
	out := make([]Transaction, 0, len(items))
	used := 0
	for _, it := range items {
		sz := it.e.Tx.BaseSize()
		if maxBytes > 0 && used+sz > maxBytes {
			break
		}
		used += sz
		out = append(out, it.e.Tx)
	}
	return out
}