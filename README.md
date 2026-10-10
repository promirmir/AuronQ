<p align="center">
  <img src="assets/brand/auronq-banner.svg" alt="AuronQ — Blockchain" width="100%">
</p>

# AuronQ (AURQ)

**Open-source Proof-of-Work cryptocurrency · Public Mainnet · MIT License**

[Project website](https://promirmir.github.io/AuronQ/) · [Discord community](https://discord.gg/rmmNY9RhA) · [Downloads](https://github.com/promirmir/AuronQ/releases) · [Roadmap](ROADMAP.md) · [Whitepaper](WHITEPAPER.md) · [Protocol](PROTOCOL.md) · [Report a problem](https://github.com/promirmir/AuronQ/issues)

**Android wallet — fast synchronization:** **[AuronQ Mobile 0.5.6 Alpha — Fast Sync (APK)](https://github.com/promirmir/AuronQ/releases/download/android-v0.5.6-alpha-fast-sync/AuronQ-Mobile-0.5.6-Alpha-Fast-Sync.apk)** · [Release notes and checksum](https://github.com/promirmir/AuronQ/releases/tag/android-v0.5.6-alpha-fast-sync) · [Installation guide](android/README.md). This Android alpha loads account information faster and preserves transaction history. It is experimental and has not undergone an independent security audit. Do not uninstall an existing funded wallet to upgrade.

**Project state — maintenance/freeze:** the Mainnet consensus and existing desktop v1.7.13 are held unchanged; no new features or node rollouts are scheduled. Security/availability fixes still require review and testing. [Release/freeze decision](PROJECT-FREEZE.md).

AuronQ is a public UTXO Layer-1 network written in Go. Full nodes independently validate transactions and blocks using AQM64 Proof of Work and ML-DSA-87 transaction signatures. **Mainnet is live, but the project remains experimental and in stabilization.**

> **Security notice:** The complete consensus, wallet, peer-to-peer and cryptographic implementations have not received an independent professional security audit. Do not store substantial value or treat this software as mature financial infrastructure.

## Downloads

These are **published builds**, not claims of defect-free or production-ready operation. Each download points to a specific GitHub Release.

| Software | Published release | Download and details |
| --- | --- | --- |
| **Desktop / Full Node** | v1.7.13 · Mainnet | [Windows x64](https://github.com/promirmir/AuronQ/releases/download/v1.7.13/AuronQ-1.7.13-Windows-x64.zip) · [Linux amd64](https://github.com/promirmir/AuronQ/releases/download/v1.7.13/AuronQ-1.7.13-Linux-amd64.tar.gz) · [Release notes](https://github.com/promirmir/AuronQ/releases/tag/v1.7.13) |
| **AuronQ Mobile** | v0.5.6 Alpha · Fast Sync · Android arm64 | **[Download APK](https://github.com/promirmir/AuronQ/releases/download/android-v0.5.6-alpha-fast-sync/AuronQ-Mobile-0.5.6-Alpha-Fast-Sync.apk)** · [Release](https://github.com/promirmir/AuronQ/releases/tag/android-v0.5.6-alpha-fast-sync) |
| **Universal Miner** | v0.4.8 Alpha · Windows x64 | [Windows GUI / GPU](https://github.com/promirmir/AuronQ/releases/download/miner-v0.4.8-alpha/AuronQ-Miner-v0.4.8-alpha-Windows-x64-GUI-GPU.zip) · [Release notes and checksum](https://github.com/promirmir/AuronQ/releases/tag/miner-v0.4.8-alpha) |
| **Universal Miner — Linux x64** | v0.4.8 Alpha · CPU/OpenCL | **[Download Linux tar.gz](https://github.com/promirmir/AuronQ/releases/download/miner-v0.4.8-alpha/AuronQ-Miner-v0.4.8-alpha-Linux-amd64.tar.gz)** · [Release and checksums](https://github.com/promirmir/AuronQ/releases/tag/miner-v0.4.8-alpha) |

**The Linux 0.4.8 package** is published in GitHub Releases; GPU performance and CUDA compatibility on physical Linux devices are not yet independently validated.

**Mobile verification and compatibility:** the currently published Android 0.5.6 Alpha is a light wallet, not a full validating node. Its fast synchronization and peer-assisted account information must not be presented as proof of full historical consensus validation. Experimental checkpoint implementations and their technical history are described in [checkpoint design notes](docs/DECENTRALIZED-CHECKPOINTS.md), not as a required or guaranteed feature of current releases. **Do not uninstall an existing funded wallet or clear application data to force an update.** Android signing certificates may differ between builds; first back up and verify wallet recovery material.

The Android wallet is a **light client, not a full validating node**. Initial verification can take time; instant payments or synchronization cannot be guaranteed. Miner GPU/CPU support and operating temperatures vary by machine. Alpha builds remain unverified for long-term reliability.

## Getting started

1. **Desktop node:** download, extract the complete archive, and follow [Start here](START-HERE.md) or the [Windows guide](README-WINDOWS.md). Full nodes validate the chain themselves.
2. **Android wallet:** read the [Android guide](android/README.md), compare the APK checksum and securely back up wallet recovery material before changes or transactions.
3. **Mining (optional):** follow the [Universal Miner guide](UNIVERSAL-MINER-GUIDE.md) and [hardware guidance](HARDWARE-SUPPORT.md). Solo mining uses a local full node; pool mining uses a third-party service.

## Network and protocol

| Property | AuronQ Mainnet |
| --- | --- |
| Ledger / consensus | UTXO · Proof of Work · cumulative-work chain selection |
| Mining algorithm | AQM64 |
| Transaction signatures | ML-DSA-87 |
| Target block interval | 600 seconds |
| Nominal maximum supply | 21,000,000 AURQ (including genesis allocation) |
| Default P2P port | TCP 18444 |

**Canonical network identity and consensus rules:** [Mainnet](MAINNET.md) · [Protocol](PROTOCOL.md) · [AQM64](AQM64.md). Live peer counts, block height and estimated hashrate are deliberately not hardcoded here: they change and must be obtained from a validating node.

Bootstrap servers are discovery hints, not consensus authorities. Network operation without the original project machines is a goal **still requiring an independent offline-bootstrap drill**, not a proven guarantee. See [Network independence](NETWORK-INDEPENDENCE.md), [Decentralization](DECENTRALIZATION.md) and [Stabilization](STABILIZATION.md).

**Stability before features:** consensus-critical parameters (genesis, Network ID, AQM64, difficulty, monetary rules and transaction/block validation) must not be changed in an ordinary maintenance release. See the [Mainnet change policy](MAINNET-CHANGE-POLICY.md).

## Independent ecosystem

These links refer to **third-party services**; they are not operated by AuronQ and may be unavailable or inaccurate. Their presence does not imply endorsement, verified compatibility, or an independent security audit.

| Category | Resources |
| --- | --- |
| Mining pools | [MeshPool — AuronQ](https://meshpool.net/pool/auronq-main) · [RPlant — AuronQ](https://pool.rplant.xyz/#auronq#connect) |
| Third-party miner | [MeshMiner 0.8.35](https://github.com/totom9000/meshminer/releases/tag/v.0.8.35) |
| Independent tracking | [MiningChamp — AuronQ](https://miningchamp.com/coin/auronq) · [CPU-Mining.info — AURQ](https://cpu-mining.info/coins/AURQ) |
| Public community Explorer | [Independent Explorer](https://mir.taild63f46.ts.net/explorer) (availability not guaranteed) |

**Use AURQ-specific pages, not data for similarly named coins.** External statistics may be stale, incomplete or mistaken (including mining rewards and network hashrate). The source of consensus truth is a locally validating full node, not a pool, aggregator or Explorer.

Permissionless independent integrations are welcome: [Pool integration](POOL-INTEGRATION.md) · [Exchange integration](EXCHANGE-INTEGRATION.md) · [Machine-readable metadata](auronq-project.json).

## Documentation

| Subject | References |
| --- | --- |
| Overview | [Whitepaper](WHITEPAPER.md) · [FAQ](FAQ.md) · [Roadmap](ROADMAP.md) |
| Installation | [Start here](START-HERE.md) · [Windows](README-WINDOWS.md) · [Android](android/README.md) · [Miner](UNIVERSAL-MINER-GUIDE.md) · [GPU guide](GPU-MINER-GUIDE.md) |
| Network | [Mainnet](MAINNET.md) · [Protocol](PROTOCOL.md) · [Network independence](NETWORK-INDEPENDENCE.md) |
| Safety and maintenance | [Security](SECURITY.md) · [Threat model](THREAT-MODEL.md) · [Test matrix](TEST-MATRIX.md) · [Stabilization](STABILIZATION.md) |

## Participate

AuronQ is being built as an independent cryptocurrency with a public Mainnet, not as a fundraising campaign or a promise of financial returns. Anyone can review the [roadmap](ROADMAP.md), operate a node or participate in mining voluntarily. Exchange listings, future market value and commercial success cannot be guaranteed.

Read the source, operate a full node, test interoperability and report reproducible defects through [GitHub Issues](https://github.com/promirmir/AuronQ/issues), [Pull Requests](https://github.com/promirmir/AuronQ/pulls), [Bitcointalk](https://bitcointalk.org/index.php?topic=5595868.0) or the [AuronQ Discord community](https://discord.gg/rmmNY9RhA).

Distributed under the [MIT License](LICENSE). ML-DSA-87 is standardized, but that alone does not establish end-to-end post-quantum security.
