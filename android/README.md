# AuronQ Mobile

## Download

**Current recommended test build: AuronQ Mobile 0.5.6 Alpha — Fast Sync**

- [Download Android APK](https://github.com/promirmir/AuronQ/releases/download/android-v0.5.6-alpha-fast-sync/AuronQ-Mobile-0.5.6-Alpha-Fast-Sync.apk)
- [Release notes and SHA-256 checksum](https://github.com/promirmir/AuronQ/releases/tag/android-v0.5.6-alpha-fast-sync)

Faster wallet synchronization while preserving wallet creation, receiving, sending, and transaction history, including pending transactions. This is an **experimental Android alpha**, not an independently audited financial application. The Fast Sync build uses a separate app identity and does not upgrade or migrate existing wallet files. Never uninstall an existing funded wallet or erase its data to install a different APK. Keep secure offline recovery material; do not test with valuable funds.

## Current architecture and limitations

AuronQ Mobile 0.5.6 Alpha is a **light wallet**, not a full validating node. It uses fast synchronization and information obtained from network peers. Newly displayed balances, transaction history and transaction state must not be confused with independent verification of the entire blockchain.

The source repository contains experimental advisory peer checkpoint components, but their existence **does not mean they are active, required or independently security-audited in every published full-node or Android binary**. Peer signatures and multiple IP groups alone cannot prove honest chain history. The technical history is available in [checkpoint design notes](../docs/DECENTRALIZED-CHECKPOINTS.md); those notes do not define a current user setup procedure.

**Never uninstall an existing wallet or erase application data to upgrade.** Verify offline recovery backups before moving funds. Separate APK signing certificates or application IDs can prevent in-place upgrades or migration.

## Independent network operation

Full validating nodes persist chain data and discovered peers. Once connected, they independently verify blocks and exchange data without GitHub. Fresh installs still need a working peer discovery path; long-term independence from founder-operated infrastructure remains to be demonstrated with an actual public-network outage/clean-install test.

[Network independence](../NETWORK-INDEPENDENCE.md) · [Current release policy](../PROJECT-FREEZE.md)

[Mainnet information](../MAINNET.md) · [Security](../SECURITY.md) · [Report a problem](https://github.com/promirmir/AuronQ/issues)
