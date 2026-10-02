# AuronQ Mobile 0.2.0 Alpha

Android wallet client for the same public AuronQ Mainnet used by AuronQ Desktop and full nodes.

## What changed in 0.2.0

- redesigned dark interface to visually match AuronQ Desktop;
- bottom navigation: Dashboard, Wallet, Send, Network;
- live AuronQ Mainnet screen updated every 5 seconds;
- current height, peer count, mempool, issued supply, tip and chain work;
- recent block feed with height, hash, timestamp and transaction count;
- automatic public-node discovery through the official GitHub bootstrap manifest;
- failover: if the current node becomes unavailable, the app attempts discovery again;
- exact Mainnet Network ID verification before the app accepts a node.

## Wallet model

The app is intentionally a mobile wallet client, not a full Android node/miner.

It:

- creates the same encrypted `.wallet` format as AuronQ Desktop;
- imports and exports compatible wallet files;
- keeps the wallet file inside Android app-private storage;
- never stores the password;
- signs ML-DSA-87 transactions locally through the same portable Go implementation used by desktop AuronQ;
- queries the public AuronQ Mainnet for status, UTXOs and balances;
- broadcasts only the already-signed transaction.

All users who connect to valid AuronQ Mainnet nodes observe the same canonical chain selected by the full nodes' consensus rules. The mobile app verifies the exact AuronQ Mainnet Network ID, but it does **not** independently validate the full blockchain like a full node.

## Security model

A remote node can misreport wallet-visible data such as balance or availability, or refuse service. It cannot sign transactions because the encrypted private seed stays on the device and signing happens locally.

This remains an alpha test build and the APK is debug-signed by GitHub Actions. Do not use substantial value before independent security review and broader network testing.
