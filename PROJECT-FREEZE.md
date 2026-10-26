# AuronQ — Mainnet feature freeze and release policy
**Decision: 10 October 2026. Status: maintenance / verification, not a claim of a completed security audit.**

## Current published Android release — 26 October 2026

- **Current APK:** [AuronQ Mobile 0.5.8 Alpha — Fast Balance + Logo](https://github.com/promirmir/AuronQ/releases/download/android-v0.5.8-alpha-fast-balance-logo/AuronQ-Mobile-0.5.8-Alpha-Fast-Balance.apk).
- **Published APK SHA-256:** `dcf8c06c87e3c3509252b247c4f889f8d383f86f71243d00109bc13c5ede9844`.
- [Official release metadata](https://github.com/promirmir/AuronQ/releases/tag/android-v0.5.8-alpha-fast-balance-logo). The app ID is `com.auronq.mobile.fast058`, signed with CI debug credentials. It **installs separately from previous funded wallet apps** and cannot automatically access or migrate their local data. Never uninstall an existing funded wallet without verified recovery.
- **Current project/release navigation:** [PROJECT-STATUS.md](PROJECT-STATUS.md). Earlier release and freeze decisions below are retained as dated historical records, not current download guidance.

## Current maintenance state — 26 October 2026

The consensus freeze remains in force, but it does **not** prohibit reviewed non-consensus security/availability/independence maintenance. Two controlled maintenance exceptions have now been merged to protected `main`:

- PR #187: GitHub-independent runtime peer discovery, verified peer persistence/gossip and updated Android fallback behavior;
- PR #189 / issue #188: verified failover for CLI thin clients and standalone/multi-GPU Universal Miner workers.

The network maintenance changes preserved Mainnet identity and consensus rules and passed cross-platform CI and Frozen Mainnet checks. Packaged applications were published as **Desktop 1.7.14, Android 0.5.7 Alpha and Miner 0.4.9 Alpha** after reviewed release PRs #192 and #193; [release verification](PROJECT-STATUS.md) records current asset hashes.

**Distribution:** verified downloads of Desktop v1.7.14, Android 0.5.7 Alpha and Universal Miner v0.4.9 Alpha include the 26 October runtime-independence work. Historic 1.7.13 / 0.5.6 / 0.4.8 installed binaries do not update themselves automatically. Publishing software does not prove professional security review or founder-independent long-duration network operation.
- Neither a version number nor a published checksum proves the light wallet has been professionally audited; the scaling and wallet-safety issues remain open.

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
- Give users one small written installation guide and a clearly marked *alpha/no-large-funds* warning.
- Preserve full-node v1.7.13 historical and v1.7.14 current binaries and at least two independent physical operator backups of complete validated history.
- Keep collecting reproducible bug reports, especially missing peers, transaction broadcast failures, signature migration errors and multi-node chain divergence.
- Avoid expanding unreviewed checkpoint trust models. Retired GitHub Release entries are not current distribution channels; preserve source code, Git tags and protocol history.

This is a **consensus/feature freeze with a narrow maintenance path**, not shutdown of the blockchain. Necessary reviewed non-consensus security, compatibility, availability and independence fixes may still be accepted under the Mainnet change policy. The chain is expected to be operated independently by its participants, with the above real-world constraints.
