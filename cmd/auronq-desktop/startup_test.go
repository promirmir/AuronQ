package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"strings"
	"time"

	aq "auronq/internal/auronq"
)

func TestDesktopStartupPhases(t *testing.T) {
	a := &App{}
	s := a.startupSnapshot()
	if s.Phase != "setup" || s.NetworkLoaded {
		t.Fatalf("missing network: %+v", s)
	}

	a.network = &aq.NetworkConfig{}
	a.nodeStarting = true
	a.nodeStage = "Otwieranie lokalnego blockchaina..."
	s = a.startupSnapshot()
	if s.Phase != "preparing" || s.Message != a.nodeStage {
		t.Fatalf("opening chain: %+v", s)
	}

	a.nodeStarting = false
	a.nodeErr = "cannot open chain"
	s = a.startupSnapshot()
	if s.Phase != "error" || s.Error != a.nodeErr {
		t.Fatalf("failed open: %+v", s)
	}

	a.nodeErr = ""
	s = a.startupSnapshot()
	if s.Phase != "offline" {
		t.Fatalf("stopped node: %+v", s)
	}
}

func TestDesktopStartupReportsObservedPeersAndHeight(t *testing.T) {
	peers := 3
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/status" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"height": 42, "peers": peers})
	}))
	defer server.Close()

	a := &App{network: &aq.NetworkConfig{}, nodeURL: server.URL, nodeRun: true}
	s := a.startupSnapshot()
	if s.Phase != "connected" || s.Height != 42 || s.Peers != 3 {
		t.Fatalf("connected status: %+v", s)
	}

	peers = 0
	s = a.startupSnapshot()
	if s.Phase != "searching" || s.Height != 42 || s.Peers != 0 {
		t.Fatalf("peer discovery status: %+v", s)
	}
}

func TestDesktopStartupAPIRespondsWhenNodeIsSlow(t *testing.T) {
	// The /api/startup handler must never wait for a 30s wallet balance call.
	node := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer node.Close()

	a := &App{network: &aq.NetworkConfig{}, nodeURL: node.URL, nodeRun: true}
	mux := http.NewServeMux()
	a.routes(mux)
	start := time.Now()
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/startup", nil))
	if time.Since(start) > 2*time.Second {
		t.Fatalf("startup API blocked for %v", time.Since(start))
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("startup API returned %d", rec.Code)
	}
	var result desktopStartupState
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Phase != "connecting" {
		t.Fatalf("slow node status: %+v", result)
	}
}

func TestDesktopStartupWindowHasIndependentStatusPolling(t *testing.T) {
	b, err := webFS.ReadFile("web/index.html")
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	for _, want := range []string{
		`id="startupPanel"`,
		`id="startupTitle"`,
		`id="startupMetrics"`,
		`function refreshStartup()`,
		`/api/startup`,
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("missing visible startup marker %q", want)
		}
	}
}
