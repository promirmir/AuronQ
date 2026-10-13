# AuronQ — Exchange Readiness and Final Preservation Gate

**Status: NOT CERTIFIED / NOT READY FOR PERMANENT ARCHIVING — 2026-10-13.**
This is a transparent evidence checklist, not an audit certificate, exchange endorsement, listing request, or claim of regulatory compliance.

AuronQ (AURQ) is a native public PoW Layer-1 coin. [Independent listings require no prior project approval](EXCHANGE-INTEGRATION.md), but every exchange remains responsible for its own legal, technical, liquidity, custody, and risk decisions.

## Non-negotiable invariants

- Preserve the existing Mainnet genesis hash, Network ID, AQM64, monetary issuance, transaction/block validity rules, difficulty, serialization and chainwork selection. Do **not** change consensus through a documentation or maintenance PR.
- No founder-controlled emergency mint, chain reset, privileged checkpoint, kill switch, or automatic code upgrade.
- No claims of an independent audit, sufficient mining decentralization, guaranteed finality, perpetual blockchain survival, or approved listing without corresponding dated, independently reviewable evidence.
- A GitHub archive makes the repository read-only **until its owner unarchives it**; it does not make independent nodes accept a new protocol or prevent others from forking source code.

## Publication gates — all required before declaring hardened/final

| Gate | Current evidence | State / acceptance evidence required |
|---|---|---|
| Wallet safety | [Desktop password/decryption report #134](https://github.com/promirmir/AuronQ/issues/134) | **OPEN:** reproduce with non-sensitive test fixtures, establish root cause, test backward-compatible recovery and signing; never ask for funded wallet seeds |
| Mobile wallet | [Android cold-start and balance report #149](https://github.com/promirmir/AuronQ/issues/149) | **OPEN:** demonstrate zero-cache install, correct unknown/stale display, safe spend refusal on disagreement, and backed-up upgrade/recovery |
| Consensus / cryptography | [Security policy](SECURITY.md), [test matrix](TEST-MATRIX.md) | **NOT INDEPENDENTLY AUDITED:** external review of UTXO, supply bounds, ML-DSA signatures, AQM64, difficulty, reorgs, serialization and P2P; publish scope, reproducible findings and remediations |
| Independent operation | [Operator handoff checklist](docs/INDEPENDENT-OPERATION.md), [network independence](NETWORK-INDEPENDENCE.md) | **NOT PROVEN:** three distinct independently controlled archival/full nodes, diverse networks, independent miners, fresh bootstrap and block relay with original services offline; dated logs/chainwork/tip proofs |
| Exchange custody integration | [Operator runbook](EXCHANGE-OPERATOR-RUNBOOK.md) | **NOT CERTIFIED:** exchange-owned full-node validation, persistent deposit indexing, reorg rollback, safe signing/withdrawals, crash and partition drills, accounting and maturity tests |
| Software provenance | [Project status](PROJECT-STATUS.md), [release checksums](.github/release-integrity-baseline.json) | **PARTIAL:** reproducible signed/provenanced source and release chain where possible; third-party byte verification; immutable future releases and protected tags |
| Mining and network risk | [Pool diversity #66](https://github.com/promirmir/AuronQ/issues/66) | **OPEN:** independently operated mining capacity and documented concentration/reorg threat assessment, not inferred solely from a pool dashboard |
| Legal and market listing | [Integration policy](EXCHANGE-INTEGRATION.md), [listing status](LISTING-STATUS.md) | **OPERATOR RESPONSIBILITY:** exchange performs its own local legal/compliance, liquidity and custody assessment; project makes no readiness or listing guarantee |

## Evidence format

For each passed item publish: date in UTC, reviewer/operator (or independently verifiable pseudonym), tested commit + binary SHA-256, node software version, test environment, expected/actual outcome, reproducible sanitized logs, relevant issue/PR, and residual risks. Mark an item passed **only after evidence is available**. A green CI badge is neither an external audit nor an end-to-end custody proof.

## Finalization procedure (only after all blocking gates pass)

1. Publish a signed, annotated, immutable final source reference and verified binary checksums; independently reproduce critical binaries where possible.
2. Confirm independent off-GitHub mirrors of source, releases, and full verified chain history; test a real restore and founder-offline bootstrap.
3. Protect release tags, prohibit force pushes/deletion, require PR and Windows/Linux + frozen-source checks, enable GitHub immutable releases for future publication where supported; record the actual settings.
4. Complete the independent security review, close or mitigate critical wallet/consensus issues, and document tests for reorg-safe exchange deposit and withdrawal workflows.
5. Publish a dated final technical acceptance record with independent evidence. Only then consider archiving the original GitHub repository as a read-only reference.
6. Keep a documented public path for independent operators to release security fixes or protocol alternatives; an archived source of truth must never be mistaken for a guarantee of an unchangeable network.

**Stop rule:** Do not archive or advertise as exchange-ready while the blockers above remain unresolved. Do not perform mainnet shutdown/offline resilience tests on the only available archival peer or seed.
