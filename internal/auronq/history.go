package auronq

import (
	"errors"
	"sort"
)

// WalletHistoryItem is a wallet-centric view of a canonical or mempool transaction.
// Amount excludes the network fee for outgoing transactions; Net includes it.
type WalletHistoryItem struct {
	TXID          string  `json:"txid"`
	Type          string  `json:"type"`   // genesis, mining, received, sent
	Status        string  `json:"status"` // pending, immature, confirmed
	Height        *uint64 `json:"height,omitempty"`
	Timestamp     int64   `json:"timestamp"`
	AmountAtoms   uint64  `json:"amount_atoms"`
	Amount        string  `json:"amount"`
	FeeAtoms      uint64  `json:"fee_atoms"`
	Fee           string  `json:"fee"`
	NetAtoms      int64   `json:"net_atoms"`
	Net           string  `json:"net"`
	Confirmations uint64  `json:"confirmations"`
	Coinbase      bool    `json:"coinbase"`
	Inputs        int     `json:"inputs"`
	Outputs       int     `json:"outputs"`
}

func outputMatches(o TxOutput, scheme uint16, kh KeyHash) bool {
	return o.Scheme == scheme && o.KeyHash == kh
}

func signedAmount(v int64) string {
	if v < 0 {
		return "-" + FormatAmount(uint64(-v))
	}
	if v > 0 {
		return "+" + FormatAmount(uint64(v))
	}
	return FormatAmount(0)
}

// HistoryForAddress scans the current canonical chain and mempool. It deliberately
// derives history from consensus data instead of maintaining a separate wallet index,
// so a reorg cannot leave stale wallet history behind.
func (c *Chain) HistoryForAddress(addr string, limit int) ([]WalletHistoryItem, error) {
	net, scheme, kh, err := DecodeAddress(addr)
	if err != nil {
		return nil, err
	}

	c.mu.RLock()
	if c.network == nil || net != c.network.NetworkByte {
		c.mu.RUnlock()
		return nil, errors.New("address network mismatch")
	}
	blocks := append([]Block(nil), c.blocks...)
	mempool := make([]MempoolEntry, 0, len(c.mempool))
	for _, e := range c.mempool {
		mempool = append(mempool, e)
	}
	tip := c.state.Height
	maturity := c.network.Maturity()
	c.mu.RUnlock()

	if limit <= 0 || limit > 1000 {
		limit = 200
	}

	// Keep every canonical output, including spent outputs, because transaction
	// history needs to resolve the ownership and value of past inputs.
	outputs := make(map[string]TxOutput)
	items := make([]WalletHistoryItem, 0)

	inspect := func(tx Transaction, height *uint64, timestamp int64, pending bool, knownFee *uint64) {
		var incoming, outgoing, allIn, allOut uint64
		for _, in := range tx.Inputs {
			if prev, ok := outputs[OutPoint{in.PrevTx, in.PrevIndex}.Key()]; ok {
				allIn += prev.Value
				if outputMatches(prev, scheme, kh) {
					outgoing += prev.Value
				}
			}
		}
		for _, out := range tx.Outputs {
			allOut += out.Value
			if outputMatches(out, scheme, kh) {
				incoming += out.Value
			}
		}
		if incoming == 0 && outgoing == 0 {
			return
		}

		fee := uint64(0)
		if !tx.Coinbase {
			if knownFee != nil {
				fee = *knownFee
			} else if allIn >= allOut {
				fee = allIn - allOut
			}
		}

		kind := "received"
		amount := incoming
		netAtoms := int64(incoming) - int64(outgoing)
		status := "confirmed"
		confirmations := uint64(0)

		if pending {
			status = "pending"
		} else if height != nil {
			confirmations = tip - *height + 1
		}

		if tx.Coinbase {
			if height != nil && *height == 0 {
				kind = "genesis"
				status = "confirmed"
			} else {
				kind = "mining"
				if height != nil && tip+1 < *height+maturity {
					status = "immature"
				}
			}
			amount = incoming
		} else if outgoing > 0 {
			kind = "sent"
			// Wallet builders return change back to the same wallet. Subtract that
			// change and the fee so Amount represents what was actually sent away.
			if outgoing > incoming+fee {
				amount = outgoing - incoming - fee
			} else if outgoing > incoming {
				amount = outgoing - incoming
			} else {
				amount = 0
			}
		}

		h := height
		items = append(items, WalletHistoryItem{
			TXID:          tx.ID().String(),
			Type:          kind,
			Status:        status,
			Height:        h,
			Timestamp:     timestamp,
			AmountAtoms:   amount,
			Amount:        FormatAmount(amount),
			FeeAtoms:      fee,
			Fee:           FormatAmount(fee),
			NetAtoms:      netAtoms,
			Net:           signedAmount(netAtoms),
			Confirmations: confirmations,
			Coinbase:      tx.Coinbase,
			Inputs:        len(tx.Inputs),
			Outputs:       len(tx.Outputs),
		})
	}

	for bi := range blocks {
		b := blocks[bi]
		h := b.Header.Height
		for ti := range b.Transactions {
			tx := b.Transactions[ti]
			inspect(tx, &h, b.Header.Timestamp, false, nil)
			id := tx.ID()
			for oi, out := range tx.Outputs {
				outputs[OutPoint{id, uint32(oi)}.Key()] = out
			}
		}
	}

	for i := range mempool {
		e := mempool[i]
		fee := e.Fee
		inspect(e.Tx, nil, e.Added, true, &fee)
	}

	sort.SliceStable(items, func(i, j int) bool {
		// Pending first, then newest canonical transactions.
		pi := items[i].Status == "pending"
		pj := items[j].Status == "pending"
		if pi != pj {
			return pi
		}
		hi, hj := uint64(0), uint64(0)
		if items[i].Height != nil {
			hi = *items[i].Height
		}
		if items[j].Height != nil {
			hj = *items[j].Height
		}
		if hi != hj {
			return hi > hj
		}
		if items[i].Timestamp != items[j].Timestamp {
			return items[i].Timestamp > items[j].Timestamp
		}
		return items[i].TXID > items[j].TXID
	})

	if len(items) > limit {
		items = items[:limit]
	}
	return items, nil
}
