# Start here — AuronQ 1.7.1 Mainnet

## Windows

1. Open the official release: https://github.com/promirmir/AuronQ/releases/tag/v1.7.1
2. Download `AuronQ-1.7.1-Windows-x64.zip`.
3. Verify SHA-256: `50a1b2d321117549d7ceab81aa5dfaa0bb072544915fb5e921eb9b2a24d68d8d`.
4. Extract the complete ZIP.
5. Run `START-AURONQ.cmd`.

A normal participant does not need Tailscale, Cloudflare, manual peer addresses, Go or Docker.

The client validates the bundled mainnet definition, consults the stable GitHub bootstrap manifest, connects to available peers, validates the chain locally, learns additional public peers and stores them for future restarts.

## Current bootstrap state

The stable manifest is:

`https://raw.githubusercontent.com/promirmir/AuronQ/main/bootstrap.json`

It currently contains one confirmed public rendezvous endpoint, `https://mir.taild63f46.ts.net`. It is not a trusted consensus server; it only gives a fresh node its first contact.

## Security

AuronQ 1.7.1 is live mainnet software but has not received an independent professional consensus/cryptographic security audit. See `SECURITY.md`.
