# AuronQ — Mainnet feature freeze and Android public download
**Decision: 10 October 2026. Status: maintenance / verification, not a claim of a completed security audit.**

## One canonical public Android build

- **Android wallet for new installations:** [AuronQ Mobile 0.5.5 Alpha P2P — APK](https://github.com/promirmir/AuronQ/releases/download/android-v0.5.5-alpha-p2p/AuronQ-Mobile-0.5.5-alpha-p2p.apk)
- APK SHA-256: `32ea0db682b1a0d9efb30806f532490a26da888537f6bd1dfbddc0835030379e`
- Release page and original checksum file: [android-v0.5.5-alpha-p2p](https://github.com/promirmir/AuronQ/releases/tag/android-v0.5.5-alpha-p2p)
- The wallet is a **light client** with locally checked AQM64 headers after a previously verified, release-pinned historical checkpoint. Node-signed P2P checkpoint observations are **advisory** and cannot substitute for verified history or authorize transaction spending.
- Do not create another APK for cosmetic changes. The Android build has not received independent security review or a full documented device-based end-to-end payment test. Do not claim it is audited, guaranteed safe or permanently final for substantial funds.
- **Existing-wallet upgrade hazard:** published Android builds use ephemeral CI debug signing certificates. They may fail an in-place upgrade with an Android signature mismatch. Do **not** uninstall an existing wallet or clear its app data to update; that may permanently lose recovery secrets. Users should confirm secure recovery backups and avoid importing wallets holding significant funds into new test installations. Future release-grade wallet updates require a carefully managed, persistent Android app signing identity; publishing an arbitrary new debug APK is not a secure upgrade mechanism.

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
- Stop expanding features and checkpoint trust models until independently reviewed. Do not remove public source/release history under the pretext of a freeze.

This is a **code and release change freeze**, not shutdown of the blockchain. The chain is expected to be operated independently by its participants, with the above real-world constraints.
