# AuronQ (AURQ) — Mainnet

AuronQ is a public UTXO proof-of-work cryptocurrency with ML-DSA-87 transaction signatures and the AuronQ-specific AQM64 proof-of-work construction.

## Mainnet identity

Network ID:

`44e62c2ace002a6660c14e252173c1aa303529c68e40c998e92da2b453f44f30b1e58c94d533587e2186004593fb856c433fcdb5418ed430ec8617e29529365c`

Genesis hash:

`5750a455c04bfe93c9edfef1a12744b05e29ac6da1a9dd5b790566629dea2080c581beb2f0324efba2067c9efb113ed7a29598fd3ffbed965ced31f265d0cec4`

Genesis founder allocation: 210,000 AURQ. Initial block subsidy: 49.5 AURQ. Coinbase maturity: 100 blocks.

## Public P2P discovery

AuronQ does not use a central blockchain server. Every full node stores and validates its own chain. New nodes bootstrap similarly to Bitcoin-style seed discovery:

- previously learned public peers are persisted locally and retried;
- fixed seed peers are supported;
- DNS seeds are supported;
- the official bounded HTTPS bootstrap manifest is an additional rendezvous source;
- after first contact, nodes exchange verified public peer addresses and persist them;
- NAT/CGNAT nodes participate through outbound connections without opening an inbound port;
- publicly reachable nodes can announce themselves and become normal gossiped peers.

Official stable bootstrap manifest:

`https://raw.githubusercontent.com/promirmir/AuronQ/main/bootstrap.json`

The two initial bootstrap entries are stable Tailscale Funnel HTTPS front-ends to ordinary full AuronQ nodes. They are only first-contact relays and have no consensus privileges. Every peer independently validates blocks and cumulative chain work.

## Windows

Normal users extract the Windows package and run `START-AURONQ.cmd`. No Tailscale, Cloudflare, manual peer entry, Go or Docker is required for a normal participant.

The two initial bootstrap operator machines run `START-SEED-NODE.cmd`, which starts the full node and backgrounds Tailscale Funnel when Tailscale is installed and Funnel is authorized.

## Verified network behavior

The rc4 networking fix was verified on two independent computers: a node that was at height 0 automatically caught up to height 4 from the stronger peer, and a newly mined block 5 then propagated so both nodes reached the same height and identical tip.

## Security status

The code passes its included Go test/vet/race suites, but AQM64 and the full consensus/network implementation have **not** received an independent professional security or cryptographic audit. Mainnet availability and decentralization also depend on multiple independent operators keeping reachable nodes online.
