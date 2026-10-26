# AuronQ network independence

AuronQ Mainnet has no founder/master node and no privileged validation server.

## What already works without the original computer

Full nodes:

- keep their own complete canonical chain;
- independently validate blocks and transactions;
- select chains by cumulative work;
- persist verified public peers to disk;
- exchange bounded public peer lists through `/p2p/hello`;
- reconnect to known peers after restart;
- relay blocks and transactions directly between peers.

Therefore, once multiple full nodes know one another, the original bootstrap computer is not required for consensus or continued block production.

## Runtime independence requirement

AuronQ applications must not require GitHub to join or continue using Mainnet. GitHub-hosted manifests are optional freshness/distribution helpers only.

- Full Node/Desktop must retain multiple bundled public seed peers and persisted/gossiped peer recovery.
- Android must retain multiple bundled public peer fallbacks, prefer cached/bundled peers before optional manifests, and persist only independently verified gossip candidates.
- CLI thin clients and standalone miners must prefer the configured/local node, then locally cached verified peers, then bundled Mainnet seed hints.
- Thin clients must verify the exact Mainnet Network ID before use, quarantine repeated endpoint failures and fail closed on clear same-height tip conflicts.
- Once any valid Mainnet peer is reached, peer gossip should provide additional discovery routes.
- Only an actual reachable full node may advertise itself as a P2P peer; light clients/miners do not pretend to be full nodes.
- Loss of GitHub must not change Network ID, consensus validation, wallet keys, balances or transaction rules.
- A fresh-install drill with GitHub-based manifests unavailable is part of the independence acceptance gate.

## Fresh-install discovery

A computer with no peer history still needs at least one rendezvous path. Current source ships a reviewed multi-peer seed snapshot and runtime logic that can use bundled/cached peers before any optional GitHub manifest. AuronQ supports:

1. local persisted peers;
2. fixed seed peers;
3. DNS seeds;
4. HTTPS bootstrap manifests;
5. peer gossip after first contact.

Bootstrap metadata is discovery-only and excluded from the Network ID. AuronQ uses two HTTPS manifest layers: a reviewed static fallback from `main` and a supplemental live registry published by the crawler on `automation/peer-registry`. They are optional discovery inputs, not runtime authorities; either can fail without changing consensus.

## Autonomous registry

`.github/workflows/peer-registry.yml` runs the AuronQ peer crawler periodically. The crawler:

- starts from the current manifest;
- requests `/p2p/hello`;
- requires protocol version 1 and the exact AuronQ Mainnet Network ID;
- follows only publicly routable peers learned from gossip;
- requires a newly learned live-registry peer to be advertised by at least two distinct peer netgroups before automatic promotion;
- treats a node's own advertise field as a self-claim, not an independent endorsement;
- rejects loopback, private, CGNAT, link-local and documentation ranges;
- callback-verifies reachability;
- limits peer count and basic network-group concentration;
- preserves existing registry entries across transient outages;
- runs the public network health checker against the generated registry before publishing it;
- pushes the verified live registry to `automation/peer-registry` when there is a change;
- may also prepare an optional reviewed snapshot for protected `main`, but fresh-node discovery no longer needs to wait for that manual review step.

This process is not a consensus oracle. A malicious or broken peer cannot make an invalid chain valid; every full node validates candidate blocks itself.

## When the original computer can be turned off

The strongest practical test is:

1. have at least 3 independent publicly reachable full nodes on different networks/providers;
2. ensure they appear in gossip and the public bootstrap registry and/or DNS seeds;
3. turn the original bootstrap computer completely off;
4. start a clean AuronQ install with no peer cache;
5. confirm it discovers other peers and reaches the same height/tip;
6. mine/relay another block and confirm propagation without the original computer.

Only after this passes should the network be described as operationally independent of the original bootstrap machine.

The live network now has multiple reachable public full nodes and the crawler is discovering additional peers through ordinary gossip. The remaining independence test is operational rather than protocol-level: perform a clean-install join and block-relay drill while the original project-operated rendezvous machines are completely offline, and continue improving provider/operator diversity.

## Packaging status — 26 October 2026

The reviewed runtime-independence implementation from PR #187 and PR #189 is included in **published Desktop v1.7.14, Android 0.5.7 Alpha and Universal Miner v0.4.9 Alpha** (26 October 2026). These builds are available in [official GitHub Releases](https://github.com/promirmir/AuronQ/releases), with published SHA-256 checksums. Previous Desktop v1.7.13, Android 0.5.6 Alpha and Miner v0.4.8 Alpha binaries remain historical and do not update themselves automatically. Packaging is verified; an independent clean-install/offline-founder resilience drill is still required to establish operational independence.
