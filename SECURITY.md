# AuronQ Security Policy and Threat Model

AuronQ is security-sensitive financial infrastructure. Passing tests is not equivalent to an independent security audit.

## Cryptographic boundary

Transaction authorization uses ML-DSA-87 (FIPS 204) through the bundled portable implementation in the normal Windows/Linux build. Wallet key derivation uses the project's RFC 7914 scrypt implementation. Seed encryption uses Go's AES-256-GCM implementation. Transaction/block identifiers use SHA-512/SHA-512/256. AQM64 proof of work uses SHAKE256 and a portable Argon2id v1.3 implementation.

The AQM64 composition and its portable Argon2id consensus implementation are **mainnet-capable release-candidate code**. Test vectors and cross-checks are required, but they do not replace an independent implementation/security review before mainnet.

## Wallet rules

A wallet's private material is a 32-byte ML-DSA seed. The file stores only encrypted seed material plus the public key/address. Default KDF parameters are scrypt `N=131072, r=8, p=1`, and authenticated encryption is AES-256-GCM.

Never commit a real wallet, password file, seed, shell history containing a password, or test secret reused in production. Keep at least two offline backups of the encrypted founder wallet and separate backups of its password.

## Consensus attack surfaces that require independent review

Independent auditors should explicitly test:

- issuance cap and halving boundary conditions;
- UTXO double spends, duplicate inputs and same-block dependencies;
- ML-DSA sighash completeness and signature malleability consequences;
- serialization ambiguity and cross-implementation determinism;
- difficulty/timestamp manipulation and low-hashrate launch behavior;
- cumulative-work arithmetic and deep reorganization behavior;
- mempool conflicts and block/mempool state transitions;
- P2P eclipse, amplification, resource-exhaustion and malformed JSON attacks;
- disk failure/crash consistency and canonical-chain recovery;
- large-key/signature memory pressure;
- AQM64 deterministic test vectors, Argon2id v1.3 correctness and cross-implementation agreement;
- AQM64 time-memory tradeoffs, GPU/FPGA/ASIC advantage and denial-of-service cost;
- ML-DSA interoperability and deterministic key derivation from the stored 32-byte seed;
- miner template freshness and stale-block handling.

## P2P assumptions

Peers are untrusted. Height, tip and claimed chain work are hints only. A candidate chain is validated before canonical replacement. P2P transport is not confidential; blocks and transactions are public data.

The built-in HTTP network layer imposes body bounds, a global concurrent-request bound and basic per-IP request/block-submission rate limits. These controls are not a substitute for host firewalling, upstream/connection limits, distributed-DoS protection and independent adversarial review before exposing high-value infrastructure.

## Local data

Canonical block files are replayed to reconstruct consensus state on node startup. UTXO/state snapshots are acceleration artifacts and are not treated as consensus truth.

## Responsible launch

Do not advertise the project as "unhackable", "quantum proof forever", or guaranteed safe. Use "post-quantum signatures" and name the concrete primitive and standard.

Do not call the genesis founder allocation a Bitcoin/Satoshi-style fair launch. It is a transparent 1% founder allocation.

Before real-value mainnet launch, publish the source commit, exact network definition, genesis hash, network ID, founder address/allocation, build instructions, independent audit report and known limitations.
