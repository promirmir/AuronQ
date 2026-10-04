# AuronQ Mobile 0.4.2 Alpha

Android light-wallet client for the public AuronQ Mainnet.

## Multi-peer model

Version 0.4.2 no longer silently relies on one selected node for its network view.

It:
- remembers public HTTPS AuronQ nodes learned from native P2P gossip;
- compares Mainnet observations from multiple replaceable nodes;
- groups peers by exact height, tip and reported chain work;
- exposes the observed/agreement count in the UI;
- prefers a state reported by multiple peers over a lone endpoint claiming much greater work;
- compares wallet balances only across peers that report the same chain state;
- refuses to present a quorum balance when same-chain peers return conflicting balances;
- broadcasts the same locally signed transaction directly to multiple reachable AuronQ nodes.

Bundled bootstrap addresses and the optional GitHub manifest are **rendezvous hints**, not consensus authorities.

## Wallet model

The encrypted wallet and private seed remain in Android app-private storage. The password is not stored. ML-DSA-87 transaction signing happens locally using the same portable Go implementation as the Desktop/full-node software.

Remote nodes receive only public addresses, queries and already-signed transactions. They cannot derive or use the wallet private key.

## What the mobile app verifies

AuronQ Mobile verifies that contacted nodes report the exact AuronQ Mainnet Network ID and it cross-checks several peer-visible states.

It is still a **light client, not a full node**. It does not yet independently replay every transaction or verify the complete AQM64 proof-of-work/header chain. Multi-peer agreement reduces single-endpoint dependence but must not be described as equivalent to full-node validation.

## Independence from the original computers

Previously learned public peers are cached locally. The app can therefore move away from the bootstrap addresses after it learns other reachable AuronQ nodes. A fresh installation still needs at least one reachable discovery route because an unknown peer address cannot be discovered from nothing.

The long-term target remains the same as Desktop: the original project machines should be removable without affecting a network that already contains independently operated reachable nodes.

## Release

- APK: [AuronQ-Mobile-0.4.2-alpha.apk](https://github.com/promirmir/AuronQ/releases/download/android-v0.4.2-alpha/AuronQ-Mobile-0.4.2-alpha.apk)
- Release: [android-v0.4.2-alpha](https://github.com/promirmir/AuronQ/releases/tag/android-v0.4.2-alpha)
- SHA-256: `0224f45dca6b721577a60dc5dcd03d5734c0468c1abcac6f4c4e183b19c59736`

This remains debug-signed alpha software and has not received an independent security audit. Do not use substantial value.
