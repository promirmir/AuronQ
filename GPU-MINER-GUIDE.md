# AuronQ GPU Miner — Windows and Linux guide

Current release: **v0.3.8 Alpha**

The official NVIDIA CUDA miner supports Windows x64 and Linux amd64.

## Packages

**Windows x64** includes the PL/EN graphical dashboard, the native CUDA worker, CUDA DLL, Mainnet metadata and documentation.

**Linux amd64** includes the native CLI miner, `libauronq-aqm64-cuda.so`, Mainnet metadata and documentation. v0.3.8 is CLI-first on Linux: native Solo CUDA mining, multi-GPU, Auto Tune, live H/s and smart thermal control are supported. The Windows dashboard GUI is not claimed as a Linux feature.

## Requirements

- x86-64 / amd64
- NVIDIA GPU with CUDA Compute Capability 7.5+
- current proprietary NVIDIA driver
- working `nvidia-smi`
- synchronized AuronQ Mainnet full node for Solo mining

The packaged release statically links the CUDA runtime. CUDA Toolkit is needed only when building the CUDA backend from source.

Check the driver:

~~~bash
nvidia-smi
~~~

## Verify downloads

Every release contains `SHA256SUMS-GPU-MINER.txt`.

Linux:

~~~bash
sha256sum -c SHA256SUMS-GPU-MINER.txt --ignore-missing
~~~

Windows PowerShell:

~~~powershell
Get-FileHash .\AuronQ-GPU-Miner-v0.3.8-alpha-Windows-x64.zip -Algorithm SHA256
~~~

Compare the value with the release checksum.

## Windows quick start

1. Download and extract the entire Windows ZIP.
2. Run `AuronQ-GPU-Miner.exe`.
3. Enter a valid AURQ Mainnet reward address.
4. Select all NVIDIA GPUs or individual cards.
5. Leave Auto Tune enabled for the first run.
6. Set the maximum GPU temperature.
7. Choose Solo or Pool.
8. Start mining.

In Solo mode the app uses an existing local AuronQ full node on `127.0.0.1:18444` or starts its own validating node.

## Linux quick start

~~~bash
tar -xzf AuronQ-GPU-Miner-v0.3.8-alpha-Linux-amd64.tar.gz
cd AuronQ-GPU-Miner-v0.3.8-alpha-Linux-amd64
chmod +x auronq-gpu-miner
~~~

Run the mandatory correctness test:

~~~bash
./auronq-gpu-miner --self-test --device 0
~~~

Expected result:

~~~text
SELF-TEST OK
~~~

Benchmark:

~~~bash
./auronq-gpu-miner --benchmark --benchmark-seconds 20 --device 0
~~~

Solo mine on all NVIDIA GPUs:

~~~bash
./auronq-gpu-miner \
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
./auronq-gpu-miner --devices all ...
~~~

Selected GPUs:

~~~bash
./auronq-gpu-miner --devices 0,1 ...
~~~

Each GPU receives a separate worker and disjoint nonce range. The parent process aggregates current H/s.

## Smart thermal control

Recommended options:

~~~text
--thermal-auto --thermal-limit 81
~~~

The configured limit is the emergency-cooldown threshold. Normal operation targets about 5 °C below it. On Windows, thermal decisions use direct local NVIDIA NVML hardware telemetry only. The Pool governor samples the shared NVML hardware state twice per second, applies short adaptive duty-cycle pulses and restores performance gradually with hysteresis instead of repeatedly jumping between full load and long pauses. MeshMiner/pool-reported temperatures are not used for safety control.

At the configured limit (81 °C in the recommended profile), the external pool miner is suspended automatically until the GPU is stably cooled below the target, then the same process resumes with a conservative duty cycle. A separate catastrophic fail-safe stops the miner at approximately limit + 1 °C if temperature keeps rising despite suspension. Stale/missing direct NVML telemetry or a suspend/resume control failure also stops mining rather than allowing operation with guessed data.

Recommended profile for laptop GPUs:

~~~text
thermal limit: 81 °C
automatic target: ~76 °C
emergency cooldown: 81 °C
catastrophic fail-safe: 83 °C
~~~

No manual restart is normally required after an ordinary thermal excursion.

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

Windows GUI Pool mode can launch a user-supplied MeshMiner 0.8.35+ binary. It supports MeshPool preset, custom compatible endpoints, CUDA/CPU/both backends, selected GPUs, CPU thread override, Retune, live H/s parsing and the autonomous AuronQ-side thermal governor with automatic cooldown/resume.

MeshMiner is independently developed and is not bundled or automatically downloaded by AuronQ.

- MeshMiner: https://github.com/totom9000/meshminer/releases/tag/v.0.8.35
- MeshPool: https://meshpool.net/pool/auronq-main
- RPlant: https://pool.rplant.xyz/#auronq#connect

On Linux, compatible third-party pool miners can be run directly. The official Linux AuronQ binary in v0.3.8 provides the native **Solo CUDA** path.

## Main CLI options

| Option | Meaning |
|---|---|
| `--node URL` | AuronQ full-node URL |
| `--address aurq1...` | Mainnet reward address |
| `--device N` | Single CUDA GPU |
| `--devices all` | All detected NVIDIA GPUs |
| `--devices 0,1` | Selected GPUs |
| `--batch N` | Manual batch; 0 uses backend recommendation |
| `--auto-tune` | Find the fastest safe batch |
| `--auto-tune-seconds N` | Time per Auto Tune candidate |
| `--self-test` | GPU/CPU AQM64 equivalence test |
| `--benchmark` | Offline AQM64 benchmark |
| `--benchmark-seconds N` | Benchmark duration |
| `--thermal-auto` | Smart thermal governor |
| `--thermal-limit N` | Hard temperature ceiling |
| `--thermal-target N` | Optional explicit thermal target |
| `--cuda-dll PATH` | CUDA backend path (.dll or .so) |

## Linux troubleshooting

If `nvidia-smi` is unavailable, fix/install the proprietary NVIDIA driver first.

If the CUDA shared library cannot be opened, keep `libauronq-aqm64-cuda.so` beside the miner or pass:

~~~bash
./auronq-gpu-miner --cuda-dll /full/path/libauronq-aqm64-cuda.so ...
~~~

If the GPU self-test fails, do not mine Mainnet with that GPU/build. Update the driver, re-extract the official package and test again.

## Build from source

Windows CUDA backend:

~~~powershell
powershell -ExecutionPolicy Bypass -File .\gpu\cuda\build-windows.ps1
~~~

Linux CUDA backend:

~~~bash
bash ./gpu/cuda/build-linux.sh
~~~

Linux package:

~~~bash
bash ./build-gpu-miner-linux.sh 0.3.8-alpha
~~~

## Status

GPU Miner remains alpha software. The CUDA implementation has not received an independent professional security or cryptographic audit. Valid Solo blocks are still subject to ordinary AuronQ full-node Mainnet validation.
