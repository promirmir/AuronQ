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
| Decentralization model | [DECENTRALIZATION.md](DECENTRALIZATION.md) |
| Network independence | [NETWORK-INDEPENDENCE.md](NETWORK-INDEPENDENCE.md) |
| Public network | [PUBLIC-NETWORK.md](PUBLIC-NETWORK.md) |
| Threat model | [THREAT-MODEL.md](THREAT-MODEL.md) |
| Security policy | [SECURITY.md](SECURITY.md) |
| FAQ | [FAQ.md](FAQ.md) |
| Contributing | [CONTRIBUTING.md](CONTRIBUTING.md) |

## Quick Start — Windows

For a normal Windows user, no manual peer configuration, Tailscale, port forwarding, Go, Docker or command line setup is required.

1. Download **[AuronQ-1.7.10-Windows-x64.zip](https://github.com/promirmir/AuronQ/releases/download/v1.7.10/AuronQ-1.7.10-Windows-x64.zip)** from this repository's official Releases page.
2. Verify the ZIP SHA-256 if possible:
   `e573c3a605fc66c33f6681d51a8c1544b4200fd207463ec8893971ebd20a134c`
3. Extract the **entire ZIP** to a new folder. Do not run files directly from inside the archive.
4. Run **`START-AURONQ.cmd` once**.
5. AuronQ Desktop will start the full node, verify the Mainnet configuration, discover the public bootstrap nodes and begin synchronization automatically.
6. In **Wallets**, create a wallet or import an existing one. Back up the wallet before using it for anything important.
7. After the node is synchronized, you can receive/send AURQ and optionally use the built-in CPU miner.

A normal user should use **`START-AURONQ.cmd`**.  
**`START-SEED-NODE.cmd` is only for operators who intentionally want to expose a public bootstrap/full node through Tailscale Funnel.**

AuronQ Desktop stores its persistent data under:

`%AppData%\AuronQ`

This includes wallets, chain data and learned peers. Updating the application does not require deleting that directory.

Windows binaries are currently not Authenticode-signed. Windows SmartScreen may show a warning on first launch; verify that the archive came from this repository and that its SHA-256 matches the value published above before running it.

For a longer Windows guide, see **[README-WINDOWS.md](README-WINDOWS.md)**.

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

Current stable desktop/full-node release: **AuronQ 1.7.10 Mainnet**.

Version 1.7.10 continues the decentralization hardening: configured seeds are replaceable startup hints rather than permanent authorities, repeatedly failing seeds are pruned, learned public peers persist independently, and the built-in Explorer is explicitly a local view of each full node's own validated canonical chain. CI also contains a distributed regression test in which the original/bootstrap node disappears permanently, surviving nodes continue the chain, and a fresh node joins through a later non-founder peer.

The network is operational, but practical decentralization is still developing. The bootstrap manifest currently contains **two externally reachable public bootstrap endpoints** verified by the repository crawler. They have no consensus privileges. Additional independently operated public nodes, miners and independent discovery routes are still required before the live network can be considered operationally independent of the original operator infrastructure.

## Download

**Windows x64:** [AuronQ-1.7.10-Windows-x64.zip](https://github.com/promirmir/AuronQ/releases/download/v1.7.10/AuronQ-1.7.10-Windows-x64.zip)

**Linux amd64:** [AuronQ-1.7.10-Linux-amd64.tar.gz](https://github.com/promirmir/AuronQ/releases/download/v1.7.10/AuronQ-1.7.10-Linux-amd64.tar.gz)

Release page: [AuronQ 1.7.10 Mainnet](https://github.com/promirmir/AuronQ/releases/tag/v1.7.10)

SHA-256:

```text
e573c3a605fc66c33f6681d51a8c1544b4200fd207463ec8893971ebd20a134c  AuronQ-1.7.10-Windows-x64.zip
af926ef826b721ff635a1740e2aaaa25515fe3d545fb53e92e70a911df24a388  AuronQ-1.7.10-Linux-amd64.tar.gz
```

Windows users: extract the whole ZIP and run `START-AURONQ.cmd`.

The current Windows binaries are **not Authenticode-signed**, so Windows SmartScreen may warn on first launch. Verify the SHA-256 above and download only from this repository's Releases page.

## Public blockchain explorer

There is **no canonical central AuronQ Explorer**. Every AuronQ full node serves
the same read-only Explorer from its own locally validated canonical chain at
`/explorer`.

One currently reachable public instance is:

`https://mir.taild63f46.ts.net/explorer`

That URL is only one full-node instance and has no special authority. If it
disappears, other full nodes and their local Explorers continue to operate.

## Built-in blockchain explorer

AuronQ 1.7.5+ full nodes include a built-in, read-only blockchain explorer at:

`/explorer`

For a publicly reachable node, append `/explorer` to its HTTPS node address. The explorer provides:

- live Mainnet height, peer count, mempool size, issued supply and estimated AQM64 network power;
- the latest canonical blocks;
- block lookup by height or block hash;
- transaction lookup by TXID, including inputs, outputs, fees and confirmations;
- address lookup with spendable/total balance and recent transaction history.

Explorer endpoints are rate-limited. Explorer v1 reads canonical chain and mempool data directly from **that node**; it has no consensus privileges and does not modify blockchain state. Desktop users therefore inspect their own independently validated chain rather than trusting a project-operated explorer service.

## Android alpha

For Android 8.0+ on ARM64, the current test wallet is **AuronQ Mobile 0.4.2 Alpha**:

- [AuronQ-Mobile-0.4.2-alpha.apk](https://github.com/promirmir/AuronQ/releases/download/android-v0.4.2-alpha/AuronQ-Mobile-0.4.2-alpha.apk)
- Release page: [android-v0.4.2-alpha](https://github.com/promirmir/AuronQ/releases/tag/android-v0.4.2-alpha)
- SHA-256: `0224f45dca6b721577a60dc5dcd03d5734c0468c1abcac6f4c4e183b19c59736`

AuronQ Mobile remains a **light wallet client, not a full node**. Private keys and ML-DSA-87 signing stay local on the phone. Version 0.4.2 compares Mainnet state across multiple replaceable HTTPS AuronQ nodes, exposes peer agreement, compares wallet balances only across peers reporting the same chain state, fails closed on conflicting same-chain balance responses, remembers P2P-learned nodes and directly fans the same signed transaction out to multiple reachable nodes.

Multi-peer agreement reduces dependence on a single endpoint, but it is **not equivalent to independently validating the complete AQM64 chain**. The APK is debug-signed alpha software and has not received an independent security audit; do not use substantial value.

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

At present the manifest advertises **two externally verified public bootstrap endpoints**:

- `https://mir.taild63f46.ts.net`
- `https://desktop-4nifg1j.taild63f46.ts.net`

These endpoints are only first-contact hints to ordinary full nodes. They have no consensus privileges and, from the decentralization-v3 hardening onward, repeatedly failing configured seeds are pruned from the running peer set instead of remaining permanent authorities. Full nodes independently validate blocks, transactions, Network ID and cumulative chain work.

AuronQ full nodes already persist verified public peers and exchange them through peer gossip. In addition, the repository now has a scheduled public-peer crawler that follows AuronQ gossip, verifies reachable Mainnet peers and can append them to `bootstrap.json` automatically. This means that, as independently operated public nodes appear, the first-contact registry can become multi-peer without relying on the founder's computer being online.

This does **not** create independent peers out of nothing: the live network currently has two crawler-verified public bootstrap endpoints under the project infrastructure. More independently operated public full nodes are still required for stronger practical resilience. A completely fresh install always needs some discovery route; AuronQ also supports DNS seeds for that purpose. The automated resilience suite now explicitly proves that, once later reachable peers exist, the original bootstrap node can disappear permanently while surviving nodes continue the chain and a fresh node joins through a non-founder peer.

NAT/CGNAT users can participate through outbound connections without port forwarding. Operators who intentionally want to expose a publicly reachable Windows node can use the opt-in `START-SEED-NODE.cmd` with Tailscale Funnel; normal users should continue to use `START-AURONQ.cmd`.

## Network behavior

The node exchanges peer metadata, transactions and blocks over its P2P HTTP transport. It validates downloaded blocks locally, selects chains by cumulative work, limits synchronization batches, rate-limits requests/block submission, persists known public peers, and repairs a same-branch peer that missed one or more block broadcasts by pushing the missing validated extension.

A real two-node test synchronized the same mainnet chain through height 5 and the same tip after mining and catch-up repair.

## Source and CI

The complete Go source is in this repository. CI runs tests and vetting on Linux and Windows, includes the race detector on Linux, and builds the Windows CLI/Desktop plus Linux CLI.

The current stable release is **v1.7.10**. Release artifacts and SHA-256 checksums are published on the GitHub Releases page.

## Security status

AuronQ 1.7.10 is live mainnet software, but **AQM64 and the overall consensus/network implementation have not received an independent professional security or cryptographic audit**. Passing internal/CI tests is not equivalent to an external audit. Do not present AuronQ as production-audited financial infrastructure until independent review has occurred.

See [SECURITY.md](SECURITY.md), [PROTOCOL.md](PROTOCOL.md), [AQM64.md](AQM64.md) and [MAINNET.md](MAINNET.md).
