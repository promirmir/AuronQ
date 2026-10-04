# AuronQ Roadmap

This roadmap lists engineering priorities, not promises or investment claims. Security and decentralization take priority over feature count.

## Delivered

- Live Mainnet with fixed Network ID and genesis.
- Windows/Linux full node and Desktop wallet.
- ML-DSA-87 transaction signing.
- AQM64 Proof-of-Work mining.
- UTXO validation and cumulative-work reorganization.
- Persistent P2P peer storage and gossip.
- Replaceable/prunable bootstrap peers.
- Native local blockchain Explorer on every full node.
- Peer-aware public Explorer links.
- 20-node partition/fork/restart convergence testing.
- Eclipse/Sybil concentration hardening through peer netgroups.
- Transaction/block fuzzing and race-detector CI.
- Reproducible-build checks.
- Scheduled public-network health monitoring.
- Founder-bootstrap-removal regression testing.
- Android light wallet with local header/AQM64 verification and verified-chain wallet-state quorum.

## Near-term priorities

### Network decentralization
- Bring additional independently operated public full nodes online.
- Add independent DNS-seed operators and additional first-contact routes.
- Measure peer diversity by operator/network rather than raw peer count.
- Continue failure testing with the original project machines absent.

### Security hardening
- Extend adversarial P2P, eclipse, Sybil and resource-exhaustion tests.
- Expand long-running chaos/stress tests.
- Continue fuzzing consensus/network parsers and state transitions.
- Review integer bounds, storage corruption recovery and reorg edge cases.

### Mobile trust minimization
- Reduce reliance on remote full-node UTXO responses further.
- Evaluate compact proofs / authenticated state mechanisms that fit AuronQ's model.
- Strengthen peer diversity and reorg handling on mobile.
- Move from debug-signed alpha packaging toward production release signing when the software is ready.

### Release integrity
- Preserve deterministic/reproducible builds.
- Add stronger artifact provenance/attestation.
- Introduce production code signing for supported platforms when operationally feasible.

## External-review gate

Before presenting AuronQ as mature financial infrastructure:

- obtain independent review of AQM64;
- obtain independent consensus/network/security review;
- address findings publicly;
- operate for an extended period under independent nodes/miners;
- demonstrate recovery from failures and hostile peer behavior.

## Longer-term

Potential future work should be evaluated only if it preserves the project's decentralization invariants:

- more efficient initial block download;
- stronger light-client proofs;
- broader platform support;
- tooling for independent node operators;
- exchange/integration documentation after network/security maturity;
- protocol improvements only through explicit versioning and public review.

See [DECENTRALIZATION.md](DECENTRALIZATION.md), [THREAT-MODEL.md](THREAT-MODEL.md) and [SECURITY.md](SECURITY.md).
