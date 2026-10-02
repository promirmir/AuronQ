# AuronQ (AURQ) — Mainnet 1.7.0

AuronQ is a public UTXO proof-of-work cryptocurrency with ML-DSA-87 transaction signatures and the AuronQ-specific AQM64 proof-of-work construction.

## Download

**Windows x64:** [AuronQ-1.7.0-Windows-x64.zip](https://github.com/promirmir/AuronQ/releases/download/v1.7.0/AuronQ-1.7.0-Windows-x64.zip)

**Linux amd64:** [AuronQ-1.7.0-Linux-amd64.tar.gz](https://github.com/promirmir/AuronQ/releases/download/v1.7.0/AuronQ-1.7.0-Linux-amd64.tar.gz)

Release page: [AuronQ 1.7.0 Mainnet](https://github.com/promirmir/AuronQ/releases/tag/v1.7.0)

SHA-256:

```text
79f9e75b61ad17f26e6e90e3d8dc07883aa3cb00f1f7882bcf96dd6291a72f37  AuronQ-1.7.0-Windows-x64.zip
9b4a022c333eb11c030a91a6325f85c7ade949958b493245a67a00bfe9014fbf  AuronQ-1.7.0-Linux-amd64.tar.gz
```

Windows users: extract the whole ZIP and run `START-AURONQ.cmd`.

The current Windows binaries are **not Authenticode-signed**, so Windows SmartScreen may warn on first launch. Verify the SHA-256 above and download only from this repository's Releases page.

## Mainnet identity

Network ID:

`44e62c2ace002a6660c14e252173c1aa303529c68e40c998e92da2b453f44f30b1e58c94d533587e2186004593fb856c433fcdb5418ed430ec8617e29529365c`

Genesis hash:

`5750a455c04bfe93c9edfef1a12744b05e29ac6da1a9dd5b790566629dea2080c581beb2f0324efba2067c9efb113ed7a29598fd3ffbed965ced31f265d0cec4`

Genesis founder allocation: 210,000 AURQ (1% of the nominal 21,000,000 AURQ cap). Initial subsidy: 49.5 AURQ. Coinbase maturity: 100 blocks. Target spacing: 600 seconds.

## Public P2P discovery

AuronQ does not use a central blockchain server. Every full node keeps and validates its own canonical chain.

A fresh node can discover peers from:

- previously learned peers stored locally;
- configured fixed seeds;
- DNS seeds;
- the bounded HTTPS bootstrap manifest;
- peer gossip after the first connection.

The official stable manifest is:

`https://raw.githubusercontent.com/promirmir/AuronQ/main/bootstrap.json`

At present the manifest advertises **one confirmed public bootstrap endpoint**:

`https://mir.taild63f46.ts.net`

That endpoint is only a first-contact relay to an ordinary full node. It has no consensus privileges. The node independently validates blocks, transactions, Network ID and cumulative chain work. More independently operated public nodes and DNS seeds should be added as the network grows so fresh installations do not depend on a single rendezvous path.

NAT/CGNAT users can participate through outbound connections without port forwarding. Publicly reachable nodes can be learned and gossiped by other peers.

## Network behavior

The node exchanges peer metadata, transactions and blocks over its P2P HTTP transport. It validates downloaded blocks locally, selects chains by cumulative work, limits synchronization batches, rate-limits requests/block submission, persists known public peers, and repairs a same-branch peer that missed one or more block broadcasts by pushing the missing validated extension.

A real two-node test synchronized the same mainnet chain through height 5 and the same tip after mining and catch-up repair.

## Source and CI

The complete Go source is in this repository. CI runs tests and vetting on Linux and Windows, includes the race detector on Linux, and builds the Windows CLI/Desktop plus Linux CLI.

The immutable `v1.7.0` release tag points to commit:

`a4f6e1ff4c2afc975831d42c45432b94fa14d5bb`

## Security status

AuronQ 1.7.0 is live mainnet software, but **AQM64 and the overall consensus/network implementation have not received an independent professional security or cryptographic audit**. Passing internal/CI tests is not equivalent to an external audit. Do not present AuronQ as production-audited financial infrastructure until independent review has occurred.

See [SECURITY.md](SECURITY.md), [PROTOCOL.md](PROTOCOL.md), [AQM64.md](AQM64.md) and [MAINNET.md](MAINNET.md).
