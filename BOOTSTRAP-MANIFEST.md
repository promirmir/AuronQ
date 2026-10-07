# AuronQ bootstrap manifest — Mainnet

A bootstrap manifest is discovery metadata. It is **not consensus** and changing it does not alter Network ID, genesis, balances, supply or proof-of-work rules.

Official discovery manifests:

- reviewed fallback: `https://raw.githubusercontent.com/promirmir/AuronQ/main/bootstrap.json`
- automatic live registry: `https://raw.githubusercontent.com/promirmir/AuronQ/automation/peer-registry/bootstrap.json`

The reviewed fallback changes through protected-main review. The live registry is produced by the peer crawler and is health-checked before its branch is published. Both are untrusted discovery metadata; neither has consensus authority.

The peer list is intentionally dynamic. It can contain project rendezvous endpoints and crawler-verified public nodes learned from the live network, so documentation should not hardcode a fixed peer count.

Example shape:

```json
{
  "network_id": "44e62c2ace002a6660c14e252173c1aa303529c68e40c998e92da2b453f44f30b1e58c94d533587e2186004593fb856c433fcdb5418ed430ec8617e29529365c",
  "peers": [
    "https://seed.example.org"
  ],
  "expires_at": 0
}
```

Manifest validation rules:

- remote manifest URLs configured in `network.json` must use HTTPS;
- manifest `network_id` must exactly match the loaded network;
- manifest size and peer count are bounded;
- `expires_at = 0` means no explicit expiry;
- literal private/loopback/CGNAT addresses from a remote manifest are rejected;
- newly learned automatic-registry peers require independent gossip endorsements from at least two distinct peer netgroups;
- DNS peer names learned from a remote manifest require HTTPS;
- manifests are refreshed periodically;
- failing learned peers can be pruned and later replaced.

For Windows releases, a `bootstrap.json` beside `AuronQ-Desktop.exe` uses the same shape. It is only accepted after the bundled network has been validated and the Network ID matches.

A healthy network should have multiple independently operated bootstrap paths and, where possible, independent DNS seeds. The live registry reduces manual maintenance but does not by itself prove operator or hosting-provider diversity. No seed or manifest is a consensus authority.
