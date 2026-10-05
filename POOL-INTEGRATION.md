# AuronQ mining pool integration

This document is for independent mining-pool operators evaluating support for **AuronQ (AURQ)**.

The goal is to make pool integration possible **without changing Mainnet consensus**. AuronQ Mainnet consensus parameters, Network ID, genesis, monetary rules and AQM64 are not part of the pool-integration surface.

> AuronQ is experimental financial software. AQM64 and the complete implementation have not received an independent professional security or cryptographic audit.

## Network summary

| Parameter | Value |
|---|---|
| Coin | AuronQ |
| Symbol | AURQ |
| Ledger | UTXO |
| Proof of Work | AQM64 |
| PoW algorithm ID | `0x0001` |
| Block header version | `2` |
| Target block interval | 600 seconds |
| Target width | 512 bits, big-endian |
| Coinbase maturity | 100 blocks |
| Mainnet Network ID | `44e62c2ace002a6660c14e252173c1aa303529c68e40c998e92da2b453f44f30b1e58c94d533587e2186004593fb856c433fcdb5418ed430ec8617e29529365c` |
| Genesis hash | `5750a455c04bfe93c9edfef1a12744b05e29ac6da1a9dd5b790566629dea2080c581beb2f0324efba2067c9efb113ed7a29598fd3ffbed965ced31f265d0cec4` |

Canonical references:

- [AQM64.md](AQM64.md)
- [PROTOCOL.md](PROTOCOL.md)
- [MAINNET.md](MAINNET.md)
- [MAINNET-CHANGE-POLICY.md](MAINNET-CHANGE-POLICY.md)

## Current node integration surface

AuronQ does **not** currently expose Bitcoin-style JSON-RPC `getblocktemplate` / `submitblock` and does **not** include a native Stratum server.

The current full node exposes a simple HTTP/JSON mining interface. A pool should normally run its own AuronQ full node locally and place its Stratum/pool adapter in front of it.

### 1. Check node status

```http
GET /v1/status
```

Relevant response fields include:

```json
{
  "network": "...",
  "network_id": "...",
  "height": 123,
  "tip": "...",
  "chain_work": "...",
  "mempool": 0,
  "peers": 8,
  "network_hashrate": 1234.5
}
```

Before serving work, the pool should verify that `network_id` exactly matches AuronQ Mainnet.

The `network_hashrate` field is an estimate derived from recent solved blocks. It is not an authoritative measurement and can lag or differ materially from instantaneous pool-side hashrate.

### 2. Request a complete block template

```http
GET /v1/template?address=<AURQ_POOL_REWARD_ADDRESS>
```

The node validates the reward address, selects mempool transactions, calculates fees and subsidy, creates the coinbase transaction, calculates the next consensus target and returns a complete JSON `Block`.

Simplified structure:

```json
{
  "header": {
    "version": 2,
    "pow_algo": 1,
    "height": 124,
    "prev_hash": "...",
    "merkle_root": "...",
    "timestamp": 0,
    "target": "...",
    "nonce": 0
  },
  "transactions": [
    {
      "version": 1,
      "coinbase": true,
      "coinbase_height": 124,
      "outputs": []
    }
  ]
}
```

The returned coinbase pays `block subsidy + included transaction fees` to the address supplied by the pool.

The simplest and safest integration is to treat the node-produced transaction set, coinbase, Merkle root, target and previous hash as authoritative template data and vary only mining work fields required by the pool implementation.

## AQM64 proof of work

AQM64 operates on the **canonical binary block header**, not on JSON.

The canonical version-2 header is:

```text
"AURONQ_BLOCK_V2\0"
version:u16be
pow_algo:u16be
height:u64be
previous_hash:64
merkle_root:64
timestamp:i64be
target:64
nonce:u64be
```

AQM64 v1:

```text
pre  = SHAKE256-512("AURONQ_AQM64_PRE_V1\0" || canonical_header)
salt = SHAKE256-256("AURONQ_AQM64_SALT_V1\0" || previous_hash || height:u64be || pow_algo:u16be)
mid  = Argon2id-v1.3(pre, salt, memory=65536 KiB, time=2, parallelism=1, out=64)
pow  = SHAKE256-512("AURONQ_AQM64_FINAL_V1\0" || pre || mid)
```

A network-valid block satisfies:

```text
unsigned_big_endian_512(pow) <= unsigned_big_endian_512(header.target)
```

Canonical implementations are in:

- `internal/auronq/pow.go`
- `internal/auronq/types.go`
- `gpu/cuda/`

Pool operators should reproduce AQM64 against the repository's deterministic/reference behavior before accepting miner shares.

## Share difficulty

The AuronQ node validates **network block candidates**, not pool shares.

A pool therefore defines its own share target/difficulty locally. The pool adapter should:

1. construct or distribute work derived from a current AuronQ block template;
2. evaluate every submitted share with AQM64;
3. compare the resulting 512-bit PoW value with the pool's share target;
4. if the same result also satisfies the template's network target, assemble the complete candidate block and submit it to the full node.

Share accounting, vardiff and miner payout accounting remain pool-side policy and are not consensus rules.

## 3. Submit a network block candidate

```http
POST /v1/block
Content-Type: application/json
```

Body: the complete JSON `Block` returned from the template path, with the solved header fields.

On success, the node returns:

```json
{
  "hash": "...",
  "height": 124
}
```

The full node independently validates the candidate, including:

- block linkage;
- height;
- version and PoW algorithm;
- AQM64;
- target/difficulty;
- timestamps;
- Merkle root;
- transactions;
- UTXO spends;
- signatures;
- fees;
- subsidy and coinbase amount;
- supply rules.

The pool must never treat its own share validator as a substitute for full-node validation.

## Stale-work handling

A template is current only while both are true:

```text
status.height == template.header.height - 1
status.tip    == template.header.prev_hash
```

A change in either value makes the job stale.

The reference remote miner checks `/v1/status` while mining and cancels a template when the canonical tip changes, including same-height reorganizations.

Pool software should do the same. Until a push/long-poll mining interface exists, polling approximately once per second is a reasonable compatibility model and matches the current reference miner's default behavior.

Pools should also refresh work periodically even without a tip change so that timestamp and mempool contents do not become unnecessarily old.

## Timestamp handling

The node-generated template timestamp already satisfies the current median-time-past rule when it is created.

Consensus requires a block timestamp to be:

- strictly greater than the median of the previous 11 blocks; and
- no more than two hours ahead of the validating node's adjusted wall clock.

A pool can avoid unnecessary complexity by regularly requesting fresh templates instead of aggressively rewriting timestamps.

If a pool changes the header timestamp itself, that changed timestamp is part of the AQM64 input and therefore requires new work.

## Coinbase and payouts

`GET /v1/template?address=...` creates the block coinbase for the supplied AuronQ address.

A simple pool deployment can use one pool-controlled reward address for block rewards and perform miner payouts later as ordinary AURQ transactions after coinbase maturity.

Mainnet coinbase maturity is **100 blocks**.

Payout scheduling, minimum payout thresholds, fees, PPLNS/PPS policy and accounting are pool policy, not AuronQ consensus.

The node exposes:

```http
POST /v1/tx
```

for broadcasting an already signed transaction. Private-key custody and payout signing should remain under the pool operator's control.

## Recommended pool architecture

```text
miners
   |
   |  Stratum / pool protocol
   v
independent pool server
   |
   |  local share validation / vardiff / accounting
   |  AQM64
   v
AuronQ pool adapter
   |
   |  HTTP/JSON
   |  /v1/status
   |  /v1/template
   |  /v1/block
   |  /v1/tx
   v
local AuronQ full node
   |
   v
AuronQ P2P network
```

For production use, the pool should operate its own full node and should not depend on a public community node for template production or candidate submission.

## Public endpoint rate limiting

The standard node applies request and block-validation rate limits to public HTTP endpoints. A production pool should place its adapter next to its own node and avoid routing high-volume share traffic through AuronQ's node API.

Only network-valid block candidates need to reach `POST /v1/block`; ordinary shares should be verified entirely by pool software.

## What is not currently provided

As of the current Mainnet implementation, AuronQ does not provide:

- native Stratum V1/V2;
- Bitcoin-compatible `getblocktemplate`;
- Bitcoin-compatible `submitblock`;
- a pool share database;
- vardiff;
- pool payout accounting;
- a remote hot-wallet signing RPC.

These are intentionally outside consensus.

If multiple independent pool operators need the same additional interface, the preferred path is a **small, documented, non-consensus compatibility API** rather than changes to consensus or AQM64.

## Compatibility policy

Pool integration must not require ordinary Mainnet nodes to change:

- genesis;
- Network ID;
- AQM64;
- difficulty rules;
- block/transaction serialization;
- monetary policy;
- chain-selection rules;
- signature validation.

Any proposed compatibility change should be limited to transport, RPC, observability or adapter tooling and must preserve validation behavior.

See [MAINNET-CHANGE-POLICY.md](MAINNET-CHANGE-POLICY.md).

## Integration checklist

Before announcing AURQ support, a pool operator should verify:

- [ ] the node reports the canonical Mainnet Network ID;
- [ ] a reward address is accepted by `/v1/template`;
- [ ] the returned template extends the current canonical tip;
- [ ] the pool's AQM64 implementation matches the canonical implementation;
- [ ] share-target comparison uses unsigned big-endian 512-bit values;
- [ ] stale jobs are invalidated on both height and tip changes;
- [ ] a locally solved test candidate is accepted by `/v1/block`;
- [ ] rejected candidates are not credited as blocks;
- [ ] coinbase maturity is handled before payouts;
- [ ] the pool runs an independently operated full node;
- [ ] pool infrastructure does not modify Mainnet consensus parameters.

## Contact / coordination

Please use public project channels so integration work remains auditable and does not depend on private contact information:

- GitHub issues: https://github.com/promirmir/AuronQ/issues
- Pool decentralization coordination: https://github.com/promirmir/AuronQ/issues/66
- Bitcointalk ANN: https://bitcointalk.org/index.php?topic=5595868.0
- Project Discord: https://discord.gg/rmmNY9RhA

When requesting node-side changes, include:

1. pool software / Stratum implementation;
2. exact missing RPC or job fields;
3. expected request/response example;
4. whether the change is required for shares, candidate blocks or payouts;
5. why an external adapter cannot provide it.

This makes it possible to evaluate a minimal non-consensus compatibility change without destabilizing Mainnet.
