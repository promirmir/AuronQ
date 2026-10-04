# AuronQ Mainnet Change Policy

AuronQ Mainnet is now a live network used by independent participants. Mainnet consensus must therefore be treated as production protocol infrastructure, not as an ordinary application feature surface.

## 1. Consensus freeze

The existing Mainnet Network ID, genesis block, monetary policy, AQM64 parameters, block/transaction serialization, validity rules, difficulty rules, timestamp rules and cumulative-work chain-selection semantics are frozen for ordinary patch releases.

Bug fixes in UI, peer discovery, logging, packaging, monitoring and non-consensus networking may ship without a fork when they do not alter which blocks or transactions are considered valid.

## 2. No silent consensus changes

A change that can cause two honest nodes on the same Mainnet history to disagree about block or transaction validity must not be merged and released as a normal patch.

Such a change requires:

1. a written protocol change;
2. dedicated testnet/devnet validation;
3. explicit compatibility and fork analysis;
4. a versioned activation mechanism and activation point if Mainnet adoption is ever approved;
5. independent technical review before Mainnet activation.

## 3. Release discipline

Official releases must:

- come from the current protected `main` commit;
- pass required Windows and Linux CI;
- pass the frozen Mainnet identity test;
- match the version encoded in both CLI and Desktop sources;
- publish archive SHA-256 checksums;
- preserve the published Network ID and genesis unless a deliberately planned new network is being created.

The release workflow is expected to fail if a tag points at a commit other than current protected `main` or if the source version does not match the tag.

## 4. Mainnet safety principles

AuronQ must not add a hidden founder override, emergency mint, privileged reorganization command, remote kill switch or centrally signed checkpoint that can silently override normal cumulative-work consensus.

Operational recovery should rely on ordinary software upgrades, transparent protocol rules and independently validating full nodes.

## 5. Experimental work

Risky consensus experiments belong on testnet/devnet first. Mainnet should favor stability over feature velocity.

Examples include changes to:

- AQM64 or its parameters;
- difficulty adjustment;
- block timestamps;
- transaction sighash/serialization;
- ML-DSA validation semantics;
- subsidy/halving/cap rules;
- reorganization/chain-work rules;
- block/transaction size limits that affect validity.

## 6. Security reality

These process controls reduce accidental chain splits and unsafe releases. They do not prove that AuronQ is secure. Independent consensus, cryptographic, wallet and P2P review remains necessary before substantial real-world value depends on the network.
