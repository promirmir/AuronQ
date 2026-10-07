# AuronQ GPU Miner v0.3.8 Alpha

> [!NOTE]
> This file documents the historical standalone GPU Miner v0.3.8 Alpha release. The current official mining package is **AuronQ Universal Miner v0.4.6 Alpha**, which uses validated CUDA → OpenCL GPU → CPU AUTO selection, includes accelerated Windows/Linux packages, and retains portable CPU-safe builds for Windows/Linux/macOS. See [UNIVERSAL-MINER-GUIDE.md](UNIVERSAL-MINER-GUIDE.md) and [HARDWARE-SUPPORT.md](HARDWARE-SUPPORT.md).

This release focuses on **hardware-truth thermal safety on Windows**. It does **not** change AuronQ Mainnet consensus, AQM64, difficulty, genesis, Network ID, transaction validation, wallet rules or monetary policy.

## Direct local hardware telemetry

Windows thermal control no longer depends on temperatures reported by MeshMiner, pool software, websites or other external data sources.

Safety decisions now use the locally installed NVIDIA driver directly through **NVML**:

- GPU temperature is read directly from the NVIDIA driver;
- NVIDIA utilization, VRAM, power, clocks, P-state and fan telemetry are read from the same local driver interface when supported;
- NVML devices are mapped to CUDA ordinals through direct CUDA UUID matching;
- on multi-GPU systems the miner refuses unsafe sensor/device guessing;
- the same 500 ms hardware sample feeds both the dashboard and the pool thermal governor;
- the dashboard displays **NVML DIRECT** and the age of the current hardware sample.

MeshMiner and pools remain sources of mining work, shares and H/s only. Their temperature data is not used for AuronQ thermal safety decisions.

## Fail-safe behavior

- Recommended thermal limit remains **81 °C**.
- Automatic target remains about **76 °C** with adaptive load control.
- At the configured limit, Pool mode performs automatic suspend/cool/resume.
- The independent catastrophic fail-safe is tightened to approximately **limit + 1 °C**.
- If direct NVML telemetry becomes stale or unavailable, mining stops instead of continuing with guessed or externally reported temperature data.
- The native Windows Solo CUDA worker now follows the same direct-telemetry rule and also stops on telemetry loss.

## Windows x64

Windows includes:

- PL/EN dashboard;
- native AQM64 CUDA Solo miner;
- MeshMiner 0.8.35+ bridge for pool mining;
- custom compatible pool endpoints;
- live H/s and direct local GPU telemetry;
- autonomous thermal control based on direct NVML hardware samples.

## Linux amd64

The Linux package remains the native CLI-first AQM64 CUDA miner. The direct NVML Windows controller described above is a Windows-specific safety change in v0.3.8; Linux behavior is otherwise unchanged from v0.3.7.

## Packages

- `AuronQ-GPU-Miner-v0.3.8-alpha-Windows-x64.zip`
- `AuronQ-GPU-Miner-v0.3.8-alpha-Linux-amd64.tar.gz`
- `SHA256SUMS-GPU-MINER.txt`

## Important

GPU Miner remains alpha software. Direct NVML telemetry is a software safety layer using the official local NVIDIA driver interface; it does not replace functioning laptop fans, unobstructed airflow, firmware thermal protection or hardware protection.
