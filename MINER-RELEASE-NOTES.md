# AuronQ Universal Miner v0.4.2 Alpha

## Strategic public-node lifecycle update

v0.4.2 makes inbound P2P reachability a **full-node lifecycle function** rather than a mining-session function.

- Auto Public TCP/18444 starts with the full node, not with the miner.
- Stopping Solo/Pool mining no longer closes the public-node mapping.
- Windows GUI retries public reachability automatically while the node remains outbound-only.
- The dashboard now reports actionable diagnostics for likely CGNAT, missing UPnP/IGD, router mapping refusal and unverified callback/firewall cases.
- The bundled portable CLI defaults to `--auto-public=true` and can be explicitly disabled with `--auto-public=false`.
- CLI Auto Public is skipped for loopback-only listeners and preserves explicit `--advertise` behavior.
- Portable-node examples now listen on an inbound-capable address instead of `127.0.0.1` when public-node operation is intended.
- Remote peers still callback-verify an advertised endpoint before admitting it to public gossip.

No Mainnet consensus rules are changed.
## Critical GUI hotfix

v0.4.2 fixes a JavaScript syntax error in the Windows GUI shipped in v0.4.0 Alpha. The broken inline script prevented the dashboard controls, refresh loop and actions from running correctly even though the compiled executables and backend self-tests were valid.

A mandatory JavaScript syntax check is now part of CI so an invalid embedded dashboard script cannot pass the Universal Miner release gate again.

This release is a miner compatibility and safety update. It does not change AuronQ Mainnet consensus.

## AUTO compute backend

- AUTO is the recommended default.
- Compatible NVIDIA CUDA on Windows/Linux is used when available.
- If CUDA, the driver, GPU or accelerator library is unavailable, the miner falls back to the built-in CPU AQM64 backend.
- Explicit CUDA mode still fails closed instead of silently changing backends.

## Native CPU fallback

- Works without CUDA and without a discrete GPU.
- Reuses fixed 64 MiB AQM64 workspaces instead of reallocating each hash.
- Default CPU profile uses about one quarter of logical CPUs, capped at 2 lanes.
- Explicit CPU concurrency is capped at 16 lanes.
- Backend output is verified byte-for-byte against canonical AQM64 in CI and by the self-test.

## Hardware safety

- Windows NVIDIA GPU thermals use direct local NVML hardware telemetry.
- Windows UI and governor use the same 500 ms hardware sample.
- Linux GPU thermal telemetry loss now stops protected GPU mining rather than silently disabling the limit.
- CPU fallback does not fabricate a CPU temperature where no trustworthy cross-vendor sensor exists; it uses conservative concurrency instead.
- Pool or web temperature values are never used as safety-control inputs.

## Windows GUI

- Solo backend selector: AUTO / NVIDIA CUDA / CPU.
- Pool backend selector adds AUTO.
- No NVIDIA GPU is no longer a fatal startup condition in AUTO mode.
- Clean CPU fallback status is shown instead of a misleading GPU error.
- Pool CPU defaults are conservative.

## Portable packages

CPU-safe CLI packages are built for:

- Windows x64
- Windows ARM64
- Linux x64
- Linux ARM64
- macOS Intel x64
- macOS Apple Silicon ARM64

Each portable package includes the AuronQ miner CLI, full-node CLI, network.json, bootstrap.json and the Universal Miner guide.

## GPU acceleration limits

Official accelerated GPU backend in v0.4.2 remains NVIDIA CUDA. AMD/Intel computers are supported through the native CPU fallback. An AMD/Intel GPU backend will only be promoted after hardware testing and byte-for-byte AQM64 validation; v0.4.2 does not ship an unvalidated OpenCL/HIP/oneAPI implementation.

## Consensus unchanged

No changes to genesis, Network ID, AQM64 parameters, difficulty, monetary policy or transaction/block validation.
