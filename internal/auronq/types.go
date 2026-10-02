package auronq

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"sort"
)

type Hash [64]byte
type KeyHash [32]byte
type Target [64]byte

func ZeroHash() Hash                        { return Hash{} }
func (h Hash) String() string               { return hex.EncodeToString(h[:]) }
func (h Hash) IsZero() bool                 { return h == Hash{} }
func (h Hash) MarshalJSON() ([]byte, error) { return json.Marshal(h.String()) }
func (h *Hash) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	x, err := hex.DecodeString(s)
	if err != nil {
		return err
	}
	if len(x) != 64 {
		return fmt.Errorf("hash must be 64 bytes")
	}
	copy(h[:], x)
	return nil
}
func (k KeyHash) String() string               { return hex.EncodeToString(k[:]) }
func (k KeyHash) MarshalJSON() ([]byte, error) { return json.Marshal(k.String()) }
func (k *KeyHash) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	x, err := hex.DecodeString(s)
	if err != nil {
		return err
	}
	if len(x) != 32 {
		return fmt.Errorf("key hash must be 32 bytes")
	}
	copy(k[:], x)
	return nil
}
func (t Target) String() string               { return hex.EncodeToString(t[:]) }
func (t Target) MarshalJSON() ([]byte, error) { return json.Marshal(t.String()) }
func (t *Target) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	x, err := hex.DecodeString(s)
	if err != nil {
		return err
	}
	if len(x) != 64 {
		return fmt.Errorf("target must be 64 bytes")
	}
	copy(t[:], x)
	return nil
}
func (t Target) Big() *big.Int { return new(big.Int).SetBytes(t[:]) }
func TargetFromBig(x *big.Int) Target {
	var t Target
	b := x.Bytes()
	if len(b) > 64 {
		b = b[len(b)-64:]
	}
	copy(t[64-len(b):], b)
	return t
}

const (
	TxVersion     uint16 = 1
	BlockVersion  uint16 = 2
	SequenceFinal uint32 = 0xffffffff
)

type OutPoint struct {
	TxID  Hash   `json:"txid"`
	Index uint32 `json:"index"`
}

func (o OutPoint) Key() string { return fmt.Sprintf("%s:%08x", o.TxID.String(), o.Index) }

type TxInput struct {
	PrevTx    Hash   `json:"prev_tx"`
	PrevIndex uint32 `json:"prev_index"`
	Sequence  uint32 `json:"sequence"`
	PublicKey []byte `json:"public_key,omitempty"`
	Signature []byte `json:"signature,omitempty"`
}

type TxOutput struct {
	Value   uint64  `json:"value"`
	Scheme  uint16  `json:"scheme"`
	KeyHash KeyHash `json:"key_hash"`
}

type Transaction struct {
	Version        uint16     `json:"version"`
	Coinbase       bool       `json:"coinbase"`
	CoinbaseHeight uint64     `json:"coinbase_height,omitempty"`
	Inputs         []TxInput  `json:"inputs,omitempty"`
	Outputs        []TxOutput `json:"outputs"`
	LockHeight     uint64     `json:"lock_height"`
}

func writeU16(w io.Writer, v uint16) {
	var b [2]byte
	binary.BigEndian.PutUint16(b[:], v)
	w.Write(b[:])
}
func writeU32(w io.Writer, v uint32) {
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], v)
	w.Write(b[:])
}
func writeU64(w io.Writer, v uint64) {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], v)
	w.Write(b[:])
}
func writeI64(w io.Writer, v int64)    { writeU64(w, uint64(v)) }
func writeBytes(w io.Writer, b []byte) { writeU32(w, uint32(len(b))); w.Write(b) }

func (tx *Transaction) serialize(includeSignatures bool) []byte {
	var b bytes.Buffer
	b.WriteString("AURONQ_TX\x00")
	writeU16(&b, tx.Version)
	if tx.Coinbase {
		b.WriteByte(1)
	} else {
		b.WriteByte(0)
	}
	writeU64(&b, tx.CoinbaseHeight)
	writeU32(&b, uint32(len(tx.Inputs)))
	for _, in := range tx.Inputs {
		b.Write(in.PrevTx[:])
		writeU32(&b, in.PrevIndex)
		writeU32(&b, in.Sequence)
		writeBytes(&b, in.PublicKey)
		if includeSignatures {
			writeBytes(&b, in.Signature)
		} else {
			writeBytes(&b, nil)
		}
	}
	writeU32(&b, uint32(len(tx.Outputs)))
	for _, out := range tx.Outputs {
		writeU64(&b, out.Value)
		writeU16(&b, out.Scheme)
		b.Write(out.KeyHash[:])
	}
	writeU64(&b, tx.LockHeight)
	return b.Bytes()
}

func (tx *Transaction) ID() Hash      { return Hash512(tx.serialize(true)) }
func (tx *Transaction) BaseSize() int { return len(tx.serialize(true)) }

func (tx *Transaction) SigHash(inputIndex int, prev TxOutput) Hash {
	var b bytes.Buffer
	b.WriteString("AURONQ_SIGHASH_V1\x00")
	b.Write(tx.serialize(false))
	writeU32(&b, uint32(inputIndex))
	writeU64(&b, prev.Value)
	writeU16(&b, prev.Scheme)
	b.Write(prev.KeyHash[:])
	return Hash512(b.Bytes())
}

func (tx *Transaction) ValidateBasic() error {
	if tx.Version != TxVersion {
		return fmt.Errorf("unsupported tx version %d", tx.Version)
	}
	if len(tx.Outputs) == 0 {
		return errors.New("transaction has no outputs")
	}
	if len(tx.Outputs) > 4096 || len(tx.Inputs) > 4096 {
		return errors.New("too many transaction inputs/outputs")
	}
	var total uint64
	for _, o := range tx.Outputs {
		if o.Value == 0 {
			return errors.New("zero-value output")
		}
		if !IsConsensusScheme(o.Scheme) {
			return fmt.Errorf("unsupported signature scheme %d", o.Scheme)
		}
		if ^uint64(0)-total < o.Value {
			return errors.New("output amount overflow")
		}
		total += o.Value
		if total > MaxSupplyAtoms {
			return errors.New("transaction output exceeds max supply")
		}
	}
	if tx.Coinbase {
		if len(tx.Inputs) != 0 {
			return errors.New("coinbase must not contain inputs")
		}
	} else {
		if len(tx.Inputs) == 0 {
			return errors.New("transaction has no inputs")
		}
		seen := map[string]bool{}
		for _, in := range tx.Inputs {
			k := OutPoint{in.PrevTx, in.PrevIndex}.Key()
			if seen[k] {
				return errors.New("duplicate input")
			}
			seen[k] = true
			if len(in.PublicKey) > MaxPubKeyBytes || len(in.Signature) > MaxSignatureBytes {
				return errors.New("oversized witness")
			}
		}
	}
	if tx.BaseSize() > MaxTxBytes {
		return fmt.Errorf("transaction exceeds %d bytes", MaxTxBytes)
	}
	return nil
}

type BlockHeader struct {
	Version    uint16 `json:"version"`
	PowAlgo    uint16 `json:"pow_algo"`
	Height     uint64 `json:"height"`
	PrevHash   Hash   `json:"prev_hash"`
	MerkleRoot Hash   `json:"merkle_root"`
	Timestamp  int64  `json:"timestamp"`
	Target     Target `json:"target"`
	Nonce      uint64 `json:"nonce"`
}

func (h BlockHeader) Bytes() []byte {
	var b bytes.Buffer
	b.WriteString("AURONQ_BLOCK_V2\x00")
	writeU16(&b, h.Version)
	writeU16(&b, h.PowAlgo)
	writeU64(&b, h.Height)
	b.Write(h.PrevHash[:])
	b.Write(h.MerkleRoot[:])
	writeI64(&b, h.Timestamp)
	b.Write(h.Target[:])
	writeU64(&b, h.Nonce)
	return b.Bytes()
}
func (h BlockHeader) Hash() Hash { return Hash512(h.Bytes()) }

type Block struct {
	Header       BlockHeader   `json:"header"`
	Transactions []Transaction `json:"transactions"`
}

func (b *Block) Hash() Hash { return b.Header.Hash() }
func (b *Block) Size() int  { x, _ := json.Marshal(b); return len(x) }

func MerkleRoot(txs []Transaction) Hash {
	if len(txs) == 0 {
		return ZeroHash()
	}
	hs := make([]Hash, len(txs))
	for i := range txs {
		hs[i] = txs[i].ID()
	}
	for len(hs) > 1 {
		if len(hs)%2 == 1 {
			hs = append(hs, hs[len(hs)-1])
		}
		next := make([]Hash, 0, len(hs)/2)
		for i := 0; i < len(hs); i += 2 {
			d := make([]byte, 0, 128+16)
			d = append(d, []byte("AURONQ_MERKLE_V1")...)
			d = append(d, hs[i][:]...)
			d = append(d, hs[i+1][:]...)
			next = append(next, Hash512(d))
		}
		hs = next
	}
	return hs[0]
}

type UTXO struct {
	Out      TxOutput `json:"out"`
	Height   uint64   `json:"height"`
	Coinbase bool     `json:"coinbase"`
}

type ChainState struct {
	Height    uint64 `json:"height"`
	Tip       Hash   `json:"tip"`
	Issued    uint64 `json:"issued"`
	ChainWork string `json:"chain_work"`
}

func WorkForTarget(t Target) *big.Int {
	denom := new(big.Int).Add(t.Big(), big.NewInt(1))
	if denom.Sign() <= 0 {
		return big.NewInt(0)
	}
	num := new(big.Int).Lsh(big.NewInt(1), 512)
	return num.Div(num, denom)
}

func SortedTxIDs(m map[string]Transaction) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}