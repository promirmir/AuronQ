# AuronQ Mobile

## Download

**Current recommended test build: AuronQ Mobile 0.5.6 Alpha — Fast Sync**

- [Download Android APK](https://github.com/promirmir/AuronQ/releases/download/android-v0.5.6-alpha-fast-sync/AuronQ-Mobile-0.5.6-Alpha-Fast-Sync.apk)
- [Release notes and SHA-256 checksum](https://github.com/promirmir/AuronQ/releases/tag/android-v0.5.6-alpha-fast-sync)

Faster wallet synchronization while preserving wallet creation, receiving, sending, and transaction history, including pending transactions. This is an **experimental Android alpha**, not an independently audited financial application. The Fast Sync build uses a separate app identity and does not upgrade or migrate existing wallet files. Never uninstall an existing funded wallet or erase its data to install a different APK. Keep secure offline recovery material; do not test with valuable funds.

## Technical notes and limitations

## No private checkpoint keys to configure

Every **updated AuronQ full node**, regardless of operator or hosting provider, autonomously derives an advisory checkpoint every 256 blocks from its own fully validated history and signs it with a locally generated Ed25519 **node-identity key**, not an AuronQ wallet key.

- It stores `peer-checkpoints/latest.json` alongside the node's full chain.
- It makes the current checkpoint available at `GET /p2p/checkpoint`.
- It needs no GitHub access, CI runners, Sigstore, founder-owned infrastructure, manually entered secrets or trusted remote checkpoint signer.
- The mobile app probes reachable node endpoints in the background and verifies node signatures, Mainnet Network ID and genesis. It shows observed signatures and peer groups as network diagnostics.
- The phone does **not** promote self-signed peer observations into authenticated chain history, wallet balance, UTXO state, or a spending authorization. Those require its existing verified-header and peer-consistency rules.
- The previous GitHub-based signed-checkpoint update *call* is disabled in the new Android build. Its saved header cache and embedded release checkpoint continue to work without GitHub.

## Why signatures from several peers are not automatically trusted

A peer signature proves only that a given peer identity signed the claim. An attacker can create many identities (Sybil attack). Even two or more different IP ranges cannot guarantee independent operators. Accepting their checkpoint as a fully authenticated historical chain without independently verifiable history could allow a fake chain. Therefore signed P2P checkpoints are deliberately **advisory** until a stronger, audited lightweight proof mechanism is available.

This is **not a promise of instantly trustless first sync** for a brand-new phone years into the chain's history. Historical checkpoint trust still depends on a release anchor or an independently checked chain, while incremental AQM64 validation remains mandatory. Do not disable full-node consensus validation to accelerate wallet bootstrapping.

## Keeping the blockchain alive without GitHub

The P2P checkpoint endpoint, node block database, peer gossip, locally stored peers and independent full-node validators do not require GitHub once installed. You still need at least one genuinely reachable validating node with complete history and miners producing new blocks. If all independent full nodes go offline or all chain backups are lost, signatures and checkpoints cannot recreate missing blocks or keep the network alive.

Full technical explanation: [Decentralized advisory checkpoints](../docs/DECENTRALIZED-CHECKPOINTS.md).

[Mainnet information](../MAINNET.md) · [Security](../SECURITY.md) · [Report a problem](https://github.com/promirmir/AuronQ/issues)
