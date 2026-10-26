# AuronQ Mobile

## Download

**Current published Android build (26 October 2026): AuronQ Mobile 0.5.9 Alpha — Light Quorum / Fast Start.**

- [Download the official Android APK](https://github.com/promirmir/AuronQ/releases/download/android-v0.5.9-alpha-light-quorum/AuronQ-Mobile-0.5.9-Alpha-Light.apk)
- [Published GitHub Release and SHA-256 checksums](https://github.com/promirmir/AuronQ/releases/tag/android-v0.5.9-alpha-light-quorum)
- APK SHA-256: `4ce9f59f07fa7544e4edb66cac0c3158c80a12808b75f820dbfc20cf54b9915f`.
- GitHub release target (protected main): `b1e0d3ba1ddfd17670cb27e4ca8c5579a46610a8`.

**Fast startup:** opening the mobile app does not initiate a replay of thousands of AQM64 Argon2id hashes or fetch checkpoints. The UI connects to public Mainnet nodes for a peer-reported view. Actual readiness depends on network responses: if at least three suitable public IPv4 /16 groups do not agree, the app will leave the balance unknown and block sending rather than falsely claiming to be synced.

**Security / trust model:**
- Reported Mainnet Network ID, height, tip and chain work must agree across at least three public IPv4 /16 network groups before reading an account state.
- The displayed fast balance additionally requires matching spendable and total amounts from at least three such groups. It is **peer-reported and not independently AQM64-verified**.
- The displayed confirmed transaction history needs matching data from two agreeing groups; unconfirmed mempool entries can differ.
- Sending requires the matching reported chain plus identical UTXO sets across three groups, local ML-DSA-87 signing and acceptance by at least one full node.
- The phone does **not** cryptographically validate AQM64 Proof-of-Work, difficulty, Merkle inclusion, complete UTXO correctness or block consensus. Three different IP ranges **do not** prove three independent operators; compromised or colluding nodes can misreport funds. Use a full node for independently verified chain history, and avoid valuable funds in this unaudited Alpha wallet.

**Compatibility:** `com.auronq.mobile.light059` is a separately installed, CI debug-signed package. It does **not** overwrite or automatically migrate wallet data from earlier builds. Verify an offline backup and recovery before attempting a new installation; **never uninstall or clear data from a funded wallet**. The 0.5.9 build and automated tests passed, but actual startup speed and transaction behavior have not been validated across physical Android devices.

Historical Android downloads: [0.5.8 Alpha — Fast Balance + Logo](https://github.com/promirmir/AuronQ/releases/tag/android-v0.5.8-alpha-fast-balance-logo) (uses separate AQM64 header validation), [0.5.7 Alpha](https://github.com/promirmir/AuronQ/releases/tag/android-v0.5.7-alpha-runtime-independence), [0.5.6 Alpha](https://github.com/promirmir/AuronQ/releases/tag/android-v0.5.6-alpha-password-confirmed).

## Current architecture and limitations

AuronQ Mobile 0.5.9 Alpha is a **peer-quorum light wallet**, not a full validating node. Fast displayed balances, history and UTXO sets are reported by peers, not independently verified on the phone. It does not run background AQM64 header replay; sending relies on a separate three-group UTXO agreement rather than the 0.5.8 verified-header gate. Peer netgroup agreement cannot prove independent operators or a trustless UTXO state.

The source repository contains experimental advisory peer checkpoint components, but their existence **does not mean they are active, required or independently security-audited in every published full-node or Android binary**. Peer signatures and multiple IP groups alone cannot prove honest chain history. The technical history is available in [checkpoint design notes](../docs/DECENTRALIZED-CHECKPOINTS.md); those notes do not define a current user setup procedure.

**Never uninstall an existing wallet or erase application data to upgrade.** Verify offline recovery backups before moving funds. Separate APK signing certificates or application IDs can prevent in-place upgrades or migration.

## Independent network operation

Full validating nodes persist chain data and discovered peers. Once connected, they independently verify blocks and exchange data without GitHub. Fresh installs still need a working peer discovery path; long-term independence from founder-operated infrastructure remains to be demonstrated with an actual public-network outage/clean-install test.

[Network independence](../NETWORK-INDEPENDENCE.md) · [Current release policy](../PROJECT-FREEZE.md)

[Mainnet information](../MAINNET.md) · [Security](../SECURITY.md) · [Report a problem](https://github.com/promirmir/AuronQ/issues)
