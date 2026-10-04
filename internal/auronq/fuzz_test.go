package auronq

import (
	"encoding/json"
	"testing"
)

func FuzzDecodeAddressNeverPanics(f *testing.F) {
	for _, s := range []string{
		"",
		"aurq1",
		"aurq1invalid",
		"test",
		"\x00\x01\x02",
	} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		_, _, _, _ = DecodeAddress(s)
	})
}

func FuzzParseAmountNeverPanics(f *testing.F) {
	for _, s := range []string{
		"0",
		"1",
		"1.00000000",
		"1.000000001",
		"-1",
		"21000000.00000000",
		"999999999999999999999999999999",
		"",
	} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		v, err := ParseAmount(s)
		if err == nil {
			// Any accepted value must survive the canonical formatter/parser
			// round trip without changing atom count.
			canonical := FormatAmount(v)
			v2, err2 := ParseAmount(canonical)
			if err2 != nil || v2 != v {
				t.Fatalf("accepted amount failed roundtrip: input=%q value=%d canonical=%q value2=%d err=%v", s, v, canonical, v2, err2)
			}
		}
	})
}


func FuzzTransactionJSONNeverPanics(f *testing.F) {
	for _, s := range []string{
		`{}`,
		`{"version":1,"coinbase":true,"outputs":[]}`,
		`{"version":1,"inputs":[],"outputs":[]}`,
		`{"version":999999,"coinbase":false}`,
	} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		if len(s) > 1<<20 {
			t.Skip()
		}
		var tx Transaction
		if err := json.Unmarshal([]byte(s), &tx); err != nil {
			return
		}
		_ = tx.ID()
		_ = tx.BaseSize()
		_ = tx.ValidateBasic()
	})
}

func FuzzBlockJSONNeverPanics(f *testing.F) {
	for _, s := range []string{
		`{}`,
		`{"header":{"version":2,"pow_algo":1,"height":0},"transactions":[]}`,
		`{"header":{"height":18446744073709551615},"transactions":[{}]}`,
	} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		if len(s) > 1<<20 {
			t.Skip()
		}
		var b Block
		if err := json.Unmarshal([]byte(s), &b); err != nil {
			return
		}
		_ = b.Hash()
		_ = b.Size()
		_ = MerkleRoot(b.Transactions)
		for i := range b.Transactions {
			_ = b.Transactions[i].ValidateBasic()
		}
	})
}
