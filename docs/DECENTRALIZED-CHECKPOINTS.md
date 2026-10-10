# AuronQ: decentralized advisory checkpoints (P2P, non-consensus)

AuronQ full nodes are the independent source of blockchain validation. Every node runs the existing AQM64/transaction/UTXO consensus checks and independently reconstructs the canonical chain from genesis when starting. The new **peer checkpoint** mechanism is an optional acceleration *hint and status witness*, not a consensus rule and not a shortcut to accepting historic chainwork or UTXOs.

## How it works

- After each **256 fully validated canonical block heights**, every participating node can derive an advisory checkpoint with header window (64 headers for LWMA/median-time validation), Network ID, genesis hash, checkpoint tip, and exact cumulative work **from its own validated chain**.
- Each node generates its own persistent **Ed25519 P2P identity** locally. A wallet or network operator does not need to provide a signing key. The seed is written to that node's private data directory as `checkpoint-node-identity.seed` (file mode 0600). This key is NOT a wallet key and is NOT a consensus signing key. Back up or rotate a node identity only if needed for stable peer recognition.
- The node produces `peer-checkpoints/latest.json` locally (background best-effort, about once per minute) and serves current canonical signed data from `GET /p2p/checkpoint`. Signing and persistence errors cannot stop mining, node sync or transaction handling.
- No GitHub API, release, Fulcio certificate, Rekor service, pool, founder key, paid hosting or DNS bootstrap manifest is required to derive or verify the peer's checkpoint. A node may advertise it directly over its usual P2P address, and a peer validates the payload signature against the advertised public key.
- Existing synced nodes retain their own chain validation results and local block database during external outages. A checkpoint is not necessary to mine, process blocks or transfer funds.

## Essential security limits

**Peer-issued signatures prove only that the checkpoint was issued by the holder of that peer's signing key.** They do **not** prove that its chain is canonical, nor are two or ten matching peer signatures inherently Sybil-resistant. Wallets MUST NOT directly import unverified peer checkpoint manifests into verified balance, UTXO, history, or spending state. Existing mobile release-pinned checkpoint and local AQM64 header validation remain unchanged; the new API adds an advisory source without silently reducing trust.

A fully trustless, instantly bootstrapping phone would need a separately assessed cryptographic light-client protocol with secure commitments/proofs or independently stored trusted history. That is **not** delivered by self-signed peer snapshots, and should not be added as a covert consensus change.

If GitHub or most nodes disappear, the chain can continue **only as long as at least one validating participant retains a compatible full history and new blocks can be mined, and nodes/wallets can discover a reachable peer**. A node that lost the genesis-to-tip history cannot reconstruct it from an advisory checkpoint with full-node security. A network with no live independent participants cannot continue on its own.

## Compatibility

- No genesis, Network ID, AQM64, mining difficulty, block or transaction format, emission, UTXO consensus, peer-handshake version or wallet private key changes.
- Existing full nodes and mobile wallets can ignore the new endpoint and continue normally.
- Existing v0.5.4 Alpha wallet is deliberately **not** converted into an unsigned/peer-vote trust-first wallet.
- Signed peer checkpoints remain independently available and regenerable from each node even when GitHub disappears.

**Next step to strengthen resilience:** add more geographically and administratively independent publicly reachable full nodes, independent seed discovery (DNS seeds/QR/manual endpoints) and multiple independent mining pools. A hundred nodes behind one administrator or one cloud provider do not provide the resilience of several independently controlled nodes.
