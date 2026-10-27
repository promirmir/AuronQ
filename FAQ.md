# AuronQ FAQ

## What is AuronQ?

AuronQ (AURQ) is an open-source UTXO Proof-of-Work cryptocurrency written in Go. Transactions use ML-DSA-87 signatures, while mining uses the AuronQ-specific AQM64 Proof-of-Work construction.

## Is AuronQ a post-quantum cryptocurrency?

AuronQ uses ML-DSA-87 for transaction signatures. ML-DSA is a standardized post-quantum digital-signature scheme. This protects the transaction-signature layer against the class of attacks ML-DSA is designed to address.

That does not establish end-to-end security of the whole implementation; consensus, networking, wallets and AQM64 require independent audit.

## What is ML-DSA-87?

ML-DSA-87 is the highest-security parameter set of the ML-DSA post-quantum digital-signature standard. AuronQ uses it to authorize transactions.

## What is AQM64?

AQM64 is AuronQ's project-specific Proof-of-Work construction. Its technical description is available in [AQM64.md](AQM64.md).

AQM64 is project-specific and requires independent review as part of the AuronQ audit.

## Is AuronQ Proof-of-Work?

Yes. AuronQ uses Proof-of-Work consensus. The target block interval is 600 seconds.

## Is AuronQ based on a UTXO model?

Yes. AuronQ uses a UTXO transaction model.

## Is the AuronQ Mainnet live?

Yes. The public Mainnet is live. Current full-node/Desktop release: Windows v1.7.14 and Linux graphical `.deb` v1.7.14.1.

## Where can I download AuronQ?

Use the official [GitHub Releases](https://github.com/promirmir/AuronQ/releases) page. Verify published SHA-256 checksums before running downloaded binaries.

## Which operating systems are supported?

The current Desktop/full-node is **Windows x64 v1.7.14** or **Linux graphical Desktop v1.7.14.1 `.deb`** (Ubuntu/Mint/Debian x86-64). The Linux v1.7.14 `.tar.gz` is CLI-only. Current Android wallet: **0.5.9 Alpha — Light Quorum / Fast Start**. See [Linux guide](README-LINUX.md) and [Android guide](android/README.md).

The current published **Universal Miner v0.4.9 Alpha** has Windows x64 GUI/GPU and Linux x64 accelerated packages. See [Releases](https://github.com/promirmir/AuronQ/releases/tag/miner-v0.4.9-alpha) and the [miner guide](UNIVERSAL-MINER-GUIDE.md).

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

**AuronQ requires an independent security audit.** Internal tests do not replace independent review.

## Is AuronQ open source?

Yes. The project is published under the MIT License.

## What language is AuronQ written in?

The core implementation is written in Go.

## How can I contribute?

Read [CONTRIBUTING.md](CONTRIBUTING.md), open a focused issue, or submit a pull request. Security vulnerabilities should be reported using the process in [SECURITY.md](SECURITY.md), not through a public issue.
