# AuronQ Mobile 0.3.0 Alpha

Android wallet client for the same public AuronQ Mainnet used by AuronQ Desktop and full nodes.

## What changed in 0.3.0

- Polish / English language switch, saved between launches;
- remembers previously discovered public HTTPS AuronQ nodes;
- prefers known working nodes before falling back to the bootstrap manifest;
- learns additional public HTTPS peers from `/p2p/hello` and keeps a bounded local cache;
- automatic failover if the current public node becomes unavailable;
- keeps the live Mainnet view from 0.2.0: height, peers, mempool, issued supply, tip, chain work and recent blocks;
- retains the Desktop-style dark interface.

## Independence from the original bootstrap computer

AuronQ full nodes already persist and gossip verified public peers. Existing full nodes can therefore continue operating with each other when the original bootstrap computer is offline, provided independent reachable peers exist.

A fresh installation still needs at least one discovery path. AuronQ supports a mutable bootstrap manifest and DNS seeds at the network configuration level. The original operator computer can be retired safely only after the public network has multiple independent reachable nodes and at least one independent discovery route remains available to fresh users.

AuronQ Mobile 0.3.0 improves this further by remembering public HTTPS nodes it has already learned. A phone that has discovered independent public nodes can reconnect to them later even if the original bootstrap node is offline.

This is resilience, not a mathematical guarantee: no peer-to-peer network can remain reachable if every public node/discovery route is offline.

## Wallet model

The app is intentionally a mobile wallet client, not a full Android node/miner.

It:
- creates the same encrypted `.wallet` format as AuronQ Desktop;
- imports and exports compatible wallet files;
- keeps the wallet file inside Android app-private storage;
- never stores the password;
- signs ML-DSA-87 transactions locally through the same portable Go implementation used by desktop AuronQ;
- queries public AuronQ Mainnet nodes for status, UTXOs and balances;
- broadcasts only the already-signed transaction;
- verifies the exact AuronQ Mainnet Network ID before accepting a node.

All users who connect to valid AuronQ Mainnet nodes observe the same canonical chain selected by the full nodes' consensus rules. The mobile app does not independently validate the entire blockchain like a full node.

## Security model

A remote node can misreport wallet-visible data such as balance or availability, or refuse service. It cannot sign transactions because the encrypted private seed stays on the device and signing happens locally.

This remains an alpha test build and the APK is debug-signed by GitHub Actions. Do not use substantial value before independent security review and broader network testing.
