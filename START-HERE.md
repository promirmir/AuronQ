# Start here — AuronQ 1.7.0 Mainnet

## Windows

1. Open the official release: https://github.com/promirmir/AuronQ/releases/tag/v1.7.0
2. Download `AuronQ-1.7.0-Windows-x64.zip`.
3. Verify SHA-256:
   `79f9e75b61ad17f26e6e90e3d8dc07883aa3cb00f1f7882bcf96dd6291a72f37`
4. Extract the complete ZIP.
5. Run `START-AURONQ.cmd`.

A normal participant does not need Tailscale, Cloudflare, manual peer addresses, Go or Docker.

The client validates the bundled mainnet definition, consults the stable GitHub bootstrap manifest, connects to available peers, validates the chain locally, learns additional public peers and stores them for future restarts.

## Current bootstrap state

The stable manifest is:

`https://raw.githubusercontent.com/promirmir/AuronQ/main/bootstrap.json`

It currently contains one confirmed public rendezvous endpoint, `https://mir.taild63f46.ts.net`. It is not a trusted consensus server; it only gives a fresh node its first contact.

## Security

AuronQ 1.7.0 is live mainnet software but has not received an independent professional consensus/cryptographic security audit. See `SECURITY.md`.
