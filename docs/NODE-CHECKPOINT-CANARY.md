# AuronQ Mainnet v1.7.14 — fail-closed checkpoint canary plan

## Rule 0: keep the currently operating Mainnet alive

The new P2P checkpoint mechanism is **advisory, off by default and not a consensus change**. No Mainnet genesis, Network ID, AQM64, difficulty, block/transaction serialization, UTXO, monetary rules or full-node canonical history validation may change.

**Never perform an in-place upgrade on every existing public node at once.** No auto-installation and no automated miner/wallet rollout. The official 1.7.13 release remains available for rollback.

## Release gates

Before publishing a 1.7.14 prerelease, require:

1. Protected-main CI pass on Linux, Windows and Android where applicable; frozen Mainnet identity and genesis checks.
2. A source diff safety gate against the previously merged P2P baseline `c1001ee2950158781f216820cc5e4911aa0044a9`: no changes to Mainnet network configuration, transaction validity or consensus code.
3. New tests that confirm the checkpoint feature is disabled by default; disabled mode does not create checkpoint signing material; peer-checkpoint request throttling cannot consume the block/header rate budget; malformed/signed-network-mismatched checkpoints are rejected; full-node canonical state is unchanged by checkpoint requests.
4. Full Go unit/vet tests and race detector, plus Windows and Linux build checks.
5. **Explicit manual execution** of the release workflow by the operator after reviewing results. A merge to `main` alone must not publish the release.

## Phase A — canary, with full rollback

- Start on a **new or noncritical validating node** with a **separate, copied full chain data directory** and a copy of wallet recovery material if relevant. Do not point a test process at production writable data simultaneously.
- Back up the node's data and store a SHA-256 inventory of the old program binary and genesis/network configuration. Do not publish wallet private keys or local node identity seeds.
- Install only the 1.7.14 binaries on the canary. Run once in default mode; the endpoint `/p2p/checkpoint` should respond with 404 and the node should continue synchronizing normally.
- On the canary only, set `AURONQ_ENABLE_P2P_CHECKPOINTS=1` in the node process environment and restart the node. This turns on the nonconsensus P2P service. The setting creates an independent node identity automatically; no checkpoint private key has to be manually provided.
- Check `/v1/status` for the same Network ID, canonical tip and cumulative chainwork as unchanged peers. Check `/p2p/checkpoint` for a valid self-signed advisory checkpoint at the highest locally validated multiple of 256. Observe `peer-checkpoints/latest.json` as the network progresses.
- Record 24 hours of block acceptance, mining, sync, peer connectivity and resource usage. A 256-block boundary can be difficult to reach in a day on a slow chain; do not fake progress or infer success from a single snapshot.

## Phase B — independent operators

- Invite **at least two real independently managed full nodes** to test, preferably different physical locations, network providers and administrators. Comparing two HTTP endpoints under one operator does not provide true decentralization.
- Confirm old 1.7.13 and new 1.7.14 nodes remain connected and agree on canonical chainwork, block heights and tips at identical timestamps. New checkpoint hints must not authorize funds or override full-chain validation.
- Disconnect one node or its Internet connection at a time; verify a remaining node continues canonical chain sync and mining with existing peers. Independently test discovery when GitHub/bootstrap endpoints are unreachable **without** removing your last historical chain copies.
- Do **not** turn on checkpoint service everywhere until independent canaries have demonstrated safety.

## Immediate rollback triggers

Stop canary service if an unexpected fork, divergence in work at the same height, failure to accept valid blocks, mining interruption, high CPU/memory use, new key/file corruption or repeatable synchronization regression occurs.

To roll back:
1. Stop **only the canary** node process.
2. Remove `AURONQ_ENABLE_P2P_CHECKPOINTS` from its environment.
3. Restart using 1.7.13 with the **original backed-up data directory**; never delete genesis or existing blockchain files.
4. Keep other Mainnet nodes unchanged and document any discrepancy. Do not reinitialize Mainnet to fix a local service problem.

## Security boundary

A node-local signature proves only which node signed the advisory checkpoint. Neither a majority vote of self-generated signing keys nor multiple IP addresses gives cryptographic proof of honest historical chainwork. **AuronQ Mobile must never silently trust these attestations as historical consensus**. Existing release-pinned mobile checkpoints are a separate trust model and must be reviewed separately.

Checkpoint hints do not make a network immortal: preserving full historical data, geographically and administratively independent full nodes, reachable peer discovery, and independent miners are essential. No GitHub, Sigstore, pool or single operator must be able to rewrite ordinary consensus.

## Status

Code PR #145 was merged after successful Linux/Windows/Android tests. This v1.7.14 release candidate keeps the P2P checkpoint canary **disabled by default** and requires a separately gated, manual release. **Building or publishing binaries does not upgrade any running node.**
