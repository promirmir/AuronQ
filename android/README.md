# AuronQ Mobile

## Download

**Current published Android build (26 October 2026): AuronQ Mobile 0.5.9 Alpha — Light Quorum / Fast Start.**

- [Download the Android APK](https://github.com/promirmir/AuronQ/releases/download/android-v0.5.9-alpha-light-quorum/AuronQ-Mobile-0.5.9-Alpha-Light.apk)
- [Release notes, APK and checksums](https://github.com/promirmir/AuronQ/releases/tag/android-v0.5.9-alpha-light-quorum)
- APK SHA-256: `4ce9f59f07fa7544e4edb66cac0c3158c80a12808b75f820dbfc20cf54b9915f`

The app opens its local wallet and network interface without waiting for on-phone AQM64 header replay or fetching checkpoints at startup. It discovers public AuronQ Mainnet nodes without requiring GitHub for routine operation. The peer-reported balance requires matching chain identity, height, tip, chain work and amounts from **three public IPv4 /16 netgroups**. History uses agreement on confirmed records from at least two groups, after the chain-state gate. Wallet keys and ML-DSA-87 signatures stay local.

**Security:** this is a faster but **less independently verifying** trust model. Three /16 networks are not proof of independent node operators. The phone does not independently verify AQM64 PoW, difficulty, Merkle inclusion, complete UTXOs or chain selection. Sending also requires three matching UTXO sets and is locally signed, but these peer claims do not become cryptographic proofs. Do not use this unaudited Alpha wallet for valuable funds. If fewer than three groups agree, the balance remains unknown and sending is blocked. Real connection and balance latency depend on reachable nodes.

**Compatibility:** debug-signed Android package `com.auronq.mobile.light059` installs **alongside** existing 0.5.6/0.5.7/0.5.8 wallets. It cannot automatically access or migrate another application's locally encrypted wallet. **Never uninstall a funded wallet or clear its data to force an upgrade.** Keep a verified offline wallet backup first.

Older builds: [0.5.8 Alpha](https://github.com/promirmir/AuronQ/releases/tag/android-v0.5.8-alpha-fast-balance-logo) · [0.5.7 Alpha](https://github.com/promirmir/AuronQ/releases/tag/android-v0.5.7-alpha-runtime-independence) · [0.5.6 Alpha](https://github.com/promirmir/AuronQ/releases/tag/android-v0.5.6-alpha-password-confirmed).

## Current architecture and limitations

AuronQ Mobile 0.5.9 Alpha is a **peer-quorum light wallet**, not a validating node. It does not calculate the AQM64 chain on the phone during routine startup. Balances, confirmed history and sending rely on matching responses from public peer groups; these are **not trustless blockchain proofs** or independently verified UTXO state.

The source repository contains experimental advisory peer checkpoint components, but their existence **does not mean they are active, required or independently security-audited in every published full-node or Android binary**. Peer signatures and multiple IP groups alone cannot prove honest chain history. The technical history is available in [checkpoint design notes](../docs/DECENTRALIZED-CHECKPOINTS.md); those notes do not define a current user setup procedure.

**Never uninstall an existing wallet or erase application data to upgrade.** Verify offline recovery backups before moving funds. Separate APK signing certificates or application IDs can prevent in-place upgrades or migration.

## Independent network operation

Full validating nodes persist chain data and discovered peers. Once connected, they independently verify blocks and exchange data without GitHub. Fresh installs still need a working peer discovery path; long-term independence from founder-operated infrastructure remains to be demonstrated with an actual public-network outage/clean-install test.

[Network independence](../NETWORK-INDEPENDENCE.md) · [Current release policy](../PROJECT-FREEZE.md)

[Mainnet information](../MAINNET.md) · [Security](../SECURITY.md) · [Report a problem](https://github.com/promirmir/AuronQ/issues)
