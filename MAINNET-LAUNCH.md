# AuronQ Mainnet launch record — 1.7.0

AuronQ Mainnet has been created and the public `v1.7.0` software release has been published.

## Immutable network identity

Network ID:

`44e62c2ace002a6660c14e252173c1aa303529c68e40c998e92da2b453f44f30b1e58c94d533587e2186004593fb856c433fcdb5418ed430ec8617e29529365c`

Genesis hash:

`5750a455c04bfe93c9edfef1a12744b05e29ac6da1a9dd5b790566629dea2080c581beb2f0324efba2067c9efb113ed7a29598fd3ffbed965ced31f265d0cec4`

Founder address:

`aurq1keaacmqgcvostprfhfejx55noyd746frjzv3qusb62lmutxurypahx7yiqq3mbg2lq`

Genesis founder allocation: 210,000 AURQ.

Coinbase maturity: 100 blocks.

Tagged release commit:

`a4f6e1ff4c2afc975831d42c45432b94fa14d5bb`

## Public bootstrap

The stable discovery manifest is:

`https://raw.githubusercontent.com/promirmir/AuronQ/main/bootstrap.json`

The currently confirmed public bootstrap is:

`https://mir.taild63f46.ts.net`

Bootstrap metadata is not consensus and may be rotated without changing Network ID.

## Windows user flow

1. Download the official Windows ZIP from GitHub Releases.
2. Verify SHA-256.
3. Extract the entire archive.
4. Run `START-AURONQ.cmd`.
5. The node validates the network definition, contacts bootstrap peers, synchronizes and learns additional peers.

## Operator priorities after launch

- keep at least one confirmed public bootstrap reachable;
- add independent seeds on different networks/providers;
- add DNS seeds when multiple stable public nodes exist;
- monitor synchronization/reorg behavior over long-running public operation;
- obtain independent consensus/cryptography/P2P/wallet review;
- sign future Windows releases with Authenticode when a code-signing certificate is available.

## Private material

Never publish `founder.wallet`, wallet passwords, private seeds, node private data or ceremony password files.

The mainnet launch does not imply independent audit or formal verification.
