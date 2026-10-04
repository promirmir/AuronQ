# AuronQ GPU Miner — CUDA prototype

This directory contains the first standalone NVIDIA/CUDA miner for AuronQ's
AQM64 proof-of-work. It is intentionally separate from AuronQ Desktop and does
not change Mainnet consensus.

## Status

Prototype / review stage. Do not treat it as a finished release yet.

Real-device validation has now passed on an NVIDIA GeForce RTX 4050 Laptop GPU
with CUDA 13.4: the mandatory self-test produced a byte-identical full AQM64
result between the CUDA backend and the canonical CPU PowHash implementation.

Measured end-to-end offline benchmark on that device:

- batch 20: 118.733 H/s
- batch 40: 226.263 H/s (30 s repeat: 226.447 H/s)
- batch 60: 316.227 H/s
- batch 64: 299.377 H/s

Batch 60 was the best of the tested values on this RTX 4050 Laptop GPU. This is
a single-device prototype measurement, not a guaranteed performance figure;
laptop power limits, thermals, clocks and batch size can materially change
throughput.

The heavy Argon2id memory graph runs on CUDA. SHAKE256 domain separation,
Argon2 initialization/final extraction, target comparison, template handling
and block submission stay in the Go miner so they reuse AuronQ's existing
consensus implementation wherever practical.

AQM64 uses 64 MiB of memory per candidate. The CUDA backend therefore mines a
batch of independent candidates concurrently and automatically recommends a
conservative batch from free VRAM and the GPU SM count.

## Requirements

- Windows x64
- NVIDIA GPU with CUDA Compute Capability 7.5+
- NVIDIA driver
- CUDA Toolkit 13.4.x (or another compatible CUDA 13.x toolkit) to build the DLL from source
- a running AuronQ full node, normally http://127.0.0.1:18444

RTX 20/30/40-class cards are the initial target. RTX 4050 Laptop GPU is supported by the current sm_89 build target. The first implementation is
correctness-first; kernel tuning comes after device self-test and real hardware
benchmarks.

## Build

Build the Go miner from the repository root:

~~~powershell
go build -o auronq-gpu-miner.exe ./cmd/auronq-gpu-miner
~~~

Build the CUDA DLL:

~~~powershell
powershell -ExecutionPolicy Bypass -File .\gpu\cuda\build-windows.ps1
copy .\gpu\cuda\auronq-aqm64-cuda.dll .\
~~~

## Windows application

The Windows application is built as `AuronQ-GPU-Miner.exe` and intentionally
uses the same visual language as AuronQ Desktop: dark sidebar, dashboard cards,
network status and live logs. The UI is available in **Polish and English**.

The application is designed to be standalone:

- if an AuronQ Mainnet full node is already listening on `127.0.0.1:18444`,
  the miner verifies its Network ID and uses it;
- otherwise it starts its own full validating node on TCP/18444;
- it loads the bundled Mainnet `network.json` and `bootstrap.json`, learns
  more peers through normal P2P gossip and persists a public peer store;
- before Mainnet mining it checks the local height against reachable bootstrap
  peers and refuses to start on an obviously stale local tip;
- when “support the network as a public node” is enabled, it attempts UPnP/IGD
  mapping for TCP/18444, sets the embedded node's public advertise endpoint and
  proactively announces that endpoint to bootstrap peers;
- if the app is using an already-running local AuronQ Desktop node, the same
  UPnP mapping is created and the endpoint is announced externally so remote
  peers can callback-verify the local node;
- CGNAT, disabled UPnP or router failure are non-fatal: the node remains
  outbound-only and mining continues.

The application provides reward-address, CUDA-device and batch controls,
Start/Stop, mandatory GPU/CPU AQM64 self-test, a 15-second offline benchmark,
live hashrate/block statistics, node height/peer/public-endpoint status and
technical logs. Non-secret preferences are stored under the user's Windows
AuronQ configuration directory.

The GUI launches the sibling `auronq-gpu-worker.exe` CUDA worker with its
console hidden. The distinct filename is required on Windows because paths are
case-insensitive by default; using only `AuronQ-GPU-Miner.exe` vs
`auronq-gpu-miner.exe` would make the GUI launch itself instead of the worker. Every candidate block is still submitted to and fully validated
by the ordinary AuronQ full node.

For a complete Windows package, run from the repository root:

~~~powershell
powershell -ExecutionPolicy Bypass -File .\build-gpu-miner-windows.ps1
~~~

The script builds the CUDA DLL, CLI worker and Windows app, copies the immutable
Mainnet configuration/bootstrap metadata, performs the mandatory GPU/CPU AQM64
self-test, and creates `AuronQ-GPU-Miner-v0.2.0-alpha-Windows-x64.zip` under
`dist`.

## Mandatory device self-test before mining

Run:

~~~powershell
.\auronq-gpu-miner.exe --self-test
~~~

The self-test performs one full 64 MiB / time-cost-2 Argon2id calculation on
the GPU and independently computes the same result with AuronQ's CPU reference.
The miner reports SELF-TEST OK only if the outputs are byte-identical.

This is important because a CUDA kernel that is merely fast but differs by one
bit from AQM64 can never produce a valid Mainnet block.

## Mine

~~~powershell
.\auronq-gpu-miner.exe --node http://127.0.0.1:18444 --address aurq1... --device 0
~~~

Optional flags:

- --batch N: candidates processed concurrently; 0 = automatic
- --cuda-dll PATH: explicit path to auronq-aqm64-cuda.dll
- --self-test: GPU/CPU equivalence test before mining
- --benchmark: offline end-to-end AQM64 throughput benchmark; does not connect to a node or submit blocks
- --benchmark-seconds N: approximate benchmark duration (default 20 seconds)

The miner obtains a block template from the local full node, searches nonces on
the GPU, detects canonical-tip changes between batches, and submits a candidate
block only after the final AQM64 digest satisfies the template target.

## Security / correctness notes

- The full node remains the authority that validates submitted blocks.
- No privileged mining endpoint or consensus shortcut is introduced.
- The CUDA code is new and has not received an independent audit.
- Real-device equivalence testing has passed on RTX 4050 Laptop GPU; more GPU models still need coverage.
- The Windows app's embedded/public-node path uses the same AuronQ full-node and P2P implementation as the main project.
- Public-node enablement is best-effort and callback-verified; it never treats UPnP success alone as consensus or peer trust.
- Performance numbers should not be advertised until measured with the offline benchmark on actual GPUs.
