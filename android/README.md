# AuronQ Mobile

## Download

**Current official download (14 October 2026): AuronQ Mobile 0.5.6 Alpha — Password Confirmed (Fast Sync).**

- [Download the official Android APK](https://github.com/promirmir/AuronQ/releases/download/android-v0.5.6-alpha-password-confirmed/AuronQ-Mobile-0.5.6-Alpha-Password-Confirmed.apk)
- [Release notes and SHA-256 checksums](https://github.com/promirmir/AuronQ/releases/tag/android-v0.5.6-alpha-password-confirmed)
- APK SHA-256: `09b8ece07746f7e0390aba812c2382f96d1df317c782793a538f0811fea0ebb5`.

This published build retains the 0.5.6 Fast Sync wallet behavior and adds two matching password fields during wallet creation. The older [0.5.6 Fast Sync release](https://github.com/promirmir/AuronQ/releases/tag/android-v0.5.6-alpha-fast-sync) is historical, **not** the current recommended download.

**Compatibility warning:** the current APK uses application ID `com.auronq.mobile.quick2stagev2.passwordcheck` and a CI debug signing certificate. It installs separately rather than updating the previous application in place, and cannot automatically access another installation's wallet files. **Do not uninstall an existing funded wallet, clear its data or assume automatic wallet migration.** Preserve and verify offline wallet recovery material first. The Android wallet is an Alpha light client without an independent professional security audit; fast synchronization is not proof of full historical validation.

## Current architecture and limitations

AuronQ Mobile 0.5.6 Alpha is a **light wallet**, not a full validating node. It uses fast synchronization and information obtained from network peers. Newly displayed balances, transaction history and transaction state must not be confused with independent verification of the entire blockchain.

The source repository contains experimental advisory peer checkpoint components, but their existence **does not mean they are active, required or independently security-audited in every published full-node or Android binary**. Peer signatures and multiple IP groups alone cannot prove honest chain history. The technical history is available in [checkpoint design notes](../docs/DECENTRALIZED-CHECKPOINTS.md); those notes do not define a current user setup procedure.

**Never uninstall an existing wallet or erase application data to upgrade.** Verify offline recovery backups before moving funds. Separate APK signing certificates or application IDs can prevent in-place upgrades or migration.

## Independent network operation

Full validating nodes persist chain data and discovered peers. Once connected, they independently verify blocks and exchange data without GitHub. Fresh installs still need a working peer discovery path; long-term independence from founder-operated infrastructure remains to be demonstrated with an actual public-network outage/clean-install test.

[Network independence](../NETWORK-INDEPENDENCE.md) · [Current release policy](../PROJECT-FREEZE.md)

[Mainnet information](../MAINNET.md) · [Security](../SECURITY.md) · [Report a problem](https://github.com/promirmir/AuronQ/issues)
