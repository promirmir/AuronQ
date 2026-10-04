package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	aq "auronq/internal/auronq"
)

const mainnetNetworkID = "44e62c2ace002a6660c14e252173c1aa303529c68e40c998e92da2b453f44f30b1e58c94d533587e2186004593fb856c433fcdb5418ed430ec8617e29529365c"

type manifest struct {
	NetworkID string   `json:"network_id"`
	Peers     []string `json:"peers"`
}

type result struct {
	Peer     string
	Hello    aq.Hello
	Explorer bool
}

func checkPeer(client *http.Client, peer string) (result, error) {
	var out result
	peer = strings.TrimRight(strings.TrimSpace(peer), "/")
	if peer == "" {
		return out, errors.New("empty peer")
	}
	out.Peer = peer

	resp, err := client.Get(peer + "/p2p/hello")
	if err != nil {
		return out, fmt.Errorf("hello: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return out, fmt.Errorf("hello: HTTP %s", resp.Status)
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 64<<10)).Decode(&out.Hello); err != nil {
		return out, fmt.Errorf("hello decode: %w", err)
	}
	if out.Hello.ProtocolVersion != 1 {
		return out, fmt.Errorf("protocol version %d", out.Hello.ProtocolVersion)
	}
	if out.Hello.NetworkID.String() != mainnetNetworkID {
		return out, fmt.Errorf("network id mismatch: %s", out.Hello.NetworkID.String())
	}

	explorer, err := client.Get(peer + "/explorer")
	if err != nil {
		return out, fmt.Errorf("explorer: %w", err)
	}
	defer explorer.Body.Close()
	if explorer.StatusCode != http.StatusOK {
		return out, fmt.Errorf("explorer: HTTP %s", explorer.Status)
	}
	body, err := io.ReadAll(io.LimitReader(explorer.Body, 512<<10))
	if err != nil {
		return out, fmt.Errorf("explorer read: %w", err)
	}
	if !strings.Contains(string(body), "AuronQ Explorer") {
		return out, errors.New("explorer signature missing")
	}
	out.Explorer = true
	return out, nil
}

func readManifest(path string) (manifest, error) {
	var m manifest
	f, err := os.Open(path)
	if err != nil {
		return m, err
	}
	defer f.Close()
	if err := json.NewDecoder(io.LimitReader(f, 1<<20)).Decode(&m); err != nil {
		return m, err
	}
	if m.NetworkID != mainnetNetworkID {
		return m, fmt.Errorf("manifest Network ID mismatch: %s", m.NetworkID)
	}
	if len(m.Peers) == 0 {
		return m, errors.New("manifest contains no peers")
	}
	return m, nil
}

func main() {
	manifestPath := flag.String("manifest", "bootstrap.json", "bootstrap manifest to verify")
	flag.Parse()

	m, err := readManifest(*manifestPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "manifest:", err)
		os.Exit(1)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	healthy := make([]result, 0, len(m.Peers))
	for _, peer := range m.Peers {
		r, err := checkPeer(client, peer)
		if err != nil {
			fmt.Printf("UNHEALTHY %s: %v\n", peer, err)
			continue
		}
		healthy = append(healthy, r)
		fmt.Printf("HEALTHY %s height=%d tip=%s peers=%d explorer=%t\n",
			r.Peer, r.Hello.Height, r.Hello.Tip.String(), len(r.Hello.Peers), r.Explorer)
	}

	if len(healthy) == 0 {
		fmt.Fprintln(os.Stderr, "no healthy public AuronQ peers")
		os.Exit(2)
	}

	sort.Slice(healthy, func(i, j int) bool { return healthy[i].Hello.Height < healthy[j].Hello.Height })
	minHeight := healthy[0].Hello.Height
	maxHeight := healthy[len(healthy)-1].Hello.Height
	if maxHeight-minHeight > 12 {
		fmt.Fprintf(os.Stderr, "public peer height divergence too large: min=%d max=%d\n", minHeight, maxHeight)
		os.Exit(3)
	}

	topTips := map[string]struct{}{}
	for _, r := range healthy {
		if r.Hello.Height == maxHeight {
			topTips[r.Hello.Tip.String()] = struct{}{}
		}
	}
	if len(topTips) > 1 {
		fmt.Fprintf(os.Stderr, "public peers disagree on tip at height %d\n", maxHeight)
		os.Exit(4)
	}

	fmt.Printf("AuronQ public network health OK: %d/%d peers healthy, height range %d-%d\n",
		len(healthy), len(m.Peers), minHeight, maxHeight)
}
