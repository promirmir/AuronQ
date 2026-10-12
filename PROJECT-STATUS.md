# AuronQ — canonical project status and preservation

**Updated: 12 October 2026.** This is the navigation and release-integrity reference for the **existing public Mainnet**. The protocol and historical chain are not being reset or migrated.

> **Status: Mainnet live, protocol change freeze in effect, stabilization incomplete.** No claim of a finished professional security audit, fully independent bootstrap, risk-free custody, or permanent immutability of GitHub account settings is made.

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
| Desktop/full node v1.7.13 Windows x64 | [AuronQ-1.7.13-Windows-x64.zip](https://github.com/promirmir/AuronQ/releases/download/v1.7.13/AuronQ-1.7.13-Windows-x64.zip) | `7a7ce3d0b31b290bb6a3e43ab357ec6977ee4e7098a7d950594eda6cb64b2ba1` |
| Desktop/full node v1.7.13 Linux amd64 | [AuronQ-1.7.13-Linux-amd64.tar.gz](https://github.com/promirmir/AuronQ/releases/download/v1.7.13/AuronQ-1.7.13-Linux-amd64.tar.gz) | `2bdd1ffa72ca99c7395a146d21943e4aedcabe24cde123b5af8ff3c6c640ecb6` |
| Official Android 0.5.6 Alpha, password confirmation | [AuronQ-Mobile-0.5.6-Alpha-Password-Confirmed.apk](https://github.com/promirmir/AuronQ/releases/download/android-v0.5.6-alpha-password-confirmed/AuronQ-Mobile-0.5.6-Alpha-Password-Confirmed.apk) | `09b8ece07746f7e0390aba812c2382f96d1df317c782793a538f0811fea0ebb5` |
| Miner v0.4.8 Alpha Windows x64 | [AuronQ-Miner-v0.4.8-alpha-Windows-x64-GUI-GPU.zip](https://github.com/promirmir/AuronQ/releases/download/miner-v0.4.8-alpha/AuronQ-Miner-v0.4.8-alpha-Windows-x64-GUI-GPU.zip) | `ad8a9aa75d734b5bf7ecbe5f3595896f2db4b17c8196395940c3a0dcfea711ee` |
| Miner v0.4.8 Alpha Linux amd64 | [AuronQ-Miner-v0.4.8-alpha-Linux-amd64.tar.gz](https://github.com/promirmir/AuronQ/releases/download/miner-v0.4.8-alpha/AuronQ-Miner-v0.4.8-alpha-Linux-amd64.tar.gz) | `ccea8334c6705f8350a267a1b6f84a24284f3fb9db51e7c1794b96e7bb86cd48` |

The Android password-confirmation APK has a **different application ID** (`com.auronq.mobile.quick2stagev2.passwordcheck`) and a CI debug signing certificate. It **cannot perform an in-place upgrade** from the former Android 0.5.6 installation. Never uninstall a funded wallet without verified recovery material. The Android light wallet is not a fully validating node.

**Release-immutability gap as of 12 October 2026:** GitHub reports `immutable: false` for the listed existing releases. SHA-256 checksums identify the bytes observed at publication but do not prevent their replacement at the same URL. Immutable Releases can be enabled for **future** releases in repository settings, but that toggle is **not retroactive**. Historic released binaries should also be copied to independent archival storage with digests and provenance. Do not reissue new, untested binaries solely to get an immutable badge.

**Read-only integrity monitoring:** [Pinned GitHub Release asset and tag baseline](.github/release-integrity-baseline.json) is checked by [Release integrity watch](.github/workflows/release-integrity-watch.yml) with [the verifier](scripts/verify_release_integrity.py). Every scheduled/manual run re-downloads the listed official assets and hashes the actual bytes against the reviewed baseline; PR runs compare tags/metadata and execute tampering regression tests. Changes are **detected, not prevented**: repository administrators can change release settings, and this check is not a substitute for immutable-release protection, independent signatures, outside-GitHub archiving or an independent audit.

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

## 6. Unresolved blockers; do not call this “final secure” yet

- [#134 — Desktop v1.7.13 wallet send/unlock failure](https://github.com/promirmir/AuronQ/issues/134): treat funded-wallet recovery safety as urgent until the root cause and tests are established.
- [#149 — Android fresh-install scalability and safe balance](https://github.com/promirmir/AuronQ/issues/149): avoid claims of instant, independently verified balances or confirmed spendability on an unaudited light client.
- [#65 — stabilization](https://github.com/promirmir/AuronQ/issues/65), [#66 — independent mining pools](https://github.com/promirmir/AuronQ/issues/66): prove independent operation and avoid hashrate concentration.
- Professional consensus, PoW, P2P and wallet audits remain outstanding, as does review of legal classification and any future project-originated exchange activities.

**Operational rule:** no new consensus features or emission changes. Security and critical wallet/data-loss fixes can still be proposed transparently, reviewed, tested and released without changing existing Mainnet rules. A total permanent repository lock is **not advised** while such issues remain open.

## 7. What “immutable” means

- **Past validated chain:** full nodes reject arbitrary historical alterations unless consensus/cumulative-work requirements are met; reorganization remains a technical possibility, especially on young PoW networks.
- **Mainnet protocol rules:** frozen by policy and tests for ordinary releases. A different protocol fork remains technically possible, but independent participants choose whether to follow it.
- **Source and releases:** Git history plus rulesets, immutable future releases, archived checksums, off-platform mirrors. No GitHub account or rule provides an unconditional, irreversible guarantee forever.
- **Network ownership:** open code and lack of a special master key are important; practical decentralization additionally requires independent miners, nodes, pools and first-contact paths.

Project communications, including exchange integration, should remain factual: independent operators may assess AURQ at their discretion subject to law; no listing or investment return is promised. This document is technical/operational, **not a determination of MiCA issuer status**.
