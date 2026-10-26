package auronq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	ErrClientNetworkMismatch = errors.New("node network ID mismatch")
	ErrClientChainConflict    = errors.New("conflicting node chain state")
)

type VerifiedClientConfig struct {
	ExpectedNetworkID Hash
	Preferred         string
	Seeds             []string
	StorePath         string
	Exclude           []string
	Timeout           time.Duration
	MaxCandidates     int
}

type VerifiedClientResolution struct {
	Client   *Client
	Endpoint string
	Status   Status
}

type clientPeerRecord struct {
	URL             string `json:"url"`
	Failures        int    `json:"failures,omitempty"`
	LastOK          int64  `json:"last_ok,omitempty"`
	QuarantineUntil int64  `json:"quarantine_until,omitempty"`
}

type clientPeerStore struct {
	Peers []clientPeerRecord `json:"peers"`
}

type verifiedEndpoint struct {
	URL    string
	Status Status
	Client *Client
}

func normalizeClientEndpoint(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", false
	}
	if !strings.Contains(raw, "://") {
		raw = "http://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || u.User != nil || u.Fragment != "" {
		return "", false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", false
	}
	if u.Path != "" && u.Path != "/" {
		return "", false
	}
	if u.RawQuery != "" {
		return "", false
	}
	return u.Scheme + "://" + u.Host, true
}

func loadClientPeerStore(path string) map[string]clientPeerRecord {
	out := map[string]clientPeerRecord{}
	if strings.TrimSpace(path) == "" {
		return out
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return out
	}
	var s clientPeerStore
	if json.Unmarshal(b, &s) == nil {
		for _, rec := range s.Peers {
			if p, ok := normalizeClientEndpoint(rec.URL); ok {
				rec.URL = p
				out[p] = rec
			}
		}
		return out
	}
	var old []string
	if json.Unmarshal(b, &old) == nil {
		for _, raw := range old {
			if p, ok := normalizeClientEndpoint(raw); ok {
				out[p] = clientPeerRecord{URL: p}
			}
		}
	}
	return out
}

func saveClientPeerStore(path string, records map[string]clientPeerRecord) {
	if strings.TrimSpace(path) == "" {
		return
	}
	list := make([]clientPeerRecord, 0, len(records))
	for _, rec := range records {
		if rec.URL == "" {
			continue
		}
		list = append(list, rec)
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].Failures != list[j].Failures {
			return list[i].Failures < list[j].Failures
		}
		if list[i].LastOK != list[j].LastOK {
			return list[i].LastOK > list[j].LastOK
		}
		return list[i].URL < list[j].URL
	})
	if len(list) > 64 {
		list = list[:64]
	}
	b, err := json.MarshalIndent(clientPeerStore{Peers: list}, "", "  ")
	if err != nil {
		return
	}
	dir := filepath.Dir(filepath.Clean(path))
	if dir != "." {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return
		}
	}
	tmp := path + ".tmp"
	if os.WriteFile(tmp, b, 0600) == nil {
		_ = os.Rename(tmp, path)
	}
}

func markClientPeerSuccess(records map[string]clientPeerRecord, endpoint string, now time.Time) {
	rec := records[endpoint]
	rec.URL = endpoint
	rec.Failures = 0
	rec.LastOK = now.Unix()
	rec.QuarantineUntil = 0
	records[endpoint] = rec
}

func markClientPeerFailure(records map[string]clientPeerRecord, endpoint string, now time.Time) {
	rec := records[endpoint]
	rec.URL = endpoint
	rec.Failures++
	if rec.Failures >= 3 {
		rec.QuarantineUntil = now.Add(30 * time.Minute).Unix()
	}
	records[endpoint] = rec
}

func clientExcluded(endpoint string, cfg VerifiedClientConfig) bool {
	for _, raw := range cfg.Exclude {
		if p, ok := normalizeClientEndpoint(raw); ok && p == endpoint {
			return true
		}
	}
	return false
}

func verifyClientEndpoint(ctx context.Context, endpoint string, cfg VerifiedClientConfig) (verifiedEndpoint, error) {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 4 * time.Second
	}
	checkCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cl := NewClient(endpoint)
	cl.HTTP.Timeout = timeout
	st, err := cl.StatusContext(checkCtx)
	if err != nil {
		return verifiedEndpoint{}, err
	}
	if st.NetworkID != cfg.ExpectedNetworkID {
		return verifiedEndpoint{}, fmt.Errorf("%w: %s reports %s", ErrClientNetworkMismatch, endpoint, st.NetworkID)
	}
	return verifiedEndpoint{URL: endpoint, Status: st, Client: cl}, nil
}

func verifyClientCandidates(ctx context.Context, candidates []string, cfg VerifiedClientConfig) ([]verifiedEndpoint, map[string]error) {
	type result struct {
		v   verifiedEndpoint
		err error
		url string
	}
	ch := make(chan result, len(candidates))
	var wg sync.WaitGroup
	for _, endpoint := range candidates {
		endpoint := endpoint
		wg.Add(1)
		go func() {
			defer wg.Done()
			v, err := verifyClientEndpoint(ctx, endpoint, cfg)
			ch <- result{v: v, err: err, url: endpoint}
		}()
	}
	wg.Wait()
	close(ch)

	ok := make([]verifiedEndpoint, 0, len(candidates))
	failed := map[string]error{}
	for r := range ch {
		if r.err != nil {
			failed[r.url] = r.err
			continue
		}
		ok = append(ok, r.v)
	}
	return ok, failed
}

func chooseVerifiedFallback(healthy []verifiedEndpoint) (verifiedEndpoint, error) {
	if len(healthy) == 0 {
		return verifiedEndpoint{}, errors.New("no reachable verified AuronQ node")
	}
	sort.Slice(healthy, func(i, j int) bool {
		if healthy[i].Status.Height != healthy[j].Status.Height {
			return healthy[i].Status.Height > healthy[j].Status.Height
		}
		return healthy[i].URL < healthy[j].URL
	})
	maxHeight := healthy[0].Status.Height
	tip := healthy[0].Status.Tip
	for _, v := range healthy[1:] {
		if v.Status.Height != maxHeight {
			break
		}
		if v.Status.Tip != tip {
			return verifiedEndpoint{}, fmt.Errorf("%w at height %d between %s and %s", ErrClientChainConflict, maxHeight, healthy[0].URL, v.URL)
		}
	}
	return healthy[0], nil
}

func learnVerifiedClientPeers(ctx context.Context, selected verifiedEndpoint, cfg VerifiedClientConfig, records map[string]clientPeerRecord) {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 4 * time.Second
	}
	helloCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	h, err := selected.Client.HelloContext(helloCtx)
	if err != nil || h.ProtocolVersion != 1 || h.NetworkID != cfg.ExpectedNetworkID {
		return
	}
	raw := make([]string, 0, len(h.Peers)+1)
	if h.Advertise != "" {
		raw = append(raw, h.Advertise)
	}
	raw = append(raw, h.Peers...)
	seen := map[string]bool{selected.URL: true}
	candidates := make([]string, 0, 16)
	for _, item := range raw {
		p, ok := normalizeClientEndpoint(item)
		if !ok || seen[p] || clientExcluded(p, cfg) {
			continue
		}
		seen[p] = true
		candidates = append(candidates, p)
		if len(candidates) >= 16 {
			break
		}
	}
	if len(candidates) == 0 {
		return
	}
	healthy, _ := verifyClientCandidates(ctx, candidates, cfg)
	now := time.Now()
	for _, v := range healthy {
		markClientPeerSuccess(records, v.URL, now)
	}
}

func ResolveVerifiedClient(ctx context.Context, cfg VerifiedClientConfig) (VerifiedClientResolution, error) {
	if cfg.MaxCandidates <= 0 || cfg.MaxCandidates > 64 {
		cfg.MaxCandidates = 32
	}
	now := time.Now()
	records := loadClientPeerStore(cfg.StorePath)
	preferred, havePreferred := normalizeClientEndpoint(cfg.Preferred)

	if havePreferred && !clientExcluded(preferred, cfg) {
		if v, err := verifyClientEndpoint(ctx, preferred, cfg); err == nil {
			markClientPeerSuccess(records, preferred, now)
			learnVerifiedClientPeers(ctx, v, cfg, records)
			saveClientPeerStore(cfg.StorePath, records)
			return VerifiedClientResolution{Client: v.Client, Endpoint: v.URL, Status: v.Status}, nil
		} else {
			markClientPeerFailure(records, preferred, now)
			if errors.Is(err, ErrClientNetworkMismatch) {
				saveClientPeerStore(cfg.StorePath, records)
				return VerifiedClientResolution{}, err
			}
		}
	}

	seen := map[string]bool{}
	if havePreferred {
		seen[preferred] = true
	}
	candidates := make([]string, 0, cfg.MaxCandidates)
	add := func(raw string) {
		if len(candidates) >= cfg.MaxCandidates {
			return
		}
		p, ok := normalizeClientEndpoint(raw)
		if !ok || seen[p] || clientExcluded(p, cfg) {
			return
		}
		seen[p] = true
		if rec, ok := records[p]; ok && rec.QuarantineUntil > now.Unix() {
			return
		}
		candidates = append(candidates, p)
	}
	storeList := make([]clientPeerRecord, 0, len(records))
	for _, rec := range records {
		storeList = append(storeList, rec)
	}
	sort.Slice(storeList, func(i, j int) bool {
		if storeList[i].LastOK != storeList[j].LastOK {
			return storeList[i].LastOK > storeList[j].LastOK
		}
		return storeList[i].Failures < storeList[j].Failures
	})
	for _, rec := range storeList {
		add(rec.URL)
	}
	for _, seed := range cfg.Seeds {
		add(seed)
	}
	if len(candidates) == 0 {
		saveClientPeerStore(cfg.StorePath, records)
		return VerifiedClientResolution{}, errors.New("no AuronQ node candidates available")
	}

	healthy, failed := verifyClientCandidates(ctx, candidates, cfg)
	for _, v := range healthy {
		markClientPeerSuccess(records, v.URL, now)
	}
	for endpoint := range failed {
		markClientPeerFailure(records, endpoint, now)
	}
	selected, err := chooseVerifiedFallback(healthy)
	if err != nil {
		saveClientPeerStore(cfg.StorePath, records)
		return VerifiedClientResolution{}, err
	}
	learnVerifiedClientPeers(ctx, selected, cfg, records)
	saveClientPeerStore(cfg.StorePath, records)
	return VerifiedClientResolution{Client: selected.Client, Endpoint: selected.URL, Status: selected.Status}, nil
}

func IsRetryableNodeError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}
	if errors.Is(err, context.Canceled) {
		return false
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return true
	}
	var httpErr *nodeHTTPError
	if errors.As(err, &httpErr) {
		return httpErr.StatusCode == 502 || httpErr.StatusCode == 503 || httpErr.StatusCode == 504
	}
	s := strings.ToLower(err.Error())
	for _, frag := range []string{"connection refused", "connection reset", "no such host", "server misbehaving", "broken pipe"} {
		if strings.Contains(s, frag) {
			return true
		}
	}
	return false
}
