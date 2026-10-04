# AuronQ (AURQ)

[![CI](https://github.com/promirmir/AuronQ/actions/workflows/ci.yml/badge.svg)](https://github.com/promirmir/AuronQ/actions/workflows/ci.yml)
[![Network hardening](https://github.com/promirmir/AuronQ/actions/workflows/network-hardening.yml/badge.svg)](https://github.com/promirmir/AuronQ/actions/workflows/network-hardening.yml)
[![Reproducible builds](https://github.com/promirmir/AuronQ/actions/workflows/reproducible-builds.yml/badge.svg)](https://github.com/promirmir/AuronQ/actions/workflows/reproducible-builds.yml)
[![Latest release](https://img.shields.io/github/v/release/promirmir/AuronQ?display_name=tag)](https://github.com/promirmir/AuronQ/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go](https://img.shields.io/github/go-mod/go-version/promirmir/AuronQ)](go.mod)

**AuronQ** is an open-source experimental cryptocurrency network written in Go. It combines a **UTXO ledger**, **Proof-of-Work**, **ML-DSA-87 post-quantum transaction signatures**, and the AuronQ-specific **AQM64** proof-of-work construction.

The AuronQ Mainnet is live. The project is designed around **independent full-node validation, replaceable peer discovery, local wallets, local explorers and no privileged founder node**.

> [!WARNING]
> AuronQ is experimental financial software. AQM64 and the complete consensus/network implementation have **not** received an independent professional security or cryptographic audit. Do not use substantial value.

## Current releases

| Component | Current release | Status | Download |
|---|---:|---|---|
| Desktop / Full Node | **v1.7.11** | Mainnet | [Windows x64](https://github.com/promirmir/AuronQ/releases/download/v1.7.11/AuronQ-1.7.11-Windows-x64.zip) · [Linux amd64](https://github.com/promirmir/AuronQ/releases/download/v1.7.11/AuronQ-1.7.11-Linux-amd64.tar.gz) |
| AuronQ Mobile | **0.5.1 Alpha** | Light wallet | [Android APK](https://github.com/promirmir/AuronQ/releases/download/android-v0.5.1-alpha/AuronQ-Mobile-0.5.1-alpha.apk) |

**Project site:** https://promirmir.github.io/AuronQ/

**Release archive:** [GitHub Releases](https://github.com/promirmir/AuronQ/releases) · **History:** [CHANGELOG.md](CHANGELOG.md)

### Current checksums

```text
cd1b0ff59fdaf28753758473336f2b9908c769374de5e1874dd0808ebd86f0ee  AuronQ-1.7.11-Windows-x64.zip
bcb87751fa6c559ed008f245a804b8341c664460068cfd5623ea5d3b13372a50  AuronQ-1.7.11-Linux-amd64.tar.gz
c224dd9f94d52132d71e7f781dfadcbbc3a7bc92c2c07bf32d38dc4929328c73  AuronQ-Mobile-0.5.1-alpha.apk
```

## What has been built

| Area | Status | What AuronQ currently provides |
|---|---|---|
| Mainnet | ✅ Live | Fixed Network ID and genesis, persistent chain storage |
| Full-node validation | ✅ | Local validation of blocks, transactions, signatures, timestamps, difficulty and cumulative work |
| Wallets | ✅ | Local ML-DSA-87 wallet creation/import, send/receive and wallet history |
| Proof-of-Work | ✅ | AQM64 CPU mining and cumulative-work chain selection |
| P2P | ✅ | Persisted peers, peer gossip, replaceable seed hints, bounded bootstrap manifests and DNS-seed support |
| Founder-node independence | ✅ Tested | Regression test removes the original bootstrap node permanently and joins a fresh node through a surviving peer |
| Explorer | ✅ | Every full node serves its own read-only explorer from its locally validated chain |
| Explorer network | ✅ | Public explorers can surface alternative HTTPS explorers learned through native P2P gossip |
| Reorg resilience | ✅ | Higher-work fork validation, immediate post-reorg sync continuation |
| Network hardening | ✅ | 20-node partition/fork/restart convergence tests, netgroup diversity, Sybil/eclipse concentration limits |
| Fuzzing / race testing | ✅ | Transaction/block parser fuzzing and Linux race detector in CI |
| Reproducible builds | ✅ | Release-style deterministic build checks on Linux and Windows |
| Public health monitoring | ✅ | Scheduled checks for Network ID, height/tip agreement and Explorer availability |
| Android light client | ✅ Alpha | Local header/AQM64 verification plus verified-chain multi-peer wallet-state quorum |
| Independent external audit | ❌ Not yet | Required before treating AuronQ as mature financial infrastructure |
| Broad independent node/miner set | ⚠️ Developing | More independently operated public nodes and miners are required |

## Architecture

AuronQ follows a **full-node-first** model.

Each AuronQ full node:

- stores and validates its own canonical blockchain;
- independently verifies AQM64 proof-of-work and cumulative work;
- validates UTXO spends and ML-DSA-87 signatures locally;
- maintains its own mempool;
- discovers and persists peers;
- relays valid transactions and blocks;
- serves its own local read-only blockchain Explorer.

Bootstrap peers are **rendezvous hints, not authorities**. They cannot approve invalid blocks, select the canonical chain, change monetary rules or override a node's validation. Repeatedly failing seed peers are pruned from the running peer set.

See [DECENTRALIZATION.md](DECENTRALIZATION.md) and [NETWORK-INDEPENDENCE.md](NETWORK-INDEPENDENCE.md).

## Quick start — Windows

1. Download the current [Windows x64 release](https://github.com/promirmir/AuronQ/releases/download/v1.7.11/AuronQ-1.7.11-Windows-x64.zip).
2. Verify its SHA-256 against the value above.
3. Extract the **entire ZIP** to a new folder.
4. Run **`START-AURONQ.cmd`**.
5. Wait for the local full node to load and synchronize.
6. Create or import a wallet and **back it up before using it**.
7. Use the built-in Explorer, send/receive AURQ, and enable CPU mining only if you intentionally want to mine.

Persistent Desktop data is stored under:

`%AppData%\AuronQ`

Windows binaries are not currently Authenticode-signed, so SmartScreen may warn on first launch.

Detailed guide: [README-WINDOWS.md](README-WINDOWS.md)

## Mainnet identity

| Parameter | Value |
|---|---|
| Symbol | **AURQ** |
| Ledger | UTXO |
| Transaction signatures | ML-DSA-87 |
| Proof-of-Work | AQM64 |
| Target block spacing | 600 seconds |
| Nominal supply cap | 21,000,000 AURQ |
| Genesis founder allocation | 210,000 AURQ |
| Coinbase maturity | 100 blocks |

**Network ID**

`44e62c2ace002a6660c14e252173c1aa303529c68e40c998e92da2b453f44f30b1e58c94d533587e2186004593fb856c433fcdb5418ed430ec8617e29529365c`

**Genesis hash**

`5750a455c04bfe93c9edfef1a12744b05e29ac6da1a9dd5b790566629dea2080c581beb2f0324efba2067c9efb113ed7a29598fd3ffbed965ced31f265d0cec4`

Canonical specification: [MAINNET.md](MAINNET.md)

## Explorer

There is **no canonical central AuronQ Explorer**.

Every full node exposes a read-only Explorer at:

`/explorer`

The Explorer reads that node's own independently validated canonical chain. A public Explorer can additionally show alternative HTTPS full-node Explorers learned through AuronQ peer gossip.

One currently reachable instance is:

`https://mir.taild63f46.ts.net/explorer`

Its availability does not determine consensus and it has no special authority.

## AuronQ Mobile

AuronQ Mobile 0.5.1 Alpha is a **light wallet, not a full node**.

It:

- keeps private keys and ML-DSA-87 signing local on the phone;
- independently validates the Mainnet header chain from embedded genesis;
- locally verifies AQM64 proof-of-work, difficulty, timestamps and hash continuity;
- accepts balance/history/UTXO state only from peers matching the verified header chain;
- compares canonical wallet state across agreeing peers and fails closed on conflicts;
- broadcasts locally signed transactions to agreeing verified-chain peers.

It does **not** reconstruct the entire UTXO set from every full block, so it must not be described as equivalent to a full node.

## Engineering milestones

- **Mainnet launch:** public Windows/Linux full-node releases with immutable Network ID and genesis.
- **Wallet safety:** guarded local wallet deletion and backup warnings.
- **Public networking:** HTTPS/DNS peer discovery, peer gossip, persistent peers and public endpoint advertisement.
- **Observability:** wallet history API and estimated network hash power.
- **Explorer:** built-in read-only Explorer, then Desktop integration and CSP hardening.
- **Resilience:** direct bootstrap fallbacks, faster reorg recovery, 20-node partition convergence and hourly public health checks.
- **P2P hardening:** peer netgroups, DNS parent-domain grouping and sync diversity.
- **Founder independence:** startup seeds made prunable/replaceable; regression tests remove the original bootstrap node.
- **Mobile trust reduction:** multi-peer agreement, then local header/AQM64 verification and verified-chain wallet-state quorum.
- **Build integrity:** deterministic/reproducible build checks on Linux and Windows.
- **Explorer decentralization:** peer-aware links between independently hosted full-node explorers.

Detailed history: [CHANGELOG.md](CHANGELOG.md)

## Roadmap

The next engineering priorities are tracked in [ROADMAP.md](ROADMAP.md). The highest-value items are:

1. more independently operated public full nodes and miners;
2. additional discovery routes, including independently operated DNS seeds;
3. continued adversarial P2P / eclipse / Sybil / DoS testing;
4. further reduction of light-client trust in remote full-node state;
5. independent professional review of AQM64, consensus and networking;
6. production-grade signing/distribution hardening for binaries and mobile releases.

## Documentation

| Topic | Document |
|---|---|
| Mainnet specification | [MAINNET.md](MAINNET.md) |
| Protocol | [PROTOCOL.md](PROTOCOL.md) |
| AQM64 Proof-of-Work | [AQM64.md](AQM64.md) |
| Decentralization model | [DECENTRALIZATION.md](DECENTRALIZATION.md) |
| Network independence | [NETWORK-INDEPENDENCE.md](NETWORK-INDEPENDENCE.md) |
| Public networking | [PUBLIC-NETWORK.md](PUBLIC-NETWORK.md) |
| Threat model | [THREAT-MODEL.md](THREAT-MODEL.md) |
| Security policy | [SECURITY.md](SECURITY.md) |
| Test matrix | [TEST-MATRIX.md](TEST-MATRIX.md) |
| FAQ | [FAQ.md](FAQ.md) |
| Windows guide | [README-WINDOWS.md](README-WINDOWS.md) |
| Contributing | [CONTRIBUTING.md](CONTRIBUTING.md) |
| Release history | [CHANGELOG.md](CHANGELOG.md) |
| Historical material | [docs/archive/README.md](docs/archive/README.md) |

## Security

AuronQ's use of ML-DSA-87 provides a standardized post-quantum signature scheme for transactions. That **does not** prove that the complete cryptocurrency is post-quantum secure or production secure.

AQM64, consensus, networking, wallet behavior and implementation details still require independent review.

Please read [SECURITY.md](SECURITY.md) before reporting or evaluating security issues.

## Contributing

Contributions, reproducible bug reports and independent review are welcome.

- [Contributing guide](CONTRIBUTING.md)
- [Security policy](SECURITY.md)
- [Issue tracker](https://github.com/promirmir/AuronQ/issues)
- [Pull requests](https://github.com/promirmir/AuronQ/pulls)

## License

AuronQ is released under the [MIT License](LICENSE).
