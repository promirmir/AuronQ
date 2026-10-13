# AuronQ — Community Independence and Maintainer Exit

**Prepared 2026-10-13. Status: proposed handoff; NOT an announcement that independence testing has passed or that the original operator has already shut down.**

AuronQ (AURQ) is an open-source public Proof-of-Work Layer-1 blockchain. The reference implementation is maintained under a Mainnet consensus change freeze. The goal is a community-operated network requiring no founder authorization to run nodes, mine, build compatible clients, audit, or independently evaluate exchange integration.

## What the community may do without asking the original author

- Run and restore independent validating nodes with their own durable full chain history.
- Mine blocks independently or operate compatible pools.
- Publish independently developed wallets, explorers, exchange gateways and alternative software builds.
- Mirror the repository, specifications, verified release artifacts and their checksums outside GitHub.
- Audit the source and report vulnerabilities responsibly.
- Propose improvements or forks. Each node operator alone chooses whether to adopt new code; neither a GitHub commit nor a community statement can force consensus changes onto independent nodes.
- Evaluate or list AURQ on exchanges independently, subject to each operator's legal, technical, liquidity, custody and regulatory review.

## Conditions before the original operator exits

The following must be completed by real independent operators and accompanied by dated, reproducible, redacted evidence. Do not mark any check complete merely because this document exists.

- [ ] Three full archival nodes under **three genuinely independent operators**, on diversified infrastructure, agree on exact Mainnet Network ID/genesis, validated tip and cumulative work.
- [ ] At least two unrelated mining operators can produce and propagate valid blocks without founder infrastructure.
- [ ] At least two off-GitHub independently maintained mirrors retain source, release binaries and hashes, and complete chain backups. An unrelated party successfully restores from one mirror.
- [ ] A fresh node with an empty peer cache bootstraps while the original project's own endpoints are offline; record exact tested release/version, start/end tips, UTC date and sanitized logs.
- [ ] A new valid block propagates to the independent nodes during this controlled test.
- [ ] Critical wallet recovery, consensus and reorganization defects have no unresolved reproducible high/critical findings; independent security review scope and outstanding risks are disclosed.
- [ ] Release tags and future artifacts are hardened using available protection and immutable release settings; final reference source and checksums are mirrored and authenticated.
- [ ] At least one independent exchange/custodian integration can verify deposits/withdrawals and reorg rollback from an independently validating node **if claiming operational exchange readiness**.
- [ ] A public final handoff announcement is posted *after* the evidence exists. No statements suggesting professional audit, guaranteed immutability, listing, investment value or legal approval without proof.

## Operational safety and freeze

- Preserve the present genesis, Network ID, AQM64, transaction signatures/validity, issuance, difficulty and cumulative-work selection. No silent Mainnet consensus change.
- Never intentionally shut down the final reachable archival peer or only remaining bootstrap seed to perform an independence test. Coordinate temporary outage windows only after validating alternate independently managed infrastructure.
- No admin backdoor, remote kill switch or centrally approved trusted checkpoint.
- Encourage independently operated mirrors; do not publish keys, wallet seeds, passwords, private IP details or credentials.
- For vulnerabilities, follow [SECURITY.md](SECURITY.md). A read-only GitHub archive does not guarantee permanent immutability or prevent independently maintained security releases.
- Keep the [exchange readiness acceptance gate](EXCHANGE-READINESS.md) public and factual until independent evidence exists.

## Administrative final step — only after acceptance

1. Verify required CI and frozen-source checks. Close/merge only legitimate documentation PRs.
2. Verify the final source commit, tags, checksums and outside-GitHub restoration.
3. Publish a short dated final handoff announcement identifying the last official reference version and unresolved risks.
4. From GitHub repository settings, use **Archive this repository** to make the reference repository read-only. The repository owner may still reverse archival later. Archival cannot prevent outside parties from changing and redistributing open-source code.
5. Leave independent full nodes and miners online. The original author may stop participating; the network's survival depends on other operators maintaining the chain.

**Readiness: handoff evidence pending. No claims of completed founder-independent operations as of this document.**
