# AuronQ Mobile — decentralized P2P checkpoint diagnostics

## Download / installation

## User-tested fast-sync APK (separate experimental installation)

**AuronQ Mobile 0.5.3 Two-Stage V2 Mempool TEST** — [Download APK](https://github.com/promirmir/AuronQ/releases/download/android-v0.5.3-two-stage-v2-mempool-test-20261010/AuronQ-Mobile-0.5.3-Two-Stage-V2-Mempool-TEST.apk) · [Release with SHA-256 file](https://github.com/promirmir/AuronQ/releases/tag/android-v0.5.3-two-stage-v2-mempool-test-20261010).

The project operator reports faster startup and working wallet functions. This build restores display of unconfirmed mempool transactions in the history feed. It is **not independently audited** and does not prove correctness of remote UTXO data; 3 distinct IP network groups do not prove 3 independent owners. Use this **isolated test package** only with nonvaluable test wallets. Do not uninstall the existing wallet or clear application data.


**Current experimental release: v0.5.5 Alpha P2P.**

- [Download standard Android APK](https://github.com/promirmir/AuronQ/releases/download/android-v0.5.5-alpha-p2p/AuronQ-Mobile-0.5.5-alpha-p2p.apk)
- [Download isolated Android test APK](https://github.com/promirmir/AuronQ/releases/download/android-v0.5.5-alpha-p2p/AuronQ-Mobile-0.5.5-alpha-p2p-Isolated-Test.apk)
- [Release information and SHA-256 checksums](https://github.com/promirmir/AuronQ/releases/tag/android-v0.5.5-alpha-p2p)
- [Previous v0.5.4 Alpha release](https://github.com/promirmir/AuronQ/releases/tag/android-v0.5.4-alpha)

SHA-256:
- Standard: `32ea0db682b1a0d9efb30806f532490a26da888537f6bd1dfbddc0835030379e`
- Isolated: `758dada75ce369b46efb76fb950c5542f50f3741bf56be4ec63a7def27b7bd78`

**Protect existing wallet data:** Builds are CI debug-signed and may have an Android certificate incompatible with prior installations. **Never uninstall an existing wallet containing funds or clear its application data just to update**. The isolated `com.auronq.mobile.p2pcheck` package is independent; it does not migrate, overwrite or read the installed wallet. Do not import private keys holding funds into experimental builds.

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
