<p align="center">
  <img src="assets/brand/auronq-banner.svg" alt="AuronQ — Blockchain" width="100%">
</p>

# AuronQ (AURQ)

**Open-source Proof-of-Work cryptocurrency · Public Mainnet · MIT License**

[Project website](https://promirmir.github.io/AuronQ/) · [Downloads](https://github.com/promirmir/AuronQ/releases) · [Whitepaper](WHITEPAPER.md) · [Protocol](PROTOCOL.md) · [Report a problem](https://github.com/promirmir/AuronQ/issues)

AuronQ is a public UTXO Layer-1 network written in Go. Full nodes independently validate transactions and blocks using AQM64 Proof of Work and ML-DSA-87 transaction signatures. **Mainnet is live, but the project remains experimental and in stabilization.**

> **Security notice:** The complete consensus, wallet, peer-to-peer and cryptographic implementations have not received an independent professional security audit. Do not store substantial value or treat this software as mature financial infrastructure.

## Downloads

These are **published builds**, not claims of defect-free or production-ready operation. Each download points to a specific GitHub Release.

| Software | Published release | Download and details |
| --- | --- | --- |
| **Desktop / Full Node** | v1.7.13 · Mainnet | [Windows x64](https://github.com/promirmir/AuronQ/releases/download/v1.7.13/AuronQ-1.7.13-Windows-x64.zip) · [Linux amd64](https://github.com/promirmir/AuronQ/releases/download/v1.7.13/AuronQ-1.7.13-Linux-amd64.tar.gz) · [Release notes](https://github.com/promirmir/AuronQ/releases/tag/v1.7.13) |
| **AuronQ Mobile** | v0.5.4 Alpha · Android | [Android APK](https://github.com/promirmir/AuronQ/releases/download/android-v0.5.4-alpha/AuronQ-Mobile-0.5.4-alpha.apk) · [Safe separate test APK](https://github.com/promirmir/AuronQ/releases/download/android-v0.5.4-alpha/AuronQ-Mobile-0.5.4-alpha-Isolated-Test.apk) · [Release notes / checksums](https://github.com/promirmir/AuronQ/releases/tag/android-v0.5.4-alpha) · [Android guide](android/README.md) |
| **Universal Miner** | v0.4.8 Alpha · Windows x64 | [Windows GUI / GPU](https://github.com/promirmir/AuronQ/releases/download/miner-v0.4.8-alpha/AuronQ-Miner-v0.4.8-alpha-Windows-x64-GUI-GPU.zip) · [Release notes and checksum](https://github.com/promirmir/AuronQ/releases/tag/miner-v0.4.8-alpha) |
| **Universal Miner, other platforms** | v0.4.6 Alpha · older builds | [Linux and portable CPU packages](https://github.com/promirmir/AuronQ/releases/tag/miner-v0.4.6-alpha) |

**Android upgrade caution:** v0.5.4 Alpha adds optional Ed25519-authenticated automatic checkpoint updates. Its automatic remote checkpoint publisher must first be activated with a private GitHub Actions signing secret; until then, the phone continues safely from its bundled checkpoint 1284. The APK is CI debug-signed and may **not** install as an in-place update of older builds. **Do not uninstall a working wallet or clear its private data to force an upgrade.** The isolated test APK uses a different application ID and is safer to evaluate first. See [Android setup and signing status](android/README.md). Previous v0.5.3 Alpha remains available as the known earlier release.

**Legacy Android APK identity (published v0.5.3 Alpha):** `AuronQ-Mobile-0.5.3-alpha.apk` · **9,574,566 bytes** · SHA-256:
`913e723a5b918ecf76bc923b599f5196b2a3340ec554897aa084b7193f1b520c`

This fingerprint identifies the **published GitHub APK**, not necessarily a different build already installed on someone's phone. A matching version name is not sufficient evidence of binary identity. **Do not uninstall an existing working wallet or clear its data to investigate a download difference. Back up recovery material first.**

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

Read the source, operate a full node, test interoperability and report reproducible defects through [GitHub Issues](https://github.com/promirmir/AuronQ/issues), [Pull Requests](https://github.com/promirmir/AuronQ/pulls) or [Bitcointalk](https://bitcointalk.org/index.php?topic=5595868.0).

Distributed under the [MIT License](LICENSE). ML-DSA-87 is standardized, but that alone does not establish end-to-end post-quantum security.
