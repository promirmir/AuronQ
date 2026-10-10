# AuronQ Mobile — Android

**Current published release: v0.5.3 Alpha · Experimental Mainnet light wallet**

This guide describes the current GitHub release. It replaces the former v0.4.2 Alpha documentation. AuronQ Mobile is **not a full validating node** and has not undergone an independent professional security audit.

## Published APK and verification

- [Download the v0.5.3 Alpha APK](https://github.com/promirmir/AuronQ/releases/download/android-v0.5.3-alpha/AuronQ-Mobile-0.5.3-alpha.apk)
- [Release notes](https://github.com/promirmir/AuronQ/releases/tag/android-v0.5.3-alpha)
- Filename: `AuronQ-Mobile-0.5.3-alpha.apk`
- Size: `9574566` bytes
- SHA-256: `913e723a5b918ecf76bc923b599f5196b2a3340ec554897aa084b7193f1b520c`

This hash identifies the **published** APK. An app previously installed under the same version label may be a different binary. To investigate differences, compare APK hashes and package metadata rather than assuming they are identical. **Never uninstall a working wallet or erase its local data before safely backing up recovery material.**

## Network model in v0.5.3 Alpha

The [published release notes](https://github.com/promirmir/AuronQ/releases/tag/android-v0.5.3-alpha) describe:

- Multiple replaceable Mainnet peers and direct globally routable public IPv4 fallbacks.
- HTTPS required for DNS-named nodes; cleartext communication limited to literal public IP addresses, excluding private, loopback, CGNAT and documentation ranges.
- Local verification of Mainnet headers, AQM64 Proof of Work, difficulty, timestamps and hash continuity.
- Persisted verification progress after successful header batches, for resuming interrupted synchronization.
- Separate display of connectivity and header verification status; transient errors should not erase verified state.
- Wallet-state queries constrained to peers agreeing on the verified chain state.

These are **release design claims**, not a substitute for independent implementation review or testing on a particular Android device. Verification from genesis may take longer as the blockchain grows. Multi-peer agreement and header verification are not equivalent to replaying and validating all transactions and blocks in a full node.

## Wallet safety

Protect passwords, local recovery material and private keys. Keep a tested offline backup before sending AURQ, changing phones, upgrading, reinstalling or clearing app data. Do not include wallet secrets in issue reports.

For independent full-chain validation use the [Desktop / Full Node](../START-HERE.md). Also see [Security policy](../SECURITY.md) and [Mainnet specification](../MAINNET.md).

## Reporting a difference between APK versions

Provide the GitHub release URL, SHA-256 of each available APK, Android version and the exact behavior or visible interface difference. Do **not** send private keys, recovery material or passwords.

[Open an issue](https://github.com/promirmir/AuronQ/issues).
