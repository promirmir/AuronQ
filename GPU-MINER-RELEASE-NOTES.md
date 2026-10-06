# AuronQ GPU Miner v0.3.5 Alpha

Windows alpha update focused on the **GPU dashboard and operator readability**. This release changes presentation and telemetry only; it does **not** change Mainnet consensus, AQM64, difficulty, genesis, Network ID, wallet rules or monetary policy.

## GPU dashboard refresh

- redesigned per-GPU cards instead of a compact text row;
- large live temperature with clear OK / WARM / HOT state;
- utilization, fan, power and VRAM shown as dedicated metrics;
- core clock, memory clock and NVIDIA performance state (P-state);
- progress bars for GPU utilization, VRAM, temperature and power;
- configured thermal target and hard limit shown directly on each card;
- overview page now includes the GPU cards as well as the mining page;
- total GPU power and live H/W efficiency are calculated alongside H/s;
- multi-GPU layouts scale automatically to the available window width.

## Telemetry

AuronQ now requests these additional NVIDIA driver values through `nvidia-smi`:

- graphics/core clock;
- memory clock;
- P-state.

Existing temperature, fan, load, power and VRAM monitoring remain unchanged.

## Mining behavior retained

v0.3.5 keeps the v0.3.4 external-pool thermal governor, MeshMiner 0.8.35 integration, custom pool endpoints, live H/s parsing, multi-GPU Solo mode and the hard thermal safety stop.

## Important

This remains alpha software. Some laptop NVIDIA drivers do not expose fan speed and may report it as unavailable/driver-controlled even while the physical cooling system is operating.
