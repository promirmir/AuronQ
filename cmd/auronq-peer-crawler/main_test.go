package main

import (
	"errors"
	"net"
	"testing"
)

func TestPublicGossipPeer(t *testing.T) {
	good := []string{
		"http://8.8.8.8:18444",
		"https://1.1.1.1:18444",
		"http://[2606:4700:4700::1111]:18444",
		"https://node.example.com",
	}
	for _, p := range good {
		if !publicGossipPeer(p) {
			t.Fatalf("expected public peer: %s", p)
		}
	}
	bad := []string{
		"http://127.0.0.1:18444",
		"http://10.0.0.5:18444",
		"http://192.168.1.5:18444",
		"http://100.64.0.1:18444",
		"http://192.0.2.1:18444",
		"http://198.51.100.2:18444",
		"http://203.0.113.3:18444",
		"http://[2001:db8::1]:18444",
	}
	if !publicGossipPeer("https://node.example.com") {
		t.Fatal("expected HTTPS DNS peer to be accepted")
	}
	for _, p := range bad {
		if publicGossipPeer(p) {
			t.Fatalf("unexpected public peer: %s", p)
		}
	}
}

func TestSafeConfiguredPeer(t *testing.T) {
	if !safeConfiguredPeer("https://seed.example.org") {
		t.Fatal("expected HTTPS DNS seed to be accepted as configured metadata")
	}
	for _, p := range []string{
		"http://seed.example.org",
		"https://localhost",
		"https://node.local",
		"http://10.0.0.1:18444",
	} {
		if safeConfiguredPeer(p) {
			t.Fatalf("unsafe configured peer accepted: %s", p)
		}
	}
}

func TestNetgroup(t *testing.T) {
	if a, b := netgroup("http://8.8.1.1:18444"), netgroup("http://8.8.9.9:18444"); a == "" || a != b {
		t.Fatalf("same IPv4 /16 should share netgroup: %q %q", a, b)
	}
	if a, b := netgroup("http://8.8.1.1:18444"), netgroup("http://1.1.1.1:18444"); a == b {
		t.Fatalf("different IPv4 /16 should differ: %q", a)
	}
}


func TestDNSNetgroupUsesParentDomain(t *testing.T) {
	if a, b := netgroup("https://a.example.com"), netgroup("https://b.example.com"); a == "" || a != b {
		t.Fatalf("same parent DNS domain should share netgroup: %q %q", a, b)
	}
	if a, b := netgroup("https://one.example.com"), netgroup("https://two.other.net"); a == b {
		t.Fatalf("different parent DNS domains should differ: %q", a)
	}
	if a, b := netgroup("https://node-a.example.co.uk"), netgroup("https://node-b.example.co.uk"); a == "" || a != b {
		t.Fatalf("same ccTLD parent DNS domain should share netgroup: %q %q", a, b)
	}
}


func TestPermanentLookupFailure(t *testing.T) {
	nxdomain := &net.DNSError{Name: "dead.example", Err: "no such host", IsNotFound: true}
	if !isPermanentLookupFailure(nxdomain) {
		t.Fatal("NXDOMAIN/no-such-host should be permanent")
	}
	if isPermanentLookupFailure(errors.New("context deadline exceeded")) {
		t.Fatal("transient timeout must not be treated as permanent")
	}
}


func TestCrawlerEndorsementsRequireDistinctNetgroups(t *testing.T) {
	c := &crawler{endorsements: map[string]map[string]bool{}}
	candidate := "http://9.9.9.9:18444"

	c.endorse(candidate, "http://8.8.1.1:18444")
	c.endorse(candidate, "http://8.8.2.2:18444") // same /16, not independent
	if got := c.endorsementCount(candidate); got != 1 {
		t.Fatalf("same-netgroup endorsements=%d want 1", got)
	}
	if c.eligibleLearnedPeer(candidate, false) {
		t.Fatal("single-netgroup learned peer became eligible")
	}

	c.endorse(candidate, "http://1.1.1.1:18444")
	if got := c.endorsementCount(candidate); got != 2 {
		t.Fatalf("independent endorsements=%d want 2", got)
	}
	if !c.eligibleLearnedPeer(candidate, false) {
		t.Fatal("two-netgroup learned peer did not become eligible")
	}
}

func TestCrawlerSelfAdvertisementIsNotAnEndorsement(t *testing.T) {
	c := &crawler{endorsements: map[string]map[string]bool{}}
	peer := "http://8.8.8.8:18444"
	c.endorse(peer, peer)
	if got := c.endorsementCount(peer); got != 0 {
		t.Fatalf("self endorsement count=%d want 0", got)
	}
}

func TestExistingManifestPeerDoesNotNeedFreshEndorsements(t *testing.T) {
	c := &crawler{endorsements: map[string]map[string]bool{}}
	if !c.eligibleLearnedPeer("http://8.8.8.8:18444", true) {
		t.Fatal("existing reviewed peer should remain eligible without fresh gossip quorum")
	}
}
