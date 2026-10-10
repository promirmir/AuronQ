# AuronQ Universal Miner — Installation and hardware support
> [!IMPORTANT]
> **Intelligent Miner update (8 October 2026):** PR [#118](https://github.com/promirmir/AuronQ/pull/118) is merged into `main` ([commit eb24ea8](https://github.com/promirmir/AuronQ/commit/eb24ea82230d06480a9c0439d5cbfc4c47436777)). Changes include a local deterministic adaptive batch agent (not a cloud AI model), two-pass accelerator autotuning, CUDA throughput refinements, NVIDIA fail-closed thermal protection, node-status watchdog and improved diagnostics; a follow-up fix [#121](https://github.com/promirmir/AuronQ/pull/121) throttles Solo status RPC calls and backs off when rate-limited. All five PR CI workflows passed, and the Windows CUDA test package was built successfully.
>
> **Latest updated Windows GPU test build:** [GitHub Actions run 37832849518](https://github.com/promirmir/AuronQ/actions/runs/37837949129) → artifact `AuronQ-Universal-Miner-Windows-x64-GPU-TEST` (login may be required; artifacts expire). **This is not the v0.4.6-alpha GitHub Release asset.** Existing release download links below continue to point to the older tagged version. Independent long-duration and multi-device verification and a permanent Release asset remain outstanding. No AQM64 consensus, Network ID, genesis or monetary rule changed.



**Current builds:** Windows x64 GUI/GPU v0.4.8 Alpha is the existing published release and remains unchanged. Linux x64 v0.4.8 Alpha has passed automated build and CPU/OpenCL packaging checks; its [CI test archive](https://github.com/promirmir/AuronQ/actions/runs/38089531174) is temporary and is not yet a permanent Release download. Historical Linux and portable CPU builds are available in the [v0.4.6 Alpha archive](https://github.com/promirmir/AuronQ/releases/tag/miner-v0.4.6-alpha). No Linux CUDA real-GPU validation has been claimed.

AuronQ Universal Miner is designed to start safely on as many ordinary computers as possible without changing AuronQ Mainnet consensus.

## AUTO first

Recommended compute backend: `--backend auto`.

AUTO follows this order:

1. Try the primary NVIDIA CUDA backend built with the current CUDA 13.2 release toolkit for modern/new NVIDIA generations.
2. If unavailable or invalid, try the packaged CUDA 12.6 legacy backend for supported Maxwell / Pascal / Volta targets.
3. If unavailable or invalid, try the packaged CUDA 11.8 Kepler backend for supported sm_35 / sm_37 devices.
4. Validate every CUDA implementation against the canonical AQM64 self-test before accepting it.
5. If the NVIDIA CUDA paths are unavailable or fail exact byte-for-byte AQM64 validation, try the vendor-neutral OpenCL GPU backend.
6. Validate OpenCL against the same canonical AQM64 CPU reference.
7. If no accelerator passes validation, fall back to the built-in CPU AQM64 backend.

This means an older NVIDIA GPU such as a GTX 10xx Pascal card can be tried through the packaged legacy CUDA backend before OpenCL or CPU fallback. AMD Radeon and Intel Arc/iGPU hardware can also be attempted through OpenCL without changing AQM64 or Mainnet consensus.

Every accelerated path is fail-closed for correctness: a GPU backend is used only after it produces the exact same AQM64 result as the canonical CPU implementation.

## Supported systems

Portable CPU-safe CLI packages are CI-built for Windows x64, Windows ARM64, Linux x64, Linux ARM64, macOS Intel x64 and macOS Apple Silicon ARM64.

The full Windows x64 graphical package includes primary CUDA, CUDA 12.6 legacy, CUDA 11.8 Kepler, and vendor-neutral OpenCL backends. The Linux x64 accelerated package includes the corresponding shared libraries. Portable Windows/Linux/macOS packages remain CPU-safe and do not assume a GPU runtime.

## GPU coverage

The accelerated Windows/Linux packages are designed around four GPU paths:

- **Primary CUDA** — newest release backend, preferred on modern/new NVIDIA hardware;
- **Legacy CUDA 12.6** — Maxwell/Pascal/Volta compatibility;
- **Kepler CUDA 11.8** — sm_35/sm_37 compatibility;
- **OpenCL** — vendor-neutral fallback for AMD, Intel and NVIDIA GPUs with a working OpenCL runtime/driver.

OpenCL support is deliberately runtime-validated instead of being claimed from a model name alone. A card/driver combination is considered usable only when the OpenCL kernel compiles locally and the full AQM64 self-test is byte-identical to the CPU reference.

The current project does **not** add an ASIC-specific work protocol, ASIC device bridge or privileged hardware path. General-purpose CPU/GPU backends all compute the same public AQM64 proof-of-work.

## CPU safe profile and telemetry

AQM64 requires about 64 MiB per active mining lane. With `--cpu-threads 0`, the miner chooses roughly one quarter of logical CPUs, capped at 2 active lanes. Manual override is 1..16.

On Windows, the GUI shows a dedicated CPU card even when an NVIDIA GPU is physically present. It reports the processor name, whole-system CPU utilization from the native Windows `GetSystemTimes` API, logical CPU count, AQM64 thread count, nominal clock and the 64 MiB-per-lane AQM64 memory estimate.

There is no single reliable cross-vendor CPU package-temperature API available on every Windows motherboard/firmware stack. AuronQ therefore shows CPU temperature as **N/A** unless a trustworthy source exists; it does not invent a value or reuse unrelated ACPI thermal-zone readings. The safe fallback is conservative concurrency; firmware/OS thermal protection remains authoritative.

## Self-test

Run:

    auronq-miner --backend auto --self-test

Expected: `SELF-TEST OK`.

## Public-node behavior

The bundled full-node CLI now treats inbound P2P reachability as a node-lifecycle feature rather than a mining feature.

By default, `auronq node` uses `--auto-public=true`. When no explicit `--advertise` address is configured and the listen address accepts inbound traffic, the node:

1. uses a directly assigned public IPv4/IPv6 address when one is available;
2. otherwise attempts UPnP TCP port mapping for the configured listen port;
3. keeps the mapping for the lifetime of the node and removes it on shutdown;
4. retries periodically if the router is temporarily unavailable.

Home users behind CGNAT may still remain outbound-only because their ISP does not provide a directly reachable WAN address. Server operators who do not want automatic UPnP can use:

    auronq node ... --auto-public=false

Do not use a loopback-only listen address such as `127.0.0.1:18444` if you intend to run a public peer.

## Solo mining

Start the bundled full node:

Windows:

    .\auronq.exe node --network .\network.json --data .\node-data --listen 0.0.0.0:18444

Linux/macOS:

    ./auronq node --network ./network.json --data ./node-data --listen 0.0.0.0:18444

Then mine:

    ./auronq-miner --backend auto --node http://127.0.0.1:18444 --address aurq1... --self-test --auto-tune --thermal-auto --thermal-limit 81

On CPU fallback, accelerator autotune and GPU thermal control are skipped automatically and the conservative CPU profile is used. On generic OpenCL hardware without a trustworthy vendor-specific local temperature source, `--thermal-auto` uses a conservative approximately 50% compute-duty profile and reports temperature as unavailable rather than fabricating a value.

## Windows GUI

The Windows x64 graphical miner uses the same Solo policy: AUTO (recommended), NVIDIA CUDA (explicit; primary → CUDA 12.x legacy → CUDA 11.8 Kepler), OpenCL GPU, or CPU. Pool mode still uses the external MeshMiner integration and therefore follows the backends supported by that third-party miner rather than the native Solo OpenCL backend.

## NVIDIA thermal safety

On Windows, GPU safety uses direct local NVIDIA NVML telemetry. The dashboard and governor share the same hardware sample. On Linux CUDA mining, temperature comes from the locally installed NVIDIA driver utility. If local temperature telemetry fails while thermal protection is enabled, mining stops.

Pool/miner websites and remote hashrate services are never trusted as temperature-control inputs.

In Windows Pool mode, the configured temperature is a **hard safety ceiling, not the normal operating target**. The controller targets about 5 °C below the limit, begins soft regulation one degree before that target, changes duty cycle in small steps, and restores performance much more slowly than it removes heat. If the hard limit is reached, the miner is suspended until the GPU is several degrees below target and then held at a conservative duty for 20 seconds before gradual recovery.

## Portable package contents

Each portable package contains the miner CLI, full-node CLI, network.json, bootstrap.json, this guide and the license when present. Portable CPU packages do not require CUDA.

## Fail-safe rule

- every native GPU backend, including legacy CUDA, must pass canonical AQM64 equivalence before mining;
- known + trustworthy local sensor -> use it for hardware-aware regulation;
- no trustworthy cross-vendor temperature source -> do not fabricate a value;
- generic OpenCL + thermal-auto -> conservative approximately 50% compute duty with temperature shown as unavailable;
- CPU fallback -> conservative concurrency;
- NVIDIA CUDA with requested hardware thermal protection and lost trusted telemetry -> stop GPU mining.

## Consensus

Universal Miner v0.4.6 does not change genesis, Network ID, AQM64 consensus parameters, difficulty rules, block/transaction validation or monetary policy.
