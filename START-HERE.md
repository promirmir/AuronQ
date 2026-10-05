# Start here — AuronQ Mainnet

## Windows

1. Open the official latest release: https://github.com/promirmir/AuronQ/releases/latest
2. Download the current `AuronQ-<version>-Windows-x64.zip` archive.
3. Download `SHA256SUMS.txt` from the same release and verify the archive SHA-256 before running it.
4. Extract the complete ZIP to a new folder.
5. Run `START-AURONQ.cmd`.

PowerShell checksum example:

```powershell
Get-FileHash .\AuronQ-<version>-Windows-x64.zip -Algorithm SHA256
```

Compare the result with the matching entry in `SHA256SUMS.txt` from the same GitHub Release.

A normal participant does not need Tailscale, Cloudflare, manual peer addresses, Go or Docker.

The client validates the bundled Mainnet definition, consults the public bootstrap manifest, connects to available peers, validates the chain locally, learns additional public peers and stores them for future restarts.

## Current bootstrap state

The public bootstrap manifest is:

`https://raw.githubusercontent.com/promirmir/AuronQ/main/bootstrap.json`

The manifest can contain multiple currently verified public rendezvous peers and may change over time. It is discovery metadata only: bootstrap peers do not determine consensus, cannot make an invalid block valid and are excluded from the immutable Mainnet identity.

After first contact, nodes learn and persist additional peers through native AuronQ peer gossip.

## Security

AuronQ Mainnet is live experimental financial software and has not received an independent professional consensus, networking or cryptographic security audit.

Verify release checksums, back up wallets before use and do not use substantial value while the project remains in the stabilization phase.

See [SECURITY.md](SECURITY.md), [MAINNET-CHANGE-POLICY.md](MAINNET-CHANGE-POLICY.md) and [STABILIZATION.md](STABILIZATION.md).
