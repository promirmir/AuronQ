package auronq

import (
	"context"
	"math/big"
	"strings"
	"testing"
	"time"
)

func consensusTestChain(t *testing.T, maturity uint64) (*Chain, *Wallet, string) {
	t.Helper()
	seed, pub, err := GenerateMLDSA87()
	if err != nil {
		t.Fatal(err)
	}
	founder := AddressFromPub(pub, SchemeMLDSA87, TestnetNetworkByte)
	g, err := CreateGenesis(founder, TestnetNetworkByte, time.Now().Unix()-120)
	if err != nil {
		t.Fatal(err)
	}
	netCfg := &NetworkConfig{
		Name:             "consensus-invariants",
		ProtocolVersion:  1,
		NetworkByte:      TestnetNetworkByte,
		CoinbaseMaturity: maturity,
		FounderAddress:   founder,
		Genesis:          g,
	}
	if err := ValidateGenesis(netCfg); err != nil {
		t.Fatal(err)
	}
	c, err := OpenChain(t.TempDir(), netCfg)
	if err != nil {
		t.Fatal(err)
	}
	w := &Wallet{
		File: WalletFile{
			NetworkByte: TestnetNetworkByte,
			Scheme:      SchemeMLDSA87,
			Address:     founder,
		},
		seed: seed,
		pub:  pub,
	}
	return c, w, founder
}

func mineTemplateForInvariant(t *testing.T, b Block) Block {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	mr, err := MineParallel(ctx, b, 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	return mr.Block
}

func TestDifficultyAdjustmentClampsOneBlockJumps(t *testing.T) {
	baseBig := new(big.Int).Div(PowLimit.Big(), big.NewInt(16))
	if baseBig.Sign() <= 0 {
		t.Fatal("invalid test target")
	}
	base := TargetFromBig(baseBig)

	fast := []Block{
		{Header: BlockHeader{Timestamp: 1000, Target: base}},
		{Header: BlockHeader{Timestamp: 1001, Target: base}},
		{Header: BlockHeader{Timestamp: 1002, Target: base}},
	}
	gotFast, err := expectedTarget(fast, 3)
	if err != nil {
		t.Fatal(err)
	}
	wantMin := new(big.Int).Div(new(big.Int).Set(baseBig), big.NewInt(4))
	if gotFast.Big().Cmp(wantMin) != 0 {
		t.Fatalf("fast adjustment=%s want clamp=%s", gotFast.Big(), wantMin)
	}

	slow := []Block{
		{Header: BlockHeader{Timestamp: 1000, Target: base}},
		{Header: BlockHeader{Timestamp: 1000 + TargetBlockSeconds*6, Target: base}},
		{Header: BlockHeader{Timestamp: 1000 + TargetBlockSeconds*12, Target: base}},
	}
	gotSlow, err := expectedTarget(slow, 3)
	if err != nil {
		t.Fatal(err)
	}
	wantMax := new(big.Int).Mul(new(big.Int).Set(baseBig), big.NewInt(4))
	if gotSlow.Big().Cmp(wantMax) != 0 {
		t.Fatalf("slow adjustment=%s want clamp=%s", gotSlow.Big(), wantMax)
	}
}

func TestBlockRejectsMedianTimePastAndFutureTimestamp(t *testing.T) {
	c, _, founder := consensusTestChain(t, 3)

	mtpTemplate, err := c.BuildTemplate(founder)
	if err != nil {
		t.Fatal(err)
	}
	mtpTemplate.Header.Timestamp = medianTimePast(c.blocks)
	mtpTemplate = mineTemplateForInvariant(t, mtpTemplate)
	if err := c.AddBlock(&mtpTemplate); err == nil || !strings.Contains(err.Error(), "median time past") {
		t.Fatalf("MTP timestamp was not rejected: %v", err)
	}

	futureTemplate, err := c.BuildTemplate(founder)
	if err != nil {
		t.Fatal(err)
	}
	futureTemplate.Header.Timestamp = time.Now().Unix() + MaxFutureSeconds + 2
	futureTemplate = mineTemplateForInvariant(t, futureTemplate)
	if err := c.AddBlock(&futureTemplate); err == nil || !strings.Contains(err.Error(), "too far in future") {
		t.Fatalf("future timestamp was not rejected: %v", err)
	}
}

func TestCoinbaseMaturityBoundary(t *testing.T) {
	c, _, founder := consensusTestChain(t, 3)

	_, minerPub, err := GenerateMLDSA87()
	if err != nil {
		t.Fatal(err)
	}
	minerAddr := AddressFromPub(minerPub, SchemeMLDSA87, TestnetNetworkByte)

	// Recreate the miner wallet with a known private key.
	minerSeed, minerPub, err := GenerateMLDSA87()
	if err != nil {
		t.Fatal(err)
	}
	minerAddr = AddressFromPub(minerPub, SchemeMLDSA87, TestnetNetworkByte)
	miner := &Wallet{
		File: WalletFile{NetworkByte: TestnetNetworkByte, Scheme: SchemeMLDSA87, Address: minerAddr},
		seed: minerSeed,
		pub:  minerPub,
	}

	b1, err := c.BuildTemplate(minerAddr)
	if err != nil {
		t.Fatal(err)
	}
	b1 = mineTemplateForInvariant(t, b1)
	if err := c.AddBlock(&b1); err != nil {
		t.Fatal(err)
	}

	_, recipientPub, err := GenerateMLDSA87()
	if err != nil {
		t.Fatal(err)
	}
	recipient := AddressFromPub(recipientPub, SchemeMLDSA87, TestnetNetworkByte)
	utxos, err := c.UTXOsForAddress(minerAddr, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(utxos) == 0 {
		t.Fatal("missing mined coinbase UTXO")
	}
	tx, _, err := miner.BuildTransaction(utxos, recipient, Coin, TestnetNetworkByte)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.AddMempool(tx); err == nil || !strings.Contains(err.Error(), "immature coinbase") {
		t.Fatalf("immature coinbase spend accepted: %v", err)
	}

	// Coinbase at height 1 with maturity 3 becomes spendable in block 4,
	// therefore the mempool may accept it once the current height is 3.
	for c.Height() < 3 {
		b, err := c.BuildTemplate(founder)
		if err != nil {
			t.Fatal(err)
		}
		b = mineTemplateForInvariant(t, b)
		if err := c.AddBlock(&b); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := c.AddMempool(tx); err != nil {
		t.Fatalf("mature coinbase spend rejected at boundary: %v", err)
	}
}

func TestBlockRejectsDoubleSpendWithinSameBlock(t *testing.T) {
	c, founderWallet, founder := consensusTestChain(t, 3)
	utxos, err := c.UTXOsForAddress(founder, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(utxos) == 0 {
		t.Fatal("missing founder UTXO")
	}

	_, aPub, err := GenerateMLDSA87()
	if err != nil {
		t.Fatal(err)
	}
	_, bPub, err := GenerateMLDSA87()
	if err != nil {
		t.Fatal(err)
	}
	addrA := AddressFromPub(aPub, SchemeMLDSA87, TestnetNetworkByte)
	addrB := AddressFromPub(bPub, SchemeMLDSA87, TestnetNetworkByte)
	txA, feeA, err := founderWallet.BuildTransaction(utxos, addrA, 2*Coin, TestnetNetworkByte)
	if err != nil {
		t.Fatal(err)
	}
	txB, feeB, err := founderWallet.BuildTransaction(utxos, addrB, 3*Coin, TestnetNetworkByte)
	if err != nil {
		t.Fatal(err)
	}

	_, scheme, kh, err := DecodeAddress(founder)
	if err != nil {
		t.Fatal(err)
	}
	height := uint64(1)
	subsidy := BlockSubsidy(height, c.State().Issued)
	cb := Transaction{
		Version:        TxVersion,
		Coinbase:       true,
		CoinbaseHeight: height,
		Outputs: []TxOutput{{
			Value:   subsidy + feeA + feeB,
			Scheme:  scheme,
			KeyHash: kh,
		}},
	}
	target, err := expectedTarget(c.blocks, height)
	if err != nil {
		t.Fatal(err)
	}
	b := Block{
		Header: BlockHeader{
			Version:   BlockVersion,
			PowAlgo:   PowAlgorithmAQM64,
			Height:    height,
			PrevHash:  c.Tip(),
			Timestamp: time.Now().Unix(),
			Target:    target,
		},
		Transactions: []Transaction{cb, txA, txB},
	}
	if b.Header.Timestamp <= medianTimePast(c.blocks) {
		b.Header.Timestamp = medianTimePast(c.blocks) + 1
	}
	b.Header.MerkleRoot = MerkleRoot(b.Transactions)
	b = mineTemplateForInvariant(t, b)

	if err := c.AddBlock(&b); err == nil || !strings.Contains(err.Error(), "missing/spent input") {
		t.Fatalf("same-block double spend was not rejected: %v", err)
	}
}
