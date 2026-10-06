# AuronQ Mainnet — immutable identity

This file records the public Mainnet identity established at launch. These consensus identifiers are version-independent and must not be confused with the current Desktop / Full Node software release.

- Network ID: `44e62c2ace002a6660c14e252173c1aa303529c68e40c998e92da2b453f44f30b1e58c94d533587e2186004593fb856c433fcdb5418ed430ec8617e29529365c`
- Genesis: `5750a455c04bfe93c9edfef1a12744b05e29ac6da1a9dd5b790566629dea2080c581beb2f0324efba2067c9efb113ed7a29598fd3ffbed965ced31f265d0cec4`
- Founder address: `aurq1keaacmqgcvostprfhfejx55noyd746frjzv3qusb62lmutxurypahx7yiqq3mbg2lq`
- Genesis founder allocation: 210,000 AURQ
- Initial block subsidy: 49.5 AURQ
- Coinbase maturity: 100 blocks
- Target block interval: 600 seconds
- Default P2P port: TCP 18444
- Launch release tag: `v1.7.0`
- Launch tagged commit: `a4f6e1ff4c2afc975831d42c45432b94fa14d5bb`

Bootstrap metadata is not part of consensus and may change without changing Network ID.

Current official manifest:

`https://raw.githubusercontent.com/promirmir/AuronQ/main/bootstrap.json`

Current bootstrap registry:

`https://raw.githubusercontent.com/promirmir/AuronQ/main/bootstrap.json`

The registry is discovery metadata and can change independently of Mainnet consensus. Do not use a hardcoded peer count as part of the Mainnet identity.

A fresh node uses bootstrap only for first contact, then validates the chain locally and learns additional peers. The bootstrap has no authority over transaction validity, block validity or chain selection.

The public network should add multiple independently operated reachable nodes and DNS seeds over time so fresh installations do not depend on one rendezvous path.

The AuronQ Mainnet implementation has not received an independent professional security or cryptographic audit.
