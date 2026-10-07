# AuronQ Miner — Windows and Linux accelerated guide

Current release: **v0.4.6 Alpha**

The v0.4.6 miner uses **AUTO** compute selection: current NVIDIA CUDA → CUDA 12.x legacy → CUDA 11.8 Kepler → validated vendor-neutral OpenCL GPU → native CPU AQM64. See `UNIVERSAL-MINER-GUIDE.md` for the complete fallback and safety policy.

## Packages

**Windows x64 full package** includes the PL/EN graphical dashboard, current CUDA, CUDA 12.x legacy, CUDA 11.8 Kepler, vendor-neutral OpenCL, CPU fallback, full-node CLI, Mainnet metadata and documentation.

**Linux amd64 full package** includes the universal CLI miner, `libauronq-aqm64-cuda.so`, `libauronq-aqm64-cuda-legacy.so`, `libauronq-aqm64-cuda-kepler.so`, `libauronq-aqm64-opencl.so`, CPU fallback, full-node CLI, Mainnet metadata and documentation. Portable CPU-only packages are additionally built for Windows x64/ARM64, Linux x64/ARM64 and macOS x64/ARM64.

## Requirements

For **CPU fallback**: a supported 64-bit Windows/Linux/macOS system and enough RAM for at least one 64 MiB AQM64 lane.

For **NVIDIA CUDA acceleration**: NVIDIA GPU supported by one of the architecture targets emitted by the CUDA Toolkit used for that release, a current proprietary driver and Windows/Linux. The v0.4.6 build scripts query `nvcc --list-gpu-code` / `--list-gpu-arch` and compile every baseline architecture target supported by that installed toolkit rather than hard-coding only three generations.

For **OpenCL GPU acceleration**: Windows/Linux plus a working vendor OpenCL runtime from AMD, Intel or NVIDIA. A compatible model name alone is not enough — AuronQ compiles the OpenCL kernel on the local driver and accepts it only after the full canonical AQM64 equivalence self-test.

Older NVIDIA cards such as GTX 1050/1050 Ti are tried through the packaged **CUDA 12.6 legacy backend**. Supported Kepler sm_35/sm_37 cards have a separate **CUDA 11.8 backend**. Neither compatibility path replaces or downgrades the current CUDA 13.2 backend used by newer GPUs. CUDA/OpenCL Toolkits are build-time dependencies only; the packaged OpenCL backend dynamically loads the system OpenCL runtime supplied by the GPU driver.

Solo mining also needs a synchronized AuronQ Mainnet full node; the v0.4.6 packages include the AuronQ node CLI.

Check NVIDIA acceleration, when applicable:

~~~bash
nvidia-smi
~~~

## Verify downloads

Every release contains `SHA256SUMS-AURONQ-MINER.txt`.

Linux:

~~~bash
sha256sum -c SHA256SUMS-AURONQ-MINER.txt --ignore-missing
~~~

Windows PowerShell:

~~~powershell
Get-FileHash .\AuronQ-Miner-v0.4.6-alpha-Windows-x64-GUI-GPU.zip -Algorithm SHA256
~~~

Compare the value with the release checksum.

## Windows quick start

1. Download and extract the entire Windows ZIP.
2. Run `AuronQ-Miner.exe`.
3. Enter a valid AURQ Mainnet reward address.
4. Leave the compute backend on **AUTO**. It tries current CUDA, CUDA 12.x legacy, CUDA 11.8 Kepler, OpenCL GPU, then CPU.
5. Leave Auto Tune enabled; CPU fallback automatically uses its conservative thread profile.
6. Set the maximum GPU temperature.
7. Choose Solo or Pool.
8. Start mining.

In Solo mode the app uses an existing local AuronQ full node on `127.0.0.1:18444` or starts its own validating node.

## Linux quick start

~~~bash
tar -xzf AuronQ-Miner-v0.4.6-alpha-Linux-x64-GPU.tar.gz
cd AuronQ-Miner-v0.4.6-alpha-Linux-x64-GPU
chmod +x auronq-miner
~~~

Run the mandatory correctness test:

~~~bash
./auronq-miner --self-test --device 0
~~~

Expected result:

~~~text
SELF-TEST OK
~~~

Benchmark:

~~~bash
./auronq-miner --benchmark --benchmark-seconds 20 --device 0
~~~

Solo mine on all NVIDIA GPUs:

~~~bash
./auronq-miner \
  --node http://127.0.0.1:18444 \
  --address aurq1... \
  --devices all \
  --self-test \
  --auto-tune \
  --thermal-auto \
  --thermal-limit 81
~~~

With an 81 °C hard limit, the automatic target is about 76 °C.

## Multi-GPU

All NVIDIA GPUs:

~~~bash
./auronq-miner --devices all ...
~~~

Selected GPUs:

~~~bash
./auronq-miner --devices 0,1 ...
~~~

Each GPU receives a separate worker and disjoint nonce range. The parent process aggregates current H/s.

## Smart thermal control

Recommended options:

~~~text
--thermal-auto --thermal-limit 81
~~~

The configured limit is the emergency-cooldown threshold. Normal operation targets about 5 °C below it. On Windows, thermal decisions use direct local NVIDIA NVML hardware telemetry only. The Pool governor samples the shared NVML hardware state twice per second, applies short adaptive duty-cycle pulses and restores performance gradually with hysteresis instead of repeatedly jumping between full load and long pauses. MeshMiner/pool-reported temperatures are not used for safety control.

At the configured limit (81 °C in the recommended profile), the external pool miner is suspended automatically until the GPU is stably cooled below the target, then the same process resumes with a conservative duty cycle. A separate catastrophic envelope stops the miner if the suspended GPU exceeds the bounded thermal-inertia allowance (currently up to about limit + 4 °C), while sustained rising temperature during suspension can stop it earlier. Stale/missing direct NVML telemetry or a suspend/resume control failure also stops mining rather than allowing operation with guessed data.

Recommended profile for laptop GPUs:

~~~text
thermal limit: 81 °C
automatic target: ~76 °C
emergency cooldown: 81 °C
catastrophic envelope: 85 °C (continued rise can stop earlier)
~~~

No manual restart is normally required after an ordinary thermal excursion.

v0.4.6 further reduces thermal oscillation: Pool mode begins regulation before the target, limits each duty-cycle change to small steps, refuses to release throttling while temperature is flat/rising near target, and uses a 20-second conservative stabilization hold after an emergency cooldown. This is designed to avoid repeated full-load → hard-pause → full-load cycles on laptop GPUs.

On Linux, driver telemetry can also be watched with:

~~~bash
watch -n 1 nvidia-smi
~~~

## Live H/s

The native miner reports approximately once per second:

~~~text
hashes=... rate=... H/s avg=... H/s current_height=... batch=...
~~~

Multi-GPU mode reports the aggregated rate.

## Pool mining / MeshMiner

Windows GUI Pool mode can launch a user-supplied MeshMiner 0.8.35+ binary. Pool GPU selection is intentionally **not filtered through AuronQ's native CUDA compatibility test**; an older NVIDIA card is passed to the external pool miner when it is locally detected, and the external miner decides whether its own CUDA implementation supports that card. Explicit CUDA/CPU/both modes, custom compatible endpoints, selected GPUs, conservative CPU threads, Retune, live H/s parsing and the autonomous AuronQ-side GPU thermal governor are supported.

MeshMiner is independently developed and is not bundled or automatically downloaded by AuronQ.

- MeshMiner: https://github.com/totom9000/meshminer/releases/tag/v.0.8.35
- MeshPool: https://meshpool.net/pool/auronq-main
- RPlant: https://pool.rplant.xyz/#auronq#connect

On Linux, compatible third-party pool miners can be run directly. The official v0.4.6 Linux miner provides **Solo AUTO** with current CUDA → CUDA 12.x legacy → CUDA 11.8 Kepler → OpenCL GPU → CPU fallback.

## Main CLI options

| Option | Meaning |
|---|---|
| `--backend auto|cuda|opencl|cpu` | AUTO is recommended: current CUDA → CUDA 12.x legacy → CUDA 11.8 Kepler → OpenCL GPU → CPU |
| `--cpu-threads N` | CPU lanes; 0 = conservative automatic profile, explicit max 16 |
| `--node URL` | AuronQ full-node URL |
| `--address aurq1...` | Mainnet reward address |
| `--device N` | Accelerator device index; CUDA/OpenCL depending on selected backend |
| `--devices all` | All detected NVIDIA GPUs |
| `--devices 0,1` | Selected GPUs |
| `--batch N` | Manual batch; 0 uses backend recommendation |
| `--auto-tune` | Find the fastest safe batch |
| `--auto-tune-seconds N` | Time per Auto Tune candidate |
| `--self-test` | Selected backend vs canonical AQM64 equivalence test |
| `--benchmark` | Offline AQM64 benchmark |
| `--benchmark-seconds N` | Benchmark duration |
| `--thermal-auto` | Smart thermal governor |
| `--thermal-limit N` | Hard temperature ceiling |
| `--thermal-target N` | Optional explicit thermal target |
| `--cuda-dll PATH` | Primary CUDA backend path (.dll or .so) |
| `--cuda-legacy-dll PATH` | CUDA 12.x Maxwell/Pascal/Volta backend path |
| `--cuda-kepler-dll PATH` | CUDA 11.8 Kepler sm_35/sm_37 backend path |
| `--opencl-dll PATH` | OpenCL backend path (.dll or .so) |

## Linux troubleshooting

If `nvidia-smi` is unavailable, fix/install the proprietary NVIDIA driver first.

If the CUDA shared library cannot be opened, keep `libauronq-aqm64-cuda.so` beside the miner or pass:

~~~bash
./auronq-miner --cuda-dll /full/path/libauronq-aqm64-cuda.so ...
~~~

If a selected GPU backend self-test fails, do not mine Mainnet with that backend. AUTO automatically tries the next backend; explicit `--backend cuda` or `--backend opencl` fails closed.

## Build from source

Windows CUDA backend:

~~~powershell
powershell -ExecutionPolicy Bypass -File .\gpu\cuda\build-windows.ps1
~~~

Linux CUDA backend:

~~~bash
bash ./gpu/cuda/build-linux.sh
~~~

Windows legacy CUDA backend (requires a CUDA 12.x toolkit that still emits the selected legacy targets):

~~~powershell
powershell -ExecutionPolicy Bypass -File .\gpu\cuda\build-windows-legacy.ps1
~~~

Linux legacy CUDA backend:

~~~bash
bash ./gpu/cuda/build-linux-legacy.sh
~~~

Windows Kepler CUDA backend:

~~~powershell
powershell -ExecutionPolicy Bypass -File .\gpu\cuda\build-windows-kepler.ps1
~~~

Linux Kepler CUDA backend:

~~~bash
bash ./gpu/cuda/build-linux-kepler.sh
~~~

Windows OpenCL backend:

~~~powershell
powershell -ExecutionPolicy Bypass -File .\gpu\opencl\build-windows.ps1
~~~

Linux OpenCL backend:

~~~bash
bash ./gpu/opencl/build-linux.sh
~~~

Linux package:

~~~bash
bash ./build-gpu-miner-linux.sh 0.4.6-alpha
~~~

## Status

Universal Miner remains alpha software. CUDA and OpenCL accelerator implementations have not received an independent professional security or cryptographic audit. OpenCL device compatibility is established at runtime by the mandatory canonical self-test, but real-world stability and performance still need validation across more vendors/models. Valid Solo blocks are always subject to ordinary AuronQ full-node Mainnet validation.

AuronQ does not ship an ASIC-specific work protocol, ASIC bridge or privileged hardware integration in this miner.
