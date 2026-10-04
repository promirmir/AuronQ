package main

import (
	"context"
	"errors"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	aq "auronq/internal/auronq"
)

const version = "1.7.13"

func main() {
	log.SetFlags(log.LstdFlags | log.LUTC)
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "init-mainnet":
		err = cmdInit(os.Args[2:], false)
	case "init-testnet":
		err = cmdInit(os.Args[2:], true)
	case "node":
		err = cmdNode(os.Args[2:])
	case "wallet-new":
		err = cmdWalletNew(os.Args[2:])
	case "wallet-info":
		err = cmdWalletInfo(os.Args[2:])
	case "balance":
		err = cmdBalance(os.Args[2:])
	case "send":
		err = cmdSend(os.Args[2:])
	case "mine":
		err = cmdMine(os.Args[2:])
	case "status":
		err = cmdStatus(os.Args[2:])
	case "verify-network":
		err = cmdVerifyNetwork(os.Args[2:])
	case "network-set-seeds":
		err = cmdNetworkSetSeeds(os.Args[2:])
	case "version":
		fmt.Println("AuronQ", version)
		return
	default:
		usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Print(`AuronQ - post-quantum proof-of-work cryptocurrency

Commands:
  init-mainnet    Create the immutable mainnet genesis (explicit risk acknowledgement required)
  init-testnet    Create a testnet genesis (10-block coinbase maturity)
  node            Run a full node and P2P/RPC server
  wallet-new      Create an encrypted ML-DSA-87 wallet
  wallet-info     Show public wallet metadata
  balance         Query wallet/address balance from a node
  send            Create, ML-DSA-sign and broadcast a transaction
  mine            CPU mine blocks from a node template
  status          Show node/chain status
  verify-network    Validate genesis and print immutable network ID
  network-set-seeds Update public bootstrap seed URLs without changing Network ID
  version           Print software version

Wallet password: set AURONQ_WALLET_PASSWORD or pass --password-file.
`)
}

func password(path string) (string, error) {
	if path != "" {
		b, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		p := strings.TrimSpace(string(b))
		if p == "" {
			return "", fmt.Errorf("password file is empty")
		}
		return p, nil
	}
	p := os.Getenv("AURONQ_WALLET_PASSWORD")
	if p == "" {
		return "", fmt.Errorf("set AURONQ_WALLET_PASSWORD or use --password-file")
	}
	return p, nil
}
func splitPeers(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	xs := strings.Split(s, ",")
	out := []string{}
	for _, x := range xs {
		if strings.TrimSpace(x) != "" {
			out = append(out, strings.TrimSpace(x))
		}
	}
	return out
}
func ctxSignals() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan os.Signal, 2)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() { <-ch; cancel() }()
	return ctx, cancel
}

func cmdInit(args []string, testnet bool) error {
	cmdName, defaultNetwork, defaultWallet, defaultName := "init-mainnet", "network.json", "founder.wallet", "AuronQ Mainnet"
	networkByte, maturity := aq.MainnetNetworkByte, aq.MainnetCoinbaseMaturity
	if testnet {
		cmdName, defaultNetwork, defaultWallet, defaultName = "init-testnet", "testnet-network.json", "testnet-founder.wallet", "AuronQ Testnet"
		networkByte, maturity = aq.TestnetNetworkByte, aq.TestnetCoinbaseMaturity
	}
	fs := flag.NewFlagSet(cmdName, flag.ContinueOnError)
	network := fs.String("network", defaultNetwork, "output network file")
	wallet := fs.String("wallet", defaultWallet, "encrypted founder wallet")
	pwfile := fs.String("password-file", "", "password file")
	name := fs.String("name", defaultName, "network display name")
	seed := fs.String("seed", "", "comma-separated seed peer URLs")
	dnsSeed := fs.String("dns-seed", "", "comma-separated DNS seed hostnames")
	bootstrapManifest := fs.String("bootstrap-manifest", "", "comma-separated HTTPS bootstrap manifest URLs")
	ackMainnet := fs.Bool("ack-unaudited-mainnet", false, "acknowledge that mainnet launch is unaudited and irreversible")
	threads := fs.Int("threads", runtime.NumCPU(), "genesis mining threads")
	timestamp := fs.Int64("timestamp", 0, "genesis Unix timestamp; default now")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if !testnet && !*ackMainnet {
		return fmt.Errorf("mainnet genesis requires --ack-unaudited-mainnet; this confirms you understand that the code/AQM64 have not received an independent security audit")
	}
	if _, err := os.Stat(*network); err == nil {
		return fmt.Errorf("network file already exists: %s", *network)
	}
	if _, err := os.Stat(*wallet); err == nil {
		return fmt.Errorf("wallet already exists: %s", *wallet)
	}
	pw, err := password(*pwfile)
	if err != nil {
		return err
	}

	// Stage both ceremony artifacts beside their final paths. This prevents an
	// interrupted genesis-mining run from leaving a final-looking wallet/network.
	walletTmp := filepath.Clean(*wallet) + ".ceremony.tmp"
	networkTmp := filepath.Clean(*network) + ".ceremony.tmp"
	_ = os.Remove(walletTmp)
	_ = os.Remove(networkTmp)
	committed := false
	defer func() {
		if !committed {
			_ = os.Remove(walletTmp)
			_ = os.Remove(networkTmp)
		}
	}()

	w, err := aq.NewWallet(walletTmp, pw, networkByte)
	if err != nil {
		return err
	}
	defer w.Close()
	g, err := aq.CreateGenesis(w.Address(), networkByte, *timestamp)
	if err != nil {
		return err
	}
	fmt.Printf("Mining immutable genesis for founder %s ...\n", w.Address())
	ctx, cancel := ctxSignals()
	defer cancel()
	res, err := aq.MineParallel(ctx, g, *threads, func(h uint64, d time.Duration) {
		if d > 0 {
			fmt.Printf("  %.0f H/s, %d hashes\n", float64(h)/d.Seconds(), h)
		}
	})
	if err != nil {
		return err
	}
	g = res.Block
	n := &aq.NetworkConfig{Name: *name, ProtocolVersion: 1, NetworkByte: networkByte, FounderAddress: w.Address(), Genesis: g, SeedPeers: splitPeers(*seed), DNSSeeds: splitPeers(*dnsSeed), BootstrapManifests: splitPeers(*bootstrapManifest), CoinbaseMaturity: maturity}
	if err := aq.ValidateGenesis(n); err != nil {
		return err
	}
	if err := aq.SaveNetwork(networkTmp, n); err != nil {
		return err
	}
	// Commit as a pair as closely as POSIX filesystems allow. If the second
	// rename fails, roll the first one back to its staged name.
	if err := os.Rename(walletTmp, *wallet); err != nil {
		return fmt.Errorf("commit founder wallet: %w", err)
	}
	if err := os.Rename(networkTmp, *network); err != nil {
		if rb := os.Rename(*wallet, walletTmp); rb != nil {
			return fmt.Errorf("commit network: %v; wallet rollback also failed: %v (recover %s immediately)", err, rb, *wallet)
		}
		return fmt.Errorf("commit network: %w", err)
	}
	committed = true
	fmt.Printf("\nAuronQ %s genesis created.\nNetwork ID: %s\nGenesis hash: %s\nFounder address: %s\nFounder allocation: %s AURQ (1%% of 21,000,000)\nCoinbase maturity: %d blocks\nWallet: %s\nNetwork file: %s\n", map[bool]string{true: "testnet", false: "mainnet"}[testnet], n.NetworkID(), g.Hash(), w.Address(), aq.FormatAmount(aq.FounderAtoms), n.Maturity(), *wallet, *network)
	fmt.Println("BACK UP the encrypted wallet and its password separately. Never publish the wallet file or password.")
	return nil
}

func cmdNetworkSetSeeds(args []string) error {
	fs := flag.NewFlagSet("network-set-seeds", flag.ContinueOnError)
	network := fs.String("network", "network.json", "network file to update")
	seed := fs.String("seed", "", "comma-separated public seed URLs")
	dnsSeed := fs.String("dns-seed", "", "comma-separated DNS seed hostnames")
	bootstrapManifest := fs.String("bootstrap-manifest", "", "comma-separated HTTPS bootstrap manifest URLs")
	if err := fs.Parse(args); err != nil {
		return err
	}
	n, err := aq.LoadNetwork(*network)
	if err != nil {
		return err
	}
	before := n.NetworkID()
	seeds := splitPeers(*seed)
	clean := make([]string, 0, len(seeds))
	seen := map[string]bool{}
	for _, raw := range seeds {
		p := strings.TrimSpace(strings.TrimRight(raw, "/"))
		if !strings.HasPrefix(p, "http://") && !strings.HasPrefix(p, "https://") {
			p = "http://" + p
		}
		u, err := url.Parse(p)
		if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
			return fmt.Errorf("invalid seed URL: %s", raw)
		}
		p = u.Scheme + "://" + u.Host
		if !seen[p] {
			seen[p] = true
			clean = append(clean, p)
		}
	}
	n.SeedPeers = clean
	dnsClean := make([]string, 0, len(splitPeers(*dnsSeed)))
	dnsSeen := map[string]bool{}
	for _, raw := range splitPeers(*dnsSeed) {
		h := strings.TrimSpace(strings.Trim(raw, "[]"))
		if h == "" || strings.ContainsAny(h, "/?#") {
			return fmt.Errorf("invalid DNS seed: %s", raw)
		}
		if !dnsSeen[h] {
			dnsSeen[h] = true
			dnsClean = append(dnsClean, h)
		}
	}
	n.DNSSeeds = dnsClean
	manifestClean := make([]string, 0, len(splitPeers(*bootstrapManifest)))
	manifestSeen := map[string]bool{}
	for _, raw := range splitPeers(*bootstrapManifest) {
		u, err := url.Parse(strings.TrimSpace(raw))
		if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || u.Fragment != "" {
			return fmt.Errorf("invalid bootstrap manifest URL (HTTPS required): %s", raw)
		}
		u.RawQuery = ""
		p := u.String()
		if !manifestSeen[p] {
			manifestSeen[p] = true
			manifestClean = append(manifestClean, p)
		}
	}
	n.BootstrapManifests = manifestClean
	if n.NetworkID() != before {
		return fmt.Errorf("internal error: seed update changed Network ID")
	}
	if err := aq.SaveNetwork(*network, n); err != nil {
		return err
	}
	fmt.Printf("Updated %d static seed peer(s), %d DNS seed(s), and %d bootstrap manifest(s). Network ID unchanged: %s\n", len(clean), len(dnsClean), len(manifestClean), before)
	return nil
}

func cmdNode(args []string) error {
	fs := flag.NewFlagSet("node", flag.ContinueOnError)
	network := fs.String("network", "network.json", "network file")
	data := fs.String("data", "data", "chain data directory")
	listen := fs.String("listen", "0.0.0.0:18444", "HTTP/P2P listen address")
	advertise := fs.String("advertise", "", "public peer URL, e.g. http://203.0.113.10:18444")
	peers := fs.String("peers", "", "comma-separated peers")
	peerStore := fs.String("peer-store", "", "persistent cache of discovered public peers")
	if err := fs.Parse(args); err != nil {
		return err
	}
	n, err := aq.LoadNetwork(*network)
	if err != nil {
		return err
	}
	c, err := aq.OpenChain(*data, n)
	if err != nil {
		return err
	}
	node := aq.NewNode(c, aq.NodeConfig{Listen: *listen, Advertise: *advertise, Peers: splitPeers(*peers), PeerStorePath: *peerStore})
	ctx, cancel := ctxSignals()
	defer cancel()
	return node.Run(ctx)
}

func cmdWalletNew(args []string) error {
	fs := flag.NewFlagSet("wallet-new", flag.ContinueOnError)
	network := fs.String("network", "network.json", "network file")
	out := fs.String("out", "wallet.json", "wallet path")
	pwfile := fs.String("password-file", "", "password file")
	if err := fs.Parse(args); err != nil {
		return err
	}
	n, err := aq.LoadNetwork(*network)
	if err != nil {
		return err
	}
	pw, err := password(*pwfile)
	if err != nil {
		return err
	}
	w, err := aq.NewWallet(*out, pw, n.NetworkByte)
	if err != nil {
		return err
	}
	defer w.Close()
	fmt.Printf("Address: %s\nWallet: %s\n", w.Address(), *out)
	return nil
}

func cmdWalletInfo(args []string) error {
	fs := flag.NewFlagSet("wallet-info", flag.ContinueOnError)
	path := fs.String("wallet", "wallet.json", "wallet path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	b, err := os.ReadFile(*path)
	if err != nil {
		return err
	}
	var w aq.WalletFile
	if err = json.Unmarshal(b, &w); err != nil {
		return err
	}
	fmt.Printf("Address: %s\nScheme: ML-DSA-87 (FIPS 204)\nCreated: %s\n", w.Address, time.Unix(w.CreatedAt, 0).UTC().Format(time.RFC3339))
	return nil
}

func cmdBalance(args []string) error {
	fs := flag.NewFlagSet("balance", flag.ContinueOnError)
	node := fs.String("node", "http://127.0.0.1:18444", "node URL")
	addr := fs.String("address", "", "AuronQ address")
	wallet := fs.String("wallet", "", "wallet file (address read without decrypting)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	a := *addr
	if a == "" && *wallet != "" {
		b, err := os.ReadFile(*wallet)
		if err != nil {
			return err
		}
		var w aq.WalletFile
		if err = json.Unmarshal(b, &w); err != nil {
			return err
		}
		a = w.Address
	}
	if a == "" {
		return fmt.Errorf("--address or --wallet is required")
	}
	r, err := aq.NewClient(*node).Balance(a)
	if err != nil {
		return err
	}
	fmt.Printf("Address: %s\nSpendable: %s AURQ\nTotal: %s AURQ\n", a, aq.FormatAmount(r.Spendable), aq.FormatAmount(r.Total))
	return nil
}

func cmdSend(args []string) error {
	fs := flag.NewFlagSet("send", flag.ContinueOnError)
	node := fs.String("node", "http://127.0.0.1:18444", "node URL")
	network := fs.String("network", "network.json", "network file")
	wallet := fs.String("wallet", "wallet.json", "wallet path")
	to := fs.String("to", "", "recipient address")
	amount := fs.String("amount", "", "amount in AURQ")
	pwfile := fs.String("password-file", "", "password file")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *to == "" || *amount == "" {
		return fmt.Errorf("--to and --amount are required")
	}
	n, err := aq.LoadNetwork(*network)
	if err != nil {
		return err
	}
	pw, err := password(*pwfile)
	if err != nil {
		return err
	}
	w, err := aq.LoadWallet(*wallet, pw)
	if err != nil {
		return err
	}
	defer w.Close()
	amt, err := aq.ParseAmount(*amount)
	if err != nil {
		return err
	}
	cl := aq.NewClient(*node)
	us, err := cl.UTXOs(w.Address())
	if err != nil {
		return err
	}
	tx, fee, err := w.BuildTransaction(us, *to, amt, n.NetworkByte)
	if err != nil {
		return err
	}
	r, err := cl.SubmitTx(tx)
	if err != nil {
		return err
	}
	fmt.Printf("TXID: %s\nAmount: %s AURQ\nFee: %s AURQ\n", r.TXID, aq.FormatAmount(amt), aq.FormatAmount(fee))
	return nil
}

func cmdMine(args []string) error {
	fs := flag.NewFlagSet("mine", flag.ContinueOnError)
	node := fs.String("node", "http://127.0.0.1:18444", "node URL")
	addr := fs.String("address", "", "reward address")
	wallet := fs.String("wallet", "", "wallet file (public metadata only)")
	threads := fs.Int("threads", runtime.NumCPU(), "CPU mining threads")
	once := fs.Bool("once", false, "mine one block and exit")
	if err := fs.Parse(args); err != nil {
		return err
	}
	a := *addr
	if a == "" && *wallet != "" {
		b, err := os.ReadFile(*wallet)
		if err != nil {
			return err
		}
		var wf aq.WalletFile
		if err = json.Unmarshal(b, &wf); err != nil {
			return err
		}
		a = wf.Address
	}
	if a == "" {
		return fmt.Errorf("--address or --wallet is required")
	}
	ctx, cancel := ctxSignals()
	defer cancel()
	cl := aq.NewClient(*node)
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		tpl, err := cl.Template(a)
		if err != nil {
			return err
		}
		fmt.Printf("Mining height %d target %s...\n", tpl.Header.Height, tpl.Header.Target.String()[:16])
		res, err := aq.MineRemoteTemplate(ctx, cl, tpl, *threads, time.Second, func(h uint64, d time.Duration) {
			if d > 0 {
				fmt.Printf("  %.0f H/s | %d hashes\n", float64(h)/d.Seconds(), h)
			}
		})
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			if errors.Is(err, aq.ErrMiningTemplateStale) {
				fmt.Println("Chain tip changed; refreshing mining template")
				continue
			}
			return err
		}
		br, err := cl.SubmitBlock(res.Block)
		if err != nil {
			fmt.Println("Block rejected; refreshing template:", err)
			continue
		}
		rate := float64(res.Hashes) / res.Duration.Seconds()
		fmt.Printf("BLOCK %d %s | %.0f H/s | %d hashes\n", br.Height, br.Hash, rate, res.Hashes)
		if *once {
			return nil
		}
	}
}

func cmdStatus(args []string) error {
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	node := fs.String("node", "http://127.0.0.1:18444", "node URL")
	if err := fs.Parse(args); err != nil {
		return err
	}
	s, err := aq.NewClient(*node).Status()
	if err != nil {
		return err
	}
	fmt.Printf("Network: %s\nNetwork ID: %s\nHeight: %d\nTip: %s\nIssued: %s AURQ / 21000000.00000000\nMempool: %d\nPeers: %d\n", s.Network, s.NetworkID, s.Height, s.Tip, aq.FormatAmount(s.Issued), s.Mempool, s.Peers)
	return nil
}
func cmdVerifyNetwork(args []string) error {
	fs := flag.NewFlagSet("verify-network", flag.ContinueOnError)
	path := fs.String("network", "network.json", "network file")
	if err := fs.Parse(args); err != nil {
		return err
	}
	n, err := aq.LoadNetwork(*path)
	if err != nil {
		return err
	}
	fmt.Printf("VALID\nName: %s\nNetwork ID: %s\nGenesis: %s\nFounder address: %s\nFounder output: %s AURQ\nCoinbase maturity: %d blocks\n", n.Name, n.NetworkID(), n.Genesis.Hash(), n.FounderAddress, aq.FormatAmount(n.Genesis.Transactions[0].Outputs[0].Value), n.Maturity())
	return nil
}
