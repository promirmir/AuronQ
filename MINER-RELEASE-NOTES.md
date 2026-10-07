# AuronQ Universal Miner v0.4.6 Alpha

## Legacy NVIDIA GPU support and correct Pool AUTO behavior

v0.4.6 fixes the older-NVIDIA path discovered on GTX 1050-class hardware.

The previous v0.4.5 release could detect a Pascal GPU through NVML, reject it
through the modern CUDA 13.x backend, fail OpenCL on some driver combinations
and then mine on CPU. In Pool mode there was an additional logic error: AuronQ
used its own native CUDA compatibility test to decide whether a GPU was allowed
to reach the external MeshMiner process.

v0.4.6 separates those concerns and adds a dedicated legacy CUDA backend.

## AUTO order

Solo AUTO now tries:

1. **Primary NVIDIA CUDA** — current CUDA 13.2 build for newer NVIDIA generations.
2. **Legacy NVIDIA CUDA** — CUDA 12.6 build for supported Maxwell / Pascal / Volta targets, including GTX 10xx-class Pascal when the installed driver and hardware pass the canonical AQM64 test.
3. **Vendor-neutral OpenCL GPU** — AMD / Intel / NVIDIA.
4. **Native CPU AQM64**.

Every native accelerator is accepted only after the full canonical byte-for-byte
AQM64 equivalence self-test.

## GTX 1050 / Pascal

The Windows and Linux accelerated packages now include a second CUDA library:

- Windows: `auronq-aqm64-cuda-legacy.dll`
- Linux: `libauronq-aqm64-cuda-legacy.so`

That library is built with CUDA 12.6 against the legacy architecture targets
still provided by that toolkit. The build script includes the supported subset
of Maxwell, Pascal and Volta target families and embeds the newest supported
legacy PTX target.

A GTX 1050 is no longer expected to depend on OpenCL first. AUTO attempts the
legacy CUDA backend before OpenCL.

The physical card still has to pass the runtime AQM64 self-test. A compiled
`sm_61` target is necessary for Pascal support, but it is not treated as proof
of correctness by itself.

## Pool mode fix

Windows Pool mode launches a user-supplied third-party miner such as MeshMiner.

v0.4.5 incorrectly filtered the selected NVIDIA device through AuronQ's own
native CUDA self-test before launching the external pool miner. That could turn
an older GPU into a CPU fallback even when the external miner supported the
card.

v0.4.6 no longer applies the native AuronQ CUDA compatibility gate to Pool
hardware selection. Pool GPU eligibility is based on the locally detected
NVIDIA device set and the capabilities of the user-supplied pool miner.

AuronQ-side Pool thermal safety still uses direct local NVML telemetry and the
existing autonomous duty/cooldown controller.

## UI accuracy

The Windows GUI no longer labels AUTO as CPU before the worker has finished
trying legacy CUDA / OpenCL.

It also records the actual AUTO fallback reason reported by the worker, so the
user can see why a backend was rejected instead of seeing a misleading generic
CPU label.

## Other hardware

AMD and Intel GPU support remains through the validated OpenCL backend.
Newer NVIDIA cards continue to prefer the primary CUDA 13.2 backend.

Portable Windows/Linux/macOS packages remain CPU-safe.

## No specialized-mining-appliance integration

The official miner continues to focus on ordinary CPU/GPU participation. No
dedicated specialized-mining-appliance interface, bridge or privileged work
path is added.

## Consensus unchanged

No changes to genesis, Network ID, AQM64 consensus parameters, difficulty,
monetary policy, chain selection, transaction validity or block validity.
