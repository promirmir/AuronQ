# AuronQ Universal Miner v0.4.4 Alpha

## Reliable AUTO fallback

v0.4.4 fixes an important Windows AUTO-selection problem found on older NVIDIA hardware such as GTX 1050-class Pascal GPUs.

Previously, the GUI could treat "visible through NVML" as equivalent to "usable by the current AuronQ CUDA backend". That is not always true: an NVIDIA GPU can expose normal telemetry while the bundled CUDA backend still cannot execute its kernels on that architecture.

v0.4.4 now validates the actual CUDA path before AUTO commits to GPU mining:

- each selected NVIDIA device is checked with the official AuronQ CUDA backend and the canonical AQM64 self-test;
- only devices that report `SELF-TEST OK` are accepted as usable CUDA devices;
- in AUTO mode, failed CUDA validation automatically switches to the native CPU AQM64 fallback;
- unsupported/older NVIDIA hardware, driver problems or CUDA backend initialization failures therefore no longer require manual CPU selection;
- explicit CUDA/BOTH remains fail-closed and reports the validation error instead of silently changing the requested backend.

The validation result is cached for the running GUI session so ordinary starts do not repeatedly benchmark the same device.

## Windows CPU telemetry

The Windows GUI now shows a dedicated CPU/AQM64 telemetry card even when an NVIDIA GPU is physically present.

Displayed CPU data comes from local Windows sources:

- processor name;
- whole-system CPU utilization from the native Windows `GetSystemTimes` API;
- logical processor count;
- AQM64 worker-thread count;
- nominal CPU clock reported by the Windows hardware registry;
- estimated AQM64 working memory at 64 MiB per active lane.

CPU package temperature is intentionally shown as **N/A** when there is no trustworthy universal sensor source. AuronQ does not invent a CPU temperature or reuse unrelated ACPI thermal-zone values.

When AUTO falls back to CPU, the dashboard clearly marks **CPU ACTIVE** while still showing any detected NVIDIA hardware separately.

## Existing GPU thermal safety retained

The stabilized v0.4.3 Pool-mode governor remains unchanged in its safety logic:

- direct local NVML telemetry remains authoritative;
- regulation begins before the target and releases load gradually;
- emergency cooldown uses suspend/cool/resume with a stabilization hold;
- a bounded post-suspend thermal-inertia overshoot is tolerated;
- continued heating while suspended can stop mining earlier;
- the independent catastrophic envelope is currently around hard limit + 4 °C.

The GUI/documentation now reports this actual envelope instead of the older +1 °C wording.

## Supported packages

The release continues to provide:

- Windows x64 GUI/CUDA package;
- Linux x64 CUDA package;
- portable CPU-safe CLI packages for Windows x64/ARM64, Linux x64/ARM64 and macOS x64/ARM64.

## Consensus unchanged

No changes to genesis, Network ID, AQM64 parameters, difficulty, monetary policy, chain selection or transaction/block validation.
