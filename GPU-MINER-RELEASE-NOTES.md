# AuronQ GPU Miner v0.3.1 Alpha

Windows alpha hotfix for v0.3.0 focused on Windows UX stability. It retains the multi-GPU, automatic tuning and observability features introduced in v0.3.0. This release changes only mining software and UI behavior; it does **not** change AuronQ Mainnet consensus, AQM64, difficulty, genesis, Network ID or monetary rules.

## Hotfix in v0.3.1

- fixes repeated Windows CMD/console flashing caused by the GUI launching `nvidia-smi` every telemetry refresh;
- periodic GPU telemetry and one-shot device discovery now run `nvidia-smi` hidden with `CREATE_NO_WINDOW`;
- temperature, fan, load, power and VRAM monitoring continue to work normally;
- no AQM64, consensus, difficulty, network, wallet or monetary-policy changes.

## Features retained from v0.3.0

- **Multi-GPU NVIDIA mining**
  - use every detected NVIDIA card or select individual device IDs;
  - the built-in solo miner launches an isolated CUDA worker per GPU;
  - each worker receives a disjoint 64-bit nonce range so multiple cards do not duplicate the same search space.

- **Automatic AQM64 batch tuning**
  - optional pre-mining autotune tests multiple safe batch sizes on each selected GPU;
  - the fastest measured batch is chosen independently per device;
  - the existing manual batch value remains available as a fallback and for controlled testing.

- **Live GPU telemetry**
  - temperature;
  - fan speed when reported by the NVIDIA driver;
  - GPU utilization;
  - power draw / power limit when available;
  - VRAM use / total VRAM.

- **Thermal emergency stop**
  - configurable stop temperature from 60–95 °C;
  - default is 85 °C;
  - if any selected GPU reaches the configured limit, the miner stops the complete worker process tree.

- **Pool mode**
  - GUI can launch a user-supplied AuronQ-compatible external pool miner;
  - MeshMiner 0.8.35+ command-line layout is supported (`--algo auronq`, pool endpoint, wallet.worker, CUDA device list);
  - MeshPool default endpoint is pre-filled as `pool.meshpool.net:3359`;
  - custom pool endpoints can be entered manually;
  - CUDA-only or CUDA+CPU backend can be selected when supported by the external miner;
  - AuronQ does **not** auto-download or bundle third-party mining binaries.

- **PL/EN GUI**
  - all-GPU toggle;
  - individual GPU selection;
  - autotune controls;
  - thermal limit;
  - live per-GPU telemetry;
  - Solo / Pool mode and pool settings.

## Existing safety/correctness behavior retained

- full GPU/CPU AQM64 equivalence self-test;
- all solo-mined candidate blocks still pass through the ordinary AuronQ full-node validation path;
- Mainnet Network ID verification before solo mining;
- local chain synchronization checks;
- optional P2P/UPnP participation;
- protection against conflicting local AuronQ nodes.

## Requirements

Built-in AQM64 CUDA miner:

- Windows x64;
- NVIDIA Turing / RTX 20-series or newer target;
- current NVIDIA driver;
- NVIDIA driver tooling (`nvidia-smi`) for device discovery and telemetry.

Pool mode additionally requires a compatible external pool miner supplied by the user. Third-party binaries, fees, pool protocols and payout rules are controlled by their respective operators.

## Important

This remains alpha software. The CUDA implementation and miner orchestration have not received an independent professional security or cryptographic audit. Multi-GPU and telemetry behavior should be validated on additional desktop and laptop GPU combinations before being treated as production-hardened.
