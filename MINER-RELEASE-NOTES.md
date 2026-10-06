# AuronQ Universal Miner v0.4.0 Alpha

This release is a miner compatibility and safety update. It does not change AuronQ Mainnet consensus.

## AUTO compute backend

- AUTO is the recommended default.
- Compatible NVIDIA CUDA on Windows/Linux is used when available.
- If CUDA, the driver, GPU or accelerator library is unavailable, the miner falls back to the built-in CPU AQM64 backend.
- Explicit CUDA mode still fails closed instead of silently changing backends.

## Native CPU fallback

- Works without CUDA and without a discrete GPU.
- Reuses fixed 64 MiB AQM64 workspaces instead of reallocating each hash.
- Default CPU profile uses about half of logical CPUs, capped at 4 lanes.
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

Official accelerated GPU backend in v0.4.0 remains NVIDIA CUDA. AMD/Intel computers are supported through the native CPU fallback. An AMD/Intel GPU backend will only be promoted after hardware testing and byte-for-byte AQM64 validation; v0.4.0 does not ship an unvalidated OpenCL/HIP/oneAPI implementation.

## Consensus unchanged

No changes to genesis, Network ID, AQM64 parameters, difficulty, monetary policy or transaction/block validation.
