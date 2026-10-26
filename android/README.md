# AuronQ Mobile

## Download

**Current published Android build (26 October 2026): AuronQ Mobile 0.5.8 Alpha — Fast Balance + Logo.**

- [Download the Android APK](https://github.com/promirmir/AuronQ/releases/download/android-v0.5.8-alpha-fast-balance-logo/AuronQ-Mobile-0.5.8-Alpha-Fast-Balance.apk)
- [Release notes and SHA-256 checksums](https://github.com/promirmir/AuronQ/releases/tag/android-v0.5.8-alpha-fast-balance-logo)
- APK SHA-256: `dcf8c06c87e3c3509252b247c4f889f8d383f86f71243d00109bc13c5ede9844`.

This 0.5.8 Alpha preserves independent network discovery/failover, restores fast **peer-reported balance** display after agreement from three public IPv4 network groups, keeps separate AQM64 verification and verified sending, and displays the official AuronQ logo. Wallet creation keeps two matching password fields and a 12-character minimum. Earlier [0.5.7](https://github.com/promirmir/AuronQ/releases/tag/android-v0.5.7-alpha-runtime-independence) and [0.5.6](https://github.com/promirmir/AuronQ/releases/tag/android-v0.5.6-alpha-password-confirmed) APKs remain historical and do not upgrade automatically.

**Compatibility warning:** this APK uses application ID `com.auronq.mobile.fast058` and a CI debug-signing certificate. It installs **beside** older versions, not over them, and cannot automatically access another app installation's encrypted local wallet. **Never uninstall a funded wallet or clear its data to force an upgrade.** Preserve and verify offline recovery material first. The app is Alpha light-client software with no independent professional security audit.

## Current architecture and limitations

AuronQ Mobile 0.5.8 Alpha is a **light wallet**, not a full validating node. Fast displayed balances are corroborated by peer groups and explicitly labelled as **not a full AQM64 proof**; sending remains gated by separate verified-chain checks. Peer netgroup agreement cannot prove independent operators or trustless UTXO state.

The source repository contains experimental advisory peer checkpoint components, but their existence **does not mean they are active, required or independently security-audited in every published full-node or Android binary**. Peer signatures and multiple IP groups alone cannot prove honest chain history. The technical history is available in [checkpoint design notes](../docs/DECENTRALIZED-CHECKPOINTS.md); those notes do not define a current user setup procedure.

**Never uninstall an existing wallet or erase application data to upgrade.** Verify offline recovery backups before moving funds. Separate APK signing certificates or application IDs can prevent in-place upgrades or migration.

## Independent network operation

Full validating nodes persist chain data and discovered peers. Once connected, they independently verify blocks and exchange data without GitHub. Fresh installs still need a working peer discovery path; long-term independence from founder-operated infrastructure remains to be demonstrated with an actual public-network outage/clean-install test.

[Network independence](../NETWORK-INDEPENDENCE.md) · [Current release policy](../PROJECT-FREEZE.md)

[Mainnet information](../MAINNET.md) · [Security](../SECURITY.md) · [Report a problem](https://github.com/promirmir/AuronQ/issues)
