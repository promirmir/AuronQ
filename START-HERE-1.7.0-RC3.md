# AuronQ 1.7.0-rc3 — zero-touch public-network release candidate

rc3 does **not** change the rc2 consensus rules or the already-created AuronQ Mainnet identity. It improves how a fresh Windows installation discovers its first peers.

For a normal user, the intended flow is:

1. Extract the official Windows ZIP.
2. Double-click `AuronQ-Desktop.exe`.
3. The bundled `network.json` is validated and imported automatically on first launch.
4. A matching `bootstrap.json` and/or bootstrap metadata in `network.json` supplies initial peers automatically.
5. The node verifies Network ID, synchronizes the chain/mempool, learns peers and stores them for later restarts.

No ordinary user should need Tailscale, port forwarding or a manually typed peer. Nodes behind NAT still participate through outbound connections; directly reachable nodes can additionally become inbound peers.

Security status remains unchanged from the rc2 internal audit: this is not an independent professional audit, and AQM64/public P2P still require long-running adversarial review before material real-world value is appropriate.