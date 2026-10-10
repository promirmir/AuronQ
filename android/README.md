# AuronQ Mobile — Android

**Current published APK: v0.5.4 Alpha.** AuronQ is experimental financial software. This application is a light wallet, not a full validating node; AQM64 and the overall wallet/network implementation have not been independently audited.

## Download (GitHub Releases)

| Package | Intended use |
| --- | --- |
| [AuronQ-Mobile-0.5.4-alpha.apk](https://github.com/promirmir/AuronQ/releases/download/android-v0.5.4-alpha/AuronQ-Mobile-0.5.4-alpha.apk) | Regular package ID `com.auronq.mobile`; **new installs only unless Android confirms signing certificate compatibility**. |
| [AuronQ-Mobile-0.5.4-alpha-Isolated-Test.apk](https://github.com/promirmir/AuronQ/releases/download/android-v0.5.4-alpha/AuronQ-Mobile-0.5.4-alpha-Isolated-Test.apk) | Installs as `com.auronq.mobile.autocp`, independently of existing AuronQ wallets, for no-funds functionality testing. |
| [GitHub Release and SHA-256](https://github.com/promirmir/AuronQ/releases/tag/android-v0.5.4-alpha) | Verify downloaded assets and see limitations. |
| [Legacy v0.5.3-alpha APK](https://github.com/promirmir/AuronQ/releases/tag/android-v0.5.3-alpha) | Unmodified previous release. |

SHA-256 (published v0.5.4):
- Regular APK: `12de641331d3db9a4415ba1a303f254bce7ba9a58d6b050ce8f6a0e590c66d7e`
- Isolated test APK: `dcf1c0a29574a9b089ff80f07daa774173138a11ee67322c874aef8ab3da42fe`

**Critical compatibility note:** the regular APK is signed with the CI debug key. Android may reject installing it over a previously installed AuronQ app signed with a different certificate. **Never uninstall a wallet holding funds or erase app data merely to force an APK update.** Back up recovery material using the existing application first. The isolated test app cannot access or replace your existing wallet data; do not import a wallet holding real funds into the test app.

## How synchronization works

The included initial, previously checked release checkpoint is height **1284**. Headers after the local trusted checkpoint are verified using the unchanged AQM64 PoW, difficulty and timestamp rules. Previously verified progress is stored atomically in Android app-private storage.

New v0.5.4 functionality:
- At most once every six hours, the application **attempts** to download a newer checkpoint from the official project GitHub checkpoint publication branch.
- It validates the **Ed25519 signature using a public key pinned in the APK**, domain separation, Mainnet Network ID and genesis, release anchor, expiration, increasing chainwork, and protection against reverting locally verified state.
- Before adopting it, the application checks the same checkpoint height/hash using **two public node network groups**. This cross-check is supplementary to the digital signature, not independent proof of historical consensus.
- Unavailable, expired, invalid, conflicting or unauthenticated checkpoint updates are **discarded**. The application continues using its existing cache and ordinary header verification.

**Trust boundary:** a newly downloaded checkpoint is trusted because it was signed by the project's checkpoint publishing authority after its CI verification; the mobile device does **not** independently repeat all historic AQM64 work before that checkpoint. Full-node validation remains available separately. This is not a consensus rule change or a guarantee of complete trustlessness.

## Automatic publisher activation

**The updater and the periodic publisher code are installed, but periodic publication cannot produce signatures until an operator provisions the private key in a GitHub Actions repository secret.** This is deliberate: the private key must not be embedded in the open-source app or stored in GitHub source control.

1. Obtain and securely store the operator's private checkpoint signing seed **outside the repository**. It must match the Ed25519 public key pinned in `mobile/bridge/signed_checkpoint.go`.
2. In repository **Settings → Secrets and variables → Actions → New repository secret**, create the exact name `AURONQ_CP_ED25519_PRIVATE_KEY`. Set its value to the base64-encoded 32-byte seed (not the filename or the entire explanatory note). Do not put this value in an issue, commit, PR, email or chat.
3. Verify that repository Actions can write the independent publication branch `automation/mobile-checkpoints`. The workflow [Signed AuronQ Mobile checkpoint publisher](../.github/workflows/mobile-signed-checkpoints.yml) is scheduled twice a day and can be run manually with `workflow_dispatch`. It signs and publishes only after PoW verification and independent peer checks pass.
4. Verify that the published `latest.json` appears on the publication branch. The app will only adopt newer candidates if their signatures and peer checks succeed.

The signing process does not accept raw unverified peer heights as checkpoints. Its Go validator runs the normal AQM64 header rules from the previous **already signed** checkpoint and ensures monotonic cumulative work.

For the full threat model and operational steps, see [Signed checkpoint operations](../mobile/bridge/SIGNED-CHECKPOINTS.md).

## What was not changed

Private keys, encrypted wallet format, address derivation, signing, `SendMultiVerified`, UTXO quorum verification, genesis, Network ID, AQM64 and Mainnet consensus. A light wallet is **not** equivalent to a locally validating full node.

[Security policy](../SECURITY.md) · [Mainnet specification](../MAINNET.md) · [Report an issue](https://github.com/promirmir/AuronQ/issues)
