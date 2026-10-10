# AuronQ — Mainnet feature freeze and Android public download
**Decision: 10 October 2026. Status: maintenance / verification, not a claim of a completed security audit.**

## Currently published Android release (verified against GitHub Releases)

- **Current Android APK:** [AuronQ Mobile 0.5.6 Alpha — Fast Sync](https://github.com/promirmir/AuronQ/releases/download/android-v0.5.6-alpha-fast-sync/AuronQ-Mobile-0.5.6-Alpha-Fast-Sync.apk)
- **APK SHA-256 (GitHub Release asset):** `e6613f0989a25b3d4d1d191b815d64388f9bac91f69c2336a1bf23c5af043ffa`
- [Release and checksum file](https://github.com/promirmir/AuronQ/releases/tag/android-v0.5.6-alpha-fast-sync).
- Previous 0.5.5 and earlier experimental release entries have been retired; do not send users to their deleted GitHub Release URLs.
- The Android wallet is a light client, not a full validating node. Fast peer-assisted account state and any optional checkpoint diagnostics are **not** proof of independent historical consensus verification.
- This version has not received an independent security audit. Do not assert guaranteed secure balances or instant fully validated payments.
- **Existing-wallet upgrade hazard:** published Android builds can use different CI debug signing certificates, preventing in-place upgrades. Do **not** uninstall an existing funded wallet or clear its application data to force an update. Verify recovery backups first.

## Full-node Mainnet and release freeze

- Keep the current publicly published **v1.7.13 full-node** binaries and their running chain data unchanged. Do not automatically install new binaries or rewrite genesis/Network ID/AQM64/difficulty/emission/transaction validity/UTXO validation/chainwork selection.
- PR [#145](https://github.com/promirmir/AuronQ/pull/145) introduced only node-self-signed, **nonconsensus** checkpoint hints. A self-signature proves the signing node's identity, not the correctness of historic work. Full validation rules stay authoritative.
- PR [#146](https://github.com/promirmir/AuronQ/pull/146) was **closed without merge** to avoid deploying another Mainnet full-node version during the freeze. No automated rollout or automatic v1.7.14 release is authorized.
- Do not delete or relocate any real participant's full chain, wallet, seed identity or peer store. Existing peer discovery and miners continue as configured.
- There is no magical guarantee of Mainnet survival if all full historical copies are lost, miners stop or no node remains reachable. Independent archival full nodes, node operator diversity, off-GitHub source+binary mirrors and proven offline bootstrap tests remain operational resilience requirements, not a reason to change consensus in a hurry.

## Security repairs are still allowed

A feature freeze is not an instruction to ignore vulnerabilities. Necessary nonconsensus security fixes are allowed only when:
1. A reproducible vulnerability/data-loss/availability defect is documented.
2. The proposed change is reviewed specifically for consensus and wallet-storage compatibility.
3. Mainnet golden-identity tests, cross-platform CI, wallet signing and AQM64 tests pass.
4. An explicit rollback plan and data backup exist. Stage nontrivial changes on a **noncritical test node** first, never all live Mainnet peers together.
5. No critical Mainnet code is changed via undocumented or unreviewed PRs. Consensus-breaking changes are **out of scope** for this maintenance freeze and would require a distinct public protocol proposal and testnet rollout.

The `main` branch is marked protected on GitHub, but exact required checks/rules may require repo administrator verification. The optional [freeze-gate workflow](.github/workflows/frozen-mainnet-gate.yml) rejects changes to frozen core files relative to the merged baseline `c1001ee2950158781f216820cc5e4911aa0044a9`. This is defense-in-depth and is not a substitute for repository ruleset/administrator protections.

## Operational checklist for finishing

- Verify the canonical Android download/HTTPS release and SHA-256 stay available.
- Give users one small written installation guide and a clearly marked *alpha/no-large-funds* warning.
- Preserve full-node v1.7.13 binary/archive and at least two independent physical operator backups of complete validated history.
- Keep collecting reproducible bug reports, especially missing peers, transaction broadcast failures, signature migration errors and multi-node chain divergence.
- Avoid expanding unreviewed checkpoint trust models. Retired GitHub Release entries are not current distribution channels; preserve source code, Git tags and protocol history.

This is a **code and release change freeze**, not shutdown of the blockchain. The chain is expected to be operated independently by its participants, with the above real-world constraints.
