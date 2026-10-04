# AuronQ 1.7.0-rc3 bootstrap/discovery audit addendum

Scope: changes from 1.7.0-rc2 to rc3 only. The rc2 consensus/security audit remains the baseline; rc3 does not claim an independent audit.

## Changes reviewed

- Windows Desktop first-run can merge a bundled `bootstrap.json` only after `network.json` validation and only when the manifest Network ID matches.
- `network.json` may list HTTPS `bootstrap_manifests`, refreshed periodically by the node.
- Remote manifests are limited to 64 KiB / 64 peers, use strict JSON parsing, reject trailing JSON, require the exact Network ID and optionally enforce expiry.
- DNS peer names learned from a remote manifest require HTTPS. Literal private, loopback, CGNAT, link-local and documentation-only IPs are rejected.
- Manifest peers are deliberately not permanent bootstrap entries; repeated failures can prune dead tunnel URLs.
- Bootstrap metadata remains outside Network ID/consensus.
- Windows Desktop node listens on `[::]:18444` for IPv6-capable dual-stack operation.

## Verification

`go test ./...`, `go vet ./...`, `go test -race ./...` pass on the frozen rc3 tree. Windows/Linux release binaries were cross-built twice with Go 1.23.2 and identical release flags; all compared byte-for-byte equal.

## Remaining limitations

Bootstrap still requires at least one reachable rendezvous path for a completely fresh installation. A single bootstrap URL is an availability/eclipsing risk even though it has no consensus authority. The launch release should therefore add independent seeds/manifests as soon as possible. Long-term public operation still needs independent consensus/P2P review, Sybil/eclipse testing, distributed DoS testing and external AQM64 analysis.
