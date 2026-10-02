# AuronQ bootstrap manifest — 1.7.0-rc4

A bootstrap manifest is optional discovery metadata. It is **not consensus** and changing it does not alter Network ID, genesis, balances, supply or proof-of-work rules.

Example:

```json
{
  "network_id": "44e62c2ace002a6660c14e252173c1aa303529c68e40c998e92da2b453f44f30b1e58c94d533587e2186004593fb856c433fcdb5418ed430ec8617e29529365c",
  "peers": [
    "https://seed.example.org",
    "http://203.0.113.10:18444"
  ],
  "expires_at": 0
}
```

Rules in rc3:

- remote manifest URLs configured in `network.json` must use HTTPS;
- the manifest `network_id` must exactly match the loaded AuronQ network;
- a manifest is limited to 64 KiB and 64 peers;
- `expires_at` is an optional Unix timestamp; zero means no explicit expiry;
- literal private/loopback/CGNAT addresses from a remote manifest are rejected;
- DNS peer names from a remote manifest are accepted only over HTTPS;
- failing manifest peers are pruned normally and can be replaced on later refreshes;
- manifests are refreshed periodically.

For Windows releases, a `bootstrap.json` file beside `AuronQ-Desktop.exe` uses the same JSON shape. It is merged locally only after the bundled `network.json` has been validated and only when the Network ID matches.

The long-term public network should use several independently operated bootstrap paths. A single manifest host or seed must never be treated as consensus authority.
