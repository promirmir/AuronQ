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
	if !strings.Contains(csp, "frame-src http://127.0.0.1:18444") {
		t.Fatalf("local explorer frame source missing from CSP: %q", csp)
	}
	if strings.Contains(csp, "frame-src *") || strings.Contains(csp, "frame-src http:") || strings.Contains(csp, "frame-src https:") {
		t.Fatalf("CSP frame source is too broad: %q", csp)
	}
}
