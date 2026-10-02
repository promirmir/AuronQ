# AuronQ Release Candidate

AuronQ is a Bitcoin-inspired UTXO/Proof-of-Work monetary network with post-quantum signatures from genesis and an algorithm-agnostic transaction format.

## Design goal

The network is deliberately conservative: fixed maximum monetary supply, transparent genesis allocation, permissionless PoW, no ICO, no smart-contract VM, and no admin key. Its main differentiator is long-lived cryptographic migration: every address and UTXO carries a signature-scheme identifier and signatures are verified through a consensus registry.

The initial consensus scheme is ML-DSA-87. Future post-quantum schemes can be added by a consensus software upgrade while retaining the same transaction/address serialization. Existing holders then move coins from old-scheme UTXOs to new-scheme UTXOs. No hidden key can move funds and no central party can silently activate an algorithm.

## Monetary policy

- Maximum supply: 21,000,000 AURQ
- Genesis founder allocation: 210,000 AURQ (1%), publicly visible
- Initial mining subsidy: 49.5 AURQ
- Halving interval: 210,000 blocks
- Target block interval: 600 seconds
- Coinbase maturity: 100 blocks

## Release gate

The code is suitable for public testnet deployment. Do **not** represent it as audited production financial infrastructure until an independent consensus/cryptography review, adversarial testnet, fuzzing campaign, and reproducible-build review have completed.

Recommended mainnet gate:

1. Public source repository and tagged testnet release.
2. At least three independently operated seed nodes.
3. Multi-week adversarial public testnet.
4. Independent review of consensus, P2P, wallet encryption and portable ML-DSA implementation.
5. Public issue tracker and vulnerability reporting channel.
6. Freeze consensus parameters and publish genesis ceremony instructions before mainnet genesis.
7. Generate the founder wallet only on the founder's offline/trusted machine.
8. Publish genesis hash, Network ID, source tag and release hashes together.

## Crypto-agility rule

Adding a new signature scheme is intentionally a consensus event. A future implementation must assign a new immutable scheme ID, define exact key/signature encodings and verification rules, add test vectors, and activate it through a publicly reviewed software release. Reusing or redefining an existing scheme ID is forbidden.

The repository-level `TEST-MATRIX.md` is the working acceptance checklist. Mainnet should not be launched while any critical/high-severity item remains unresolved.

## 1.7.0-rc2 mainnet-capable release

1.7.0-rc2 keeps block header version 2 and AQM64 (Argon2id v1.3 64 MiB / t=2 / p=1 + SHAKE256-512), and enables `init-mainnet` with an explicit `--ack-unaudited-mainnet` acknowledgement. rc2 additionally hardens fork synchronization against fabricated-work replay DoS, adds per-IP request/block throttling, expands Network ID commitment, hardens genesis validation, fixes the optional OpenSSL crypto build path, and supports bundled `network.json` first-run Desktop onboarding. A final mainnet must use a newly generated immutable genesis/network definition created by this frozen consensus version. See `AUDIT-2026-10-02.md`.