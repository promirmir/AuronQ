# AuronQ Mobile — Android

**Latest proposed version: v0.5.5 Alpha** — automated keyless AQM64 header checkpoints. AuronQ is experimental software; this is a light wallet, not a full validating node.

## Installation and wallet safety

- [Download AuronQ Mobile v0.5.5 Alpha (Android APK)](https://github.com/promirmir/AuronQ/releases/download/android-v0.5.5-alpha/AuronQ-Mobile-0.5.5-alpha.apk)
- [Download safely installable isolated test APK](https://github.com/promirmir/AuronQ/releases/download/android-v0.5.5-alpha/AuronQ-Mobile-0.5.5-alpha-Keyless-Isolated-Test.apk)
- [Release page and SHA-256 checksums](https://github.com/promirmir/AuronQ/releases/tag/android-v0.5.5-alpha)
- [Prior stable checkpoint release v0.5.4 Alpha](https://github.com/promirmir/AuronQ/releases/tag/android-v0.5.4-alpha)

**Wallet warning:** Android may refuse to install a normal debug-signed build over an older version signed by another debug certificate. Never delete, uninstall, or clear a wallet holding funds just to force an upgrade. Test the separately installed `com.auronq.mobile.keyless` build first without importing any valuable wallet. Existing wallets, private keys, addresses and the Mainnet consensus protocol remain unchanged.

## Completely automatic checkpoint handling — no personal keys

1. The original Android release contains an immutable verified AQM64 historical anchor at height **1284**.
2. On GitHub Actions, [the keyless publisher](../.github/workflows/mobile-keyless-checkpoints.yml) runs every **four hours** and on relevant Mainnet code pushes. It publishes a *new* checkpoint only after **256 additional blocks** can be independently validated from the last authenticated checkpoint, followed by corroboration across different public peer groups.
3. GitHub Actions automatically receives an ephemeral GitHub OpenID Connect identity. [Sigstore/Cosign](https://docs.sigstore.dev/cosign/verifying/verify/) signs the exact checkpoint bytes and publishes a transparency-backed bundle. **No manual private signing key, wallet key, or repository secret is required.**
4. Android checks for an update periodically and verifies **the Sigstore bundle, Fulcio/Rekor evidence, exact expected GitHub Actions workflow identity, payload digest, AuronQ Mainnet/genesis, expiry, accumulated work and nonrollback conditions** before considering any update. It also checks multiple peer groups.
5. If the publisher, GitHub, or the Sigstore trust roots are unavailable, the app refuses an unverified update and continues from its existing verified cache and incremental AQM64 validation.

### Trust and decentralization boundaries

Keyless signing authenticates **which specific GitHub Actions workflow issued a checkpoint**, not that its operator is incapable of misbehavior. As with the previous signed release scheme, trusting a historical checkpoint is a light-wallet compromise: the phone does not revalidate every AQM64 header before that checkpoint, but continues validating newer headers independently. A separate full AuronQ node remains the option for full transaction/UTXO validation. GitHub/Sigstore are publishing infrastructure and cannot change the consensus of running AuronQ full nodes.

The signing certificates and ephemeral private keys are generated automatically during the workflow and are not shared with users. Do not attempt to configure `AURONQ_CP_ED25519_PRIVATE_KEY` for this version: it is **not used**.

## Operator monitoring

Check the [keyless publisher workflow](../.github/workflows/mobile-keyless-checkpoints.yml), its last successful job and the `automation/mobile-checkpoints` branch. The published `latest.json` and `latest.sigstore.json` must both be present; mismatched files are rejected.

GitHub scheduled workflow execution is best-effort and can be delayed. No endpoint, signing authority, or certificate is assumed available at all times. The wallet never depends on a signing service to preserve the local private keys.

[Security](../SECURITY.md) · [Mainnet protocol](../MAINNET.md) · [Issues](https://github.com/promirmir/AuronQ/issues)
