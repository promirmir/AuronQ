# AuronQ Changelog

This file is the concise public history of AuronQ. Detailed historical release notes and old checksum files are preserved under [docs/archive/](docs/archive/).

## Recent Android updates

### Android 0.5.3 Alpha

- fresh installs can use independently verified public IPv4 full nodes advertised as `http://<public-ip>:18444`, not only the two HTTPS bootstrap endpoints;
- cleartext mobile peers are accepted only as literal globally routable IP addresses; private, loopback, CGNAT and documentation ranges remain rejected, DNS peers still require HTTPS, and redirects remain disabled;
- bundles the three newly crawler-verified public IPv4 Mainnet peers as direct fallbacks;
- shows “network reachable / verifying AQM64” immediately instead of presenting first-install header verification as a connection failure;
- a temporary refresh failure no longer erases an already verified mobile state or flashes the app back to “offline” every few seconds;
- header verification progress is persisted after every successful batch so a flaky peer or app restart resumes from the last verified height instead of replaying AQM64 from genesis;
- no Mainnet consensus, Network ID, genesis, transaction, PoW or monetary changes.

### Android 0.5.2 Alpha

- prevents temporary node timeouts or a lagging peer from wiping the locally verified header cache and forcing an expensive full AQM64 replay;
- mobile peer probing returns quickly after a valid Mainnet peer answers while keeping a short grace window for multi-peer agreement;
- skips the remote bootstrap-manifest retry when a bundled/learned Mainnet peer is already reachable;
- updates the UI to show a verified network snapshot immediately instead of waiting for slower wallet balance/history quorum requests;
- no Mainnet consensus, Network ID, genesis, transaction, PoW or monetary changes.

## Current releases

### Universal Miner v0.4.6 Alpha

- Adds a packaged CUDA 12.6 legacy backend for supported Maxwell/Pascal/Volta targets, including Pascal `sm_61` used by GTX 1050/1050 Ti/1060/1070/1080-class cards.
- Solo AUTO order is now primary CUDA → legacy CUDA → OpenCL GPU → CPU, with canonical AQM64 self-test validation on every native accelerator.
- Fixes Windows Pool AUTO incorrectly rejecting older NVIDIA GPUs through AuronQ's native CUDA compatibility gate before launching an external pool miner.
- Pool mode now passes locally detected NVIDIA devices to the user-supplied pool miner and lets that miner determine its own CUDA compatibility.
- Fixes misleading GUI labeling that could show CPU before AUTO finished testing legacy CUDA/OpenCL.
- Records the actual worker fallback reason in the Windows GUI.
- Primary CUDA 13.2 remains unchanged for newer NVIDIA generations; no downgrade is required.
- No Mainnet consensus, genesis, Network ID, AQM64, difficulty, monetary policy or validity-rule changes.

### Universal Miner v0.4.5 Alpha

- AUTO now validates and tries NVIDIA CUDA, then vendor-neutral OpenCL GPU, then native CPU AQM64.
- Adds a runtime-loaded OpenCL backend for Windows/Linux ordinary GPUs from AMD, Intel and NVIDIA.
- OpenCL devices are accepted only after local kernel compilation and byte-identical canonical AQM64 self-test.
- CUDA build scripts generate all maintained baseline architecture targets supported by the installed CUDA Toolkit instead of hard-coding only three generations.
- Generic OpenCL thermal-auto uses a conservative approximately 50% compute-duty profile when no trustworthy cross-vendor temperature source exists; no temperature is fabricated.
- Windows Solo GUI adds explicit OpenCL GPU selection while AUTO remains recommended.
- Multi-GPU child workers can carry AUTO/OpenCL fallback for selected device indices.
- Native OpenCL remains a Solo backend; Pool mode continues to depend on the capabilities of the user-supplied external pool miner.
- No dedicated specialized-miner integration is added.
- No Mainnet consensus, genesis, Network ID, AQM64, difficulty, transaction validation or monetary-policy changes.

### Universal Miner v0.4.4 Alpha

- AUTO now validates the real AuronQ CUDA backend with the canonical AQM64 self-test before selecting GPU mining.
- Older/unsupported NVIDIA hardware or CUDA initialization/runtime failures automatically fall back to native CPU mining instead of requiring manual intervention.
- Explicit CUDA/BOTH remains fail-closed.
- Windows dashboard adds real CPU telemetry: processor name, whole-system utilization, logical CPUs, AQM64 threads, nominal clock and 64 MiB-per-lane memory estimate.
- CPU package temperature remains N/A when no trustworthy universal Windows sensor exists; no guessed thermal value is displayed.
- CPU telemetry is visible even when an unsupported NVIDIA GPU is physically present.
- Corrects the GUI/documentation wording for the existing Pool catastrophic thermal envelope.
- No Mainnet consensus, genesis, Network ID, AQM64, difficulty, transaction validation or monetary-policy changes.

### Universal Miner v0.4.3 Alpha

- Reworks Windows Pool-mode thermal regulation to reduce oscillation and repeated suspend/resume cycles on thermally constrained GPUs.
- Starts throttling before the target temperature and changes duty cycle in smaller steps.
- Prevents throttle release while temperature is flat/rising near the target.
- After emergency cooldown, resumes farther below target and holds a conservative duty floor for 20 seconds before gradually restoring performance.
- Allows a small bounded post-suspend thermal overshoot caused by normal thermal inertia, while continued heating still fails closed.
- Direct local NVML telemetry remains the only source used for GPU safety decisions.
- No Mainnet consensus, genesis, Network ID, AQM64, difficulty, transaction validation or monetary-policy changes.

### Universal Miner v0.4.2 Alpha

- Public TCP/18444 is now a **full-node lifecycle service**, independent of mining.
- Stopping Solo/Pool mining no longer removes the node's public UPnP mapping.
- Windows Miner retries Auto Public while the node remains outbound-only and reports CGNAT/UPnP/firewall diagnostics.
- AuronQ Desktop source now follows the same node-scoped public-node lifecycle instead of enabling UPnP only from CPU mining.
- Portable `auronq node` defaults to `--auto-public=true`; operators can explicitly opt out.
- Portable public-node examples use inbound-capable listen addresses; loopback-only listeners are never auto-advertised.
- Public endpoints remain callback-verified by remote peers before entering peer gossip.
- CI syntax-checks both embedded Miner and Desktop JavaScript.
- No Mainnet consensus, genesis, Network ID, AQM64, difficulty, transaction validation or monetary-policy changes.


### Universal Miner v0.4.1 Alpha

- Critical Windows GUI hotfix: repaired an invalid JavaScript string in the CPU fallback card that prevented the embedded dashboard script from parsing.
- Restores dashboard refresh, buttons, settings, mining actions and live status updates in the Windows GUI.
- Adds a mandatory `node --check` gate for the embedded miner JavaScript so syntax-broken GUI releases cannot pass CI again.
- No Mainnet consensus, AQM64 parameters, difficulty, Network ID, genesis, transaction validation or monetary-policy changes.


### Universal Miner v0.4.0 Alpha

- Adds **AUTO** compute selection: official NVIDIA CUDA acceleration when available, otherwise the native AQM64 CPU fallback.
- Adds a reusable CPU AQM64 backend verified against the canonical proof-of-work implementation.
- Conservative automatic CPU profile uses about one quarter of logical CPUs, capped at two 64 MiB lanes; explicit override is capped at 16 lanes.
- Windows GUI adds AUTO / NVIDIA CUDA / CPU modes and clean no-GPU fallback behavior.
- Pool AUTO selects CUDA when a usable NVIDIA GPU is present and otherwise uses CPU with conservative thread defaults.
- Windows NVIDIA thermal safety remains driven by direct local NVML hardware samples; external pool/miner temperature reports are not trusted.
- Linux protected GPU mining now stops if local temperature telemetry disappears instead of silently continuing.
- Portable CPU-safe CLI builds are CI-compiled for Windows x64/ARM64, Linux x64/ARM64 and macOS x64/ARM64, together with the AuronQ full-node CLI.
- Source build scripts can produce usable CPU-safe packages even when the CUDA Toolkit is not installed.
- Ordinary `auronq mine` now defaults to a conservative CPU thread profile instead of consuming every logical CPU.
- No Mainnet consensus, AQM64 parameters, difficulty, Network ID, genesis, transaction validation or monetary-policy changes.


### GPU Miner v0.3.8 Alpha

- Windows thermal safety now uses direct local NVIDIA NVML hardware telemetry instead of temperatures reported by pool/miner software.
- NVML sensors are mapped to CUDA devices through direct CUDA UUID matching; ambiguous multi-GPU mappings fail safe.
- Dashboard and Pool governor share one 500 ms hardware sample and expose source/sample age.
- Missing or stale direct telemetry now stops mining instead of silently continuing.
- Native Windows Solo CUDA thermal control uses the same direct NVML backend.
- Pool catastrophic fail-safe tightened to the configured limit + 1 °C.
- No Mainnet consensus, AQM64, difficulty, Network ID, genesis or monetary-policy changes.


### GPU Miner v0.3.7 Alpha

- Reworked Windows Pool thermal control for autonomous long-running mining.
- NVIDIA telemetry is sampled every 500 ms and external MeshMiner load is regulated with short adaptive duty-cycle pulses.
- Added hysteresis: throttling increases quickly on heat but is released gradually after cooling to prevent full-load / long-pause oscillation.
- Recommended/default thermal limit for new settings is now 81 °C with an automatic target around 76 °C.
- Reaching the configured limit now triggers an automatic suspend/cool/resume cycle instead of immediately requiring a manual restart.
- Added catastrophic fail-safe stop if temperature continues to rise despite suspension, if telemetry is lost repeatedly, or if suspend/resume control fails.
- Updated PL/EN dashboard, documentation and release packaging.
- No Mainnet consensus, AQM64, difficulty, Network ID, genesis or monetary-policy changes.


### GPU Miner v0.3.6 Alpha

- Added official Linux amd64 NVIDIA CUDA package alongside Windows x64.
- Linux supports native AQM64 Solo CUDA mining, multi-GPU, Auto Tune, rolling H/s, self-test, benchmark and smart thermal control.
- Added Linux CUDA shared-library loader and `libauronq-aqm64-cuda.so` package.
- Added Linux CUDA/package build scripts and CI compilation coverage.
- Release workflow now publishes Windows ZIP + Linux tar.gz with one checksum file.
- Added a complete Windows/Linux GPU Miner user guide.
- Windows keeps the PL/EN dashboard, MeshMiner integration, custom pools and external-miner thermal governor.
- No Mainnet consensus, AQM64, difficulty, Network ID, genesis or monetary-policy changes.

### Desktop / Full Node v1.7.13 — Mainnet patch

- Fixed repeated retries of dead DNS peers learned through P2P gossip or bootstrap metadata.
- DNS peers returning a permanent not-found result are removed immediately and quarantined for 30 minutes.
- Repeatedly failing peers are quarantined so gossip/bootstrap refresh cannot instantly resurrect them.
- The public peer-registry crawler now drops existing manifest DNS endpoints that return a permanent DNS not-found result instead of preserving them forever.
- Public peer discovery runs hourly and groups DNS peers by parent domain for better infrastructure diversity.
- Directly addressed public VPS/server nodes can auto-advertise their public interface endpoint when `--advertise` is omitted; peers still callback-verify reachability before gossip admission.
- Windows release packaging now includes `START-PUBLIC-NODE.cmd` for directly reachable independent full nodes.
- Added regression tests for dead-peer quarantine, manifest re-add prevention and DNS peer grouping.
- Consensus, AQM64, difficulty, Network ID, genesis, monetary policy and transaction rules are unchanged.

### Desktop / Full Node v1.7.12 — Mainnet

- Fixed stale mining templates in CLI and Desktop mining.
- Miners now watch the connected node's canonical tip while hashing.
- If another miner advances the chain, obsolete work is cancelled and a fresh template is fetched promptly.
- A same-height reorganization also invalidates the current template through PrevHash comparison.
- Added regression coverage for tip advances, same-height reorgs and caller cancellation.
- Corrected the CLI-reported software version to 1.7.12.
- Consensus, AQM64, difficulty, Network ID, genesis, monetary policy and transaction rules are unchanged.

### Desktop / Full Node v1.7.11 — Previous Mainnet patch

- Built-in Explorer lists alternative HTTPS full-node Explorers learned through native AuronQ P2P gossip.
- No Explorer is canonical or trusted; each full node serves data from its own locally validated chain.
- Consensus, Network ID, genesis, monetary policy, AQM64, signature and transaction rules are unchanged from the preceding 1.7.x hardening releases.

### AuronQ Mobile 0.5.1 Alpha

- Independently validates the Mainnet header chain from embedded genesis.
- Locally verifies AQM64 proof-of-work, difficulty, timestamps and chain continuity.
- Accepts balance, history and spendable UTXO state only from peers matching the locally verified header state.
- Compares canonical wallet state across agreeing peers and fails closed on conflicts.
- Keeps ML-DSA-87 private keys and signing local.
- Broadcasts signed transactions directly to agreeing verified-chain peers.
- Remains a light client rather than a full UTXO-validating node.

## Desktop / Full Node history

### v1.7.10
- Made configured seed/manual peers replaceable startup hints instead of permanent runtime authorities.
- Repeatedly failing seed peers are pruned.
- Added founder-node-removal regression coverage: the original bootstrap disappears permanently and a fresh node joins through surviving non-founder peers.
- Clarified Explorer as a local view of each node's own validated canonical chain.

### v1.7.9
- Strengthened eclipse/Sybil resistance through peer netgroups and DNS parent-domain grouping.
- Added sync-order diversity across peer groups.
- Added consensus invariant tests for difficulty clamps, median-time-past, future timestamps, coinbase maturity and same-block double spends.
- Added deterministic/reproducible build verification.

### v1.7.8
- Added direct bootstrap fallbacks in addition to the manifest.
- Improved higher-work reorg recovery and immediate post-reorg synchronization.
- Added a 20-node / four-partition convergence and restart test.
- Expanded fuzzing to transaction and block JSON paths.
- Added hourly public Mainnet health monitoring.
- Updated automated peer-registry changes to respect protected `main`.

### v1.7.7
- Fixed Desktop Explorer iframe CSP compatibility while keeping frame access restricted to the local AuronQ node.

### v1.7.6
- Added the dedicated Explorer tab to AuronQ Desktop.

### v1.7.5
- Added the built-in read-only blockchain Explorer at `/explorer`.
- Added search by address, TXID, block height and block hash plus block/transaction/address detail views.

### v1.7.4
- Added wallet/light-client history API.
- Added estimated AQM64 network hash power.
- Added Desktop network-power telemetry and Android history support.
- Added Android CI.

### v1.7.3
- Fixed Windows seed-node launcher handling for Tailscale discovery/Funnel setup.

### v1.7.2
- Added verified HTTPS/DNS peer discovery and gossip with resolver filtering.
- Added public endpoint advertisement support and autonomous peer-crawler verification.

### v1.7.1
- Added guarded Desktop wallet deletion with exact-name confirmation, active-miner protection and backup warnings.

### v1.7.0 — Mainnet launch
- Promoted the tested release-candidate consensus/network code to public Mainnet without changing the established genesis or Network ID.
- Shipped persistent peer storage, peer gossip, fixed seeds, DNS-seed support and bounded HTTPS bootstrap manifests.
- Published Windows/Linux release packages and zero-touch Windows startup tooling.

## Mobile history

### 0.4.2 Alpha
- Introduced multi-peer Mainnet state agreement.
- Compared balances only across peers reporting the same chain state.
- Failed closed on conflicting same-chain balances.
- Added direct transaction fanout after local signing.
- Still depended on full nodes for chain validation.

### 0.4.0 Alpha
- Added mobile wallet history against public 1.7.4+ nodes.

### 0.1.0–0.3.0 Alpha
- Early Android wallet/bridge releases preserved in the release archive.

## Earlier development

Versions 1.2.x through 1.6.x and 1.7.0 release-candidate material are historical development builds. Their original notes are preserved under:

- [docs/archive/releases/](docs/archive/releases/)
- [docs/archive/rc/](docs/archive/rc/)
- [docs/archive/checksums/](docs/archive/checksums/)

These historical files are retained for provenance, not as recommended downloads.

For current software, always use the release links in the main [README](README.md).
