# AuronQ Mainnet Stabilization and Finalization Gate

AuronQ Mainnet is live. The project is now in a stabilization phase: preserve consensus compatibility, observe the network under independent use, fix real defects conservatively, and avoid feature-driven protocol churn.

This document defines the conditions that should be met before the project is treated as ready for final repository/release hardening.

## Current invariant

The existing Mainnet consensus rules are already frozen for ordinary patch releases under [MAINNET-CHANGE-POLICY.md](MAINNET-CHANGE-POLICY.md).

Do not change genesis, Network ID, monetary policy, AQM64 consensus behavior, difficulty/target rules, block or transaction validity rules, timestamp rules, serialization or cumulative-work chain selection as an ordinary maintenance change.

A consensus-affecting change is exceptional and requires explicit protocol/fork analysis, testnet or devnet validation, compatibility planning and independent review.

## Stabilization gate

Final hardening should wait until all critical items below are satisfied.

### Network operation

- [ ] Mainnet has operated stably for a sustained observation period without unresolved consensus-critical failures.
- [ ] Multiple independently controlled public full nodes remain reachable over time.
- [ ] Peer discovery still works when the original project-operated machines are unavailable.
- [ ] Public health monitoring repeatedly confirms common chain history and tip agreement among healthy peers.
- [ ] Restart, catch-up and reorganization behavior remains stable under real network churn.

### Consensus and implementation confidence

- [ ] No unresolved critical/high-severity consensus or chain-state defect is known.
- [ ] Required Windows/Linux CI remains consistently green.
- [ ] Mainnet identity/frozen-consensus tests remain green.
- [ ] Reproducible-build checks remain green for supported desktop targets.
- [ ] AQM64 independent implementation/review work has progressed enough to reduce single-implementation risk.
- [ ] Independent consensus/cryptography/P2P/security review findings, when available, are addressed or explicitly documented.

### Release and supply-chain integrity

- [ ] Current supported binaries have published SHA-256 checksums.
- [ ] Release build procedure remains documented and reproducible.
- [ ] Production signing/provenance is added where operationally feasible, or the remaining limitation is explicitly documented.
- [ ] Final release tags are protected from deletion or retargeting.
- [ ] Immutable/provenance-preserving release controls are enabled for the final hardened release where supported.

### Operational independence

- [ ] The network can continue operating if the original developer's machines are offline.
- [ ] No privileged founder node, hidden override, remote kill switch or centrally trusted checkpoint is required.
- [ ] Wallets and full nodes validate consensus locally.
- [ ] Bootstrap infrastructure is replaceable and does not determine consensus.
- [ ] At least one practical mining path remains available without relying on a single project-controlled service.

## What is still allowed during stabilization

Changes may continue when they preserve consensus, especially:

- security fixes;
- P2P reliability and peer-diversity improvements;
- monitoring and diagnostics;
- wallet/UI/Explorer fixes;
- packaging, build and signing improvements;
- documentation;
- miner compatibility fixes that do not alter block validity.

Every change should be evaluated for compatibility with existing nodes, wallets, miners and chain history.

## Finalization hardening

When the stabilization gate is satisfied:

1. create a clearly identified final/hardened release and record its source commit and checksums;
2. protect the corresponding release tag against deletion or retargeting;
3. enable the strongest practical immutable/provenance controls for published release artifacts;
4. keep `main` protected by pull requests and required CI;
5. document the final consensus identity and long-term maintenance policy;
6. avoid further protocol changes unless a demonstrable security emergency makes them unavoidable.

Finalization does not mean abandoning security maintenance. It means that compatibility and preservation of the live chain take precedence over feature development.

## Stop rule

If a proposed change is not required for security, compatibility, independence or a demonstrated defect, and it creates any material risk of consensus divergence or network disruption, do not merge it into Mainnet merely because it appears to be an improvement.
