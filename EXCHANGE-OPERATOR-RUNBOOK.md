# AuronQ (AURQ) — Independent Exchange Operator Runbook

**Purpose:** a self-service technical evaluation path for exchanges, custodians, indexers and infrastructure operators. **No project approval, founder contact, token allocation or listing payment is needed to read the protocol or build an independent integration.** A decision to offer trading, and its regulatory obligations, belongs to the operator.

**Important:** AuronQ is a new, experimental Mainnet without an independent professional security/consensus audit. This document is **not** a readiness certification, regulatory opinion, recommendation to list AURQ or guarantee that deposits/withdrawals are safe. See [security policy](SECURITY.md), [stabilization gate](STABILIZATION.md) and [existing unresolved reports](https://github.com/promirmir/AuronQ/issues).

## 1. Verify what you are connecting to

- Native UTXO Layer-1 coin **AURQ**; not an ERC-20 token. Do **not** substitute a same-symbol token contract.
- Consensus: **AQM64** Proof of Work; transaction signatures: **ML-DSA-87**.
- Mainnet Network ID: `44e62c2ace002a6660c14e252173c1aa303529c68e40c998e92da2b453f44f30b1e58c94d533587e2186004593fb856c433fcdb5418ed430ec8617e29529365c`
- Genesis hash: `5750a455c04bfe93c9edfef1a12744b05e29ac6da1a9dd5b790566629dea2080c581beb2f0324efba2067c9efb113ed7a29598fd3ffbed965ced31f265d0cec4`
- Nominal supply cap: **21,000,000 AURQ**, including a publicly disclosed **210,000 AURQ founder genesis allocation (1%)**.
- Coinbase maturity: **100 blocks** (not a guarantee that 100 confirmations are sufficient for third-party exchange risk).

Always verify against the [canonical Mainnet specification](MAINNET.md), [protocol](PROTOCOL.md), [AQM64 specification](AQM64.md) and implementation. A website, explorer, seed, API provider or this guide cannot determine consensus truth.

## 2. Operate and independently validate your own full node

1. Review and obtain source or a versioned release from the [public repository](https://github.com/promirmir/AuronQ). Check the published checksum, source revision and provenance where available. Do not treat Alpha binaries as audited.
2. Follow [Start here](START-HERE.md) and the [public node instructions](PUBLIC-NETWORK.md) to run an **operator-controlled full node** with durable storage and secure backups.
3. Keep the node's HTTP interface private or tightly firewalled. The same service exposes both read-only endpoints and transaction/block submission endpoints; **do not expose it as an unrestricted custody or mining backend**.
4. Confirm the local node has independently validated the chain, matches the **exact Network ID** above, and agrees with other **independently operated** full nodes on a common chain history and tip. Matching two public explorer pages is not independent validation.
5. Test loss of all original project-operated rendezvous hosts and recovery with fresh peer discovery. This is an **outstanding operational verification goal**, not a guarantee. See [network independence](NETWORK-INDEPENDENCE.md).

Read-only check against **your own** running node (the port is the default local HTTP/P2P interface):

```sh
curl --fail --show-error http://127.0.0.1:18444/v1/status
curl --fail --show-error 'http://127.0.0.1:18444/p2p/blockhash?height=0'
curl --fail --show-error 'http://127.0.0.1:18444/v1/explorer/blocks?limit=10'
```

These checks are **diagnostics only**. A returned HTTP 200 or matching reported height does not prove the node has safely validated all prior history. Check local logs, persisted state and consensus-validation behavior.

## 3. Current integration interfaces

These paths are implemented in `internal/auronq/node.go` and are **AuronQ-specific HTTP/JSON endpoints, not Bitcoin Core JSON-RPC**:

| Method | Endpoint | Intended use |
| --- | --- | --- |
| `GET` | `/v1/status` | Local network ID, height, tip, chainwork and node diagnostics |
| `GET` | `/p2p/blockhash?height=N` | Block hash by height for locally verified chain tracking |
| `GET` | `/p2p/getblock?height=N` | Full block for a given height |
| `GET` | `/v1/explorer/block?height=N` or `?hash=...` | Block inspection |
| `GET` | `/v1/explorer/tx?txid=...` | Transaction inspection |
| `GET` | `/v1/balance?address=...` | Address balance query |
| `GET` | `/v1/utxos?address=...` | Address UTXO query |
| `GET` | `/v1/history?address=...&limit=100` | **Bounded** address history (maximum limit 250, not a paginated full-chain deposit index) |
| `POST` | `/v1/tx` | Submit an **already signed** transaction encoded as the native JSON `Transaction` |
| `GET` | `/v1/template?address=...` | Mining template, not an exchange deposit interface |
| `POST` | `/v1/block` | Submit a solved block, not a transaction submission endpoint |

Do not assume the node offers `getblocktemplate`, `sendrawtransaction`, `gettransaction` or a Bitcoin-compatible wallet RPC. Review actual request and response structs in the source before using any endpoint.

## 4. Deposits, withdrawals and chain reorganizations

A production integration needs an operator-owned **persistent block/transaction index** and deterministic recovery. Do not rely on the bounded `/v1/history` response, a third-party explorer, or a peer-reported balance as the sole source for crediting customer deposits.

Before enabling deposits or withdrawals, demonstrate:

- [ ] Exact Mainnet identity and reproducible full-node verification from genesis.
- [ ] Independently operated node(s), monitored for sync state and divergence, with no founder-host dependency.
- [ ] A durable, idempotent ingestion process that records block height **and hash** and tracks outputs to operator-controlled deposit addresses.
- [ ] Confirmation counting using the **currently validated canonical chain**, not wall-clock time, pool reports or only an API response.
- [ ] Reorganization rollback and replay (including same-height tip replacements) without double-crediting or losing deposits.
- [ ] Restart, storage-corruption, node-loss and network-partition drills with reconciliation before reopening withdrawals.
- [ ] Correct atomic-unit accounting, fee handling, dust/minimum rules, signing, change outputs and replay/duplicate-send protection against real implementation vectors.
- [ ] Coinbase rewards are not treated as spendable before **100-block maturity**.
- [ ] Hot-key custody, withdrawal limits, backups and audit logging designed and reviewed by the operator; no private keys or wallet files uploaded to public issues.
- [ ] End-to-end signed-transaction acceptance **and later chain confirmation** tested on a safe environment; mempool acceptance is **not** settlement or irreversible finality.
- [ ] Confirmation and custody policies reflect the current hashrate concentration and deep-reorganization risk; no fixed number is asserted universally safe.
- [ ] Independent legal, sanctions/AML, consumer-protection, operational and cybersecurity review completed where applicable.

**Release hold point:** Do not switch on customer deposits/withdrawals just because an RPC works or a listing contact agrees. First resolve or accept with documented mitigations any release-blocking wallet/security issues, test custody recovery and record sign-off for the exact node/release version.

**Current public reports to review (12 October 2026):** [desktop wallet unlock/send failure #134](https://github.com/promirmir/AuronQ/issues/134) and [Android fresh-install scalability blocker #149](https://github.com/promirmir/AuronQ/issues/149). These are unresolved engineering reports, **not proof of a consensus failure**. An exchange must independently validate its own wallet custody stack rather than assuming user-facing wallet issues are irrelevant.


## 5. Autonomous integration and disclosure

An independent exchange may assess and integrate AURQ **without obtaining a signature, letter, approval, allocation or payment from the AuronQ project**. Public source, canonical network identity and independently validated nodes are the technical reference points.

This open technical policy **does not** grant a legal exemption, establish that AURQ has no identifiable issuer, or waive operator obligations. The project's **own** listing applications and negotiations are [paused pending legal review](LISTING-STATUS.md); that does not place a network-level restriction on third-party integration.

A third-party operator should disclose that it is independently operated, identify any custody and listing risks accurately, and avoid claims of project endorsement or an independent security audit.

For reproducible technical defects, use [public GitHub issues](https://github.com/promirmir/AuronQ/issues) **without secrets, customer data or credentials**. For security vulnerabilities, follow the private reporting process in [SECURITY.md](SECURITY.md). Neither route is a prerequisite for independent integration.
