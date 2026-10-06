# AuronQ GPU Miner v0.3.6 Alpha

This release adds the first **official native Linux amd64 CUDA miner package** and synchronizes the miner documentation for Windows and Linux. It does **not** change AuronQ Mainnet consensus, AQM64, difficulty, genesis, Network ID, wallet rules or monetary policy.

## Linux amd64

The official Linux package contains:

- `auronq-gpu-miner` — native AQM64 CUDA CLI miner;
- `libauronq-aqm64-cuda.so` — Linux CUDA backend;
- Mainnet `network.json` and `bootstrap.json`;
- complete GPU Miner documentation.

Linux v0.3.6 supports:

- native Solo AQM64 CUDA mining;
- NVIDIA multi-GPU with disjoint nonce ranges;
- Auto Tune;
- rolling live H/s;
- GPU/CPU AQM64 equivalence self-test;
- offline benchmark;
- smart thermal governor using NVIDIA temperature telemetry;
- explicit GPU selection or `--devices all`.

The Linux release is intentionally **CLI-first**. The Windows PL/EN dashboard GUI is not advertised as a Linux feature.

## Windows x64

Windows keeps the v0.3.5 GPU dashboard and v0.3.4 external-pool thermal governor, including:

- rich per-GPU telemetry dashboard;
- Solo mining;
- MeshMiner 0.8.35 integration;
- custom compatible pool endpoints;
- live H/s, power and H/W efficiency;
- adaptive thermal control and hard safety limit.

## Cross-platform packaging

The release workflow now builds and publishes both:

- `AuronQ-GPU-Miner-v0.3.6-alpha-Windows-x64.zip`
- `AuronQ-GPU-Miner-v0.3.6-alpha-Linux-amd64.tar.gz`

A single `SHA256SUMS-GPU-MINER.txt` covers both official packages.

The packaged CUDA runtime is statically linked. A CUDA Toolkit is required to build from source, but not merely to run the packaged miner. A current proprietary NVIDIA driver is still required.

## Documentation

New complete guide:

- [GPU-MINER-GUIDE.md](GPU-MINER-GUIDE.md)

It documents Windows and Linux setup, self-test, benchmark, Solo mining, multi-GPU, thermal control, live H/s, Pool/MeshMiner behavior, troubleshooting and source builds.

## Important

GPU Miner remains alpha software. The CUDA implementation has not received an independent professional security or cryptographic audit. Real-device Linux validation is still needed across multiple distributions, NVIDIA driver versions and GPU models before Linux support should be treated as production-hardened.
