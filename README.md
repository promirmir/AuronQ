# AuronQ (AURQ) — Mainnet 1.7.0

AuronQ is a public UTXO proof-of-work cryptocurrency with ML-DSA-87 transaction signatures and the AuronQ-specific AQM64 proof-of-work construction.

## Mainnet identity

Network ID:

`44e62c2ace002a6660c14e252173c1aa303529c68e40c998e92da2b453f44f30b1e58c94d533587e2186004593fb856c433fcdb5418ed430ec8617e29529365c`

Genesis hash:

`5750a455c04bfe93c9edfef1a12744b05e29ac6da1a9dd5b790566629dea2080c581beb2f0324efba2067c9efb113ed7a29598fd3ffbed965ced31f265d0cec4`

Genesis founder allocation: 210,000 AURQ (1% of the nominal 21,000,000 AURQ cap). Initial subsidy: 49.5 AURQ. Coinbase maturity: 100 blocks. Target spacing: 600 seconds.

## Bitcoin-like peer discovery model

AuronQ does not use a central blockchain server. Every full node keeps and validates its own canonical chain. First contact uses several bootstrap mechanisms, conceptually similar to Bitcoin's seed process:

- previously learned public peers are persisted locally and retried on restart;
- configured fixed seeds are supported;
- DNS seeds are supported;
- a bounded HTTPS bootstrap manifest is supported as an additional rendezvous source;
- after contact, nodes exchange verified public peer addresses and persist them;
- NAT/CGNAT nodes can fully participate through outbound connections without opening an inbound port;
- publicly reachable nodes can announce themselves and become part of normal peer gossip.

The official stable manifest is:

`https://raw.githubusercontent.com/promirmir/AuronQ/main/bootstrap.json`

At launch it contains two operator bootstrap endpoints exposed through stable Tailscale Funnel hostnames. Those endpoints are rendezvous paths only: they are not consensus authorities and do not hold privileged chain state. Once additional publicly reachable peers exist, nodes learn them directly from the P2P network.

## Windows

Download the `AuronQ-1.7.0-Windows-x64.zip` release asset, extract the whole archive and run `START-AURONQ.cmd`. A normal user does not need Tailscale, Cloudflare, Go, Docker or manual peer configuration.

The two official operator machines use `START-SEED-NODE.cmd`; it starts AuronQ and backgrounds Tailscale Funnel if Tailscale is installed/authorized.

## P2P behavior

The node exchanges hello/peer metadata, transactions and blocks over its HTTP P2P transport. It selects chains by cumulative work, validates downloaded blocks locally, limits sync batches, applies request/block rate limits, persists known public peers, and repairs a same-branch peer that missed one or more block broadcasts by pushing the missing validated extension.

A two-node real-world test synchronized the same mainnet chain through height 5 and the same tip after mining and catch-up repair.

## Build and verification

The repository CI runs `go test ./...`, `go vet ./...`, Linux CLI builds and Windows CLI/Desktop builds. The release workflow produces Windows and Linux mainnet artifacts from the tagged source.

## Security status

AuronQ 1.7.0 is the public mainnet software, but **AQM64 and the overall consensus/network implementation have not received an independent professional security or cryptographic audit**. The code passing its own tests is not equivalent to an external audit. Do not present the project as production-audited financial infrastructure until independent review has occurred.
