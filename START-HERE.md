# Start here — AuronQ Mainnet

## Linux — desktop by click (Ubuntu / Mint / Debian, x86-64)

1. Download [AuronQ Desktop Linux `.deb`](https://github.com/promirmir/AuronQ/releases/download/linux-desktop-v1.7.14.1/AuronQ-Desktop-1.7.14-1-Linux-amd64.deb) and compare SHA-256: `79f2859dc1ced75beb321b98bc9ed1edd0e5ab0c18a07be2c0d8ddd7ea90d73a`.
2. **Double-click** the downloaded `.deb`, then choose **Install** in the graphical package installer.
3. From the Linux applications menu, click **AuronQ Desktop**. Its wallet, miner, node and Explorer are the same Desktop UI as on Windows, opened by the default Linux browser.
4. Wait for the initial local blockchain validation/synchronization; this can postpone opening the browser.

See the [complete Linux one-click guide](README-LINUX.md). The older `AuronQ-1.7.14-Linux-amd64.tar.gz` archive contains only CLI, not the graphical Desktop.

## Windows

1. Open the official Desktop/full-node v1.7.14 release: https://github.com/promirmir/AuronQ/releases/tag/v1.7.14
2. Download `AuronQ-1.7.14-Windows-x64.zip`.
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

## Audit status

AuronQ requires an independent security audit. [Security policy](SECURITY.md).

Verify checksums and preserve wallet backups. For Android, [install 0.5.9](android/README.md) without removing an older funded wallet.

See [SECURITY.md](SECURITY.md), [MAINNET-CHANGE-POLICY.md](MAINNET-CHANGE-POLICY.md) and [STABILIZATION.md](STABILIZATION.md).
