package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	aq "auronq/internal/auronq"
)

const (
	mainnetNetworkID = "44e62c2ace002a6660c14e252173c1aa303529c68e40c998e92da2b453f44f30b1e58c94d533587e2186004593fb856c433fcdb5418ed430ec8617e29529365c"
	maxManifestPeers = 64
	maxCrawlCandidates = 256
	maxPerNetgroup = 2
)

type manifest struct {
	NetworkID string   `json:"network_id"`
	Peers     []string `json:"peers"`
	ExpiresAt int64    `json:"expires_at"`
}

type crawler struct {
	client *http.Client
	seen map[string]bool
	queue []string
	verified map[string]aq.Hello
	permanentDead map[string]bool
}

func normalizePeer(raw string) string {
	raw = strings.TrimSpace(strings.TrimRight(raw, "/"))
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil || u.User != nil || u.Fragment != "" || (u.Scheme != "http" && u.Scheme != "https") {
		return ""
	}
	if u.Path != "" && u.Path != "/" {
		return ""
	}
	u.Path = ""
	u.RawPath = ""
	u.RawQuery = ""
	return strings.TrimRight(u.String(), "/")
}

func nonPublicIP(ip net.IP) bool {
	if ip == nil || !ip.IsGlobalUnicast() || ip.IsLoopback() || ip.IsUnspecified() || ip.IsMulticast() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}
	if v4 := ip.To4(); v4 != nil {
		if v4[0] == 100 && v4[1] >= 64 && v4[1] <= 127 {
			return true
		}
		if (v4[0] == 192 && v4[1] == 0 && v4[2] == 2) ||
			(v4[0] == 198 && v4[1] == 51 && v4[2] == 100) ||
			(v4[0] == 203 && v4[1] == 0 && v4[2] == 113) {
			return true
		}
	}
	if v6 := ip.To16(); v6 != nil && ip.To4() == nil && v6[0] == 0x20 && v6[1] == 0x01 && v6[2] == 0x0d && v6[3] == 0xb8 {
		return true
	}
	return false
}

func safeDNSHost(host string) bool {
	host = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(host), "."))
	if host == "" || net.ParseIP(strings.Trim(host, "[]")) != nil || !strings.Contains(host, ".") {
		return false
	}
	return host != "localhost" &&
		!strings.HasSuffix(host, ".localhost") &&
		!strings.HasSuffix(host, ".local") &&
		!strings.HasSuffix(host, ".internal") &&
		!strings.HasSuffix(host, ".home.arpa")
}

func publicGossipPeer(raw string) bool {
	p := normalizePeer(raw)
	if p == "" {
		return false
	}
	u, err := url.Parse(p)
	if err != nil || u.User != nil || u.Fragment != "" || u.RawQuery != "" || (u.Path != "" && u.Path != "/") {
		return false
	}
	host := strings.Trim(u.Hostname(), "[]")
	if ip := net.ParseIP(host); ip != nil {
		return !nonPublicIP(ip)
	}
	return u.Scheme == "https" && safeDNSHost(host)
}

func isPermanentLookupFailure(err error) bool {
	var dnsErr *net.DNSError
	return errors.As(err, &dnsErr) && dnsErr.IsNotFound
}

func safeHTTPClient() *http.Client {
	dialer := &net.Dialer{Timeout: 5 * time.Second, KeepAlive: 30 * time.Second}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, err
		}
		host = strings.Trim(host, "[]")
		if ip := net.ParseIP(host); ip != nil {
			if nonPublicIP(ip) {
				return nil, fmt.Errorf("refusing non-public peer address %s", ip)
			}
			return dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
		}
		if !safeDNSHost(host) {
			return nil, fmt.Errorf("refusing unsafe peer hostname %q", host)
		}
		ips, err := net.DefaultResolver.LookupHost(ctx, host)
		if err != nil {
			return nil, err
		}
		var lastErr error
		for _, rawIP := range ips {
			ip := net.ParseIP(strings.TrimSpace(rawIP))
			if nonPublicIP(ip) {
				continue
			}
			conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
			if err == nil {
				return conn, nil
			}
			lastErr = err
		}
		if lastErr == nil {
			lastErr = fmt.Errorf("peer hostname %q resolved to no public addresses", host)
		}
		return nil, lastErr
	}
	return &http.Client{
		Timeout:   8 * time.Second,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func safeConfiguredPeer(raw string) bool {
	p := normalizePeer(raw)
	if p == "" {
		return false
	}
	u, err := url.Parse(p)
	if err != nil {
		return false
	}
	host := strings.Trim(strings.ToLower(u.Hostname()), "[]")
	if ip := net.ParseIP(host); ip != nil {
		return !nonPublicIP(ip)
	}
	// DNS names are allowed only when they are already explicitly present in
	// the signed/reviewed repository manifest and use HTTPS.
	return u.Scheme == "https" && strings.Contains(host, ".") &&
		host != "localhost" && !strings.HasSuffix(host, ".localhost") &&
		!strings.HasSuffix(host, ".local") && !strings.HasSuffix(host, ".internal") &&
		!strings.HasSuffix(host, ".home.arpa")
}

func netgroup(raw string) string {
	p := normalizePeer(raw)
	u, err := url.Parse(p)
	if err != nil {
		return ""
	}
	ip := net.ParseIP(strings.Trim(u.Hostname(), "[]"))
	if ip == nil {
		host := strings.ToLower(strings.TrimSuffix(u.Hostname(), "."))
		parts := strings.Split(host, ".")
		if len(parts) >= 3 {
			// Keep common second-level ccTLDs together, otherwise group by the
			// registrable-looking parent domain. This is diversity only, not trust.
			secondLevel := map[string]bool{"co": true, "com": true, "net": true, "org": true, "gov": true, "ac": true}
			if len(parts[len(parts)-1]) == 2 && secondLevel[parts[len(parts)-2]] && len(parts) >= 3 {
				host = strings.Join(parts[len(parts)-3:], ".")
			} else {
				host = strings.Join(parts[len(parts)-2:], ".")
			}
		}
		return "dns:" + host
	}
	if v4 := ip.To4(); v4 != nil {
		return fmt.Sprintf("v4:%d.%d", v4[0], v4[1])
	}
	v6 := ip.To16()
	if v6 == nil {
		return ""
	}
	return fmt.Sprintf("v6:%02x%02x:%02x%02x", v6[0], v6[1], v6[2], v6[3])
}

func (c *crawler) enqueue(p string) {
	p = normalizePeer(p)
	if p == "" || c.seen[p] || len(c.seen) >= maxCrawlCandidates {
		return
	}
	c.seen[p] = true
	c.queue = append(c.queue, p)
}

func (c *crawler) hello(peer string) (aq.Hello, error) {
	var h aq.Hello
	req, err := http.NewRequest(http.MethodGet, peer+"/p2p/hello", nil)
	if err != nil {
		return h, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "AuronQ-Peer-Crawler/1")
	resp, err := c.client.Do(req)
	if err != nil {
		return h, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return h, fmt.Errorf("HTTP %s", resp.Status)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
	if err != nil {
		return h, err
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&h); err != nil {
		return h, err
	}
	if h.ProtocolVersion != 1 || h.NetworkID.String() != mainnetNetworkID {
		return h, errors.New("network/protocol mismatch")
	}
	return h, nil
}

func readManifest(path string) (manifest, error) {
	var m manifest
	b, err := os.ReadFile(path)
	if err != nil {
		return m, err
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&m); err != nil {
		return m, err
	}
	if m.NetworkID != mainnetNetworkID {
		return m, errors.New("manifest Network ID mismatch")
	}
	return m, nil
}

func writeManifest(path string, m manifest) error {
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(path, b, 0644)
}

func main() {
	manifestPath := flag.String("manifest", "bootstrap.json", "bootstrap manifest path")
	mirrorPath := flag.String("mirror", "bootstrap.mainnet.json", "optional mirror manifest path")
	flag.Parse()

	m, err := readManifest(*manifestPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "manifest:", err)
		os.Exit(1)
	}

	c := &crawler{
		client: safeHTTPClient(),
		seen: map[string]bool{},
		verified: map[string]aq.Hello{},
		permanentDead: map[string]bool{},
	}

	// Existing manifest entries are trusted only as initial rendezvous metadata.
	// Learned peers must be literal globally-routable IP endpoints.
	for _, p := range m.Peers {
		if safeConfiguredPeer(p) {
			c.enqueue(p)
		}
	}

	for len(c.queue) > 0 {
		p := c.queue[0]
		c.queue = c.queue[1:]
		h, err := c.hello(p)
		if err != nil {
			if isPermanentLookupFailure(err) {
				c.permanentDead[p] = true
				fmt.Printf("permanently unreachable %s: %v\n", p, err)
			} else {
				fmt.Printf("unreachable %s: %v\n", p, err)
			}
			continue
		}
		c.verified[p] = h
		fmt.Printf("verified %s height=%d peers=%d\n", p, h.Height, len(h.Peers))

		learned := append([]string{}, h.Peers...)
		if h.Advertise != "" {
			learned = append(learned, h.Advertise)
		}
		for _, q := range learned {
			if publicGossipPeer(q) {
				c.enqueue(q)
			}
		}
	}

	// Preserve existing entries so a transient outage does not delete a known
	// rendezvous path. Add only independently reachable, Network-ID verified
	// public peers, with basic netgroup diversity.
	out := make([]string, 0, maxManifestPeers)
	seenOut := map[string]bool{}
	groups := map[string]int{}
	add := func(p string, enforceGroup bool) {
		p = normalizePeer(p)
		if p == "" || seenOut[p] || len(out) >= maxManifestPeers {
			return
		}
		if enforceGroup {
			g := netgroup(p)
			if g == "" || groups[g] >= maxPerNetgroup {
				return
			}
			groups[g]++
		}
		seenOut[p] = true
		out = append(out, p)
	}
	for _, p := range m.Peers {
		if safeConfiguredPeer(p) && !c.permanentDead[normalizePeer(p)] {
			add(p, false)
		}
	}

	verified := make([]string, 0, len(c.verified))
	for p := range c.verified {
		if publicGossipPeer(p) {
			verified = append(verified, p)
		}
	}
	sort.Strings(verified)
	for _, p := range verified {
		add(p, true)
	}

	if len(out) == 0 {
		fmt.Fprintln(os.Stderr, "refusing to write an empty bootstrap manifest")
		os.Exit(1)
	}

	m.Peers = out
	m.ExpiresAt = 0
	if err := writeManifest(*manifestPath, m); err != nil {
		fmt.Fprintln(os.Stderr, "write manifest:", err)
		os.Exit(1)
	}
	if *mirrorPath != "" {
		if err := writeManifest(*mirrorPath, m); err != nil {
			fmt.Fprintln(os.Stderr, "write mirror:", err)
			os.Exit(1)
		}
	}
	fmt.Printf("bootstrap registry now contains %d peers (%d verified this run)\n", len(out), len(c.verified))
}
