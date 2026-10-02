# AuronQ Security Policy and Threat Model

AuronQ 1.7.0 is live mainnet software. It is security-sensitive financial infrastructure, but it has **not** received an independent professional security or cryptographic audit. Passing tests is not equivalent to an audit.

## Cryptographic boundary

Transaction authorization uses ML-DSA-87 (FIPS 204) through the bundled portable implementation in the normal Windows/Linux build. Wallet key derivation uses the project's RFC 7914 scrypt implementation. Seed encryption uses Go AES-256-GCM. Transaction/block identifiers use SHA-512/SHA-512/256. AQM64 proof of work uses SHAKE256 and a portable Argon2id v1.3 implementation.

AQM64 is a project-specific consensus construction. Its deterministic vectors and implementation checks reduce accidental incompatibility risk, but they do not establish cryptographic security, ASIC resistance or immunity to time-memory tradeoffs.

## Wallet rules

A wallet's private material is a 32-byte ML-DSA seed. The file stores encrypted seed material plus the public key/address. Default KDF parameters are scrypt `N=131072, r=8, p=1`; authenticated encryption is AES-256-GCM.

Never commit a real wallet, password file, seed, shell history containing a password, or test secret reused in production. Keep at least two offline backups of the encrypted founder wallet and separate backups of its password.

## Main unresolved review areas

Independent reviewers should explicitly test:

- issuance cap and halving boundary conditions;
- UTXO double spends, duplicate inputs and same-block dependencies;
- ML-DSA sighash completeness, encodings and interoperability;
- serialization ambiguity and cross-implementation determinism;
- difficulty/timestamp manipulation and low-hashrate behavior;
- cumulative-work arithmetic and deep reorganizations;
- mempool conflicts and reorg state transitions;
- P2P eclipse, Sybil, amplification, malformed-input and resource-exhaustion attacks;
- disk failure/crash consistency and canonical-chain recovery;
- large-key/signature memory pressure;
- AQM64/Argon2id correctness and cross-implementation agreement;
- AQM64 time-memory tradeoffs and GPU/FPGA/ASIC advantage;
- miner template freshness and stale-block handling.

## P2P assumptions

Peers are untrusted. Height, tip and claimed chain work are hints only. Blocks are validated locally before canonical adoption.

The built-in HTTP layer imposes body bounds, a global concurrent-request bound and basic per-IP request/block-submission rate limits. These controls are not a substitute for upstream DDoS protection, stronger peer scoring/diversity and independent adversarial testing.

Bootstrap services are discovery infrastructure only. They have no consensus authority.

## Local data

Canonical block files are replayed to reconstruct consensus state on node startup. UTXO/state snapshots are acceleration artifacts and are not treated as consensus truth.

## Windows release trust

Current Windows binaries are not Authenticode-signed. Users should download only from the official GitHub Releases page and verify the published SHA-256 checksum before execution.

## Responsible claims

Do not advertise AuronQ as "unhackable", "quantum proof forever", independently audited, formally verified or ASIC-proof.

It is accurate to say that AuronQ uses ML-DSA-87 post-quantum transaction signatures and the AQM64 proof-of-work construction.

The genesis includes a transparent 210,000 AURQ founder allocation (1% of the nominal cap); it must not be presented as a no-allocation/fair-launch equivalent.

Before substantial real-world value depends on AuronQ, obtain independent consensus, cryptography, wallet and P2P review and operate multiple independently controlled public nodes/seeds.
