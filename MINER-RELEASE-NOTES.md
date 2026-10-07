# AuronQ Universal Miner v0.4.6 Alpha

## Legacy NVIDIA GPU support and correct Pool AUTO behavior

v0.4.6 fixes the older-NVIDIA path discovered on GTX 1050-class hardware.

The previous v0.4.5 release could detect a Pascal GPU through NVML, reject it
through the modern CUDA 13.x backend, fail OpenCL on some driver combinations
and then mine on CPU. In Pool mode there was an additional logic error: AuronQ
used its own native CUDA compatibility test to decide whether a GPU was allowed
to reach the external MeshMiner process.

v0.4.6 separates those concerns and uses **three NVIDIA CUDA generations in parallel** so older-card support does not reduce support for new cards.

## AUTO order

Solo AUTO now tries:

1. **Primary NVIDIA CUDA** — CUDA 13.2 build for current/newer NVIDIA generations; its build script emits every maintained architecture target supported by the release toolkit, including current Blackwell-era targets.
2. **Legacy NVIDIA CUDA** — CUDA 12.6 build for supported Maxwell / Pascal / Volta targets, including GTX 10xx-class Pascal.
3. **Kepler NVIDIA CUDA** — CUDA 11.8 build for supported Kepler `sm_35` / `sm_37` targets.
4. **Vendor-neutral OpenCL GPU** — AMD / Intel / NVIDIA.
5. **Native CPU AQM64**.

Every native accelerator is accepted only after the full canonical byte-for-byte
AQM64 equivalence self-test.

## GTX 1050 / Pascal

The Windows and Linux accelerated packages now include generation-specific CUDA libraries rather than downgrading the primary backend:

- primary Windows/Linux CUDA 13.2 library for newer GPUs;
- Windows: `auronq-aqm64-cuda-legacy.dll` (CUDA 12.6 Maxwell/Pascal/Volta);
- Linux: `libauronq-aqm64-cuda-legacy.so`;
- Windows: `auronq-aqm64-cuda-kepler.dll` (CUDA 11.8 Kepler);
- Linux: `libauronq-aqm64-cuda-kepler.so`.

That library is built with CUDA 12.6 against the legacy architecture targets
still provided by that toolkit. The build script includes the supported subset
of Maxwell, Pascal and Volta target families and embeds the newest supported
legacy PTX target.

A GTX 1050 is no longer expected to depend on OpenCL first. AUTO attempts the
CUDA 12.6 legacy backend before OpenCL. Older supported Kepler cards get their
own CUDA 11.8 path.

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
New NVIDIA cards continue to prefer the primary CUDA 13.2 backend and are not
forced onto any legacy runtime. Future CUDA architecture targets can be picked
up by the primary build script when the release toolkit exposes them.

Portable Windows/Linux/macOS packages remain CPU-safe.

## No specialized-mining-appliance integration

The official miner continues to focus on ordinary CPU/GPU participation. No
dedicated specialized-mining-appliance interface, bridge or privileged work
path is added.

## Consensus unchanged

No changes to genesis, Network ID, AQM64 consensus parameters, difficulty,
monetary policy, chain selection, transaction validity or block validity.
