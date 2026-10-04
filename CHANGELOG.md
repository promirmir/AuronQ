# AuronQ Changelog

This file is the concise public history of AuronQ. Detailed historical release notes and old checksum files are preserved under [docs/archive/](docs/archive/).

## Current releases

### Desktop / Full Node v1.7.13 — Mainnet patch

- Fixed repeated retries of dead DNS peers learned through P2P gossip or bootstrap metadata.
- DNS peers returning a permanent not-found result are removed immediately and quarantined for 30 minutes.
- Repeatedly failing peers are quarantined so gossip/bootstrap refresh cannot instantly resurrect them.
- The public peer-registry crawler now drops existing manifest DNS endpoints that return a permanent DNS not-found result instead of preserving them forever.
- Public peer discovery runs hourly and groups DNS peers by parent domain for better infrastructure diversity.
- Directly addressed public VPS/server nodes can auto-advertise their public interface endpoint when `--advertise` is omitted; peers still callback-verify reachability before gossip admission.
- Windows release packaging now includes `START-PUBLIC-NODE.cmd` for directly reachable independent full nodes.
- Added regression tests for dead-peer quarantine, manifest re-add prevention and DNS peer grouping.
- Consensus, AQM64, difficulty, Network ID, genesis, monetary policy and transaction rules are unchanged.

### Desktop / Full Node v1.7.12 — Mainnet

- Fixed stale mining templates in CLI and Desktop mining.
- Miners now watch the connected node's canonical tip while hashing.
- If another miner advances the chain, obsolete work is cancelled and a fresh template is fetched promptly.
- A same-height reorganization also invalidates the current template through PrevHash comparison.
- Added regression coverage for tip advances, same-height reorgs and caller cancellation.
- Corrected the CLI-reported software version to 1.7.12.
- Consensus, AQM64, difficulty, Network ID, genesis, monetary policy and transaction rules are unchanged.

### Desktop / Full Node v1.7.11 — Previous Mainnet patch

- Built-in Explorer lists alternative HTTPS full-node Explorers learned through native AuronQ P2P gossip.
- No Explorer is canonical or trusted; each full node serves data from its own locally validated chain.
- Consensus, Network ID, genesis, monetary policy, AQM64, signature and transaction rules are unchanged from the preceding 1.7.x hardening releases.

### AuronQ Mobile 0.5.1 Alpha

- Independently validates the Mainnet header chain from embedded genesis.
- Locally verifies AQM64 proof-of-work, difficulty, timestamps and chain continuity.
- Accepts balance, history and spendable UTXO state only from peers matching the locally verified header state.
- Compares canonical wallet state across agreeing peers and fails closed on conflicts.
- Keeps ML-DSA-87 private keys and signing local.
- Broadcasts signed transactions directly to agreeing verified-chain peers.
- Remains a light client rather than a full UTXO-validating node.

## Desktop / Full Node history

### v1.7.10
- Made configured seed/manual peers replaceable startup hints instead of permanent runtime authorities.
- Repeatedly failing seed peers are pruned.
- Added founder-node-removal regression coverage: the original bootstrap disappears permanently and a fresh node joins through surviving non-founder peers.
- Clarified Explorer as a local view of each node's own validated canonical chain.

### v1.7.9
- Strengthened eclipse/Sybil resistance through peer netgroups and DNS parent-domain grouping.
- Added sync-order diversity across peer groups.
- Added consensus invariant tests for difficulty clamps, median-time-past, future timestamps, coinbase maturity and same-block double spends.
- Added deterministic/reproducible build verification.

### v1.7.8
- Added direct bootstrap fallbacks in addition to the manifest.
- Improved higher-work reorg recovery and immediate post-reorg synchronization.
- Added a 20-node / four-partition convergence and restart test.
- Expanded fuzzing to transaction and block JSON paths.
- Added hourly public Mainnet health monitoring.
- Updated automated peer-registry changes to respect protected `main`.

### v1.7.7
- Fixed Desktop Explorer iframe CSP compatibility while keeping frame access restricted to the local AuronQ node.

### v1.7.6
- Added the dedicated Explorer tab to AuronQ Desktop.

### v1.7.5
- Added the built-in read-only blockchain Explorer at `/explorer`.
- Added search by address, TXID, block height and block hash plus block/transaction/address detail views.

### v1.7.4
- Added wallet/light-client history API.
- Added estimated AQM64 network hash power.
- Added Desktop network-power telemetry and Android history support.
- Added Android CI.

### v1.7.3
- Fixed Windows seed-node launcher handling for Tailscale discovery/Funnel setup.

### v1.7.2
- Added verified HTTPS/DNS peer discovery and gossip with resolver filtering.
- Added public endpoint advertisement support and autonomous peer-crawler verification.

### v1.7.1
- Added guarded Desktop wallet deletion with exact-name confirmation, active-miner protection and backup warnings.

### v1.7.0 — Mainnet launch
- Promoted the tested release-candidate consensus/network code to public Mainnet without changing the established genesis or Network ID.
- Shipped persistent peer storage, peer gossip, fixed seeds, DNS-seed support and bounded HTTPS bootstrap manifests.
- Published Windows/Linux release packages and zero-touch Windows startup tooling.

## Mobile history

### 0.4.2 Alpha
- Introduced multi-peer Mainnet state agreement.
- Compared balances only across peers reporting the same chain state.
- Failed closed on conflicting same-chain balances.
- Added direct transaction fanout after local signing.
- Still depended on full nodes for chain validation.

### 0.4.0 Alpha
- Added mobile wallet history against public 1.7.4+ nodes.

### 0.1.0–0.3.0 Alpha
- Early Android wallet/bridge releases preserved in the release archive.

## Earlier development

Versions 1.2.x through 1.6.x and 1.7.0 release-candidate material are historical development builds. Their original notes are preserved under:

- [docs/archive/releases/](docs/archive/releases/)
- [docs/archive/rc/](docs/archive/rc/)
- [docs/archive/checksums/](docs/archive/checksums/)

These historical files are retained for provenance, not as recommended downloads.

For current software, always use the release links in the main [README](README.md).
