# AuronQ Mobile 0.1.0 Alpha

Android companion wallet for AuronQ Mainnet.

## Scope

This alpha is intentionally a **wallet client**, not a full Android node or miner.

It:

- creates the same encrypted `.wallet` format as AuronQ Desktop;
- imports and exports compatible wallet files;
- keeps the wallet file inside Android app-private storage;
- never stores the password;
- derives/signs ML-DSA-87 transactions locally through the same portable Go code used by desktop AuronQ;
- discovers an HTTPS AuronQ Mainnet node through the official GitHub bootstrap manifest;
- checks the node's Network ID before using it;
- shows spendable/total balance;
- submits locally signed transactions;
- can back up and delete the local wallet file.

## Security model

The mobile alpha is not a full node and does not independently validate the entire blockchain. A remote node can misreport balance/availability or refuse service. It cannot sign transactions because the private seed remains encrypted on the phone and signing happens locally.

The app is an **alpha test build** and the APK is debug-signed by GitHub Actions. Do not use it for substantial value before independent project/security review.

## Build

GitHub Actions generates an Android AAR from `mobile/bridge` with `gomobile bind`, then builds the Java Android application.
