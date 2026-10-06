# AuronQ GPU Miner v0.3.2 Alpha

Windows alpha update focused on automatic thermal management, useful live mining statistics and pool flexibility. This release changes miner/UI behavior only; it does **not** change AuronQ Mainnet consensus, AQM64, difficulty, genesis, Network ID, wallet rules or monetary policy.

## New in v0.3.2

- **Smart thermal governor**
  - the configured temperature is now a hard safety ceiling, not the normal operating target;
  - the built-in CUDA miner automatically targets about **5 °C below the hard limit**;
  - it reduces batch size as temperature rises;
  - it also introduces short adaptive duty-cycle pauses, because smaller batches alone may still leave a GPU continuously saturated;
  - after sustained cooling it gradually restores performance;
  - if the GPU still reaches the configured hard limit, mining stops as a final safety measure.

- **Reliable live H/s**
  - the built-in miner reports an interval hashrate about once per second instead of waiting for a fixed number of large batches;
  - multi-GPU mode aggregates the current rate from all GPU workers;
  - the GUI prefers current `rate=` data and falls back to average H/s;
  - pool-mode log parsing recognizes H/s, kH/s, MH/s, GH/s, TH/s and PH/s values emitted by compatible external miners.

- **Other pool endpoints**
  - MeshPool remains a convenient preset;
  - users can select **Custom / other pool** and enter any compatible `host:port` or `stratum+tcp://host:port` endpoint;
  - the UI explicitly supports copying a current endpoint from another AURQ pool such as RPlant;
  - no third-party pool binary is downloaded automatically.

## Retained from v0.3.x

- multi-NVIDIA-GPU mining with disjoint nonce ranges;
- per-GPU automatic batch tuning;
- live temperature, fan, load, power and VRAM telemetry;
- configurable hard thermal safety limit;
- Solo and Pool modes;
- PL/EN GUI;
- GPU/CPU AQM64 equivalence self-test;
- ordinary full-node validation of Solo-mined blocks.

## Important

This remains alpha software. The CUDA implementation, adaptive governor and multi-GPU orchestration have not received an independent professional audit. Test temperature behavior on your own cooling system before leaving mining unattended.
