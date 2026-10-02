package auronq

import (
	"context"
	"testing"
	"time"
)

func TestMLDSA87SignVerify(t *testing.T) {
	seed, pub, err := GenerateMLDSA87()
	if err != nil {
		t.Fatal(err)
	}
	msg := Hash512([]byte("AuronQ consensus test vector"))
	sig, err := SignMLDSA87(seed, msg[:])
	if err != nil {
		t.Fatal(err)
	}
	if len(pub) != MLDSA87PublicKeySize || len(sig) != MLDSA87SignatureSize {
		t.Fatalf("unexpected sizes pub=%d sig=%d", len(pub), len(sig))
	}
	if !VerifyMLDSA87(pub, msg[:], sig) {
		t.Fatal("valid ML-DSA signature rejected")
	}
	msg[0] ^= 1
	if VerifyMLDSA87(pub, msg[:], sig) {
		t.Fatal("modified message accepted")
	}
}

func TestAddressRoundTrip(t *testing.T) {
	_, pub, err := GenerateMLDSA87()
	if err != nil {
		t.Fatal(err)
	}
	a := AddressFromPub(pub, SchemeMLDSA87, MainnetNetworkByte)
	n, s, k, err := DecodeAddress(a)
	if err != nil {
		t.Fatal(err)
	}
	if n != MainnetNetworkByte || s != SchemeMLDSA87 || k != Hash256(pub) {
		t.Fatal("address roundtrip mismatch")
	}
	bad := a[:len(a)-1] + "a"
	if bad == a {
		bad = a[:len(a)-1] + "b"
	}
	if _, _, _, err := DecodeAddress(bad); err == nil {
		t.Fatal("bad checksum accepted")
	}
}

func TestAmounts(t *testing.T) {
	v, err := ParseAmount("210000.12345678")
	if err != nil {
		t.Fatal(err)
	}
	if FormatAmount(v) != "210000.12345678" {
		t.Fatal(FormatAmount(v))
	}
	if _, err := ParseAmount("1.000000001"); err == nil {
		t.Fatal("accepted >8 decimals")
	}
}

func TestEndToEndSpend(t *testing.T) {
	founderSeed, founderPub, err := GenerateMLDSA87()
	if err != nil {
		t.Fatal(err)
	}
	_, recipientPub, err := GenerateMLDSA87()
	if err != nil {
		t.Fatal(err)
	}
	founderAddr := AddressFromPub(founderPub, SchemeMLDSA87, MainnetNetworkByte)
	recipientAddr := AddressFromPub(recipientPub, SchemeMLDSA87, MainnetNetworkByte)
	g, err := CreateGenesis(founderAddr, MainnetNetworkByte, time.Now().Unix()-20)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	mr, err := MineParallel(ctx, g, 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	g = mr.Block
	net := &NetworkConfig{Name: "test", ProtocolVersion: 1, NetworkByte: MainnetNetworkByte, FounderAddress: founderAddr, Genesis: g}
	if err := ValidateGenesis(net); err != nil {
		t.Fatal(err)
	}
	c, err := OpenChain(t.TempDir(), net)
	if err != nil {
		t.Fatal(err)
	}
	us, err := c.UTXOsForAddress(founderAddr, false)
	if err != nil {
		t.Fatal(err)
	}
	w := &Wallet{File: WalletFile{NetworkByte: MainnetNetworkByte, Scheme: SchemeMLDSA87, Address: founderAddr}, seed: founderSeed, pub: founderPub}
	amount := uint64(25)*Coin + Coin/2
	tx, _, err := w.BuildTransaction(us, recipientAddr, amount, MainnetNetworkByte)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.AddMempool(tx); err != nil {
		t.Fatal(err)
	}
	tpl, err := c.BuildTemplate(founderAddr)
	if err != nil {
		t.Fatal(err)
	}
	mr, err = MineParallel(ctx, tpl, 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.AddBlock(&mr.Block); err != nil {
		t.Fatal(err)
	}
	spend, total, err := c.Balance(recipientAddr)
	if err != nil {
		t.Fatal(err)
	}
	if spend != amount || total != amount {
		t.Fatalf("recipient got %s/%s", FormatAmount(spend), FormatAmount(total))
	}
	if c.State().Issued != FounderAtoms+InitialSubsidyAtoms {
		t.Fatalf("issued supply mismatch: %s", FormatAmount(c.State().Issued))
	}
}

func TestMonetaryPolicy(t *testing.T) {
	if BlockSubsidy(1, FounderAtoms) != InitialSubsidyAtoms {
		t.Fatalf("height 1 subsidy mismatch")
	}
	if BlockSubsidy(HalvingInterval, FounderAtoms) != InitialSubsidyAtoms {
		t.Fatalf("first epoch must contain exactly %d mined blocks", HalvingInterval)
	}
	if BlockSubsidy(HalvingInterval+1, FounderAtoms) != InitialSubsidyAtoms/2 {
		t.Fatalf("first halving boundary mismatch")
	}

	issued := FounderAtoms
	for h := uint64(1); ; h++ {
		s := BlockSubsidy(h, issued)
		if s == 0 {
			break
		}
		if MaxSupplyAtoms-issued < s {
			t.Fatalf("subsidy would exceed maximum supply")
		}
		issued += s
		if h > HalvingInterval*64 {
			t.Fatalf("emission did not terminate")
		}
	}
	if issued > MaxSupplyAtoms {
		t.Fatalf("issued %d exceeds cap %d", issued, MaxSupplyAtoms)
	}
	// Integer halvings intentionally leave only a tiny remainder below the hard cap.
	if MaxSupplyAtoms-issued > Coin {
		t.Fatalf("terminal issuance too far below cap: %s AURQ unissued", FormatAmount(MaxSupplyAtoms-issued))
	}
}

func TestScryptCompatibilityVector(t *testing.T) {
	got, err := ScryptKey("test-password", []byte("0123456789abcdef"), 1<<14, 8, 1, 32)
	if err != nil {
		t.Fatal(err)
	}
	const want = "3fb2ced15c9ea23345f8fccf4b4545e2d62bd0ae71c54cec93aa91b18666a9dd"
	if Hex(got) != want {
		t.Fatalf("scrypt mismatch: %s", Hex(got))
	}
}

func TestSignatureSchemeRegistry(t *testing.T) {
	info, ok := SignatureScheme(SchemeMLDSA87)
	if !ok {
		t.Fatal("ML-DSA-87 missing from registry")
	}
	if info.PublicKeySize != MLDSA87PublicKeySize || info.SignatureSize != MLDSA87SignatureSize {
		t.Fatal("ML-DSA-87 registry sizes mismatch")
	}
	if IsConsensusScheme(0xffff) {
		t.Fatal("unknown scheme accepted")
	}
}

func TestTransactionRejectsUnknownScheme(t *testing.T) {
	tx := Transaction{
		Version:        TxVersion,
		Coinbase:       true,
		CoinbaseHeight: 1,
		Outputs:        []TxOutput{{Value: Coin, Scheme: 0xffff, KeyHash: KeyHash{1}}},
	}
	if err := tx.ValidateBasic(); err == nil {
		t.Fatal("transaction with unknown signature scheme accepted")
	}
}

func TestSchemeBoundInAddressAndSigHash(t *testing.T) {
	_, pub, err := GenerateMLDSA87()
	if err != nil {
		t.Fatal(err)
	}
	a := AddressFromPub(pub, SchemeMLDSA87, MainnetNetworkByte)
	_, scheme, kh, err := DecodeAddress(a)
	if err != nil {
		t.Fatal(err)
	}
	if scheme != SchemeMLDSA87 || kh != Hash256(pub) {
		t.Fatal("address is not cryptographically bound to scheme/public key")
	}
	tx := Transaction{Version: TxVersion, Inputs: []TxInput{{Sequence: SequenceFinal}}, Outputs: []TxOutput{{Value: Coin, Scheme: SchemeMLDSA87, KeyHash: kh}}}
	prev1 := TxOutput{Value: 2 * Coin, Scheme: SchemeMLDSA87, KeyHash: kh}
	prev2 := prev1
	prev2.Scheme = SchemeMLDSA87 + 1
	if tx.SigHash(0, prev1) == tx.SigHash(0, prev2) {
		t.Fatal("signature hash is not bound to previous output scheme")
	}
}

func TestNetworkCoinbaseMaturityProfiles(t *testing.T) {
	main := &NetworkConfig{NetworkByte: MainnetNetworkByte, CoinbaseMaturity: MainnetCoinbaseMaturity}
	if got := main.Maturity(); got != 100 {
		t.Fatalf("mainnet maturity = %d, want 100", got)
	}
	test := &NetworkConfig{NetworkByte: TestnetNetworkByte, CoinbaseMaturity: TestnetCoinbaseMaturity}
	if got := test.Maturity(); got != 10 {
		t.Fatalf("testnet maturity = %d, want 10", got)
	}
	legacyTest := &NetworkConfig{NetworkByte: TestnetNetworkByte}
	if got := legacyTest.Maturity(); got != 10 {
		t.Fatalf("legacy testnet maturity = %d, want 10", got)
	}
}

func TestWalletHistoryTracksGenesisPendingAndConfirmed(t *testing.T) {
	founderSeed, founderPub, err := GenerateMLDSA87()
	if err != nil {
		t.Fatal(err)
	}
	_, recipientPub, err := GenerateMLDSA87()
	if err != nil {
		t.Fatal(err)
	}
	founderAddr := AddressFromPub(founderPub, SchemeMLDSA87, TestnetNetworkByte)
	recipientAddr := AddressFromPub(recipientPub, SchemeMLDSA87, TestnetNetworkByte)
	g, err := CreateGenesis(founderAddr, TestnetNetworkByte, time.Now().Unix()-20)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	mr, err := MineParallel(ctx, g, 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	g = mr.Block
	net := &NetworkConfig{Name: "history-test", ProtocolVersion: 1, NetworkByte: TestnetNetworkByte, CoinbaseMaturity: 10, FounderAddress: founderAddr, Genesis: g}
	c, err := OpenChain(t.TempDir(), net)
	if err != nil {
		t.Fatal(err)
	}

	h, err := c.HistoryForAddress(founderAddr, 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(h) != 1 || h[0].Type != "genesis" || h[0].AmountAtoms != FounderAtoms || h[0].Status != "confirmed" {
		t.Fatalf("unexpected founder genesis history: %+v", h)
	}

	us, err := c.UTXOsForAddress(founderAddr, false)
	if err != nil {
		t.Fatal(err)
	}
	w := &Wallet{File: WalletFile{NetworkByte: TestnetNetworkByte, Scheme: SchemeMLDSA87, Address: founderAddr}, seed: founderSeed, pub: founderPub}
	amount := uint64(3)*Coin + Coin/4
	tx, fee, err := w.BuildTransaction(us, recipientAddr, amount, TestnetNetworkByte)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.AddMempool(tx); err != nil {
		t.Fatal(err)
	}
	h, err = c.HistoryForAddress(founderAddr, 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(h) < 2 || h[0].Status != "pending" || h[0].Type != "sent" || h[0].AmountAtoms != amount || h[0].FeeAtoms != fee {
		t.Fatalf("unexpected pending history: %+v", h)
	}

	tpl, err := c.BuildTemplate(founderAddr)
	if err != nil {
		t.Fatal(err)
	}
	mr, err = MineParallel(ctx, tpl, 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.AddBlock(&mr.Block); err != nil {
		t.Fatal(err)
	}
	hr, err := c.HistoryForAddress(recipientAddr, 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(hr) != 1 || hr[0].Type != "received" || hr[0].Status != "confirmed" || hr[0].AmountAtoms != amount || hr[0].Height == nil || *hr[0].Height != 1 {
		t.Fatalf("unexpected recipient history: %+v", hr)
	}
}