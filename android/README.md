# AuronQ Mobile

## Download

**Current published Android build (26 October 2026): AuronQ Mobile 0.5.8 Alpha — Fast Balance + Logo.**

- [Download the Android APK](https://github.com/promirmir/AuronQ/releases/download/android-v0.5.8-alpha-fast-balance-logo/AuronQ-Mobile-0.5.8-Alpha-Fast-Balance.apk)
- [Release notes and SHA-256 checksums](https://github.com/promirmir/AuronQ/releases/tag/android-v0.5.8-alpha-fast-balance-logo)
- APK SHA-256: `dcf8c06c87e3c3509252b247c4f889f8d383f86f71243d00109bc13c5ede9844`.

This 0.5.8 Alpha preserves independent network discovery/failover, restores fast **peer-reported balance** display after agreement from three public IPv4 network groups, keeps separate AQM64 verification and verified sending, and displays the official AuronQ logo. Wallet creation keeps two matching password fields and a 12-character minimum. Earlier [0.5.7](https://github.com/promirmir/AuronQ/releases/tag/android-v0.5.7-alpha-runtime-independence) and [0.5.6](https://github.com/promirmir/AuronQ/releases/tag/android-v0.5.6-alpha-password-confirmed) APKs remain historical and do not upgrade automatically.

**Compatibility warning:** this APK uses application ID `com.auronq.mobile.fast058` and a CI debug-signing certificate. It installs **beside** older versions, not over them, and cannot automatically access another app installation's encrypted local wallet. **Never uninstall a funded wallet or clear its data to force an upgrade.** Preserve and verify offline recovery material first. The app is Alpha light-client software with no independent professional security audit.

## 0.5.9 lightweight-quorum candidate (not yet published)

This branch contains an **experimental side-by-side Android APK** (`com.auronq.mobile.light059`). It can be built and downloaded from the Android CI artifacts after the pull request passes. It is **not** an update of the funded 0.5.8 wallet, and no official 0.5.9 release is claimed until actual builds and tests pass.

- Startup and ordinary refresh never invoke AQM64 header replay or download checkpoints. The wallet address remains available from encrypted local wallet data.
- The network screen is a peer-reported view. The balance needs matching chain tip/work plus identical spendable and total amounts across three publicly routable IPv4 /16 netgroups.
- History needs matching confirmed history from at least two netgroups, after chain status is consistent across three groups. Pending entries may vary.
- Sending first requires matching Mainnet network ID, tip, work, and height across three separate public /16 groups. Before *locally* signing a transaction, it also requires three identical UTXO lists from those groups, then broadcasts to validating full nodes.
- **Security tradeoff:** these checks do not verify Argon2id/AQM64 Proof-of-Work, difficulty, Merkle inclusion, chain selection, transaction confirmations or the full UTXO state cryptographically on the phone. Distinct /16 networks are not evidence of independent operators; colluding or compromised full nodes could misreport account data and UTXOs. The recipient and amount still must be verified by the user. For independent validation, run a full AuronQ node. Do not describe this light mode as trustless SPV.
- When fewer than three qualifying groups agree, balances stay unknown and sending is blocked. There is no silent downgrade to a single peer.

**Migration caution:** 0.5.9 TEST is debug-signed and uses a new application ID. Do not uninstall a funded version or assume wallet-file migration occurs automatically. Keep verified offline backups.

## Current architecture and limitations

AuronQ Mobile 0.5.8 Alpha is a **light wallet**, not a full validating node. Fast displayed balances are corroborated by peer groups and explicitly labelled as **not a full AQM64 proof**; sending remains gated by separate verified-chain checks. Peer netgroup agreement cannot prove independent operators or trustless UTXO state.

The source repository contains experimental advisory peer checkpoint components, but their existence **does not mean they are active, required or independently security-audited in every published full-node or Android binary**. Peer signatures and multiple IP groups alone cannot prove honest chain history. The technical history is available in [checkpoint design notes](../docs/DECENTRALIZED-CHECKPOINTS.md); those notes do not define a current user setup procedure.

**Never uninstall an existing wallet or erase application data to upgrade.** Verify offline recovery backups before moving funds. Separate APK signing certificates or application IDs can prevent in-place upgrades or migration.

## Independent network operation

Full validating nodes persist chain data and discovered peers. Once connected, they independently verify blocks and exchange data without GitHub. Fresh installs still need a working peer discovery path; long-term independence from founder-operated infrastructure remains to be demonstrated with an actual public-network outage/clean-install test.

[Network independence](../NETWORK-INDEPENDENCE.md) · [Current release policy](../PROJECT-FREEZE.md)

[Mainnet information](../MAINNET.md) · [Security](../SECURITY.md) · [Report a problem](https://github.com/promirmir/AuronQ/issues)
