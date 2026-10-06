# AuronQ Universal Miner v0.4.3 Alpha

## Thermal stability update

v0.4.3 focuses on making Windows Pool-mode GPU regulation **smooth and stable instead of oscillatory**.

The previous controller could react too late near the configured limit, then make large duty-cycle changes and repeatedly enter suspend/cool/resume cycles on thermally constrained laptop GPUs. v0.4.3 changes the control law without changing AQM64 or Mainnet consensus.

### Smoother adaptive governor

- regulation now begins **before** the thermal target instead of waiting until the GPU is already at or above it;
- duty-cycle changes are limited to small steps: throttling can increase by about 10 percentage points per 500 ms sample, while performance is restored much more slowly;
- throttling is not released while temperature is flat or still rising near the target;
- the controller therefore converges toward a stable operating point instead of repeatedly jumping between near-full load and deep throttling;
- all safety decisions still use direct local NVIDIA NVML telemetry, never pool/web-reported temperatures.

### Stable recovery after a hard-limit cooldown

If the configured hard limit is reached:

- the external pool miner is suspended and the GPU is allowed to cool;
- restart now waits farther below the target and requires more consecutive cool samples;
- after resume, a conservative minimum throttle is held for 20 seconds so heat soak can settle before performance is restored;
- a brief 1–3 °C post-suspend thermal overshoot is tolerated as normal thermal inertia, but continued heating while suspended still fails closed;
- missing/stale local telemetry or suspend/resume control failure still stops mining.

This specifically addresses the repeated **75 °C → suspend → 67 °C → resume → 75 °C** pattern seen on laptop GPUs.

## AUTO compute backend

- AUTO remains the recommended default.
- Compatible NVIDIA CUDA on Windows/Linux is used when available.
- If CUDA, the NVIDIA driver, GPU or accelerator library is unavailable, the miner falls back to the built-in CPU AQM64 backend.
- Explicit CUDA mode still fails closed instead of silently changing backends.

## Native CPU fallback

- Works without CUDA and without a discrete GPU.
- Uses the canonical AQM64 implementation and fixed 64 MiB workspaces.
- Default CPU profile remains conservative.
- Backend output is verified against canonical AQM64 by CI and self-test.

## Public-node lifecycle

The v0.4.2 node-lifecycle behavior remains in v0.4.3:

- Auto Public TCP/18444 belongs to the full-node lifecycle, not to a mining session.
- Stopping Solo/Pool mining does not remove the node's public UPnP mapping.
- Windows GUI retries public reachability while the node remains outbound-only.
- Portable `auronq node` defaults to `--auto-public=true` and can be explicitly disabled.

## Supported packages

The release continues to provide:

- Windows x64 GUI/CUDA package;
- Linux x64 CUDA package;
- portable CPU-safe CLI packages for Windows x64/ARM64, Linux x64/ARM64 and macOS x64/ARM64.

## Consensus unchanged

No changes to genesis, Network ID, AQM64 parameters, difficulty, monetary policy, chain selection or transaction/block validation.
