# AuronQ — Mainnet feature freeze and release policy
**Decision: 10 October 2026. Status: maintenance / verification, not a claim of a completed security audit.**

## Current public software and maintenance — 27 October 2026

The Mainnet consensus freeze remains in effect. Reviewed maintenance can correct non-consensus availability, security, compatibility and release issues without changing genesis, Network ID, AQM64, monetary policy or validity rules.

| Component | Public release | Installation |
| --- | --- | --- |
| Windows Desktop/full node | [v1.7.14](https://github.com/promirmir/AuronQ/releases/tag/v1.7.14) | [Windows](README-WINDOWS.md) |
| Linux graphical Desktop | [v1.7.14.1 DEB](https://github.com/promirmir/AuronQ/releases/tag/linux-desktop-v1.7.14.1) | [Linux](README-LINUX.md) |
| Android light wallet | [0.5.9 Alpha Light Quorum](https://github.com/promirmir/AuronQ/releases/tag/android-v0.5.9-alpha-light-quorum) | [Android](android/README.md) |
| Universal Miner | [0.4.9 Alpha](https://github.com/promirmir/AuronQ/releases/tag/miner-v0.4.9-alpha) | [Miner guide](UNIVERSAL-MINER-GUIDE.md) |

Android 0.5.9 is a **separately installed, debug-signed light wallet** (app ID `com.auronq.mobile.light059`). It does not upgrade or automatically import previously installed funded wallets. Its balance/history/UTXO views rely on agreeing full-node reports, not independent on-device AQM64 chain verification. Preserve backups and do not delete earlier funded wallet data.

Mainnet runtime-independent discovery/failover was reviewed in PRs #187 and #189; Linux graphical packaging was merged in PR #202. The published binaries are not updated automatically on existing devices. **AuronQ requires an independent security audit.** [Current release checksums](PROJECT-STATUS.md).

## Historical 10 October 2026 Android release snapshot (superseded)

- **Current Android APK:** [AuronQ Mobile 0.5.6 Alpha — Fast Sync](https://github.com/promirmir/AuronQ/releases/download/android-v0.5.6-alpha-fast-sync/AuronQ-Mobile-0.5.6-Alpha-Fast-Sync.apk)
- **APK SHA-256 (GitHub Release asset):** `e6613f0989a25b3d4d1d191b815d64388f9bac91f69c2336a1bf23c5af043ffa`
- [Release and checksum file](https://github.com/promirmir/AuronQ/releases/tag/android-v0.5.6-alpha-fast-sync).
- Previous 0.5.5 and earlier experimental release entries have been retired; do not send users to their deleted GitHub Release URLs.
- The Android wallet is a light client, not a full validating node. Fast peer-assisted account state and any optional checkpoint diagnostics are **not** proof of independent historical consensus verification.
- This version has not received an independent security audit. Do not assert guaranteed secure balances or instant fully validated payments.
- **Existing-wallet upgrade hazard:** published Android builds can use different CI debug signing certificates, preventing in-place upgrades. Do **not** uninstall an existing funded wallet or clear its application data to force an update. Verify recovery backups first.

## Full-node Mainnet and release freeze

- Preserve both historical **v1.7.13** and current **v1.7.14** full-node release archives and all participants' chain data. Do not automatically update users' binaries or rewrite genesis/Network ID/AQM64/difficulty/emission/transaction validity/UTXO validation/chainwork selection.
- PR [#145](https://github.com/promirmir/AuronQ/pull/145) introduced only node-self-signed, **nonconsensus** checkpoint hints. A self-signature proves the signing node's identity, not the correctness of historic work. Full validation rules stay authoritative.
- **Historical decision (10 October 2026):** PR [#146](https://github.com/promirmir/AuronQ/pull/146) was closed without merge, and at that time automatic release of v1.7.14 was not authorized. This restriction on that earlier proposal was superseded by the reviewed and tested maintenance-release approvals in [PR #192](https://github.com/promirmir/AuronQ/pull/192) and [PR #193](https://github.com/promirmir/AuronQ/pull/193), without changing Mainnet consensus.
- Do not delete or relocate any real participant's full chain, wallet, seed identity or peer store. Existing peer discovery and miners continue as configured.
- There is no magical guarantee of Mainnet survival if all full historical copies are lost, miners stop or no node remains reachable. Independent archival full nodes, node operator diversity, off-GitHub source+binary mirrors and proven offline bootstrap tests remain operational resilience requirements, not a reason to change consensus in a hurry.

## Security repairs are still allowed

A feature freeze is not an instruction to ignore vulnerabilities. Necessary nonconsensus security fixes are allowed only when:
1. A reproducible vulnerability/data-loss/availability defect is documented.
2. The proposed change is reviewed specifically for consensus and wallet-storage compatibility.
3. Mainnet golden-identity tests, cross-platform CI, wallet signing and AQM64 tests pass.
4. An explicit rollback plan and data backup exist. Stage nontrivial changes on a **noncritical test node** first, never all live Mainnet peers together.
5. No critical Mainnet code is changed via undocumented or unreviewed PRs. Consensus-breaking changes are **out of scope** for this maintenance freeze and would require a distinct public protocol proposal and testnet rollout.

The `main` branch is protected on GitHub. The [Frozen Mainnet source guard](.github/workflows/frozen-mainnet-gate.yml) rejects unreviewed changes to protected paths against its latest accepted maintenance baseline (release exception #191 commit `5163e07ac5c0148e83745e6fffed7da69a3aeb74`) and independently verifies immutable Mainnet identity. Protection requires further independent review for consensus-critical changes.

## Operational checklist for finishing

- Verify the canonical Android download/HTTPS release and SHA-256 stay available.
- Keep one current set of [Windows](README-WINDOWS.md), [Linux](README-LINUX.md) and [Android](android/README.md) installation instructions and a clear audit-status statement.
- Preserve full-node v1.7.13 historical and v1.7.14 current binaries and at least two independent physical operator backups of complete validated history.
- Keep collecting reproducible bug reports, especially missing peers, transaction broadcast failures, signature migration errors and multi-node chain divergence.
- Avoid expanding unreviewed checkpoint trust models. Retired GitHub Release entries are not current distribution channels; preserve source code, Git tags and protocol history.

This is a **consensus/feature freeze with a narrow maintenance path**, not shutdown of the blockchain. Necessary reviewed non-consensus security, compatibility, availability and independence fixes may still be accepted under the Mainnet change policy. The chain is expected to be operated independently by its participants, with the above real-world constraints.
