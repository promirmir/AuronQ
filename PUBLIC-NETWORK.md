# AuronQ public network deployment

AuronQ 1.7.0-rc3 keeps ordinary Internet P2P as the network layer. Normal users do not need to configure Tailscale, VPNs, router ports or manual peer addresses.

## What a normal user does

A released `network.json` can carry three replaceable bootstrap mechanisms:

The official AuronQ mainnet manifest is published at `https://raw.githubusercontent.com/promirmir/AuronQ/main/bootstrap.json`. It is discovery metadata, not a consensus authority.

- `dns_seeds`: DNS names whose A/AAAA answers point at currently reachable public AuronQ nodes;
- `seed_peers`: fallback HTTP/HTTPS peer URLs;
- `bootstrap_manifests`: HTTPS JSON documents containing the expected Network ID plus a bounded peer list.

The Windows ZIP may additionally include `bootstrap.json` beside `AuronQ-Desktop.exe`. On launch rc3 merges that list into the local bootstrap cache only when its Network ID exactly matches the loaded mainnet.

A fresh node uses those entries only to find its first peers. It verifies the immutable AuronQ Network ID, synchronizes the greatest-work chain and mempool, learns additional public peers through P2P gossip, stores a validated peer cache, and uses that cache on later starts.

The bootstrap list is discovery infrastructure, not consensus. A seed cannot create valid coins, rewrite blocks, approve transactions, or override greatest-work chain selection.

## Automatic participation

Every running Desktop instance is a full validating node. It can make outbound P2P requests, relay transactions/blocks, mine, and learn peers without port forwarding.

AuronQ 1.7.0-rc3 also automatically announces its listen port to connected peers. The receiving peer derives the sender's observed public source IP and attempts a callback to that IP and port. Only if the callback succeeds and the Network ID matches is the address admitted to public peer gossip. This lets directly reachable nodes become discoverable without asking the user to type `--advertise`.

A node behind NAT/CGNAT with no inbound mapping still participates through outbound P2P. It simply cannot be used as an inbound bootstrap target until TCP 18444 becomes publicly reachable.

## Minimum real deployment

For a public testnet/mainnet release, publish several independent entry points. A practical minimum is three reachable nodes operated on different hosts/providers, plus at least two DNS seed names controlled independently if possible. Each public node runs the same `network.json` and exposes TCP 18444.

Example public node:

```bash
./auronq node \
  --network ./network.json \
  --data /var/lib/auronq \
  --listen 0.0.0.0:18444 \
  --peer-store /var/lib/auronq/public-peers.json
```

`--advertise http://PUBLIC_IP:18444` remains available as an explicit override for server operators, but it is no longer required when peers can observe and verify the node's public address automatically.

Configure bootstrap metadata without changing Network ID:

```bash
./auronq network-set-seeds \
  --network ./network.json \
  --seed http://seed1.example.org:18444,http://seed2.example.org:18444 \
  --dns-seed dnsseed1.example.org,dnsseed2.example.org \
  --bootstrap-manifest https://example.org/auronq/bootstrap.json
```

`seed_peers`, `dns_seeds` and `bootstrap_manifests` are intentionally excluded from the Network ID, so discovery infrastructure can be rotated without a consensus fork. Manifest entries are accepted only when the manifest carries the exact Network ID; DNS peer endpoints learned from a manifest must use HTTPS. Publish the resulting public `network.json` with the GitHub release.

## Peer discovery and relay

- static seeds may use DNS names or public IPs;
- DNS seeds are resolved periodically and may return multiple IPv4/IPv6 addresses;
- P2P hello exchanges Network ID, tip/work and a bounded public peer list;
- directly reachable nodes can be learned automatically from the source IP observed by a peer;
- learned gossip rejects loopback, RFC1918/private, CGNAT, link-local and documentation-only addresses;
- discovered public peers are persisted locally;
- repeatedly failing learned peers are pruned while configured bootstrap entries remain available;
- blocks and transactions are pushed to known peers and also recovered by periodic chain/mempool synchronization.

## Decentralization boundary

No Internet protocol can let a brand-new installation discover an unknown private network with literally zero prior rendezvous information. AuronQ therefore uses multiple replaceable bootstrap paths, then moves discovery to ordinary peers. The goal is that a user experience is simply: download, run, sync — while no bootstrap operator has consensus authority.

## GitHub release contents

Publish source, `network.json`, release binaries, SHA-256 hashes, protocol/security documents and reproducible build instructions. Never publish the founder wallet, its password, private seed material, or local node data.

## Security status

This is suitable for continued public-testnet engineering, not a claim of audited mainnet security. Before attaching meaningful real-world value, obtain independent review of consensus/reorg logic, difficulty/timestamp manipulation, P2P eclipse/Sybil/DoS resistance, wallet/key handling, serialization and ML-DSA implementation. Long-running adversarial testnet operation is required.

## Public-network hardening retained from 1.5.0

The public peer set now applies basic network-group diversity to learned peers, so one IPv4 /16 or IPv6 /32 cannot trivially occupy the whole peer table. Sync work is rotated across a bounded number of peers per round, long initial synchronization is performed in bounded block batches, and small metadata responses use tighter read limits. Mempool memory is policy-bounded and valid transactions disconnected by a higher-work reorganization are revalidated and returned to the mempool.

These controls reduce obvious resource-exhaustion and eclipse failure modes but do not replace a professional P2P security review. A production network still needs adversarial distributed-DoS testing, stronger peer-quality scoring/diversity, headers-first synchronization, and long-running public-network observation. rc3 retains the rc2 per-IP request and block-submission rate limiting, but that is not sufficient against botnets or large Sybil sets.