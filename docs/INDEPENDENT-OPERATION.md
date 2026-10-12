# AuronQ — independent operation and maintainer handoff

**State: work in progress, 12 October 2026.** This is a technical and governance handoff checklist. It does **not** certify that AuronQ currently survives a complete loss of the original operator, that third parties independently control the network, or that the code is professionally audited.

## What can stay unchanged

The existing public Mainnet network ID, genesis, PoW, block and transaction validity, issuance, difficulty and chain-selection rules remain **frozen for ordinary maintenance** under [Mainnet change policy](../MAINNET-CHANGE-POLICY.md). No privileged founder node, unilateral mint function, hidden chain override or central permission for running a full node should be introduced.

Independent nodes choose the consensus software they run. Neither a GitHub account nor a repository policy can technically prevent unrelated people from publishing forks. A live PoW chain may reorganize if another valid chain has more cumulative work. An archival Git tag also cannot cryptographically stop others from distributing modified source.

## Independent operator acceptance test

Each check requires recorded evidence. **Unchecked is not passed.**

- [ ] At least **three independently controlled**, publicly reachable full nodes run on genuinely different networks/providers; confirm with the operators, not merely with IP count.
- [ ] Each node validates the **same canonical Network ID**, genesis, history, cumulative work and chain tip independently.
- [ ] At least two unrelated mining operators/pools can obtain templates and propagate valid blocks without the original project's machines.
- [ ] Public bootstrap/discovery supplies multiple non-founder paths. A first-time installation with an empty peer cache can find them.
- [ ] With the original project's machines and GitHub-based discovery temporarily unavailable, a fresh install still reaches the independently validated tip.
- [ ] During the offline drill, an independently mined valid block propagates and is validated by all participating nodes.
- [ ] After simulated peer loss, restart, divergent tips and recovery, the nodes safely converge under cumulative-work consensus.
- [ ] Source, canonical network definitions, full-chain data and verified release binaries have independent off-GitHub copies. A separate operator demonstrates restoration from such a copy.
- [ ] Mobile and desktop wallet safety blockers have been resolved or formally documented with acceptable mitigations **before** any production-custody assurance claim.
- [ ] Independent reviewers examine consensus, AQM64 cryptography, wallet storage/signing and networking; publish scope and findings honestly.

**Stop rule:** if a new node cannot discover peers or verified nodes disagree about canonical history, the independence test fails. Restore original availability as needed and record the defect; do not change genesis or insert a trusted checkpoint as a shortcut.

See [NETWORK-INDEPENDENCE.md](../NETWORK-INDEPENDENCE.md) and [STABILIZATION.md](../STABILIZATION.md). Do not deliberately take offline the **only** remaining reachable seed or archival copy.

## Ongoing operations without the original author

A third party should be able to:

1. Download or restore an archived snapshot of the open-source repository and release checksums, without an administrator sharing personal credentials.
2. Build a node from the public source and independently verify the Mainnet identity.
3. Obtain or independently distribute peer addresses and synchronize from independently controlled full nodes.
4. Join mining or build a compatible pool without special approval.
5. Diagnose faults using local node logs, chain validation and publicly reproducible tests, without relying on any official explorer.
6. Review public issues, propose compatible fixes and distribute its own build; existing operators independently decide whether to install it.

**Release governance:** A reference repository owner may protect branches/tags and publish immutable release artifacts, but administrators retain control over repository settings, and an archived repository can still be unarchived. For long-term survival, distribute source and chain copies among independent operators, not just on the original author's GitHub account.

No one may claim to act as the project or bind the original operator to listing, legal, liquidity or token-transfer commitments without a separate lawful authorization. Exchanges are free to assess protocol integration independently, subject to their own legal and security duties.

## Maintenance and security exception

**No routine Mainnet consensus changes.** Future contributors may submit documented non-consensus fixes for data loss, security, node discovery, wallet interoperability or packaging. Require reproducible evidence, explicit compatibility analysis, tests, independent review when available and a staged rollback plan. Consensus-affecting proposals must follow the exceptional protocol/fork process rather than silently replacing Mainnet.

A permanent refusal to patch known wallet vulnerabilities would increase risk for users; that is why this handoff is a **maintenance freeze**, not a promise that no byte of any future implementation can change.

## Evidence log

When an operator performs a drill, add a dated record stating the node software/version, independent operator count, network/provider diversity (without publishing personal data), starting and ending height/tip hashes and chain work, offline endpoints, recovery results and reproducible logs with secrets redacted. Until such evidence exists, describe network independence as a **goal**, not an accomplished fact.

**Never upload** `founder.wallet`, passwords, recovery seeds, private keys, private network credentials or users' personal data.
