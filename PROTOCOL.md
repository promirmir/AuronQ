# AuronQ Consensus Protocol

This document specifies the consensus-relevant design implemented by the accompanying source tree. Implementations must reproduce canonical serialization and validation behavior exactly.

## 1. Monetary units and issuance

`1 AURQ = 100,000,000 atoms`.

Maximum issued supply is 2,100,000,000,000,000 atoms (21,000,000 AURQ). Genesis creates exactly 21,000,000,000,000 atoms (210,000 AURQ, 1%) in one founder UTXO. This amount counts toward the cap.

At height 1 the nominal block subsidy is 49.5 AURQ. The 1% reduction from a Bitcoin-style 50 AURQ is deliberate: the missing 1% is the founder allocation already created in genesis, so it is carved out of issuance rather than added on top. Each subsidy level applies for exactly 210,000 mined blocks (heights 1..210,000 for the first epoch) and then halves using integer right shift. Subsidy creation is always capped by remaining unissued supply. Integer halving leaves terminal issuance slightly below the 21,000,000 AURQ maximum. Transaction fees are transferred, not newly issued.

A normal block's coinbase output sum must equal exactly `subsidy + transaction fees`. Coinbase outputs require 100 confirmations before spending. The genesis founder UTXO is an allocation output and is spendable immediately.

## 2. UTXO model

Each spend references `(64-byte transaction ID, uint32 output index)`. An output contains:

- unsigned 64-bit value in atoms;
- uint16 signature-scheme identifier;
- 32-byte hash of the authorized public key.

Scheme `0x0001` is ML-DSA-87. Unknown schemes are rejected by the current consensus implementation. The explicit field reserves clean protocol space for future post-quantum scheme activation.

A normal input reveals its ML-DSA-87 public key and signature. The node hashes the public key with SHA-512/256 and requires equality with the referenced output's key hash.

## 3. Signatures

ML-DSA-87 is the FIPS 204 category-5 parameter set. AuronQ uses Pure ML-DSA semantics. The normal Windows/Linux build uses the bundled portable implementation so consensus behavior does not depend on an operating-system crypto provider. Consensus validation is defined by FIPS 204 behavior rather than a provider API.

The signed message is the 64-byte SHA-512 result of:

```text
"AURONQ_SIGHASH_V1\0"
|| canonical transaction with all signatures empty
|| input_index:u32be
|| referenced_value:u64be
|| referenced_scheme:u16be
|| referenced_key_hash:32
```

Public keys remain present in the signature-less transaction serialization. This commits every signature to all inputs, all public keys, all outputs, lock height and the exact referenced output being authorized.

## 4. Transaction ID

Transaction ID is SHA-512 over canonical transaction serialization including signatures, prefixed internally by the transaction serialization domain `AURONQ_TX\0`.

All integer serialization is fixed-width big-endian. Variable byte arrays use a uint32 big-endian byte length followed by bytes. There is no JSON inside consensus hashing; JSON is transport only.

## 5. Addresses

Human-readable addresses begin with `aurq1` followed by unpadded lowercase RFC 4648 base32 of:

```text
network_byte:1
scheme:u16be
key_hash:32
checksum:6
```

`checksum` is the first six bytes of SHA-512/256 over:

```text
"AURONQ_ADDR_V1" || network_byte || scheme || key_hash
```

Decoders require canonical base32 encoding, preventing alternate textual encodings of the same payload.

## 6. Block header, block ID and proof of work

The canonical version-2 header contains:

```text
"AURONQ_BLOCK_V2\0"
version:u16be
pow_algo:u16be
height:u64be
previous_hash:64
merkle_root:64
timestamp:i64be
target:64
nonce:u64be
```

`pow_algo = 0x0001` selects **AuronQ-PoW v1 (AQM64)**. The block ID is the inexpensive SHA-512 hash of the canonical header; it is used for `previous_hash` linkage and identification. The proof-of-work value is separate and is computed by AQM64.

AQM64 v1 is:

```text
pre  = SHAKE256-512("AURONQ_AQM64_PRE_V1\0" || canonical_header)
salt = SHAKE256-256("AURONQ_AQM64_SALT_V1\0" || previous_hash || height:u64be || pow_algo:u16be)
mid  = Argon2id-v1.3(pre, salt, memory=65536 KiB, time=2, parallelism=1, out=64)
pow  = SHAKE256-512("AURONQ_AQM64_FINAL_V1\0" || pre || mid)
```

A block is valid when the unsigned big-endian 512-bit integer represented by `pow` is less than or equal to the header target. AQM64 deliberately composes standardized primitives instead of defining a new hash primitive. The composition itself is AuronQ-specific, has not yet received independent cryptographic review, and is not claimed to be ASIC-proof.

## 7. Merkle tree

Leaves are transaction IDs. If a level has an odd number of hashes, the final hash is duplicated. Each parent is:

```text
SHA-512("AURONQ_MERKLE_V1" || left:64 || right:64)
```

## 8. Timestamp rules

A new block timestamp must be strictly greater than the median of the previous 11 block timestamps and may not exceed the validating node's adjusted wall clock by more than two hours.

## 9. Difficulty

The consensus target is represented directly as a 512-bit unsigned integer. AuronQ 1.7.0-rc2 freezes the bootstrap proof-of-work limit at 10 leading zero bits for both launch ceremony and network identity. This is a compatibility parameter, not a claim that the calibration is economically optimal.

For height 1, the previous/genesis target is used. Thereafter, target adjustment occurs every block using up to the last 60 solved blocks:

1. Each solve time is `timestamp[i] - timestamp[i-1]`, clamped to `[1 second, 6 * 600 seconds]`.
2. Solve times receive increasing linear weights, with the newest interval receiving the largest weight.
3. Targets in the same window are arithmetically averaged.
4. `next_target = average_target * weighted_average_solve_time / 600`.
5. The result is clamped to one quarter through four times the previous target.
6. Target never exceeds the launch PoW limit and never reaches zero.

This is LWMA-inspired rather than Bitcoin's 2016-block retarget because a newly launched network can otherwise spend days at a badly calibrated initial difficulty.

## 10. Chain selection and reorganization

Per-block work is:

```text
floor(2^512 / (target + 1))
```

Canonical chain selection uses the greatest sum of per-block work. A peer-advertised chain is never trusted from its claimed height/work alone. The node retrieves the branch, validates every candidate block and reconstructs its UTXO state before replacing the canonical chain.

## 11. Network identity

Network ID is not merely the genesis hash. It is SHA-512 over a domain tag, genesis hash and a version-3 consensus fingerprint containing monetary constants, maturity, target timing, difficulty window, maximum block size, PoW limit, AQM64 algorithm identifier/parameters and the initial signature-scheme identifier.

This prevents nodes compiled with materially different fixed consensus parameters from silently presenting the same network identity.

## 12. Genesis

Genesis is mined and contains exactly one allocation transaction with exactly one output of 210,000 AURQ to the founder address. The network definition stores the public founder address and validation requires that its network byte, scheme and key hash match the genesis output.

The final mainnet genesis is intentionally not hardcoded in the source archive: the owner must generate the founder ML-DSA seed locally. The resulting `network.json` becomes the immutable public network definition after the launch ceremony.

## 13. Limits

- maximum canonical block body: 4 MiB;
- maximum canonical transaction: 128 KiB;
- maximum transaction inputs: 4096;
- maximum transaction outputs: 4096;
- minimum relay fee: 1,000 atoms per started KiB;
- public key witness bound: 4096 bytes;
- signature witness bound: 8192 bytes.

The current ML-DSA-87 sizes are 2592-byte public keys and 4627-byte signatures.

## 14. Cryptographic scope

SHA-512, SHA-512/256 and SHAKE256 are symmetric hash/XOF constructions and do not rely on the discrete-log problem threatened by Shor's algorithm. ML-DSA-87 is intended for a post-quantum signature threat model. Argon2id is used here for memory-hard proof-of-work cost, not for transaction signatures. This protocol does not claim mathematical immunity to unknown future cryptanalysis or unknown future computing models.


## Coinbase maturity
Mainnet uses 100 blocks. The built-in testnet profile uses 10 blocks to speed up testing. The value is committed into the network configuration and Network ID.
