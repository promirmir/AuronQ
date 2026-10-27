# AuronQ Mobile

## Download and installation

**Current published Android build (26 October 2026): AuronQ Mobile 0.5.9 Alpha — Light Quorum / Fast Start.**

- [Download the Android APK](https://github.com/promirmir/AuronQ/releases/download/android-v0.5.9-alpha-light-quorum/AuronQ-Mobile-0.5.9-Alpha-Light.apk)
- [Release notes, APK and checksums](https://github.com/promirmir/AuronQ/releases/tag/android-v0.5.9-alpha-light-quorum)
- APK SHA-256: `4ce9f59f07fa7544e4edb66cac0c3158c80a12808b75f820dbfc20cf54b9915f`

1. Open the APK download link on an Android arm64 device.
2. Tap the downloaded APK. If Android requests permission to install from your browser or file manager, grant it only for this installation.
3. Open **AuronQ Mobile 0.5.9 Light** from the apps menu. The wallet UI loads without on-device AQM64 header replay; network balance availability depends on agreeing peers.
4. Create a new wallet or import one from your verified recovery backup. Keep any previously installed wallet and its data; this APK is a separate installation.

The app opens its local wallet and network interface without waiting for on-phone AQM64 header replay or fetching checkpoints at startup. It discovers public AuronQ Mainnet nodes without requiring GitHub for routine operation. The peer-reported balance requires matching chain identity, height, tip, chain work and amounts from **three public IPv4 /16 netgroups**. History uses agreement on confirmed records from at least two groups, after the chain-state gate. Wallet keys and ML-DSA-87 signatures stay local.

**How light verification works:** the 0.5.9 client uses matching Mainnet state from three public IPv4 /16 peer groups; three IP groups are not proof of independent node operators. The phone does not independently verify AQM64 PoW, difficulty, Merkle inclusion or complete UTXOs. Sending requires matching peer-reported UTXO sets and local signing. With insufficient agreement, balance remains unknown and sending is blocked. **AuronQ requires an independent security audit.**

**Compatibility:** debug-signed Android package `com.auronq.mobile.light059` installs **alongside** existing 0.5.6/0.5.7/0.5.8 wallets. It cannot automatically access or migrate another application's locally encrypted wallet. **Never uninstall a funded wallet or clear its data to force an upgrade.** Keep a verified offline wallet backup first.

Older builds: [0.5.8 Alpha](https://github.com/promirmir/AuronQ/releases/tag/android-v0.5.8-alpha-fast-balance-logo) · [0.5.7 Alpha](https://github.com/promirmir/AuronQ/releases/tag/android-v0.5.7-alpha-runtime-independence) · [0.5.6 Alpha](https://github.com/promirmir/AuronQ/releases/tag/android-v0.5.6-alpha-password-confirmed).

## Current architecture and limitations

AuronQ Mobile 0.5.9 Alpha is a **peer-quorum light wallet**, not a validating node. It does not calculate the AQM64 chain on the phone during routine startup. Balances, confirmed history and sending rely on matching responses from public peer groups; these are **not trustless blockchain proofs** or independently verified UTXO state.

Older advisory checkpoint designs are documented in [technical notes](../docs/DECENTRALIZED-CHECKPOINTS.md); they are not required to install or run AuronQ Mobile 0.5.9.

**Never uninstall an existing wallet or erase application data to upgrade.** Verify offline recovery backups before moving funds. Separate APK signing certificates or application IDs can prevent in-place upgrades or migration.

## Independent network operation

Full validating nodes persist chain data and discovered peers. Once connected, they independently verify blocks and exchange data without GitHub. Fresh installs still need a working peer discovery path; long-term independence from founder-operated infrastructure remains to be demonstrated with an actual public-network outage/clean-install test.

[Network independence](../NETWORK-INDEPENDENCE.md) · [Current release policy](../PROJECT-FREEZE.md)

[Mainnet information](../MAINNET.md) · [Security](../SECURITY.md) · [Report a problem](https://github.com/promirmir/AuronQ/issues)
