# AuronQ (AURQ) — Exchange, Indexer & Third-Party Integration Policy

AuronQ is a public, open-source Proof-of-Work cryptocurrency network.

## Permissionless integration

Independent third parties may, without prior project approval:

- list or create markets for **AURQ**;
- integrate AURQ deposits and withdrawals;
- run full nodes, wallets, explorers, indexers and monitoring services;
- operate mining pools or compatible mining software;
- build analytics, custody, payment or other services around the public network;
- independently review, reproduce and audit the public implementation.

No exclusive integration partner exists, and no third party receives authority over AuronQ consensus by integrating the network.

Any exchange, service or developer remains independently responsible for legal/regulatory compliance, security review, custody design, operational risk, user disclosures and its own listing standards. Inclusion here is **not** an endorsement of any third-party service and does not guarantee a listing.

## Public verification

AuronQ is intended to be integrated from verifiable public information rather than private documentation.

- Source code: https://github.com/promirmir/AuronQ
- License: MIT — see [LICENSE](LICENSE)
- Mainnet specification: [MAINNET.md](MAINNET.md)
- Protocol specification: [PROTOCOL.md](PROTOCOL.md)
- AQM64 Proof-of-Work: [AQM64.md](AQM64.md)
- Pool integration: [POOL-INTEGRATION.md](POOL-INTEGRATION.md)
- Releases: https://github.com/promirmir/AuronQ/releases
- Public project site: https://promirmir.github.io/AuronQ/
- Public explorer example: https://mir.taild63f46.ts.net/explorer
- Machine-readable project metadata: [auronq-project.json](auronq-project.json)

## Mainnet identity

| Field | Value |
|---|---|
| Project | AuronQ |
| Symbol | AURQ |
| Network | Mainnet |
| Ledger | UTXO |
| Consensus | Proof-of-Work |
| Proof-of-Work | AQM64 |
| Transaction signatures | ML-DSA-87 |
| Target block spacing | 600 seconds |
| Nominal supply cap | 21,000,000 AURQ |
| Network ID | `44e62c2ace002a6660c14e252173c1aa303529c68e40c998e92da2b453f44f30b1e58c94d533587e2186004593fb856c433fcdb5418ed430ec8617e29529365c` |
| Genesis hash | `5750a455c04bfe93c9edfef1a12744b05e29ac6da1a9dd5b790566629dea2080c581beb2f0324efba2067c9efb113ed7a29598fd3ffbed965ced31f265d0cec4` |

These identifiers should be checked against the canonical Mainnet specification and source before production integration.

## Integration principle

AuronQ does not require an exchange, indexer, pool or infrastructure provider to obtain permission merely to interoperate with the public network.

The preferred process is:

1. verify the source and canonical Mainnet identity;
2. build and run an independent full node;
3. validate synchronization and chain-tip agreement using more than one independent source where practical;
4. test deposits, withdrawals, reorg handling and confirmation policy before production use;
5. disclose that the integration is independently operated.

AuronQ is still a young network and has not yet received an independent professional security or cryptographic audit. Production integrations should apply conservative confirmation, custody and risk policies until the network has accumulated a longer operating history and independent review.
