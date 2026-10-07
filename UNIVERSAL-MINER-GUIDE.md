# AuronQ Universal Miner v0.4.5 Alpha

AuronQ Universal Miner is designed to start safely on as many ordinary computers as possible without changing AuronQ Mainnet consensus.

## AUTO first

Recommended compute backend: `--backend auto`.

AUTO follows this order:

1. On Windows/Linux, try the official AuronQ NVIDIA CUDA backend.
2. Before AUTO commits to GPU mining, validate the selected CUDA device against the canonical AQM64 self-test.
3. If the driver/DLL/device/kernel cannot produce the exact canonical AQM64 result, close the CUDA backend and fall back to the built-in CPU AQM64 backend automatically.
4. Never substitute pool/web-reported temperatures for local hardware sensors.
5. If GPU thermal safety was requested and trustworthy local GPU telemetry disappears, stop GPU mining rather than continue blind.

The Windows GUI applies the same validation before it resolves Pool AUTO to CUDA, so an NVIDIA card that is visible through NVML but unsupported by the current CUDA kernel no longer requires manual CPU selection.

The CPU fallback uses the same AQM64 initialization/finalization logic and is checked against the canonical AuronQ CPU proof-of-work implementation.

## Supported systems

Portable CPU-safe CLI packages are CI-built for Windows x64, Windows ARM64, Linux x64, Linux ARM64, macOS Intel x64 and macOS Apple Silicon ARM64.

The full Windows x64 graphical package additionally includes the official NVIDIA CUDA accelerator and direct NVML hardware telemetry. The Linux x64 CUDA package additionally includes the official CUDA shared library.

## AMD / Intel graphics

v0.4.5 does not pretend that an unvalidated AMD/Intel GPU accelerator exists. On AMD Radeon, Intel Arc/iGPU, unsupported NVIDIA, missing CUDA, or no discrete GPU, AUTO uses the native CPU backend.

AMD/Intel GPU acceleration can be added later only after byte-for-byte AQM64 self-tests and hardware validation.

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

On CPU fallback, CUDA autotune and GPU thermal control are skipped automatically and the conservative CPU profile is used.

## Windows GUI

The Windows x64 graphical miner uses the same policy: AUTO (recommended), NVIDIA CUDA (explicit), or CPU (force universal fallback). Pool mode also supports AUTO.

## NVIDIA thermal safety

On Windows, GPU safety uses direct local NVIDIA NVML telemetry. The dashboard and governor share the same hardware sample. On Linux CUDA mining, temperature comes from the locally installed NVIDIA driver utility. If local temperature telemetry fails while thermal protection is enabled, mining stops.

Pool/miner websites and remote hashrate services are never trusted as temperature-control inputs.

In Windows Pool mode, the configured temperature is a **hard safety ceiling, not the normal operating target**. The controller targets about 5 °C below the limit, begins soft regulation one degree before that target, changes duty cycle in small steps, and restores performance much more slowly than it removes heat. If the hard limit is reached, the miner is suspended until the GPU is several degrees below target and then held at a conservative duty for 20 seconds before gradual recovery.

## Portable package contents

Each portable package contains the miner CLI, full-node CLI, network.json, bootstrap.json, this guide and the license when present. Portable CPU packages do not require CUDA.

## Fail-safe rule

- known + trustworthy local sensor -> use it;
- no trustworthy sensor -> do not fabricate a value;
- CPU fallback -> conservative concurrency;
- requested GPU thermal protection without trustworthy telemetry -> stop GPU mining.

## Consensus

Universal Miner v0.4.5 does not change genesis, Network ID, AQM64 consensus parameters, difficulty rules, block/transaction validation or monetary policy.
