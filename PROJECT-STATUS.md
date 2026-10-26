# AuronQ — canonical project status and preservation

**Updated: 26 October 2026 (runtime-independence and release-status reconciliation).** This is the navigation and release-integrity reference for the **existing public Mainnet**. The protocol and historical chain are not being reset or migrated.

> **Status: Mainnet live, protocol change freeze in effect, stabilization incomplete.** Protected `main` includes reviewed non-consensus runtime-independence improvements, but no claim of a finished professional security audit, witnessed founder-independent bootstrap, risk-free custody, or permanent immutability of GitHub account settings is made.

## 1. One source of truth for each subject

| Subject | Canonical documentation |
| --- | --- |
| Mainnet identifiers and genesis | [MAINNET.md](MAINNET.md) |
| Block and transaction consensus format | [PROTOCOL.md](PROTOCOL.md) |
| AQM64 proof of work | [AQM64.md](AQM64.md) |
| Economic issuance and genesis allocation | [WHITEPAPER.md](WHITEPAPER.md) section 9; [MAINNET.md](MAINNET.md) |
| Restrictions on consensus changes | [MAINNET-CHANGE-POLICY.md](MAINNET-CHANGE-POLICY.md) |
| Current maintenance and acceptance gates | [PROJECT-FREEZE.md](PROJECT-FREEZE.md); [STABILIZATION.md](STABILIZATION.md) |
| Independently operated nodes and survival testing | [NETWORK-INDEPENDENCE.md](NETWORK-INDEPENDENCE.md) |
| Operator handoff and founder-free acceptance | [Independent operation checklist](docs/INDEPENDENT-OPERATION.md) |
| Browse all project documentation | [Categorized documentation index](docs/README.md) |
| Security limitations and private reports | [SECURITY.md](SECURITY.md) |
| Current software / installation guidance | [README.md](README.md); [GitHub Releases](https://github.com/promirmir/AuronQ/releases) |
| Independent exchange technical integration | [EXCHANGE-INTEGRATION.md](EXCHANGE-INTEGRATION.md); [EXCHANGE-OPERATOR-RUNBOOK.md](EXCHANGE-OPERATOR-RUNBOOK.md) |
| Project-originated exchange correspondence | [LISTING-STATUS.md](LISTING-STATUS.md) (historic dated operational decision only) |
| Old versions and historical notes | [CHANGELOG.md](CHANGELOG.md); [docs/archive/](docs/archive/) |

Older files and commit history are evidence of past decisions; they do not override the canonical Mainnet specification or current release references.

## 2. Mainnet identity — preserve, never silently replace

- **Consensus model:** UTXO, cumulative-work chain selection, AQM64 PoW, ML-DSA-87 transaction signatures.
- **Network ID:** `44e62c2ace002a6660c14e252173c1aa303529c68e40c998e92da2b453f44f30b1e58c94d533587e2186004593fb856c433fcdb5418ed430ec8617e29529365c`
- **Genesis hash:** `5750a455c04bfe93c9edfef1a12744b05e29ac6da1a9dd5b790566629dea2080c581beb2f0324efba2067c9efb113ed7a29598fd3ffbed965ced31f265d0cec4`
- **Nominal cap:** 21,000,000 AURQ including the openly documented 210,000 AURQ (1%) founder genesis allocation.
- **Default peer port:** TCP 18444.

Neither a GitHub administrator nor an explorer can unilaterally rewrite blocks accepted by independently validating full nodes. Conversely, no publication can prevent others from distributing forks or proposing incompatible code: each node operator chooses the software and network rules it runs. Mainnet consensus parameters are frozen for ordinary maintenance; exceptional changes would require public technical review, compatibility/fork assessment and independent adoption, **not a quiet repository edit**.

## 3. Public release artifacts — identity and verification

These are the **published** files identified on GitHub as of this document's date, not certified secure production releases. Verify downloaded bytes locally and verify the release page before use.

| Component | Canonical release | Asset SHA-256 (hex) |
| --- | --- | --- |
| Desktop/full node v1.7.14 Windows x64 | [AuronQ-1.7.14-Windows-x64.zip](https://github.com/promirmir/AuronQ/releases/download/v1.7.14/AuronQ-1.7.14-Windows-x64.zip) | `5e6c3537286aefa5a59d9074e6ad810bd903ad755d238c8d98224356c3455a97` |
| Desktop/full node v1.7.14 Linux amd64 | [AuronQ-1.7.14-Linux-amd64.tar.gz](https://github.com/promirmir/AuronQ/releases/download/v1.7.14/AuronQ-1.7.14-Linux-amd64.tar.gz) | `a86a0f7f9814de1fe2a0c85fd5d30e7bc8882e2f579b93e59acfa2e9cb6a160f` |
| Android 0.5.9 Alpha — Light Quorum / Fast Start (current) | [AuronQ-Mobile-0.5.9-Alpha-Light.apk](https://github.com/promirmir/AuronQ/releases/download/android-v0.5.9-alpha-light-quorum/AuronQ-Mobile-0.5.9-Alpha-Light.apk) | `4ce9f59f07fa7544e4edb66cac0c3158c80a12808b75f820dbfc20cf54b9915f` |
| Android 0.5.8 Alpha — Fast Balance + Logo | [AuronQ-Mobile-0.5.8-Alpha-Fast-Balance.apk](https://github.com/promirmir/AuronQ/releases/download/android-v0.5.8-alpha-fast-balance-logo/AuronQ-Mobile-0.5.8-Alpha-Fast-Balance.apk) | `dcf8c06c87e3c3509252b247c4f889f8d383f86f71243d00109bc13c5ede9844` |
| Miner v0.4.9 Alpha Windows GUI/GPU x64 | [AuronQ-Miner-v0.4.9-alpha-Windows-x64-GUI-GPU.zip](https://github.com/promirmir/AuronQ/releases/download/miner-v0.4.9-alpha/AuronQ-Miner-v0.4.9-alpha-Windows-x64-GUI-GPU.zip) | `2ba02cb17c56359f54da28fb52474e36a90465a30f82c00e5d34c7154f219547` |
| Miner v0.4.9 Alpha Linux GPU x64 | [AuronQ-Miner-v0.4.9-alpha-Linux-x64-GPU.tar.gz](https://github.com/promirmir/AuronQ/releases/download/miner-v0.4.9-alpha/AuronQ-Miner-v0.4.9-alpha-Linux-x64-GPU.tar.gz) | `4939ac1e444c12628f816cd8deb1324ccf01fd7b28204dc65d66b144d096f2ee` |

**Current Android 0.5.9 release (26 October 2026):** Public prerelease [0.5.9 Light Quorum](https://github.com/promirmir/AuronQ/releases/tag/android-v0.5.9-alpha-light-quorum), built from protected main commit `b1e0d3ba1ddfd17670cb27e4ca8c5579a46610a8` after passing PR Go, Windows/Linux CI, race detector, frozen-source check and Android CI. The phone no longer automatically replays memory-hard AQM64 headers or fetches checkpoints on startup. It displays a peer-reported balance only after agreement of three public IPv4 /16 netgroups, and sends only after matching UTXO replies from three groups and local signing. **These checks do not cryptographically validate PoW or UTXO state; the same operator may control several network groups.** No physical-device readiness guarantee or independent security audit exists. The CI debug-signed app ID `com.auronq.mobile.light059` is separate, with no automatic migration of older funded wallet data. Preserve verified offline backups; never uninstall a funded app as an update workaround.

Android 0.5.8 (historical) uses application ID `com.auronq.mobile.fast058` and a CI debug signing certificate. **It does not upgrade or automatically migrate wallets from Android 0.5.6/0.5.7.** The fast balance view is peer-reported from three public IPv4 network groups, not independently verified AQM64; transaction sending retains separate validation. Never uninstall funded older apps without verified offline recovery.

### Runtime-independence changes shipped in the current binaries

As of 26 October 2026, reviewed network-independence changes PRs #187/#189 are packaged in Desktop 1.7.14, Android 0.5.9 Alpha and Miner 0.4.9 Alpha. Desktop builds derive from `6db1784b054c56540cee8dc68751552d6a5f0abc`, Android 0.5.9 from `b1e0d3ba1ddfd17670cb27e4ca8c5579a46610a8`, and Miner 0.4.9 from `6269ac5ec128668775ecf59f277a3df696e2c3c8`. Android 0.5.9 adds lightweight three-netgroup peer-reported wallet operations, with no consensus or wallet-format changes:

- full-node/Desktop discovery separates candidates from callback-verified public peers and persists/gossips only verified peers;
- Android prefers cached/bundled Mainnet peers before optional GitHub manifests and verifies gossip candidates before persistence;
- CLI thin clients and standalone/miner workers can fail over across locally cached and bundled Mainnet endpoints after exact Network ID verification;
- repeated endpoint failures are locally quarantined, and clear same-height conflicting tips fail closed;
- the Universal Miner multi-GPU path propagates the same network/failover settings to child workers.

The source changes passed PR CI plus post-merge Windows/Linux CI, race detection, Mainnet identity/safety checks, Android build, reproducible-build tests and Miner compatibility checks. All three current packages were built and published through successful GitHub Actions workflows; the older 1.7.13/0.5.6/0.4.8 packages are historical, and installations do not update themselves automatically. Passing CI and publishing packages **do not establish adversarial resistance, physical-device reliability, or an independent professional audit**.

**Release-immutability gap (verified 26 October 2026):** Some earlier GitHub Releases reported `immutable: false`. Do not assume the 0.5.9 release is immutable without checking the GitHub release metadata/settings. Published SHA-256 identifies observed bytes but does not prohibit replacement at the same URL. Independent off-GitHub archives and signing/provenance controls remain desirable; do not reissue untested binaries to change release settings.

**Read-only release integrity monitoring:** [The pinned baseline](.github/release-integrity-baseline.json), [integrity watch](.github/workflows/release-integrity-watch.yml) and [verifier](scripts/verify_release_integrity.py) currently monitor the three historically pinned Desktop 1.7.13 / Android 0.5.6 / Miner 0.4.8 releases. Newer Desktop 1.7.14, Android 0.5.9 and Miner 0.4.9 have published checksums but are not yet in the periodic full-download baseline. Monitoring detects changes; it cannot prevent them.

## 4. Repository controls — verified and remaining

**Verified current protections (12 October 2026):** the active GitHub ruleset [“ochrona main”](https://github.com/promirmir/AuronQ/rules/24452047) targets `main`, rejects branch deletion and non-fast-forward updates, requires a pull request and resolved conversations, allows squash merges only, and requires `test (ubuntu-latest)` plus `test (windows-latest)` with the branch up to date. The ruleset has no listed bypass actor. **Required approving reviews: 0**, so successful CI is not a substitute for independent review.

[The frozen Mainnet guard](.github/workflows/frozen-mainnet-gate.yml) independently checks a frozen source baseline and Mainnet identity. At the date of this document, it **runs in CI but is not itself included in the ruleset's required-status-check list**.

**Repository administrator action still required** to make protections stricter:

1. Keep direct writes, force pushes and deletions blocked for `main`. Require the **Frozen Mainnet source guard / frozen-source** job in addition to current Windows/Linux checks, after verifying the exact check name shown in a passing PR.
2. If independent reviewers are available, require at least **one approving review** and restrict bypasses; do not configure a review requirement that permanently deadlocks a sole-maintainer project with no reviewer.
3. Create a **tag ruleset** blocking deletion and retargeting of release tags and requiring deliberate approvals for new official tags. Keep release history; do not delete or overwrite previous official tags.
4. In **Settings → General → Releases**, enable **release immutability for future releases**; publish each only after attaching all expected binaries and checksums.
5. Confirm ruleset state in GitHub settings, because repository administrators can ultimately edit repository controls; this is governance, not cryptographic impossibility.

These are actions requiring administrative settings access. Documentation or an ordinary commit **does not apply those settings**.

## 5. Preservation outside GitHub

Long-term survival requires **independent copies**, not only permissions on an account:

- Keep at least two independently controlled copies of complete, verified blockchain history on separate physical systems/providers; periodically test restoration to a fresh node.
- Preserve a bare Git repository mirror, canonical protocol docs, verified release binaries and SHA-256 digests outside GitHub. A single GitHub account, website and domain are insufficient.
- Prove a fresh node can find peers and synchronize **while all original project-operated rendezvous endpoints are offline**. Continue block propagation/mining during the drill.
- Never publish wallet seeds, private keys, passwords, or encrypted wallet files in public repositories or audit logs.

These preservation checks are **not yet certified complete**. See [NETWORK-INDEPENDENCE.md](NETWORK-INDEPENDENCE.md).

## 6. Open engineering items and closed support case; do not call this “final secure” yet

- **Open:** [#149 — Android fresh-install scalability and safe balance](https://github.com/promirmir/AuronQ/issues/149). Avoid claims of instant, independently verified balances or confirmed spendability on a light client without independent verification.
- **Open:** [#65 — stabilization](https://github.com/promirmir/AuronQ/issues/65) and [#66 — independent mining pools](https://github.com/promirmir/AuronQ/issues/66). Independent operation and reduced hashrate concentration still need evidence.
- **Closed 13 October 2026:** [#134 — Desktop wallet password-unavailable support case](https://github.com/promirmir/AuronQ/issues/134). GitHub records this report as closed without a confirmed software defect; no source-code wallet fix is claimed from this closure. Password recovery remains important, but this is **not an open release-blocking bug report**.
- Independent professional consensus, PoW, P2P and wallet audits remain outstanding, as does review of legal classification and any future project-originated exchange activities.

**Operational rule:** no new consensus features or emission changes. Security and critical wallet/data-loss fixes can still be proposed transparently, reviewed, tested and released without changing existing Mainnet rules. A total permanent repository lock is **not advised** while such issues remain open.

## 7. What “immutable” means

- **Past validated chain:** full nodes reject arbitrary historical alterations unless consensus/cumulative-work requirements are met; reorganization remains a technical possibility, especially on young PoW networks.
- **Mainnet protocol rules:** frozen by policy and tests for ordinary releases. A different protocol fork remains technically possible, but independent participants choose whether to follow it.
- **Source and releases:** Git history plus rulesets, immutable future releases, archived checksums, off-platform mirrors. No GitHub account or rule provides an unconditional, irreversible guarantee forever.
- **Network ownership:** open code and lack of a special master key are important; practical decentralization additionally requires independent miners, nodes, pools and first-contact paths.

Project communications, including exchange integration, should remain factual: independent operators may assess AURQ at their discretion subject to law; no listing or investment return is promised. This document is technical/operational, **not a determination of MiCA issuer status**.
