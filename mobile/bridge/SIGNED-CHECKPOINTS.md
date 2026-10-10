# AuronQ Mobile: signed checkpoint operations

## Architecture

AuronQ Mobile verifies an Ed25519 signature over a Mainnet-bound checkpoint manifest before using that candidate as a historical trust anchor. The app also checks expiry, cumulative work, local-chain monotonicity and agreement across two network groups. It validates subsequent AQM64 headers locally and preserves previously verified chain state on any update failure.

This is a light-client trust assumption: a historical checkpoint signed by the project is not equivalent to full independent validation from genesis.

## Periodic publishing

The workflow in `.github/workflows/mobile-signed-checkpoints.yml` is scheduled twice daily. It verifies newer block headers using the existing consensus validator and corroborates them against publicly reachable peers. It only publishes a new signed manifest when a compatible private checkpoint signing key is configured as the GitHub Actions repository secret `AURONQ_CP_ED25519_PRIVATE_KEY`. Without that configuration it publishes nothing, and wallets continue from their previously trusted cache.

Never commit or disclose private signing material. Changes to the publishing workflow should require code review and green CI checks. If the signer is compromised, disable publishing and rotate the key through a reviewed Android update.

## Operational considerations

Publication uses the `automation/mobile-checkpoints` GitHub branch, but that branch does not control Mainnet consensus. If GitHub is unavailable, existing wallets retain their local validated cache and the embedded release checkpoint. A new installation may need a newer APK or a separately authenticated distribution path in the future.

See [Android instructions](../../android/README.md) and [Mainnet change policy](../../MAINNET-CHANGE-POLICY.md).
