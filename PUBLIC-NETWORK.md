# AuronQ public network

AuronQ Mainnet is a permissionless full-node network. A normal user downloads the release, starts the client and participates through ordinary outbound P2P connections. There is no central blockchain database or privileged validation server.

## Discovery sources

A node can use four discovery sources:

- persisted peers learned during earlier sessions;
- fixed `seed_peers`;
- `dns_seeds`;
- bounded HTTPS `bootstrap_manifests`.

AuronQ uses two official HTTPS discovery manifests:

- reviewed fallback: `https://raw.githubusercontent.com/promirmir/AuronQ/main/bootstrap.json`
- live supplemental registry: `https://raw.githubusercontent.com/promirmir/AuronQ/automation/peer-registry/bootstrap.json`

Bootstrap metadata is **not consensus** and is excluded from Network ID. It can therefore be rotated without a hard fork.

The reviewed manifest is the durable fallback. Releases also carry multiple reviewed public seed hints in `network.json`, so first contact does not require GitHub to be online. The live registry is regenerated from reachable Mainnet gossip peers and health-checked before publication, so a fresh release can learn newly reachable public nodes without waiting for a manual protected-main merge. It may contain project-operated HTTPS rendezvous endpoints together with crawler-verified public IP nodes learned from the live network.

Do not copy a fixed peer count from this document: the registry can change as public nodes appear, disappear or fail verification. Every listed endpoint is still only a rendezvous path to an ordinary full node; it cannot create coins, approve invalid blocks or override cumulative-work selection.

## First-start flow

1. Validate the bundled `network.json`.
2. Load persisted/configured peers.
3. Fetch the HTTPS bootstrap manifest and require an exact Network ID match.
4. Contact available peers and compare chain metadata.
5. Download blocks in bounded batches and validate each block locally.
6. Learn verified public peer addresses through gossip.
7. Persist useful public peers for later restarts.

A node behind NAT/CGNAT participates normally through outbound connections. A directly reachable node may be callback-verified and admitted to public gossip.

## Availability and independence

AuronQ does not have a founder/master node. Once full nodes know each other, they persist verified public peers locally, exchange those peers in `/p2p/hello`, reconnect after restarts, synchronize by cumulative work and continue validating/mining even if the original bootstrap computer is offline.

A completely fresh installation still needs at least one discovery route. To reduce dependence on manual operator maintenance, the repository now contains an autonomous peer-registry crawler. GitHub Actions runs the registry crawler hourly. It starts from the current `bootstrap.json`, follows AuronQ peer gossip, callback-checks reachable candidates, requires the exact Mainnet Network ID/protocol, rejects private/CGNAT/documentation addresses, applies network-group diversity (including grouping DNS peers by parent domain) and appends verified public peers to the bootstrap manifest.

The crawler never changes consensus and never makes a peer trusted for blocks: every full node still validates the chain locally. A newly learned peer must be reachable, report the exact Mainnet Network ID and be advertised by at least two distinct peer netgroups before it can enter the automatic live registry; self-advertisement does not count as an independent endorsement. Before the live registry is published, the normal public-network health checker verifies height coherence, tip agreement at the top height and common canonical history among reachable peers. The reviewed `main` manifest remains an independent fallback.

The public registry now includes project rendezvous endpoints and crawler-verified public nodes learned from the network. That improves first-contact resilience, but raw endpoint count is not the same as operator independence. The stronger target is sustained diversity across independently controlled operators, networks and discovery routes. Existing nodes can already continue with persisted/gossiped peers when any particular bootstrap endpoint is unavailable.

AuronQ also supports DNS seeds. Adding independently operated DNS seeds later gives a second discovery mechanism that does not depend on the repository manifest.

## Public explorer

A public AuronQ Explorer is currently reachable at:

`https://mir.taild63f46.ts.net/explorer`

It is served directly by an AuronQ full node and exposes read-only blockchain data. The explorer can be used to inspect recent blocks, block details, transactions, addresses, balances, transaction history and live network telemetry. It has no consensus privileges and cannot modify blockchain state.

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

## Becoming an independent public node

Public reachability is now treated as a **full-node lifecycle function**, not a mining feature. AuronQ Desktop and the Windows Universal Miner automatically try to keep TCP/18444 reachable for as long as their full node is running. The portable CLI defaults to `--auto-public=true`.

The automatic path is:
1. use a directly assigned public IPv4/IPv6 address when one is available;
2. otherwise try UPnP/IGD TCP mapping for the node listen port;
3. advertise the endpoint;
4. let remote peers callback-verify it before accepting it into public gossip;
5. retry periodically while the node remains outbound-only.

Stopping a miner no longer closes the public-node mapping. Stopping the full node does.

A router/public ISP path must still allow inbound TCP/18444. If the router has no public WAN address (typical CGNAT), UPnP cannot create Internet reachability and the node correctly remains outbound-only. Windows Firewall may also need to allow the AuronQ executable on the active network profile. Public reachability is verified by another node before that address is accepted into peer gossip. Once verified and gossiped, the hourly registry crawler can discover the endpoint and publish it to the supplemental live registry automatically after the health gate. A reviewed snapshot can still be promoted to protected `main`, but it is no longer on the critical path for fresh-node discovery.

Normal users do not need inbound connectivity and can keep using `START-AURONQ.cmd`.

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

AuronQ Mainnet is live public software. It is **not independently audited**. AQM64, consensus/reorg logic, wallet handling and the P2P layer should receive independent review before meaningful real-world value depends on the network.


### Automatic direct-public advertisement

When `--advertise` is not supplied and the node is listening on a public interface, AuronQ now attempts to advertise that directly assigned public IP and listening port automatically. The receiving peer still callback-verifies the endpoint before admitting it to gossip, so detection alone does not make an endpoint trusted.

This directly helps VPS/server operators whose public IP is assigned to an interface. Home NAT users can now become public automatically when their router supports UPnP/IGD and its WAN address is globally routable. CGNAT users still need a public ISP address, IPv6 reachability, a manual/public relay, or equivalent operator-controlled ingress.
