# AuronQ threat model

## Protected assets

- UTXO ownership and signatures
- total supply and subsidy schedule
- canonical chain selection
- wallet seeds and passwords
- genesis/network identity

## Assumed adversaries

- malicious peers sending malformed blocks/transactions
- miners attempting inflation, double spends or timestamp manipulation
- malware attempting to steal wallet files
- future cryptanalytic advances against an approved signature scheme

## Current controls

- deterministic transaction/block hashing with domain separation
- signature scheme ID committed into addresses, outputs and signature hashes
- ML-DSA-87 ownership signatures
- fixed monetary cap and consensus subsidy calculation
- coinbase maturity
- cumulative-work chain selection
- bounded HTTP bodies and transaction/block sizes
- basic per-IP request and expensive block-submission rate limiting
- validated fork-envelope PoW/work checks before costly canonical reorganization
- encrypted wallet seeds using scrypt + AES-256-GCM
- consensus replay from block files on node startup
- algorithm registry designed for migration to future post-quantum schemes

## Explicit non-goals / remaining work

- resistance to every unknown future cryptanalytic technique cannot be guaranteed
- no claim of formal verification
- no independent audit has yet been performed
- public P2P layer still requires adversarial load/Sybil/DoS testing before production launch
- desktop binaries should be code-signed and reproducibly built before broad distribution
