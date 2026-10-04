package main

import (
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

