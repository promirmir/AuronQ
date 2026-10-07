# AuronQ Universal Miner — NVIDIA CUDA backend

This directory contains the native NVIDIA CUDA accelerator used by **AuronQ
Universal Miner v0.4.5 Alpha** for AQM64 Solo mining.

The CUDA backend is an implementation detail of the miner. It does **not**
change AuronQ Mainnet consensus, AQM64 parameters, difficulty, genesis,
Network ID, monetary policy, transaction validity or block validity.

For the complete hardware matrix and AUTO fallback policy, see
[../../HARDWARE-SUPPORT.md](../../HARDWARE-SUPPORT.md) and
[../../UNIVERSAL-MINER-GUIDE.md](../../UNIVERSAL-MINER-GUIDE.md).

## Current release status

Official accelerated packages:

- Windows x64 GUI/GPU:
  `AuronQ-Miner-v0.4.5-alpha-Windows-x64-GUI-GPU.zip`
- Linux x64 GPU:
  `AuronQ-Miner-v0.4.5-alpha-Linux-x64-GPU.tar.gz`

Both accelerated packages contain:

- the native CUDA backend;
- the vendor-neutral OpenCL backend;
- native CPU AQM64 fallback.

AUTO order is:

1. validated NVIDIA CUDA;
2. validated OpenCL GPU;
3. native CPU AQM64.

A CUDA device is accepted only after the mandatory full AQM64 self-test
produces a byte-identical result to the canonical CPU implementation.

## CUDA architecture coverage in v0.4.5

The v0.4.5 release is built with CUDA Toolkit **13.2**.

The release CI queried `nvcc --list-gpu-code` / `--list-gpu-arch` and emitted
these real architecture targets:

- `sm_75`
- `sm_80`
- `sm_86`
- `sm_87`
- `sm_88`
- `sm_89`
- `sm_90`
- `sm_100`
- `sm_103`
- `sm_110`
- `sm_120`
- `sm_121`

It also embeds forward-compatible PTX for the newest virtual architecture
available from that toolkit (`compute_121` in the v0.4.5 release build).

The build scripts do not hard-code only three generations. They ask the
installed CUDA compiler which maintained targets it can actually build and emit
the supported subset.

### Older NVIDIA cards

CUDA 13.2 does not emit Pascal `sm_61` code in this release. Therefore cards
such as GTX 1050/1050 Ti/1060/1070/1080 are **not claimed as native CUDA targets
of the v0.4.5 package**.

AUTO can still try those cards through the packaged **OpenCL GPU backend** when
the installed NVIDIA driver exposes a compatible OpenCL runtime. The OpenCL path
must pass the same canonical AQM64 self-test before mining; otherwise AUTO falls
back to CPU.

Do not infer compatibility only from a GPU model name or from NVML detection.

## Verified correctness

The CUDA implementation is correctness-first.

Before mining, the worker verifies:

- the accelerated Argon2id boundary;
- the complete AQM64 pipeline;
- byte-for-byte equivalence with the canonical CPU `PowHash`.

A backend that differs by even one bit is rejected.

Real-device validation has passed on an **NVIDIA GeForce RTX 4050 Laptop GPU**.
Historical offline benchmark data on that device:

- batch 20: 118.733 H/s
- batch 40: 226.263 H/s
- batch 60: 316.227 H/s
- batch 64: 299.377 H/s

These are single-device measurements, not guaranteed performance figures.
Laptop power limits, cooling, clocks, drivers and batch size can materially
change throughput.

## Memory model

AQM64 uses approximately **64 MiB per candidate lane**.

The CUDA backend runs independent candidates concurrently and chooses a
conservative starting batch from available VRAM and GPU SM count. Auto Tune can
then benchmark usable batch sizes on the actual device.

## Runtime requirements

To run the packaged CUDA backend:

- Windows x64 or Linux x64 accelerated package;
- compatible NVIDIA GPU;
- current NVIDIA driver.

The CUDA Toolkit is **not required on the mining computer** for the packaged
release; the CUDA runtime is linked into the backend. A toolkit is required only
when rebuilding the CUDA library from source.

Solo mining also requires a synchronized AuronQ Mainnet full node.

## Build from source

Windows:

~~~powershell
powershell -ExecutionPolicy Bypass -File .\gpu\cuda\build-windows.ps1
~~~

Linux:

~~~bash
bash ./gpu/cuda/build-linux.sh
~~~

The scripts query the installed `nvcc` architecture list and compile the
maintained targets supported by that toolkit.

## Self-test

Windows worker:

~~~powershell
.\auronq-miner-worker.exe --backend cuda --self-test
~~~

Linux:

~~~bash
./auronq-miner --backend cuda --self-test --device 0
~~~

Expected result:

~~~text
SELF-TEST OK
~~~

If explicit CUDA fails, do not force it on Mainnet. Use AUTO so the miner can
try OpenCL and then CPU.

## Mining examples

Single NVIDIA device:

~~~bash
./auronq-miner --backend auto --device 0 --node http://127.0.0.1:18444 --address aurq1... --self-test --auto-tune --thermal-auto --thermal-limit 81
~~~

All NVIDIA devices detected by the NVIDIA multi-GPU path:

~~~bash
./auronq-miner --backend auto --devices all --node http://127.0.0.1:18444 --address aurq1... --self-test --auto-tune --thermal-auto --thermal-limit 81
~~~

Each multi-GPU child receives a disjoint nonce range.

## Thermal safety

On supported NVIDIA systems, AuronQ uses **direct local NVIDIA telemetry** for
the CUDA thermal governor. Pool/web/miner-reported temperatures are not trusted
for safety control.

The configured hard limit is not the normal target. The controller aims several
degrees below it, reduces work gradually as temperature rises, performs
automatic cooldown/resume at the limit and fails closed when trusted telemetry
is lost.

Generic OpenCL hardware follows a separate conservative no-sensor policy
documented in [../../HARDWARE-SUPPORT.md](../../HARDWARE-SUPPORT.md).

## Security status

The CUDA and OpenCL accelerator implementations remain **alpha software** and
have not received an independent professional security/cryptographic audit.

Every candidate block is still submitted to an ordinary AuronQ full node and
must pass the same Mainnet validation rules as a block found by any other
implementation.

Complete user guide: [../../GPU-MINER-GUIDE.md](../../GPU-MINER-GUIDE.md)
