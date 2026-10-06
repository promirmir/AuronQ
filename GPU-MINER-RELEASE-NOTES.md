# AuronQ GPU Miner v0.3.7 Alpha

This release focuses on **safe autonomous pool mining on Windows**, especially laptop GPUs. It does **not** change AuronQ Mainnet consensus, AQM64, difficulty, genesis, Network ID, wallet rules or monetary policy.

## Autonomous pool thermal governor

The Windows GUI now wraps external MeshMiner pool mining with a smoother AuronQ-side thermal controller:

- NVIDIA temperature telemetry is sampled every 500 ms;
- load is reduced with short adaptive duty-cycle pulses instead of long 400–850 ms one-shot pauses;
- throttling increases quickly when temperature rises, but is released gradually after cooling to avoid thermal oscillation;
- the recommended/default safety limit is now **81 °C**, with an automatic operating target around **76 °C**;
- at 81 °C the miner is suspended for an automatic cooldown instead of immediately requiring a manual restart;
- mining resumes automatically after the GPU is stably cooled below the target;
- a separate catastrophic fail-safe stops the external miner at approximately **limit + 2 °C** if temperature still rises despite suspension;
- loss of NVIDIA thermal telemetry or failure of process suspend/resume now fails safe by stopping the miner rather than continuing blind.

The goal is stable long-running mining with less hashrate loss and no repeated full-load / long-pause temperature oscillation.

## Dashboard

The PL/EN dashboard now presents the thermal profile as an **AUTO** mode and uses 81 °C as the safe default for new settings. Existing user settings remain under user control.

## Pool mining

MeshMiner 0.8.35+ integration remains external and optional. AuronQ does not bundle or silently download third-party pool miners. Pool support still includes:

- MeshPool preset;
- custom AuronQ-compatible endpoints;
- CUDA / CPU / both backends;
- multi-GPU selection;
- Retune;
- live H/s parsing;
- AuronQ-side safety control independent of MeshMiner fan reporting.

## Linux

The native Linux amd64 Solo CUDA miner remains available in the release package. v0.3.7 does not claim the Windows external-pool suspend/resume governor as a Linux GUI feature.

## Packages

- `AuronQ-GPU-Miner-v0.3.7-alpha-Windows-x64.zip`
- `AuronQ-GPU-Miner-v0.3.7-alpha-Linux-amd64.tar.gz`
- `SHA256SUMS-GPU-MINER.txt`

## Important

GPU Miner remains alpha software. The CUDA implementation and the external-pool thermal governor have not received an independent professional security audit. The autonomous thermal controller is a software safety layer and does not replace working laptop fans, unobstructed airflow, the NVIDIA driver or hardware thermal protection.
