# AuronQ Mining Hardware Support

AuronQ Mainnet consensus defines **AQM64**, not a specific hardware vendor.

A valid miner may compute AQM64 on any implementation that produces the exact
canonical result accepted by ordinary full-node validation. The official miner
therefore treats hardware support as a **runtime correctness property**, not a
marketing list of model names.

## Official Universal Miner v0.4.5 Alpha

AUTO order on accelerated Windows/Linux packages:

1. NVIDIA CUDA
2. vendor-neutral OpenCL GPU
3. native CPU AQM64

Every GPU backend must pass the canonical byte-for-byte AQM64 self-test before
it is allowed to mine.

## Hardware matrix

| Hardware | Backend | Current status |
|---|---|---|
| x86-64 / ARM64 CPU | Native Go AQM64 | Supported; canonical reference/fallback |
| NVIDIA RTX / modern CUDA-capable GPU | CUDA | Preferred when the packaged CUDA target and driver pass AQM64 self-test |
| Older NVIDIA, including CUDA generations not present in the packaged CUDA binary | OpenCL fallback | Attempted automatically when the installed NVIDIA OpenCL runtime is available; must pass local AQM64 self-test |
| AMD Radeon | OpenCL | Runtime-detected and accepted only after local kernel compile + canonical AQM64 self-test |
| Intel Arc / compatible Intel GPU runtimes | OpenCL | Runtime-detected and accepted only after local kernel compile + canonical AQM64 self-test |
| macOS GPU | — | Current portable macOS package remains CPU-only |
| Third-party pool miner hardware | Third-party | Depends on that miner/operator; not part of official backend validation |

## CUDA coverage

The CUDA build scripts query the installed `nvcc` for available real and
virtual architecture targets and compile the maintained baseline targets that
the installed CUDA Toolkit actually supports. This avoids hard-coding support
to only one GPU generation.

The release package still verifies the backend at runtime. A compiled target
does not by itself prove correct AQM64 output.

## OpenCL coverage

The OpenCL backend is runtime-loaded and does not require an OpenCL SDK on the
mining machine. It uses the OpenCL runtime supplied by the GPU driver.

The local driver must:

- expose a GPU OpenCL device;
- support a 128-work-item group and enough local/global memory;
- compile the AQM64 OpenCL kernel;
- produce a byte-identical AQM64 result to the canonical CPU implementation.

If any of those checks fail, AUTO continues to the next backend.

## Thermal policy

AuronQ does not invent hardware telemetry.

- NVIDIA CUDA on Windows uses direct local NVIDIA telemetry for the existing
  autonomous thermal governor.
- Generic OpenCL has no universally trustworthy temperature API across all
  AMD/Intel/NVIDIA drivers. With `--thermal-auto`, the official miner therefore
  uses a conservative approximately 50% compute-duty profile and reports
  temperature as unavailable.
- CPU fallback uses conservative concurrency and does not fabricate a CPU
  package temperature.

## Specialized mining hardware

The official v0.4.5 miner is intentionally focused on general-purpose CPU/GPU
participation. It does not add a dedicated integration for specialized mining
appliances.

This is a miner/software policy only. It is **not** a consensus rule that can
prove what physical hardware produced a valid AQM64 proof.

## Validation status

Compilation CI is not the same as real-device validation.

The project can verify:

- source/build correctness in CI;
- canonical OpenCL kernel equivalence through an OpenCL software ICD in CI;
- runtime self-test on every user's actual GPU before mining.

Real-device performance/stability claims should be added only after a specific
card/driver combination has actually passed the self-test and sustained mining
test.
