package auronq

import (
	"encoding/hex"
	"errors"
	"sort"
	"strings"
)

// ExplorerBlockSummary is a compact, read-only view of a canonical block.
type ExplorerBlockSummary struct {
	Height       uint64 `json:"height"`
	Hash         string `json:"hash"`
	PrevHash     string `json:"prev_hash"`
	Timestamp    int64  `json:"timestamp"`
	Transactions int    `json:"transactions"`
	Size         int    `json:"size"`
	Target       string `json:"target"`
	Work         string `json:"work"`
}

// ExplorerInput resolves a transaction input to the previous canonical output when possible.
type ExplorerInput struct {
	PrevTx     string `json:"prev_tx"`
	PrevIndex  uint32 `json:"prev_index"`
	ValueAtoms uint64 `json:"value_atoms,omitempty"`
	Value      string `json:"value,omitempty"`
	Address    string `json:"address,omitempty"`
}

// ExplorerOutput is a human-readable transaction output.
type ExplorerOutput struct {
	Index      uint32 `json:"index"`
	ValueAtoms uint64 `json:"value_atoms"`
	Value      string `json:"value"`
	Address    string `json:"address"`
	Scheme     uint16 `json:"scheme"`
}

// ExplorerTransaction is a read-only transaction view for the public explorer.
type ExplorerTransaction struct {
	TXID           string           `json:"txid"`
	Status         string           `json:"status"`
	Height         *uint64          `json:"height,omitempty"`
	Timestamp      int64            `json:"timestamp"`
	Confirmations  uint64           `json:"confirmations"`
	Coinbase       bool             `json:"coinbase"`
	CoinbaseHeight uint64           `json:"coinbase_height,omitempty"`
	FeeAtoms       uint64           `json:"fee_atoms"`
	Fee            string           `json:"fee"`
	Size           int              `json:"size"`
	Inputs         []ExplorerInput  `json:"inputs"`
	Outputs        []ExplorerOutput `json:"outputs"`
}

// ExplorerBlock is a detailed canonical block view.
type ExplorerBlock struct {
	Summary      ExplorerBlockSummary `json:"summary"`
	MerkleRoot   string               `json:"merkle_root"`
	PowAlgo      uint16               `json:"pow_algo"`
	Nonce        uint64               `json:"nonce"`
	Transactions []string             `json:"transactions"`
}

func explorerBlockSummary(b Block) ExplorerBlockSummary {
	return ExplorerBlockSummary{
		Height:       b.Header.Height,
		Hash:         b.Hash().String(),
		PrevHash:     b.Header.PrevHash.String(),
		Timestamp:    b.Header.Timestamp,
		Transactions: len(b.Transactions),
		Size:         b.Size(),
		Target:       b.Header.Target.String(),
		Work:         WorkForTarget(b.Header.Target).Text(16),
	}
}

func (c *Chain) ExplorerRecentBlocks(limit int) []ExplorerBlockSummary {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	c.mu.RLock()
	count := len(c.blocks)
	start := count - limit
	if start < 0 {
		start = 0
	}
	blocks := append([]Block(nil), c.blocks[start:]...)
	c.mu.RUnlock()

	out := make([]ExplorerBlockSummary, 0, len(blocks))
	for i := len(blocks) - 1; i >= 0; i-- {
		out = append(out, explorerBlockSummary(blocks[i]))
	}
	return out
}

func validExplorerHash(s string) bool {
	if len(s) != 128 {
		return false
	}
	_, err := hex.DecodeString(s)
	return err == nil
}

func (c *Chain) ExplorerBlock(height *uint64, hash string) (ExplorerBlock, bool, error) {
	hash = strings.ToLower(strings.TrimSpace(hash))
	if height == nil && hash == "" {
		return ExplorerBlock{}, false, errors.New("height or hash required")
	}
	if hash != "" && !validExplorerHash(hash) {
		return ExplorerBlock{}, false, errors.New("invalid block hash")
	}

	c.mu.RLock()
	defer c.mu.RUnlock()
	var b Block
	found := false
	if height != nil {
		if *height < uint64(len(c.blocks)) {
			b = c.blocks[*height]
			found = true
		}
	} else {
		for i := range c.blocks {
			if c.blocks[i].Hash().String() == hash {
				b = c.blocks[i]
				found = true
				break
			}
		}
	}
	if !found {
		return ExplorerBlock{}, false, nil
	}
	txids := make([]string, len(b.Transactions))
	for i := range b.Transactions {
		txids[i] = b.Transactions[i].ID().String()
	}
	return ExplorerBlock{
		Summary:      explorerBlockSummary(b),
		MerkleRoot:   b.Header.MerkleRoot.String(),
		PowAlgo:      b.Header.PowAlgo,
		Nonce:        b.Header.Nonce,
		Transactions: txids,
	}, true, nil
}

func explorerOutput(out TxOutput, index int, network byte) ExplorerOutput {
	return ExplorerOutput{
		Index:      uint32(index),
		ValueAtoms: out.Value,
		Value:      FormatAmount(out.Value),
		Address:    AddressFromKeyHash(out.KeyHash, out.Scheme, network),
		Scheme:     out.Scheme,
	}
}

func buildExplorerTransaction(tx Transaction, height *uint64, timestamp int64, tip uint64, network byte, outputs map[string]TxOutput, status string, knownFee *uint64) ExplorerTransaction {
	ins := make([]ExplorerInput, 0, len(tx.Inputs))
	var totalIn, totalOut uint64
	for _, in := range tx.Inputs {
		item := ExplorerInput{PrevTx: in.PrevTx.String(), PrevIndex: in.PrevIndex}
		if prev, ok := outputs[OutPoint{in.PrevTx, in.PrevIndex}.Key()]; ok {
			item.ValueAtoms = prev.Value
			item.Value = FormatAmount(prev.Value)
			item.Address = AddressFromKeyHash(prev.KeyHash, prev.Scheme, network)
			totalIn += prev.Value
		}
		ins = append(ins, item)
	}
	outs := make([]ExplorerOutput, len(tx.Outputs))
	for i, out := range tx.Outputs {
		outs[i] = explorerOutput(out, i, network)
		totalOut += out.Value
	}
	fee := uint64(0)
	if !tx.Coinbase {
		if knownFee != nil {
			fee = *knownFee
		} else if totalIn >= totalOut {
			fee = totalIn - totalOut
		}
	}
	confirmations := uint64(0)
	if height != nil && tip >= *height {
		confirmations = tip - *height + 1
	}
	return ExplorerTransaction{
		TXID:           tx.ID().String(),
		Status:         status,
		Height:         height,
		Timestamp:      timestamp,
		Confirmations:  confirmations,
		Coinbase:       tx.Coinbase,
		CoinbaseHeight: tx.CoinbaseHeight,
		FeeAtoms:       fee,
		Fee:            FormatAmount(fee),
		Size:           tx.BaseSize(),
		Inputs:         ins,
		Outputs:        outs,
	}
}

func (c *Chain) ExplorerTransaction(txid string) (ExplorerTransaction, bool, error) {
	txid = strings.ToLower(strings.TrimSpace(txid))
	if !validExplorerHash(txid) {
		return ExplorerTransaction{}, false, errors.New("invalid transaction id")
	}

	c.mu.RLock()
	blocks := append([]Block(nil), c.blocks...)
	mempool := make([]MempoolEntry, 0, len(c.mempool))
	for _, e := range c.mempool {
		mempool = append(mempool, e)
	}
	tip := c.state.Height
	network := byte(0)
	if c.network != nil {
		network = c.network.NetworkByte
	}
	c.mu.RUnlock()

	outputs := make(map[string]TxOutput)
	for bi := range blocks {
		b := blocks[bi]
		h := b.Header.Height
		for ti := range b.Transactions {
			tx := b.Transactions[ti]
			id := tx.ID()
			if id.String() == txid {
				return buildExplorerTransaction(tx, &h, b.Header.Timestamp, tip, network, outputs, "confirmed", nil), true, nil
			}
			for oi, out := range tx.Outputs {
				outputs[OutPoint{id, uint32(oi)}.Key()] = out
			}
		}
	}

	sort.Slice(mempool, func(i, j int) bool {
		if mempool[i].Added != mempool[j].Added {
			return mempool[i].Added < mempool[j].Added
		}
		return mempool[i].Tx.ID().String() < mempool[j].Tx.ID().String()
	})
	for i := range mempool {
		e := mempool[i]
		id := e.Tx.ID()
		if id.String() == txid {
			return buildExplorerTransaction(e.Tx, nil, e.Added, tip, network, outputs, "pending", &e.Fee), true, nil
		}
		for oi, out := range e.Tx.Outputs {
			outputs[OutPoint{id, uint32(oi)}.Key()] = out
		}
	}
	return ExplorerTransaction{}, false, nil
}
