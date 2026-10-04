package main

import (
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	aq "auronq/internal/auronq"
)

func TestCheckPeerAcceptsMainnetHelloAndExplorer(t *testing.T) {
	var id aq.Hash
	raw, err := hex.DecodeString(mainnetNetworkID)
	if err != nil {
		t.Fatal(err)
	}
	copy(id[:], raw)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/p2p/hello":
			_ = json.NewEncoder(w).Encode(aq.Hello{ProtocolVersion: 1, NetworkID: id, Height: 123})
		case "/explorer":
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte("<title>AuronQ Explorer</title>"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	got, err := checkPeer(srv.Client(), srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if got.Hello.Height != 123 || !got.Explorer {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestCheckPeerRejectsWrongNetwork(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/p2p/hello" {
			_ = json.NewEncoder(w).Encode(aq.Hello{ProtocolVersion: 1, Height: 999})
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	if _, err := checkPeer(srv.Client(), srv.URL); err == nil {
		t.Fatal("wrong network was accepted")
	}
}


func headerServer(t *testing.T, headers map[uint64]aq.BlockHeader) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/p2p/headers" {
			http.NotFound(w, r)
			return
		}
		height, err := strconv.ParseUint(r.URL.Query().Get("start"), 10, 64)
		if err != nil {
			t.Fatalf("bad start: %v", err)
		}
		h, ok := headers[height]
		if !ok {
			_ = json.NewEncoder(w).Encode(map[string]any{"start": height, "headers": []aq.BlockHeader{}})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"start": height, "headers": []aq.BlockHeader{h}})
	}))
}

func TestVerifyCommonHistoryAcceptsLaggingPeerOnSameChain(t *testing.T) {
	common := aq.BlockHeader{Height: 10, Timestamp: 1000, Nonce: 7}
	a := headerServer(t, map[uint64]aq.BlockHeader{10: common})
	defer a.Close()
	b := headerServer(t, map[uint64]aq.BlockHeader{10: common})
	defer b.Close()

	healthy := []result{
		{Peer: a.URL, Hello: aq.Hello{Height: 10}},
		{Peer: b.URL, Hello: aq.Hello{Height: 12}},
	}
	if err := verifyCommonHistory(a.Client(), healthy); err != nil {
		t.Fatalf("same history rejected: %v", err)
	}
}

func TestVerifyCommonHistoryDetectsForkAtCommonHeight(t *testing.T) {
	aHeader := aq.BlockHeader{Height: 10, Timestamp: 1000, Nonce: 1}
	bHeader := aq.BlockHeader{Height: 10, Timestamp: 1000, Nonce: 2}
	a := headerServer(t, map[uint64]aq.BlockHeader{10: aHeader})
	defer a.Close()
	b := headerServer(t, map[uint64]aq.BlockHeader{10: bHeader})
	defer b.Close()

	healthy := []result{
		{Peer: a.URL, Hello: aq.Hello{Height: 10}},
		{Peer: b.URL, Hello: aq.Hello{Height: 11}},
	}
	if err := verifyCommonHistory(a.Client(), healthy); err == nil {
		t.Fatal("forked common history was not detected")
	}
}
