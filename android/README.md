# AuronQ Mobile

## Download

**Current published Android build (26 October 2026): AuronQ Mobile 0.5.7 Alpha — Runtime Independence.**

- [Download the Android APK](https://github.com/promirmir/AuronQ/releases/download/android-v0.5.7-alpha-runtime-independence/AuronQ-Mobile-0.5.7-Alpha-Runtime-Independence.apk)
- [Release notes and SHA-256 checksums](https://github.com/promirmir/AuronQ/releases/tag/android-v0.5.7-alpha-runtime-independence)
- APK SHA-256: `95666cf3d4872224d6af71a5034d955a2987c497f084bf75188f6eae9dd26a07`.

This 0.5.7 Alpha packages the reviewed GitHub-independent discovery/failover source changes, while keeping two matching password fields and a 12-character minimum when creating a wallet. Prior [0.5.6 password-confirmed](https://github.com/promirmir/AuronQ/releases/tag/android-v0.5.6-alpha-password-confirmed) and Fast Sync APKs are historical downloads and cannot upgrade automatically.

**Compatibility warning:** the current APK uses application ID `com.auronq.mobile.runtime057` and a CI debug signing certificate. It installs separately rather than updating older versions in place, and cannot automatically access another installation's wallet files. **Do not uninstall an existing funded wallet, clear its data or assume automatic wallet migration.** Preserve and verify offline wallet recovery material first. This is Alpha light-client software without an independent professional security audit; fast synchronization is not proof of full historical validation.

## Current architecture and limitations

AuronQ Mobile 0.5.7 Alpha is a **light wallet**, not a full validating node. It uses fast synchronization and information obtained from network peers. Newly displayed balances, transaction history and transaction state must not be confused with independent verification of the entire blockchain.

The source repository contains experimental advisory peer checkpoint components, but their existence **does not mean they are active, required or independently security-audited in every published full-node or Android binary**. Peer signatures and multiple IP groups alone cannot prove honest chain history. The technical history is available in [checkpoint design notes](../docs/DECENTRALIZED-CHECKPOINTS.md); those notes do not define a current user setup procedure.

**Never uninstall an existing wallet or erase application data to upgrade.** Verify offline recovery backups before moving funds. Separate APK signing certificates or application IDs can prevent in-place upgrades or migration.

## Independent network operation

Full validating nodes persist chain data and discovered peers. Once connected, they independently verify blocks and exchange data without GitHub. Fresh installs still need a working peer discovery path; long-term independence from founder-operated infrastructure remains to be demonstrated with an actual public-network outage/clean-install test.

[Network independence](../NETWORK-INDEPENDENCE.md) · [Current release policy](../PROJECT-FREEZE.md)

[Mainnet information](../MAINNET.md) · [Security](../SECURITY.md) · [Report a problem](https://github.com/promirmir/AuronQ/issues)
