# AuronQ 1.7.0-rc2 — Mainnet-capable release candidate

This build can create the immutable AuronQ mainnet genesis with the explicit `--ack-unaudited-mainnet` acknowledgement. It contains the full validating node, AQM64 CPU miner, ML-DSA-87 wallets, UTXO consensus, cumulative-work reorg, mempool, public P2P bootstrap/gossip, persistent chain data and Windows Desktop application.

Read in this order:

1. `AUDIT-2026-10-02.md` — what was checked, fixed and what remains unverified.
2. `README.md` — architecture, monetary policy and normal usage.
3. `PROTOCOL.md` — consensus rules.
4. `MAINNET-LAUNCH.md` — exact launch sequence.
5. `PUBLIC-NETWORK.md` — bootstrap/seed deployment.
6. `SECURITY.md` and `THREAT-MODEL.md` — security boundaries.

For the founder/mainnet operator: do **not** publish `founder.wallet`, its password or seed. Generate them only on your trusted machine. The final public artifact is `network.json`, not the founder wallet.

For a normal Windows user after mainnet exists: the release directory should contain `AuronQ-Desktop.exe` and the final public `network.json`. Desktop rc2 can import that bundled network automatically on a clean first launch and discover peers from its public bootstrap entries.

Security status: internal rc2 audit/hardening completed; independent professional audit, long hostile public test operation, distributed DoS/Sybil testing and independent AQM64 economic/cryptographic review remain outstanding.