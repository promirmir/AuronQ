# AuronQ network independence

AuronQ Mainnet has no founder/master node and no privileged validation server.

## What already works without the original computer

Full nodes:

- keep their own complete canonical chain;
- independently validate blocks and transactions;
- select chains by cumulative work;
- persist verified public peers to disk;
- exchange bounded public peer lists through `/p2p/hello`;
- reconnect to known peers after restart;
- relay blocks and transactions directly between peers.

Therefore, once multiple full nodes know one another, the original bootstrap computer is not required for consensus or continued block production.

## Fresh-install discovery

A computer with no peer history still needs at least one rendezvous path. Current release metadata also ships a reviewed multi-peer seed snapshot, so the CLI is not dependent on a successful GitHub manifest fetch for first contact. AuronQ supports:

1. local persisted peers;
2. fixed seed peers;
3. DNS seeds;
4. HTTPS bootstrap manifests;
5. peer gossip after first contact.

Bootstrap metadata is discovery-only and excluded from the Network ID. AuronQ now uses two HTTPS manifest layers: a reviewed static fallback from `main` and a supplemental live registry published by the crawler on `automation/peer-registry`. Either can fail without changing consensus.

## Autonomous registry

`.github/workflows/peer-registry.yml` runs the AuronQ peer crawler periodically. The crawler:

- starts from the current manifest;
- requests `/p2p/hello`;
- requires protocol version 1 and the exact AuronQ Mainnet Network ID;
- follows only publicly routable peers learned from gossip;
- requires a newly learned live-registry peer to be advertised by at least two distinct peer netgroups before automatic promotion;
- treats a node's own advertise field as a self-claim, not an independent endorsement;
- rejects loopback, private, CGNAT, link-local and documentation ranges;
- callback-verifies reachability;
- limits peer count and basic network-group concentration;
- preserves existing registry entries across transient outages;
- runs the public network health checker against the generated registry before publishing it;
- pushes the verified live registry to `automation/peer-registry` when there is a change;
- may also prepare an optional reviewed snapshot for protected `main`, but fresh-node discovery no longer needs to wait for that manual review step.

This process is not a consensus oracle. A malicious or broken peer cannot make an invalid chain valid; every full node validates candidate blocks itself.

## When the original computer can be turned off

The strongest practical test is:

1. have at least 3 independent publicly reachable full nodes on different networks/providers;
2. ensure they appear in gossip and the public bootstrap registry and/or DNS seeds;
3. turn the original bootstrap computer completely off;
4. start a clean AuronQ install with no peer cache;
5. confirm it discovers other peers and reaches the same height/tip;
6. mine/relay another block and confirm propagation without the original computer.

Only after this passes should the network be described as operationally independent of the original bootstrap machine.

The live network now has multiple reachable public full nodes and the crawler is discovering additional peers through ordinary gossip. The remaining independence test is operational rather than protocol-level: perform a clean-install join and block-relay drill while the original project-operated rendezvous machines are completely offline, and continue improving provider/operator diversity.
