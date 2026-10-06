package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDesktopExplorerTabIsWired(t *testing.T) {
	b, err := webFS.ReadFile("web/index.html")
	if err != nil {
		t.Fatal(err)
	}
	html := string(b)
	checks := []string{
		`data-view="explorer"`,
		`id="view-explorer"`,
		`http://127.0.0.1:18444/explorer`,
		`function updateExplorerView`,
		`explorer:t('Explorer')`,
	}
	for _, want := range checks {
		if !strings.Contains(html, want) {
			t.Fatalf("desktop explorer wiring missing %q", want)
		}
	}
}

func TestDesktopCSPAllowsOnlyLocalExplorerFrame(t *testing.T) {
	h := securityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:18445/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	csp := rec.Header().Get("Content-Security-Policy")
	var frameDirective string
	for _, part := range strings.Split(csp, ";") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "frame-src ") {
			frameDirective = part
			break
		}
	}
	if frameDirective != "frame-src http://127.0.0.1:18444" {
		t.Fatalf("unexpected frame-src directive: %q (full CSP: %q)", frameDirective, csp)
	}
}

func TestDesktopShowsPublicNodeReadiness(t *testing.T) {
	b, err := webFS.ReadFile("web/index.html")
	if err != nil {
		t.Fatal(err)
	}
	html := string(b)
	for _, want := range []string{
		`id="publicNodeBadge"`,
		`id="publicNodeEndpoint"`,
		`state.public_endpoint`,
		`TCP/18444`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("public-node readiness UI missing %q", want)
		}
	}
}

func TestDesktopPublicNodeIsIndependentOfMining(t *testing.T) {
	b, err := webFS.ReadFile("web/index.html")
	if err != nil {
		t.Fatal(err)
	}
	html := string(b)
	for _, want := range []string{
		`Publiczny node działa niezależnie od kopania.`,
		`Zatrzymanie minera nie wyłącza publicznego peera.`,
		`PublicNode:true`,
		`TCP/18444`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("node-scoped public-node UI missing %q", want)
		}
	}
	if strings.Contains(html, "Wspieraj sieć jako publiczny node podczas kopania") {
		t.Fatal("legacy mining-scoped public-node control is still visible")
	}
}
