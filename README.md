# AuronQ (AURQ) — Post-Quantum UTXO Proof-of-Work Cryptocurrency

[![CI](https://github.com/promirmir/AuronQ/actions/workflows/ci.yml/badge.svg)](https://github.com/promirmir/AuronQ/actions/workflows/ci.yml)
[![Latest release](https://img.shields.io/github/v/release/promirmir/AuronQ?display_name=tag)](https://github.com/promirmir/AuronQ/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go](https://img.shields.io/github/go-mod/go-version/promirmir/AuronQ)](go.mod)

**AuronQ (AURQ)** is an open-source cryptocurrency and blockchain project written in Go. It uses a **UTXO** ledger, **Proof-of-Work**, **ML-DSA-87 post-quantum transaction signatures**, and the AuronQ-specific **AQM64** proof-of-work construction.

The AuronQ **Mainnet is live**. Full-node software is available for Windows x64 and Linux amd64, and an Android wallet client is available as an alpha release.

> **Security note:** AuronQ uses a standardized post-quantum signature scheme for transactions, but neither AQM64 nor the complete consensus/network implementation has received an independent professional security or cryptographic audit. “Post-quantum signatures” should not be interpreted as a guarantee that the entire system is immune to all present or future attacks.

## Quick links

| Resource | Link |
|---|---|
| Latest releases | [GitHub Releases](https://github.com/promirmir/AuronQ/releases) |
| Mainnet specification | [MAINNET.md](MAINNET.md) |
| Protocol | [PROTOCOL.md](PROTOCOL.md) |
| AQM64 Proof-of-Work | [AQM64.md](AQM64.md) |
| Network independence | [NETWORK-INDEPENDENCE.md](NETWORK-INDEPENDENCE.md) |
| Public network | [PUBLIC-NETWORK.md](PUBLIC-NETWORK.md) |
| Threat model | [THREAT-MODEL.md](THREAT-MODEL.md) |
| Security policy | [SECURITY.md](SECURITY.md) |
| FAQ | [FAQ.md](FAQ.md) |
| Contributing | [CONTRIBUTING.md](CONTRIBUTING.md) |

## What makes AuronQ technically distinct?

- **Post-quantum transaction signatures:** ML-DSA-87 is used for transaction signing.
- **UTXO accounting model:** transactions consume and create unspent transaction outputs.
- **Proof-of-Work consensus:** mining uses the AuronQ-specific AQM64 construction.
- **Independent full-node validation:** each full node validates blocks, transactions, Network ID and chain work locally.
- **Peer-to-peer networking:** nodes discover peers through persisted peers, configured seeds, DNS seeds, an HTTPS bootstrap manifest and peer gossip.
- **Open-source implementation:** the node, wallet-related code, protocol documentation and build/CI configuration are public in this repository.
- **Go implementation:** CI tests Linux and Windows builds, vetting, and a race-detector run on Linux.

## Who is this repository for?

AuronQ may be relevant to developers, miners, node operators and researchers looking for an experimental **post-quantum cryptocurrency**, **ML-DSA blockchain implementation**, **UTXO Proof-of-Work network**, **Go cryptocurrency node**, or a public example of integrating post-quantum signatures into a cryptocurrency transaction system.

## Project status

Current stable desktop/full-node release: **AuronQ 1.7.1 Mainnet**.

The network is operational, but decentralization is still developing. At the time documented in this repository, the bootstrap manifest contains one confirmed public bootstrap endpoint. That endpoint has no consensus privileges; independent nodes validate the canonical chain themselves. Additional independently operated public nodes and independent discovery routes are needed to reduce first-contact dependency.

## Download

**Windows x64:** [AuronQ-1.7.1-Windows-x64.zip](https://github.com/promirmir/AuronQ/releases/download/v1.7.1/AuronQ-1.7.1-Windows-x64.zip)

**Linux amd64:** [AuronQ-1.7.1-Linux-amd64.tar.gz](https://github.com/promirmir/AuronQ/releases/download/v1.7.1/AuronQ-1.7.1-Linux-amd64.tar.gz)

Release page: [AuronQ 1.7.1 Mainnet](https://github.com/promirmir/AuronQ/releases/tag/v1.7.1)

SHA-256:

```text
50a1b2d321117549d7ceab81aa5dfaa0bb072544915fb5e921eb9b2a24d68d8d  AuronQ-1.7.1-Windows-x64.zip
6ddd97167f5e587cf7245ba9827d5d8c2219a96581a6ecb21a4a6d76a90270a0  AuronQ-1.7.1-Linux-amd64.tar.gz
```

Windows users: extract the whole ZIP and run `START-AURONQ.cmd`.

The current Windows binaries are **not Authenticode-signed**, so Windows SmartScreen may warn on first launch. Verify the SHA-256 above and download only from this repository's Releases page.

## Android alpha

For Android 8.0+ on ARM64, the current test wallet is **AuronQ Mobile 0.3.0 Alpha**:

- [AuronQ-Mobile-0.3.0-alpha.apk](https://github.com/promirmir/AuronQ/releases/download/android-v0.3.0-alpha/AuronQ-Mobile-0.3.0-alpha.apk)
- Release page: [android-v0.3.0-alpha](https://github.com/promirmir/AuronQ/releases/tag/android-v0.3.0-alpha)
- SHA-256: `9f92008d5174f831bced33bb11b91c08f11142e0d49610a99a29cfc159e1b9ba`

The Android alpha is a wallet client, not a full node. It keeps/signs with the encrypted AuronQ wallet locally and uses an HTTPS AuronQ Mainnet node for balances, UTXOs and transaction submission. Version 0.3.0 adds PL/EN switching, live AuronQ Mainnet telemetry, recent blocks, remembered public-node discovery, failover and a mobile interface styled after AuronQ Desktop. It is debug-signed test software; do not use substantial value.

## Mainnet identity

Network ID:

`44e62c2ace002a6660c14e252173c1aa303529c68e40c998e92da2b453f44f30b1e58c94d533587e2186004593fb856c433fcdb5418ed430ec8617e29529365c`

Genesis hash:

`5750a455c04bfe93c9edfef1a12744b05e29ac6da1a9dd5b790566629dea2080c581beb2f0324efba2067c9efb113ed7a29598fd3ffbed965ced31f265d0cec4`

Genesis founder allocation: 210,000 AURQ (1% of the nominal 21,000,000 AURQ cap). Initial subsidy: 49.5 AURQ. Coinbase maturity: 100 blocks. Target spacing: 600 seconds.

## Public P2P discovery

AuronQ does not use a central blockchain server. Every full node keeps and validates its own canonical chain.

A fresh node can discover peers from:

- previously learned peers stored locally;
- configured fixed seeds;
- DNS seeds;
- the bounded HTTPS bootstrap manifest;
- peer gossip after the first connection.

The official stable manifest is:

`https://raw.githubusercontent.com/promirmir/AuronQ/main/bootstrap.json`

At present the manifest advertises **one confirmed public bootstrap endpoint**:

`https://mir.taild63f46.ts.net`

That endpoint is only a first-contact relay to an ordinary full node. It has no consensus privileges. Full nodes independently validate blocks, transactions, Network ID and cumulative chain work.

AuronQ full nodes already persist verified public peers and exchange them through peer gossip. In addition, the repository now has a scheduled public-peer crawler that follows AuronQ gossip, verifies reachable Mainnet peers and can append them to `bootstrap.json` automatically. This means that, as independently operated public nodes appear, the first-contact registry can become multi-peer without relying on the founder's computer being online.

This does **not** create independent peers out of nothing: today there is still only one confirmed public endpoint. Practical independence from the original computer requires at least a few other publicly reachable full nodes. A completely fresh install always needs some discovery route; AuronQ also supports DNS seeds for that purpose.

NAT/CGNAT users can participate through outbound connections without port forwarding. Users who want to contribute a publicly reachable Windows full node can use the opt-in `START-PUBLIC-NODE.cmd` in future release packages (or run the node directly with public TCP/18444 reachability).

## Network behavior

The node exchanges peer metadata, transactions and blocks over its P2P HTTP transport. It validates downloaded blocks locally, selects chains by cumulative work, limits synchronization batches, rate-limits requests/block submission, persists known public peers, and repairs a same-branch peer that missed one or more block broadcasts by pushing the missing validated extension.

A real two-node test synchronized the same mainnet chain through height 5 and the same tip after mining and catch-up repair.

## Source and CI

The complete Go source is in this repository. CI runs tests and vetting on Linux and Windows, includes the race detector on Linux, and builds the Windows CLI/Desktop plus Linux CLI.

The immutable `v1.7.1` release tag points to commit:

`776401065ac454ae4ad2246b9efdf7d5f134efa0`

## Security status

AuronQ 1.7.1 is live mainnet software, but **AQM64 and the overall consensus/network implementation have not received an independent professional security or cryptographic audit**. Passing internal/CI tests is not equivalent to an external audit. Do not present AuronQ as production-audited financial infrastructure until independent review has occurred.

See [SECURITY.md](SECURITY.md), [PROTOCOL.md](PROTOCOL.md), [AQM64.md](AQM64.md) and [MAINNET.md](MAINNET.md).
