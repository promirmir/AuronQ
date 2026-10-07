# AuronQ (AURQ) Technical Whitepaper

**Version 1.0 — October 2026**

AuronQ is an open-source public Proof-of-Work cryptocurrency network built around a UTXO ledger, AQM64 proof of work, ML-DSA-87 transaction signatures, independently validating full nodes and a deliberately conservative Mainnet change policy.

This document is a technical whitepaper describing the implemented AuronQ Mainnet. It is not an investment prospectus, does not promise financial return, and is not a substitute for independent legal, regulatory, security or cryptographic review.

---

## 1. Abstract

AuronQ is designed as a public Layer-1 cryptocurrency network in which validity is determined by independently operated full nodes rather than by a privileged server, explorer, founder node or hosted API.

The network combines:

- a **UTXO ledger**;
- **Proof-of-Work** consensus;
- the AuronQ-specific **AQM64** memory-hard proof-of-work construction;
- **ML-DSA-87** post-quantum transaction signatures;
- cumulative-work chain selection;
- replaceable peer discovery and locally validated full-node state;
- a fixed Mainnet identity committed to its genesis and consensus fingerprint.

The project prioritizes **stability, independent verification and long-term compatibility** over frequent consensus changes. AuronQ Mainnet is live, but remains young infrastructure and has not yet received an independent professional security or cryptographic audit.

---

## 2. Design objectives

AuronQ was developed around the following objectives:

1. **Independent validation** — every full node should be capable of validating the blockchain without trusting a central explorer or API.
2. **Public verifiability** — source code, consensus parameters, Network ID, genesis hash and protocol documentation are public.
3. **Stable Mainnet identity** — genesis, Network ID, monetary rules and consensus-critical behavior should not be casually changed after launch.
4. **Post-quantum transaction signatures** — transaction authorization uses ML-DSA-87, standardized by NIST in FIPS 204.
5. **Resource-intensive Proof-of-Work** — AQM64 combines SHAKE256 with Argon2id so mining attempts require both computation and memory.
6. **Replaceable infrastructure** — bootstrap endpoints, explorers, miners and pools are not consensus authorities and may be replaced independently.
7. **Permissionless participation** — independent users may run nodes, mine, build explorers, wallets, pools, indexers or other compatible services.

AuronQ does **not** claim to be mathematically future-proof, ASIC-proof, unhackable, formally verified or independently audited.

---

## 3. Mainnet identity

The AuronQ Mainnet has a fixed public identity.

| Parameter | Value |
|---|---|
| Project | AuronQ |
| Symbol | AURQ |
| Ledger model | UTXO |
| Consensus | Proof-of-Work |
| Proof-of-Work | AQM64 v1 |
| Transaction signature scheme | ML-DSA-87 |
| Target block interval | 600 seconds |
| Default P2P port | TCP 18444 |
| Coinbase maturity | 100 blocks |
| Nominal supply cap | 21,000,000 AURQ |
| Genesis allocation | 210,000 AURQ |
| Initial block subsidy | 49.5 AURQ |
| Launch release | v1.7.0 |

**Network ID**

`44e62c2ace002a6660c14e252173c1aa303529c68e40c998e92da2b453f44f30b1e58c94d533587e2186004593fb856c433fcdb5418ed430ec8617e29529365c`

**Genesis hash**

`5750a455c04bfe93c9edfef1a12744b05e29ac6da1a9dd5b790566629dea2080c581beb2f0324efba2067c9efb113ed7a29598fd3ffbed965ced31f265d0cec4`

The Network ID is not merely the genesis hash. It commits to a consensus fingerprint that includes monetary parameters, maturity rules, timing, difficulty behavior, block limits, PoW parameters and the initial signature scheme.

Canonical reference: [MAINNET.md](MAINNET.md).

---

## 4. Ledger model

AuronQ uses a **UTXO (Unspent Transaction Output)** ledger.

Each spend references:

- a 64-byte transaction ID;
- a 32-bit output index.

Each output contains:

- a value denominated in atoms;
- a signature-scheme identifier;
- a 32-byte hash of the authorized public key.

The base monetary unit is:

`1 AURQ = 100,000,000 atoms`

Nodes independently verify that:

- referenced outputs exist and remain unspent;
- signatures authorize the referenced outputs;
- transaction values are valid;
- duplicate spends are rejected;
- fees and coinbase issuance follow consensus rules.

Transaction IDs are SHA-512 over canonical transaction serialization with explicit domain separation.

Canonical protocol details: [PROTOCOL.md](PROTOCOL.md).

---

## 5. Transaction signatures

AuronQ Mainnet uses **ML-DSA-87**, the category-5 parameter set standardized in NIST FIPS 204.

A normal transaction input reveals its public key and signature. The validating node hashes the public key using SHA-512/256 and verifies that it matches the key hash committed by the referenced UTXO.

The signature commits to:

- the canonical unsigned transaction;
- the input index;
- the value of the referenced output;
- its signature-scheme identifier;
- its authorized public-key hash.

This prevents a valid signature from being detached from the exact UTXO context it authorizes.

Use of ML-DSA-87 is intended to address the threat posed by large-scale quantum attacks against classical public-key signature systems. It does not imply that the entire AuronQ system is automatically quantum-secure against all future attacks or implementation failures.

---

## 6. AQM64 Proof-of-Work

AuronQ uses **AQM64 v1**, an AuronQ-specific proof-of-work construction composed from standardized primitives.

For each candidate header:

```text
pre  = SHAKE256-512("AURONQ_AQM64_PRE_V1\0" || canonical_header)
salt = SHAKE256-256("AURONQ_AQM64_SALT_V1\0" || previous_hash || height || pow_algo)
mid  = Argon2id-v1.3(pre, salt, memory=65536 KiB, time=2, parallelism=1, out=64)
pow  = SHAKE256-512("AURONQ_AQM64_FINAL_V1\0" || pre || mid)
```

Consensus parameters include:

- Argon2id v1.3;
- 65,536 KiB memory per mining lane;
- time cost 2;
- parallelism 1;
- 64-byte Argon2 output;
- 512-bit final proof.

A block is valid when the unsigned big-endian integer represented by the AQM64 result is less than or equal to the target committed in the block header.

AQM64 was designed to make each nonce attempt materially memory-intensive rather than a simple high-throughput hash loop. It is **not claimed to be ASIC-proof**. GPU, FPGA, ASIC and time-memory-tradeoff analysis remain legitimate areas for independent research.

Technical reference: [AQM64.md](AQM64.md).

---

## 7. Block structure and chain selection

The canonical block header includes:

- protocol version;
- PoW algorithm identifier;
- block height;
- previous block hash;
- Merkle root;
- timestamp;
- target;
- nonce.

The block identifier is SHA-512 over the canonical block header. The AQM64 proof value is computed separately.

AuronQ selects the canonical chain by **greatest cumulative proof of work**, not by advertised height alone.

Per-block work is defined as:

`floor(2^512 / (target + 1))`

A peer cannot cause a full node to accept a chain merely by claiming a high height or chain-work value. Candidate blocks must be downloaded and locally validated before a competing branch can become canonical.

---

## 8. Difficulty adjustment

AuronQ targets one block every **600 seconds**.

Difficulty adjusts every block using a rolling window of up to 60 solved blocks. Solve times are bounded, weighted toward more recent intervals and combined with historical targets.

The next target is constrained so that a single adjustment cannot exceed a factor of four in either direction relative to the previous target.

This faster adjustment model was chosen for a young network where hashrate may change rapidly and a Bitcoin-style 2016-block retarget would respond too slowly.

The launch proof-of-work limit is part of Mainnet compatibility and must not be treated as an independently optimized economic calibration.

---

## 9. Monetary policy

AuronQ has a nominal maximum supply of:

**21,000,000 AURQ**

Genesis created:

**210,000 AURQ (1% of the nominal cap)**

in one founder UTXO.

The initial block subsidy is:

**49.5 AURQ**

The reduction from a Bitcoin-style 50-unit initial subsidy is deliberate: the 1% genesis allocation is carved out of issuance rather than added on top of the nominal cap.

Each subsidy level applies for 210,000 mined blocks and then halves using integer right shift. Transaction fees are transfers and do not create new supply.

Coinbase outputs require 100 confirmations before they can be spent.

Integer halving means terminal issuance is expected to finish slightly below the nominal maximum rather than exceeding it.

---

## 10. Networking and decentralization model

AuronQ follows a **full-node-first** architecture.

A full node:

- stores the canonical chain locally;
- validates blocks, transactions, signatures, timestamps and difficulty;
- reconstructs and validates UTXO state;
- maintains its own mempool;
- discovers and persists peers;
- relays valid transactions and blocks;
- can serve its own local read-only blockchain explorer.

Bootstrap endpoints are **rendezvous hints only**. They cannot:

- authorize transactions;
- approve invalid blocks;
- select the canonical chain;
- change monetary policy;
- override local consensus validation.

A fresh installation requires some initial discovery path because it cannot discover unknown peer addresses from nothing. After joining the network, peers may be learned through native P2P gossip and persisted locally.

Long-term decentralization depends not only on code design, but on real independent operation: more unrelated public nodes, miners, pools, discovery routes and explorers.

See [DECENTRALIZATION.md](DECENTRALIZATION.md) and [NETWORK-INDEPENDENCE.md](NETWORK-INDEPENDENCE.md).

---

## 11. Wallets and addresses

Human-readable AuronQ addresses begin with:

`aurq1`

The encoded payload contains:

- a network byte;
- signature-scheme identifier;
- public-key hash;
- checksum.

Wallet private material is based on a 32-byte ML-DSA seed.

The current wallet implementation encrypts seed material using AES-256-GCM with a key derived by scrypt. Users remain responsible for password strength, secure backups and protecting private material from malware or physical compromise.

The Android application is a light wallet rather than a full node. It verifies the Mainnet header chain locally and validates AQM64 proof of work, difficulty, timestamps and hash continuity, while relying on agreeing verified-chain peers for wallet-state information.

---

## 12. Software ecosystem

The AuronQ project publishes multiple components:

### Full node / Desktop

The full node provides:

- Mainnet synchronization;
- local blockchain validation;
- wallet functionality;
- transaction relay;
- built-in explorer;
- mining interfaces.

### Universal Miner

The official mining software supports CPU mining and multiple accelerator paths. Third-party mining software may also implement AQM64 independently.

### Mining pools

AURQ may be mined through independently operated pools. Pools do not participate in consensus authority; submitted blocks remain subject to ordinary full-node validation.

### Explorers and indexers

Every full node can expose its own local explorer. Third-party explorers and indexers may independently interpret the public chain but do not define consensus truth.

### Third-party integrations

Exchanges, wallets, pools, analytics providers and infrastructure operators may integrate the public network without requiring prior project approval, subject to their own technical, legal and security review.

---

## 13. Mainnet change policy

AuronQ Mainnet is intended to be a persistent network rather than a disposable test chain.

The project therefore follows a conservative change philosophy:

- stability before new features;
- preserve genesis and Network ID;
- avoid unnecessary changes to consensus;
- maintain backward compatibility wherever possible;
- prefer fixes outside consensus when an issue can be solved without altering chain rules;
- require exceptional scrutiny for any change capable of splitting the network.

Consensus changes should be treated as extraordinary events because they can affect existing nodes, wallets, miners, exchanges, pools and historical chain validity.

See [MAINNET-CHANGE-POLICY.md](MAINNET-CHANGE-POLICY.md) and [STABILIZATION.md](STABILIZATION.md).

---

## 14. Security model and known limitations

AuronQ Mainnet is live financial software, but it is still a young network.

The following facts are important:

- AQM64 and the complete consensus implementation have **not** received an independent professional cryptographic/security audit.
- ML-DSA-87 is standardized, but correct use of a standardized primitive does not prove that the complete application is secure.
- AQM64 is project-specific and requires continued independent analysis.
- A small or concentrated mining network may be economically vulnerable to majority-hashrate attacks or deep reorganizations.
- P2P systems remain exposed to Sybil, eclipse, denial-of-service and malformed-input risks.
- Wallet users remain exposed to endpoint compromise, malware, password loss and backup failure.
- Software bugs can exist even when automated tests pass.
- Current Windows binaries may trigger SmartScreen because they are not presently Authenticode-signed.

The project explicitly discourages claims that AuronQ is unhackable, permanently quantum-proof, formally verified or ASIC-proof.

Security policy: [SECURITY.md](SECURITY.md).

---

## 15. Open source and independent verification

AuronQ source code is public under the **MIT License**.

The public repository is intended to allow independent parties to:

- inspect consensus logic;
- reproduce builds;
- compare protocol behavior;
- run independent nodes;
- implement miners;
- build external wallets and explorers;
- verify network identity;
- perform security review.

Open source increases transparency and auditability, but it does not itself guarantee correctness or security.

Repository:

https://github.com/promirmir/AuronQ

---

## 16. Governance and operational independence

AuronQ does not use an on-chain governance token or privileged validator committee.

Consensus behavior is determined by the software independently chosen and executed by network participants.

The GitHub repository and project maintainers can publish code and releases, but they cannot directly force independently operated nodes to accept invalid blocks or altered consensus rules.

This distinction is fundamental: source-code stewardship and network consensus are related operationally, but they are not the same authority.

As the ecosystem matures, resilience improves when:

- several independent parties operate reachable nodes;
- mining power is distributed;
- multiple pools exist;
- discovery infrastructure is diversified;
- third-party implementations and monitoring tools appear;
- users can independently verify releases and network state.

---

## 17. Integration guidance

A third-party service integrating AURQ should independently verify at least:

1. Network ID and genesis hash;
2. source repository and release provenance;
3. full-node synchronization;
4. transaction serialization and address handling;
5. deposit/withdrawal behavior;
6. coinbase maturity;
7. reorganization handling;
8. confirmation policy appropriate to observed network security;
9. custody and key-management design;
10. operational monitoring and incident response.

AuronQ is a native Layer-1 asset, not an ERC-20 or similar smart-contract token. Exchanges must therefore integrate the blockchain itself rather than only adding a token contract.

---

## 18. Development direction

The immediate direction is intentionally conservative.

Primary priorities are:

- accumulate a longer stable Mainnet operating history;
- grow independently operated nodes and mining infrastructure;
- improve discovery diversity;
- continue adversarial P2P and reorganization testing;
- expand real-hardware miner testing;
- reduce light-client trust further;
- obtain independent professional review;
- improve release signing and distribution integrity.

These goals concern reliability and decentralization rather than promises of token price, market capitalization or financial return.

---

## 19. Risks

Participation in a young cryptocurrency network involves material risk.

Examples include:

- software defects;
- cryptographic implementation errors;
- chain reorganizations;
- concentrated mining power;
- network outages;
- lost or compromised keys;
- incompatible third-party integrations;
- regulatory restrictions;
- insufficient liquidity;
- project abandonment;
- economic attacks;
- hardware or storage failure.

Anyone operating infrastructure or assigning financial value to AURQ should perform independent due diligence.

---

## 20. Public references

- Repository: https://github.com/promirmir/AuronQ
- Releases: https://github.com/promirmir/AuronQ/releases
- Mainnet identity: [MAINNET.md](MAINNET.md)
- Consensus protocol: [PROTOCOL.md](PROTOCOL.md)
- AQM64 specification: [AQM64.md](AQM64.md)
- Decentralization model: [DECENTRALIZATION.md](DECENTRALIZATION.md)
- Network independence: [NETWORK-INDEPENDENCE.md](NETWORK-INDEPENDENCE.md)
- Security policy: [SECURITY.md](SECURITY.md)
- Mainnet change policy: [MAINNET-CHANGE-POLICY.md](MAINNET-CHANGE-POLICY.md)
- Stabilization criteria: [STABILIZATION.md](STABILIZATION.md)
- Roadmap: [ROADMAP.md](ROADMAP.md)
- Project site: https://promirmir.github.io/AuronQ/

---

## 21. Status of this document

This whitepaper documents the implemented AuronQ Mainnet as of October 2026.

Where this document and byte-exact consensus documentation differ, the canonical protocol specification, Mainnet definition and actual consensus implementation take precedence.

Future revisions of the whitepaper should preserve historical versioning and clearly distinguish implemented behavior from proposed changes.

---

## Disclaimer

AURQ is experimental digital infrastructure. This document is provided for technical information and transparency. It does not constitute investment, financial, legal, tax or regulatory advice, does not guarantee listing by any exchange, and does not promise that AURQ will have or retain monetary value.
