# AuronQ release test matrix

This file is the engineering gate for moving from public testnet to mainnet. A passing unit test suite alone is not enough.

## Consensus and chain selection

- valid block extension
- invalid previous hash
- invalid proof of work
- incorrect target
- bad timestamp / median-time-past violation
- future timestamp rejection
- incorrect coinbase height
- incorrect subsidy or fee claim
- duplicate transaction in a block
- block double spend
- immature coinbase spend
- maximum-supply enforcement
- halving boundaries
- higher-work reorganization
- lower/equal-work fork rejection
- disconnected transaction revalidation after reorg

## Transactions and wallet

- valid ML-DSA-87 spend
- modified-message signature rejection
- wrong public key rejection
- duplicate input rejection
- mempool double-spend rejection
- minimum relay fee
- timelock enforcement
- wallet backup/import round trip
- encrypted wallet wrong-password failure
- transaction history: pending, confirmed, maturing, reorged

## P2P

- fresh node bootstraps from DNS seeds
- fallback seed peer works when DNS seed is unavailable
- peer gossip learns only public routable addresses
- peer netgroup diversity cap
- dead learned peers are pruned
- bootstrap peers are retained
- block relay across at least three nodes
- mempool relay and catch-up
- long initial sync in bounded batches
- fork/reorg propagation
- restart and peer-cache recovery
- malformed / oversized request handling
- slow or unresponsive peer behavior
- eclipse/Sybil simulation
- connection-flood / request-flood testing

## Platforms and deployment

- Windows 10 x64
- Windows 11 x64
- current Ubuntu LTS x64
- two or more independent residential ISPs
- public VPS nodes on at least two providers
- NAT and CGNAT clients
- public IPv4 and IPv6 where available
- clean install from GitHub release assets with no manual peer entry

## Release engineering

- `go test ./...`
- `go vet ./...`
- smoke test
- two-node integration test
- fuzzing campaign for parsers/serialization/P2P handlers
- dependency/license review
- release SHA-256 checksums
- documented build procedure
- reproducible-build investigation
- independent consensus/cryptography/security review

Mainnet genesis is created only after the consensus rules are frozen and this matrix has no unresolved critical or high-severity failures.

## AQM64 gate

- [x] deterministic AQM64 test vector in the Go test suite
- [x] Argon2id known-answer cross-check during development
- [ ] independent AQM64 implementation reproduces identical vectors
- [ ] benchmark on multiple Intel and AMD CPUs
- [ ] GPU benchmark/optimization study
- [ ] memory/time tradeoff review
- [ ] malicious-block CPU/RAM exhaustion test
- [x] rc2 compatibility freeze commits the launch PoW limit into Network ID (security/economic validation remains outstanding)


## 2026-10-02 rc2 internal verification

Completed in the internal pre-release audit: Go unit/integration suite, `go vet`, race-enabled suite during hardening, Windows/Linux cross-builds, production-parameter AQM64 vector, ML-DSA-87 cross-backend interoperability against OpenSSL 3.5.5, bounded-sync/fork hardening review and basic per-IP rate limiting.

Still external/open: independent security review, extended fuzzing, multi-provider hostile public-network soak, distributed request-flood/eclipsing/Sybil tests, independent AQM64 implementation/economic study, and independently reproduced byte-identical release builds. See `AUDIT-2026-10-02.md`.

## 2026-10-02 rc3 zero-touch bootstrap verification

- [x] bundled `bootstrap.json` is accepted only for the exact Network ID
- [x] remote HTTPS bootstrap manifest support added without changing Network ID inputs
- [x] remote manifest size and peer-count bounded
- [x] strict JSON / trailing-data rejection
- [x] wrong-network and expired manifests rejected by tests
- [x] private/CGNAT literal peers from remote manifests rejected
- [x] DNS peers from remote manifests require HTTPS
- [x] dead manifest peers remain pruneable
- [x] Desktop dual-stack listen changed to `[::]:18444`
- [x] go test ./...
- [x] go vet ./...
- [x] go test -race ./...
- [x] deterministic Windows/Linux release rebuild comparison


## 2026-10-02 rc4 live two-node mainnet verification

- [x] reproduced rc3 missed-broadcast recovery gap with node A at height 0 and node B at height 4
- [x] rc4 same-branch repair pushed the four missing validated blocks from the stronger outbound node to the lagging peer
- [x] both independent nodes converged to height 4 and identical tip
- [x] mined block 5 on node B propagated to node A
- [x] both nodes converged to height 5, identical tip, chain work and issuance
- [x] upgrade bootstrap merge is limited to the exact same Network ID and is idempotent

The live test proves the exercised two-node path; it does not replace independent security review, large-scale adversarial testing, or multi-operator bootstrap redundancy.
