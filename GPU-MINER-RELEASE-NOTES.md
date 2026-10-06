# AuronQ GPU Miner v0.3.4 Alpha

Windows alpha update focused on **real thermal regulation of external pool miners** after field testing with MeshMiner 0.8.35. This is miner/UI behavior only; there are **no changes** to Mainnet consensus, AQM64, difficulty, genesis, Network ID, wallet rules or monetary policy.

## External pool thermal governor

Field logs showed that MeshMiner 0.8.35 can mine AURQ correctly and report stable live H/s, accepted shares, temperature, clocks and power, while some laptop NVIDIA configurations still expose fan control as `driver default` / `0%`. In that case `--fan auto` alone is not enough to hold a user-selected thermal ceiling.

AuronQ v0.3.4 therefore adds an **AuronQ-side thermal governor for external GPU pool miners**:

- normal target is about **5 °C below** the configured hard limit;
- when temperature approaches the target, AuronQ briefly suspends and resumes the external miner process to reduce GPU duty cycle;
- the pause increases progressively as temperature rises;
- after cooling, full duty cycle is restored automatically;
- the configured temperature remains the final hard emergency stop;
- the mechanism does not modify MeshMiner binaries or undocumented internal parameters.

This applies to GPU-backed Pool mode, including MeshMiner CUDA and CUDA+CPU modes. CPU-only Pool mode is unaffected.

## MeshMiner 0.8.35 integration retained

- AURQ / AQM64 via `--algo auronq`;
- CUDA / CPU / CUDA+CPU backends;
- multiple NVIDIA GPUs;
- `--threads`, `--fan auto`, `--retune` and `--tune-only`;
- MeshPool preset and compatible custom pool endpoints;
- live H/s parsing;
- explicit third-party miner path or automatic discovery next to AuronQ.

MeshMiner remains independently developed and distributed:
https://github.com/totom9000/meshminer/releases/tag/v.0.8.35

MeshMiner 0.8.35 reports a 0.5% developer fee on MeshPool and 1.2% elsewhere.

## Important

This remains alpha software. The new external duty-cycle governor uses short Windows process suspend/resume pulses and should be tested on different GPUs, drivers and pool conditions before unattended long-duration use.
