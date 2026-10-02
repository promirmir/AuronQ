package auronq

import (
	"errors"
	"time"
)

func CreateGenesis(founderAddress string, network byte, timestamp int64) (Block, error) {
	n, scheme, kh, err := DecodeAddress(founderAddress)
	if err != nil {
		return Block{}, err
	}
	if n != network {
		return Block{}, errors.New("founder address network mismatch")
	}
	if timestamp == 0 {
		timestamp = time.Now().Unix()
	}
	tx := Transaction{Version: TxVersion, Coinbase: true, CoinbaseHeight: 0, Outputs: []TxOutput{{Value: FounderAtoms, Scheme: scheme, KeyHash: kh}}}
	b := Block{Header: BlockHeader{Version: BlockVersion, PowAlgo: PowAlgorithmAQM64, Height: 0, PrevHash: ZeroHash(), Timestamp: timestamp, Target: PowLimit}, Transactions: []Transaction{tx}}
	b.Header.MerkleRoot = MerkleRoot(b.Transactions)
	return b, nil
}

func MineBlock(b *Block, stop <-chan struct{}) (uint64, error) {
	if b.Header.PowAlgo != PowAlgorithmAQM64 {
		return 0, errors.New("unsupported proof-of-work algorithm")
	}
	hasher := newAQM64Hasher()
	target := b.Header.Target.Big()
	for {
		h := hasher.hash(b.Header)
		if h.Big().Cmp(target) <= 0 {
			return b.Header.Nonce, nil
		}
		b.Header.Nonce++
		if b.Header.Nonce == 0 {
			b.Header.Timestamp = time.Now().Unix()
		}
		if stop != nil {
			select {
			case <-stop:
				return 0, errors.New("mining cancelled")
			default:
			}
		}
	}
}
