package main

import (
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
