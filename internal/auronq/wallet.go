package auronq

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"time"
)

type WalletFile struct {
	Format      string `json:"format"`
	CreatedAt   int64  `json:"created_at"`
	NetworkByte byte   `json:"network_byte"`
	Scheme      uint16 `json:"scheme"`
	Address     string `json:"address"`
	PublicKey   string `json:"public_key"`
	KDF         string `json:"kdf"`
	KDFN        uint64 `json:"kdf_n"`
	KDFR        uint64 `json:"kdf_r"`
	KDFP        uint64 `json:"kdf_p"`
	Salt        string `json:"salt"`
	Nonce       string `json:"nonce"`
	Ciphertext  string `json:"ciphertext"`
}

type Wallet struct {
	File WalletFile
	seed []byte
	pub  []byte
}

func NewWallet(path, password string, network byte) (*Wallet, error) {
	if _, err := os.Stat(path); err == nil {
		return nil, errors.New("wallet file already exists")
	}
	seed, pub, err := GenerateMLDSA87()
	if err != nil {
		return nil, err
	}
	salt, nonce, ct, err := EncryptSeed(seed, password)
	if err != nil {
		return nil, err
	}
	wf := WalletFile{Format: "auronq-wallet-v1", CreatedAt: time.Now().Unix(), NetworkByte: network, Scheme: SchemeMLDSA87, Address: AddressFromPub(pub, SchemeMLDSA87, network), PublicKey: hex.EncodeToString(pub), KDF: "scrypt", KDFN: 1 << 17, KDFR: 8, KDFP: 1, Salt: hex.EncodeToString(salt), Nonce: hex.EncodeToString(nonce), Ciphertext: hex.EncodeToString(ct)}
	b, _ := json.MarshalIndent(wf, "", "  ")
	if err := atomicWrite(path, b, 0600); err != nil {
		return nil, err
	}
	return &Wallet{File: wf, seed: seed, pub: pub}, nil
}

func LoadWallet(path, password string) (*Wallet, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var wf WalletFile
	if err = json.Unmarshal(b, &wf); err != nil {
		return nil, err
	}
	if wf.Format != "auronq-wallet-v1" || wf.Scheme != SchemeMLDSA87 || wf.KDF != "scrypt" {
		return nil, errors.New("unsupported wallet format")
	}
	if wf.KDFN != 1<<17 || wf.KDFR != 8 || wf.KDFP != 1 {
		return nil, errors.New("unsupported wallet KDF parameters")
	}
	salt, err := hex.DecodeString(wf.Salt)
	if err != nil {
		return nil, err
	}
	nonce, err := hex.DecodeString(wf.Nonce)
	if err != nil {
		return nil, err
	}
	ct, err := hex.DecodeString(wf.Ciphertext)
	if err != nil {
		return nil, err
	}
	seed, err := DecryptSeed(salt, nonce, ct, password)
	if err != nil {
		return nil, err
	}
	pub, err := PublicFromSeed(seed)
	if err != nil {
		return nil, err
	}
	if hex.EncodeToString(pub) != wf.PublicKey {
		return nil, errors.New("wallet public key mismatch")
	}
	if AddressFromPub(pub, wf.Scheme, wf.NetworkByte) != wf.Address {
		return nil, errors.New("wallet address mismatch")
	}
	return &Wallet{File: wf, seed: seed, pub: pub}, nil
}
func (w *Wallet) Close() {
	for i := range w.seed {
		w.seed[i] = 0
	}
}
func (w *Wallet) Address() string   { return w.File.Address }
func (w *Wallet) PublicKey() []byte { return append([]byte(nil), w.pub...) }

func EstimateFee(inputCount, outputCount int) uint64 {
	tx := Transaction{Version: TxVersion, Inputs: make([]TxInput, inputCount), Outputs: make([]TxOutput, outputCount)}
	for i := range tx.Inputs {
		tx.Inputs[i].PublicKey = make([]byte, MLDSA87PublicKeySize)
		tx.Inputs[i].Signature = make([]byte, MLDSA87SignatureSize)
		tx.Inputs[i].Sequence = SequenceFinal
	}
	size := tx.BaseSize()
	return uint64((size+999)/1000) * MinRelayFeePerKB
}

func (w *Wallet) BuildTransaction(utxos []UTXORecord, to string, amount uint64, network byte) (Transaction, uint64, error) {
	if amount == 0 {
		return Transaction{}, 0, errors.New("amount must be positive")
	}
	n, scheme, kh, err := DecodeAddress(to)
	if err != nil {
		return Transaction{}, 0, err
	}
	if n != network {
		return Transaction{}, 0, errors.New("recipient address network mismatch")
	}
	if !IsConsensusScheme(scheme) {
		return Transaction{}, 0, errors.New("recipient scheme unsupported")
	}
	if w.File.NetworkByte != network {
		return Transaction{}, 0, errors.New("wallet network mismatch")
	}
	sort.Slice(utxos, func(i, j int) bool { return utxos[i].UTXO.Out.Value < utxos[j].UTXO.Out.Value })
	selected := []UTXORecord{}
	var total uint64
	var fee uint64
	for _, u := range utxos {
		selected = append(selected, u)
		total += u.UTXO.Out.Value
		// assume change output; if exact no-change becomes slightly smaller, overpay is acceptable but we recompute below.
		fee = EstimateFee(len(selected), 2)
		if total >= amount+fee {
			break
		}
	}
	if total < amount+fee {
		return Transaction{}, 0, fmt.Errorf("insufficient funds: have %.8f, need %.8f + fee", float64(total)/float64(Coin), float64(amount)/float64(Coin))
	}
	change := total - amount - fee
	outputs := []TxOutput{{Value: amount, Scheme: scheme, KeyHash: kh}}
	if change > 0 {
		_, cs, ck, _ := DecodeAddress(w.Address())
		outputs = append(outputs, TxOutput{Value: change, Scheme: cs, KeyHash: ck})
	} else {
		fee = EstimateFee(len(selected), 1)
		if total < amount+fee {
			return Transaction{}, 0, errors.New("insufficient funds after fee recalculation")
		}
		change = total - amount - fee
		if change > 0 {
			_, cs, ck, _ := DecodeAddress(w.Address())
			outputs = append(outputs, TxOutput{Value: change, Scheme: cs, KeyHash: ck})
		}
	}
	tx := Transaction{Version: TxVersion, Coinbase: false, LockHeight: 0, Outputs: outputs}
	tx.Inputs = make([]TxInput, len(selected))
	for i, u := range selected {
		tx.Inputs[i] = TxInput{PrevTx: u.OutPoint.TxID, PrevIndex: u.OutPoint.Index, Sequence: SequenceFinal, PublicKey: w.PublicKey()}
	}
	for i, u := range selected {
		h := tx.SigHash(i, u.UTXO.Out)
		sig, err := SignWithScheme(w.File.Scheme, w.seed, h[:])
		if err != nil {
			return Transaction{}, 0, err
		}
		tx.Inputs[i].Signature = sig
	}
	actualFee := total
	for _, o := range tx.Outputs {
		actualFee -= o.Value
	}
	min := uint64((tx.BaseSize()+999)/1000) * MinRelayFeePerKB
	if actualFee < min {
		return Transaction{}, 0, fmt.Errorf("internal fee estimate %d below required %d", actualFee, min)
	}
	return tx, actualFee, nil
}
