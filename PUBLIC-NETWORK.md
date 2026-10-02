# AuronQ 1.7.0 public network

AuronQ Mainnet is a permissionless full-node network. A normal user downloads the release, starts the client and participates through ordinary outbound P2P connections. There is no central blockchain database or privileged validation server.

## Discovery sources

A node can use four discovery sources:

- persisted peers learned during earlier sessions;
- fixed `seed_peers`;
- `dns_seeds`;
- bounded HTTPS `bootstrap_manifests`.

The official mainnet manifest is:

`https://raw.githubusercontent.com/promirmir/AuronQ/main/bootstrap.json`

Bootstrap metadata is **not consensus** and is excluded from Network ID. It can therefore be rotated without a hard fork.

At present the manifest advertises one confirmed public endpoint:

`https://mir.taild63f46.ts.net`

This is a rendezvous path to an ordinary full node. It cannot create coins, approve invalid blocks or override cumulative-work selection.

## First-start flow

1. Validate the bundled `network.json`.
2. Load persisted/configured peers.
3. Fetch the HTTPS bootstrap manifest and require an exact Network ID match.
4. Contact available peers and compare chain metadata.
5. Download blocks in bounded batches and validate each block locally.
6. Learn verified public peer addresses through gossip.
7. Persist useful public peers for later restarts.

A node behind NAT/CGNAT participates normally through outbound connections. A directly reachable node may be callback-verified and admitted to public gossip.

## Availability boundary

A brand-new node always needs at least one reachable rendezvous path or previously known peer. This is also why mature cryptocurrency networks operate multiple independent seeds.

AuronQ currently has one confirmed bootstrap path. The next operational priority is to add independent public nodes on different networks/providers and DNS seeds. If the only bootstrap is offline, already connected/previously seeded nodes can continue operating, but a completely fresh installation with no known peers may be unable to find the network until a rendezvous path returns.

## P2P behavior

- hello exchanges Network ID, chain metadata and a bounded peer list;
- learned gossip rejects loopback, RFC1918/private, CGNAT, link-local and documentation-only literal addresses;
- peer storage applies basic network-group diversity;
- synchronization is bounded and rotated across peers;
- candidate chains are validated locally and selected by cumulative work;
- transactions and blocks are pushed to peers and recovered by periodic sync;
- same-branch lagging peers can be repaired by pushing missing validated blocks;
- request/body/concurrency limits and basic per-IP throttles reduce obvious resource-exhaustion attacks.

These controls do not replace professional adversarial review or DDoS/Sybil testing.

## Operator example

A directly reachable server may run:

```bash
./auronq node \
  --network ./network.json \
  --data /var/lib/auronq \
  --listen 0.0.0.0:18444 \
  --peer-store /var/lib/auronq/public-peers.json \
  --advertise http://PUBLIC_IP:18444
```

Discovery metadata can be changed without changing Network ID.

## Mainnet security status

AuronQ 1.7.0 is live mainnet software. It is **not independently audited**. AQM64, consensus/reorg logic, wallet handling and the P2P layer should receive independent review before meaningful real-world value depends on the network.
