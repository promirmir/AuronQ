package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	aq "auronq/internal/auronq"
)

func TestCheckPeerAcceptsMainnetHelloAndExplorer(t *testing.T) {
	var id aq.Hash
	b, err := hexHash(mainnetNetworkID)
	if err != nil {
		t.Fatal(err)
	}
	id = b

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

func hexHash(s string) (aq.Hash, error) {
	var h aq.Hash
	b := make([]byte, len(h))
	for i := range b {
		var v byte
		for j := 0; j < 2; j++ {
			c := s[i*2+j]
			v <<= 4
			switch {
			case c >= '0' && c <= '9':
				v |= c - '0'
			case c >= 'a' && c <= 'f':
				v |= c - 'a' + 10
			case c >= 'A' && c <= 'F':
				v |= c - 'A' + 10
			default:
				return h, &hexError{c}
			}
		}
		b[i] = v
	}
	copy(h[:], b)
	return h, nil
}

type hexError struct{ c byte }
func (e *hexError) Error() string { return "invalid hex character" }
