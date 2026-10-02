# AuronQ Mainnet 1.7.0

This file records the public mainnet identity and launch model.

- Network ID: `44e62c2ace002a6660c14e252173c1aa303529c68e40c998e92da2b453f44f30b1e58c94d533587e2186004593fb856c433fcdb5418ed430ec8617e29529365c`
- Genesis: `5750a455c04bfe93c9edfef1a12744b05e29ac6da1a9dd5b790566629dea2080c581beb2f0324efba2067c9efb113ed7a29598fd3ffbed965ced31f265d0cec4`
- Founder address: `aurq1keaacmqgcvostprfhfejx55noyd746frjzv3qusb62lmutxurypahx7yiqq3mbg2lq`
- Genesis allocation: 210,000 AURQ
- Coinbase maturity: 100 blocks
- Default P2P port: TCP 18444

Bootstrap metadata is not part of consensus and may change without changing the Network ID. New nodes use persisted peers, fixed seeds, DNS seeds and the official HTTPS manifest, then learn additional public peers from normal P2P gossip.

The initial bootstrap endpoints are relay front-ends to ordinary full nodes. They have no consensus privileges. A healthy public network should accumulate multiple independently operated publicly reachable nodes and DNS seeds over time.
