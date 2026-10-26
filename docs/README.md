# AuronQ documentation

**Current status (26 October 2026):** AuronQ Mainnet is live; consensus changes are frozen for normal releases. Protected source `main` includes GitHub-independent peer discovery and verified thin-client/miner failover, while the currently published Desktop v1.7.13, Android 0.5.6 Alpha and Miner v0.4.8 Alpha predate those final changes. Stability, wallet correctness, independent security review and witnessed founder-independent operation **are not yet fully certified**.

**Start with [canonical project status](../PROJECT-STATUS.md).** This index helps readers navigate without digging through historical experiment files. It does not replace the protocol.

| Start here | What to read |
| --- | --- |
| Run AuronQ | [Quick start](../START-HERE.md) · [Windows](../README-WINDOWS.md) · [Public node](../PUBLIC-NETWORK.md) |
| Software downloads | [Official releases](https://github.com/promirmir/AuronQ/releases) · [Version and checksum ledger](../PROJECT-STATUS.md) |
| Protocol identity | [Mainnet](../MAINNET.md) · [Protocol](../PROTOCOL.md) · [AQM64](../AQM64.md) · [Technical Whitepaper](../WHITEPAPER.md) |
| Coin supply | [Canonical Mainnet definition](../MAINNET.md) · [Whitepaper issuance](../WHITEPAPER.md) |
| Independent network | [Network independence](../NETWORK-INDEPENDENCE.md) · [Decentralization](../DECENTRALIZATION.md) · [Operator handoff](INDEPENDENT-OPERATION.md) |
| Wallets | [Android documentation](../android/README.md) · [Current cautions](../SECURITY.md) |
| Mining | [Universal miner](../UNIVERSAL-MINER-GUIDE.md) · [GPU guidance](../GPU-MINER-GUIDE.md) · [Hardware support](../HARDWARE-SUPPORT.md) |
| Integrations | [Pool operator](../POOL-INTEGRATION.md) · [Independent exchange policy](../EXCHANGE-INTEGRATION.md) · [Exchange runbook](../EXCHANGE-OPERATOR-RUNBOOK.md) |
| Security | [Security policy](../SECURITY.md) · [Threat model](../THREAT-MODEL.md) · [Test matrix](../TEST-MATRIX.md) |
| Freeze and change policy | [Mainnet change policy](../MAINNET-CHANGE-POLICY.md) · [Maintenance freeze](../PROJECT-FREEZE.md) · [Stabilization criteria](../STABILIZATION.md) |
| Project participation | [Contributing](../CONTRIBUTING.md) · [Code of conduct](../CODE_OF_CONDUCT.md) · [Support](../SUPPORT.md) · [FAQ](../FAQ.md) |
| Historical records | [Changelog](../CHANGELOG.md) · [Archived materials](archive/README.md) · [Original Mainnet launch evidence](archive/mainnet-launch/README.md) |

## Scope of this reorganization

- Current machine-readable configuration (`network.json`, `bootstrap.json`, `auronq-project.json`) and current install/run scripts remain at their established paths.
- Canonical protocol, security, installer and integration references keep their existing root paths to avoid breaking downloads, external links, CI and tool references.
- Historical documents for the original Mainnet ceremony, launch and October 2 internal audit are grouped under `docs/archive/mainnet-launch/` without altering their historical text.
- Historical records are **not** current setup instructions. Use the active README, Mainnet specification and official releases instead.

**Do not publish private wallets, recovery seeds or passwords.** All AuronQ source and technical integration material remains publicly reviewable. Independent platforms must complete their own security and legal analysis.
