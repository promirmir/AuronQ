# AuronQ FAQ

## What is AuronQ?

AuronQ (AURQ) is an open-source UTXO Proof-of-Work cryptocurrency written in Go. Transactions use ML-DSA-87 signatures, while mining uses the AuronQ-specific AQM64 Proof-of-Work construction.

## Is AuronQ a post-quantum cryptocurrency?

AuronQ uses ML-DSA-87 for transaction signatures. ML-DSA is a standardized post-quantum digital-signature scheme. This protects the transaction-signature layer against the class of attacks ML-DSA is designed to address.

That does **not** mean the entire cryptocurrency is proven “quantum-proof.” Consensus, networking, implementation security, wallet security and AQM64 must be evaluated separately. AuronQ has not received an independent professional cryptographic or security audit.

## What is ML-DSA-87?

ML-DSA-87 is the highest-security parameter set of the ML-DSA post-quantum digital-signature standard. AuronQ uses it to authorize transactions.

## What is AQM64?

AQM64 is AuronQ's project-specific Proof-of-Work construction. Its technical description is available in [AQM64.md](AQM64.md).

AQM64 is not presented as an externally audited cryptographic primitive. Independent review is encouraged.

## Is AuronQ Proof-of-Work?

Yes. AuronQ uses Proof-of-Work consensus. The target block interval is 600 seconds.

## Is AuronQ based on a UTXO model?

Yes. AuronQ uses a UTXO transaction model.

## Is the AuronQ Mainnet live?

Yes. The current stable full-node/Desktop release documented by this repository is AuronQ 1.7.13 Mainnet.

## Where can I download AuronQ?

Use the official [GitHub Releases](https://github.com/promirmir/AuronQ/releases) page. Verify published SHA-256 checksums before running downloaded binaries.

## Which operating systems are supported?

The current Desktop/full-node release includes Windows x64 and Linux amd64 builds. An Android wallet client is available as an alpha release.

The official Universal Miner v0.4.0 Alpha adds portable CPU-safe mining builds for Windows x64/ARM64, Linux x64/ARM64 and macOS x64/ARM64. Windows x64 also has the PL/EN GUI with NVIDIA CUDA acceleration; Linux x64 has an accelerated CUDA package. On unsupported/non-NVIDIA graphics, AUTO safely falls back to native CPU mining. See [UNIVERSAL-MINER-GUIDE.md](UNIVERSAL-MINER-GUIDE.md) and [GPU-MINER-GUIDE.md](GPU-MINER-GUIDE.md).

## Is the Android app a full node?

No. The Android application is a wallet client. Private keys and transaction signing remain local, while blockchain information is obtained from AuronQ full nodes.

## Can I mine AuronQ?

Yes. Mining is part of the Proof-of-Work network. Universal Miner AUTO uses validated NVIDIA CUDA acceleration when available and otherwise falls back to the native CPU AQM64 backend. The selected backend should pass the built-in self-test before Mainnet mining. See [UNIVERSAL-MINER-GUIDE.md](UNIVERSAL-MINER-GUIDE.md).

## Does every node validate the blockchain independently?

Yes. Full nodes independently validate blocks, transactions, Network ID and cumulative chain work.

## Does AuronQ depend on one central server?

Consensus does not depend on a central server. A fresh node still needs a discovery path to find its first peers. AuronQ supports persisted peers, fixed seeds, DNS seeds, an HTTPS bootstrap manifest and peer gossip.

The repository currently documents one confirmed public bootstrap endpoint. More independently operated public nodes and independent discovery routes are needed to reduce first-contact dependency.

## Can I run a public AuronQ node?

Yes. Publicly reachable full nodes improve peer discovery and network resilience. See [PUBLIC-NETWORK.md](PUBLIC-NETWORK.md) and [NETWORK-INDEPENDENCE.md](NETWORK-INDEPENDENCE.md).

## Is AuronQ audited?

No independent professional security or cryptographic audit has been completed. Internal tests, CI and project documentation are not substitutes for independent review.

## Is AuronQ open source?

Yes. The project is published under the MIT License.

## What language is AuronQ written in?

The core implementation is written in Go.

## How can I contribute?

Read [CONTRIBUTING.md](CONTRIBUTING.md), open a focused issue, or submit a pull request. Security vulnerabilities should be reported using the process in [SECURITY.md](SECURITY.md), not through a public issue.
