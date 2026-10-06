//go:build windows

package main

import (
	"errors"
	"strings"
	"testing"
)

func TestClassifyPublicNodeError(t *testing.T) {
	cases := []struct {
		err string
		want string
	}{
		{"router does not expose a public WAN IP (NAT/CGNAT likely)", "CGNAT"},
		{"UPnP/IGD not available on this network", "UPnP"},
		{"UPnP AddPortMapping HTTP 500", "TCP/18444"},
	}
	for _, tc := range cases {
		got := classifyPublicNodeError(errors.New(tc.err))
		if !strings.Contains(got, tc.want) {
			t.Fatalf("classify %q = %q, want substring %q", tc.err, got, tc.want)
		}
	}
}

func TestPublicNodeUIIsNodeScoped(t *testing.T) {
	b, err := webFS.ReadFile("web/index.html")
	if err != nil { t.Fatal(err) }
	html := string(b)
	if strings.Contains(html, "Wspieraj sieć jako publiczny node podczas kopania") {
		t.Fatal("legacy mining-scoped public-node label still present")
	}
	for _, needle := range []string{"public_state", "nPublicState", "nPublicError", "publicLabel(s)"} {
		if !strings.Contains(html, needle) {
			t.Fatalf("public-node diagnostics UI missing %q", needle)
		}
	}
}
