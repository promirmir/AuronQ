# AuronQ Universal Miner v0.4.5 Alpha

## Broad general-purpose GPU support

v0.4.5 expands the official Solo miner from a CUDA-first implementation into a validated multi-backend miner while leaving AQM64 and Mainnet consensus unchanged.

### AUTO backend order

AUTO now tries compute backends in this order:

1. **NVIDIA CUDA**
2. **Vendor-neutral OpenCL GPU**
3. **Native CPU AQM64**

A GPU backend is never accepted from a model name or driver detection alone. Before mining, the selected accelerator must complete the canonical AQM64 equivalence self-test and produce a byte-identical result to the CPU reference implementation.

### OpenCL GPU backend

The accelerated Windows/Linux packages now include a runtime-loaded OpenCL backend designed for ordinary GPUs from:

- AMD Radeon;
- Intel Arc and compatible Intel GPU runtimes;
- NVIDIA GPUs whose CUDA path is unavailable or unsupported.

The OpenCL library does not require the OpenCL SDK on the mining machine. It dynamically loads the OpenCL runtime supplied by the installed GPU driver, compiles the AQM64 kernel locally, then runs the mandatory canonical self-test.

A card/driver combination is considered usable only if the local OpenCL kernel compiles and the full AQM64 result matches the canonical CPU implementation.

### Wider NVIDIA build coverage

CUDA build scripts no longer hard-code only sm_75, sm_86 and sm_89. They query the installed CUDA compiler for supported real/virtual architectures and emit all compatible baseline targets from the maintained target list.

This allows a release toolkit to include additional NVIDIA generations when supported by that toolkit, while unsupported generations can still fall through to OpenCL and then CPU.

### Generic GPU safety

Direct NVIDIA CUDA thermal control still uses trustworthy local NVIDIA telemetry.

There is no single trustworthy cross-vendor temperature API available for every AMD/Intel/NVIDIA OpenCL driver. AuronQ therefore does not invent a temperature for generic OpenCL hardware.

When `--thermal-auto` is enabled on the generic OpenCL backend, the miner uses a conservative approximately 50% compute-duty profile instead of pretending to know device temperature. Device firmware/driver thermal protection remains authoritative.

### Windows GUI

The Windows Solo backend selector now exposes:

- AUTO;
- NVIDIA CUDA;
- OpenCL GPU — AMD / Intel / NVIDIA;
- CPU.

AUTO remains recommended.

CPU telemetry introduced in v0.4.4 remains visible independently from GPU detection.

### Multi-GPU

The existing child-worker orchestration can carry AUTO/OpenCL fallback for explicitly selected device indices. NVIDIA `--devices all` discovery remains available through the NVIDIA path.

### Pool mode

The native OpenCL backend is currently a **Solo mining backend**. Windows Pool mode still launches a user-supplied third-party compatible pool miner and therefore supports only the backends implemented by that external miner.

## Hardware policy

This release broadens access for ordinary CPUs and GPUs. It does not add a dedicated integration for specialized mining appliances.

All supported general-purpose backends compute the same public AQM64 proof-of-work and remain subject to ordinary full-node block validation.

## Consensus unchanged

No changes to genesis, Network ID, AQM64 parameters, difficulty, monetary policy, chain selection, transaction validity or block validity.
