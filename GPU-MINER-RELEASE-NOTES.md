# AuronQ GPU Miner v0.3.3 Alpha

Windows alpha update adding first-class support for **MeshMiner 0.8.35+** in Pool mode. This remains a miner/UI update only; there are **no changes** to Mainnet consensus, AQM64, difficulty, genesis, Network ID, wallet rules or monetary policy.

## MeshMiner 0.8.35 integration

AuronQ can now launch MeshMiner 0.8.35's documented AURQ path directly with its command-line interface:

- `--algo auronq`;
- `--pool host:port`;
- `--user aurq1....worker`;
- `--backend cuda|cpu|both`;
- `--device 0,1,...` for selected NVIDIA cards;
- `--fan auto` as an optional/default GPU cooling mode;
- `--threads N` for CPU/both mode;
- `--retune` when the operator explicitly wants to remeasure cards;
- dedicated **Retune MeshMiner now** action using `--tune-only --retune`.

The GUI automatically looks for `meshpool-miner.exe` or `meshminer.exe` next to AuronQ-GPU-Miner.exe and also inside a neighboring folder whose name contains MeshMiner/MeshPool. A manual executable path remains available.

## Thermal behavior with external MeshMiner

The built-in AuronQ CUDA Solo miner can regulate workload itself. An external closed binary cannot be safely given the same internal batch governor, so AuronQ uses the controls MeshMiner actually documents:

- MeshMiner `--fan auto` can be enabled directly from AuronQ and is enabled by default for GPU pool mining;
- MeshMiner performs its own per-card tuning and remembers results;
- AuronQ continues to monitor NVIDIA temperature independently;
- the configured AuronQ temperature remains a **hard emergency stop** for Pool mode.

This avoids pretending that AuronQ controls undocumented MeshMiner internals.

## Pools

MeshPool remains the preset endpoint. The user can also select **Custom / other pool** and enter another compatible AURQ pool endpoint. Compatibility still depends on the external pool/miner protocol; entering an endpoint does not make an incompatible Stratum implementation compatible.

## Existing v0.3.2 behavior retained

- live rolling H/s in the built-in miner;
- multi-GPU aggregation;
- generic pool H/s parsing;
- smart thermal governor for the built-in CUDA miner;
- multi-NVIDIA-GPU Solo mining;
- PL/EN interface;
- GPU/CPU AQM64 equivalence self-test.

## Third-party software

MeshMiner is independently developed and distributed:

https://github.com/totom9000/meshminer/releases/tag/v.0.8.35

AuronQ does not silently download, bundle, modify or redistribute the MeshMiner binary. MeshMiner 0.8.35 reports a 0.5% developer fee on MeshPool and 1.2% elsewhere. Verify third-party releases, fees and pool rules independently.
