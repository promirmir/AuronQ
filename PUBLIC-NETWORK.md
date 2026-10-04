# AuronQ public network

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

At present the manifest advertises two externally verified public endpoints:

- `https://mir.taild63f46.ts.net`
- `https://desktop-4nifg1j.taild63f46.ts.net`

These are rendezvous paths to ordinary full nodes. They cannot create coins, approve invalid blocks or override cumulative-work selection.

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

The crawler never changes consensus and never makes a peer trusted for blocks: every full node still validates the chain locally. It also deliberately preserves existing manifest entries during transient outages instead of deleting the registry.

At present there are two confirmed public bootstrap endpoints, but they are still operated within the same project infrastructure. The network is therefore **not yet operationally independent** for first-time installs. Independence is achieved in practice after multiple independently operated, publicly reachable nodes have been learned/published. Existing nodes can already continue with persisted/gossiped peers when the original bootstrap is unavailable.

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

Windows source/release packaging includes an opt-in `START-PUBLIC-NODE.cmd`. It opens TCP/18444 in Windows Firewall and starts AuronQ Desktop. A router/public ISP path must still allow inbound TCP/18444; CGNAT normally requires a VPS or public relay. Public reachability is verified by another node before that address is accepted into peer gossip. Once verified and gossiped, the hourly registry crawler can discover the endpoint and add it to the public bootstrap registry through the protected PR/CI process.

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

AuronQ 1.7.0 is live mainnet software. It is **not independently audited**. AQM64, consensus/reorg logic, wallet handling and the P2P layer should receive independent review before meaningful real-world value depends on the network.


### Automatic direct-public advertisement

When `--advertise` is not supplied and the node is listening on a public interface, AuronQ now attempts to advertise that directly assigned public IP and listening port automatically. The receiving peer still callback-verifies the endpoint before admitting it to gossip, so detection alone does not make an endpoint trusted.

This primarily helps VPS/server operators whose public IP is assigned directly to a network interface. Nodes behind home NAT/CGNAT still need port forwarding plus an explicit `--advertise` URL, or another public relay mechanism.
