<p align="center">
  <img src="assets/brand/auronq-banner.svg" alt="AuronQ — Blockchain" width="100%">
</p>

# AuronQ (AURQ)

**Open-source Proof-of-Work cryptocurrency · Public mainnet · MIT License**

[Project website](https://promirmir.github.io/AuronQ/) · [Downloads](https://github.com/promirmir/AuronQ/releases) · [Whitepaper](WHITEPAPER.md) · [Protocol](PROTOCOL.md) · [Report an issue](https://github.com/promirmir/AuronQ/issues)

AuronQ is a public UTXO cryptocurrency written in Go. The network combines **AQM64 Proof-of-Work**, **ML-DSA-87 transaction signatures**, independent full-node validation and local wallets. Its mainnet is running; the software and wider network are still at an early stage of maturity.

> **Security notice:** AQM64 and the complete AuronQ consensus, wallet and networking implementations have not received an independent professional security or cryptographic audit. Do not treat the network as mature financial infrastructure or store substantial value without understanding the risks.

## Download

All packages are published through GitHub Releases. Use the release page to verify checksums and read compatibility notes.

| Software | Version | Download |
| --- | --- | --- |
| **Desktop / Full Node** | v1.7.13 · Mainnet | [Windows x64](https://github.com/promirmir/AuronQ/releases/download/v1.7.13/AuronQ-1.7.13-Windows-x64.zip) · [Linux amd64](https://github.com/promirmir/AuronQ/releases/download/v1.7.13/AuronQ-1.7.13-Linux-amd64.tar.gz) · [Release notes](https://github.com/promirmir/AuronQ/releases/tag/v1.7.13) |
| **Universal Miner** | v0.4.8 Alpha · Windows x64 | [Windows GUI / GPU](https://github.com/promirmir/AuronQ/releases/download/miner-v0.4.8-alpha/AuronQ-Miner-v0.4.8-alpha-Windows-x64-GUI-GPU.zip) · [Release notes and SHA-256](https://github.com/promirmir/AuronQ/releases/tag/miner-v0.4.8-alpha) |
| **Universal Miner, older platform builds** | v0.4.6 Alpha | [Linux GPU and portable CPU packages](https://github.com/promirmir/AuronQ/releases/tag/miner-v0.4.6-alpha) |
| **AuronQ Mobile** | v0.5.3 Alpha · Android | [Android APK](https://github.com/promirmir/AuronQ/releases/download/android-v0.5.3-alpha/AuronQ-Mobile-0.5.3-alpha.apk) · [Release notes](https://github.com/promirmir/AuronQ/releases/tag/android-v0.5.3-alpha) |

The **Windows miner v0.4.8 Alpha** includes supported CUDA/OpenCL GPU backends, CPU fallback, local adaptive tuning and safety controls. Hardware support varies; alpha status means long-duration reliability is not yet established. The **Android wallet is a light client, not a validating full node**. Desktop Windows binaries are not currently Authenticode-signed.

## Get started

1. **Run a node:** download the full-node package, extract the entire archive and follow the [Windows guide](README-WINDOWS.md) or the release instructions for your platform.
2. **Synchronize:** allow the node to discover peers and validate the blockchain locally.
3. **Create a wallet:** back up recovery material securely before receiving or sending coins.
4. **Mine, if you choose:** follow the [Universal Miner guide](UNIVERSAL-MINER-GUIDE.md) and [hardware support notes](HARDWARE-SUPPORT.md). Solo mining uses a local full node; pool mining depends on a compatible independent service.

Full nodes can serve their own read-only Explorer at the **/explorer** endpoint. An independently hosted [public Explorer](https://mir.taild63f46.ts.net/explorer) may also be available; it is not a source of consensus authority.

## Protocol overview

| Property | AuronQ |
| --- | --- |
| Ledger model | UTXO |
| Consensus | Proof of Work with cumulative-work chain selection |
| Mining construction | AQM64 |
| Transaction signatures | ML-DSA-87 |
| Target block interval | 600 seconds |
| Nominal supply cap | 21,000,000 AURQ |
| Full-node implementation | Go |

For the **canonical Network ID, genesis hash, monetary rules and validation details**, use [MAINNET.md](MAINNET.md) and [PROTOCOL.md](PROTOCOL.md). A summary here must never replace the specifications used by nodes.

Every full node is intended to validate chain history, transactions and signatures independently. Bootstrap peers are discovery hints, not authorities that determine valid blocks or chain selection. See [NETWORK-INDEPENDENCE.md](NETWORK-INDEPENDENCE.md), [DECENTRALIZATION.md](DECENTRALIZATION.md) and the [mainnet change policy](MAINNET-CHANGE-POLICY.md).

**Stability comes first:** updates should preserve mainnet compatibility. Genesis, Network ID, AQM64, monetary rules and consensus validation must not be changed casually. Current stabilization criteria are recorded in [STABILIZATION.md](STABILIZATION.md).

## Mining pools and third-party integrations

AURQ can be mined solo or through third-party services. The following services operate independently of this repository:

- [MeshPool — AuronQ](https://meshpool.net/pool/auronq-main)
- [RPlant — AuronQ connection details](https://pool.rplant.xyz/#auronq#connect)
- [MeshMiner 0.8.35 — third-party miner](https://github.com/totom9000/meshminer/releases/tag/v.0.8.35)

Review fees, payout policies, hardware compatibility, downloads and connection parameters directly with the operator. The official Universal Miner does **not** silently install third-party mining binaries.

Independent pools, wallets, exchanges and explorers may integrate the public network without prior project approval. See [pool integration](POOL-INTEGRATION.md), [exchange integration](EXCHANGE-INTEGRATION.md) and [machine-readable project metadata](auronq-project.json). An integration does not imply endorsement by AuronQ.

## Independent network tracking

External services have started collecting AuronQ mining and network statistics. These independently operated data sources are useful for comparing observations outside the project's own Explorer:

- **[CPU-Mining.info — AURQ network statistics](https://cpu-mining.info/):** lists AuronQ (AURQ), the AQM64 algorithm, estimated network hashrate and reported block height.
- **[MiningChamp — RPlant pool statistics](https://miningchamp.com/pool/rplant):** its third-party pool aggregator includes AURQ/AQM64 mining activity, pool hashrate and worker counts.

These are **independent observations, not official consensus data, an independent security audit or an endorsement**. Reported figures may be delayed, incomplete or calculated differently. Compare them with data from a locally validating full node. Thank you to the people and services monitoring the network independently.

## Documentation

| Topic | Reference |
| --- | --- |
| Project and economics | [Whitepaper](WHITEPAPER.md) · [Mainnet specification](MAINNET.md) |
| Consensus and mining | [Protocol](PROTOCOL.md) · [AQM64](AQM64.md) |
| Software setup | [Windows](README-WINDOWS.md) · [Miner](UNIVERSAL-MINER-GUIDE.md) · [GPU guide](GPU-MINER-GUIDE.md) |
| Network design | [Decentralization](DECENTRALIZATION.md) · [Network independence](NETWORK-INDEPENDENCE.md) |
| Engineering evidence | [Test matrix](TEST-MATRIX.md) · [Threat model](THREAT-MODEL.md) · [Changelog](CHANGELOG.md) |
| Project direction | [Roadmap](ROADMAP.md) · [Stabilization](STABILIZATION.md) |
| Community and security | [Contributing](CONTRIBUTING.md) · [Security policy](SECURITY.md) · [FAQ](FAQ.md) |

The standardized ML-DSA-87 signature scheme does not, by itself, establish that the complete cryptocurrency is post-quantum secure. Independent cryptographic and implementation reviews remain a priority.

## Participate

Read the code, run an independent node, report reproducible defects and test interoperability. You can use the [issue tracker](https://github.com/promirmir/AuronQ/issues), [pull requests](https://github.com/promirmir/AuronQ/pulls) or [Bitcointalk project discussion](https://bitcointalk.org/index.php?topic=5595868.0).

Released under the [MIT License](LICENSE).
