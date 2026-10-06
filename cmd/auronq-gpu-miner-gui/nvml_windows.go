//go:build windows

package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

const nvmlSuccess = 0

type nvmlAPI struct {
	dll           *syscall.DLL
	init          *syscall.Proc
	deviceCount   *syscall.Proc
	deviceHandle  *syscall.Proc
	deviceName    *syscall.Proc
	deviceUUID    *syscall.Proc
	temperature   *syscall.Proc
	utilization   *syscall.Proc
	memoryInfo    *syscall.Proc
	powerUsage    *syscall.Proc
	powerLimit    *syscall.Proc
	clockInfo     *syscall.Proc
	perfState     *syscall.Proc
	fanSpeed      *syscall.Proc
}

type cudaDriverAPI struct {
	dll            *syscall.DLL
	init           *syscall.Proc
	deviceCount    *syscall.Proc
	deviceGet      *syscall.Proc
	deviceGetUUID  *syscall.Proc
}

type nvmlUtilization struct {
	GPU    uint32
	Memory uint32
}

type nvmlMemory struct {
	Total uint64
	Free  uint64
	Used  uint64
}

var (
	nvmlOnce sync.Once
	nvmlInst *nvmlAPI
	nvmlErr  error

	cudaOnce sync.Once
	cudaInst *cudaDriverAPI
	cudaErr  error
)

func trustedWindowsDLLCandidates(name string, extra ...string) []string {
	var out []string
	if root := strings.TrimSpace(os.Getenv("SystemRoot")); root != "" {
		out = append(out, filepath.Join(root, "System32", name))
	}
	for _, base := range []string{os.Getenv("ProgramW6432"), os.Getenv("ProgramFiles")} {
		if strings.TrimSpace(base) != "" {
			out = append(out, filepath.Join(base, "NVIDIA Corporation", "NVSMI", name))
		}
	}
	out = append(out, extra...)
	return out
}

func loadTrustedDLL(name string, candidates ...string) (*syscall.DLL, error) {
	var last error
	for _, p := range candidates {
		if strings.TrimSpace(p) == "" {
			continue
		}
		if st, err := os.Stat(p); err != nil || st.IsDir() {
			continue
		}
		dll, err := syscall.LoadDLL(p)
		if err == nil {
			return dll, nil
		}
		last = err
	}
	if last != nil {
		return nil, fmt.Errorf("%s: %w", name, last)
	}
	return nil, fmt.Errorf("%s not found in trusted NVIDIA/Windows locations", name)
}

func procAny(dll *syscall.DLL, names ...string) (*syscall.Proc, error) {
	for _, name := range names {
		if p, err := dll.FindProc(name); err == nil {
			return p, nil
		}
	}
	return nil, fmt.Errorf("missing procedure: %s", strings.Join(names, " or "))
}

func optionalProc(dll *syscall.DLL, names ...string) *syscall.Proc {
	p, _ := procAny(dll, names...)
	return p
}

func directNVML() (*nvmlAPI, error) {
	nvmlOnce.Do(func() {
		dll, err := loadTrustedDLL("nvml.dll", trustedWindowsDLLCandidates("nvml.dll")...)
		if err != nil {
			nvmlErr = err
			return
		}
		api := &nvmlAPI{dll: dll}
		if api.init, err = procAny(dll, "nvmlInit_v2", "nvmlInit"); err != nil {
			nvmlErr = err
			return
		}
		if api.deviceCount, err = procAny(dll, "nvmlDeviceGetCount_v2", "nvmlDeviceGetCount"); err != nil {
			nvmlErr = err
			return
		}
		if api.deviceHandle, err = procAny(dll, "nvmlDeviceGetHandleByIndex_v2", "nvmlDeviceGetHandleByIndex"); err != nil {
			nvmlErr = err
			return
		}
		if api.temperature, err = procAny(dll, "nvmlDeviceGetTemperature"); err != nil {
			nvmlErr = err
			return
		}
		api.deviceName = optionalProc(dll, "nvmlDeviceGetName")
		api.deviceUUID = optionalProc(dll, "nvmlDeviceGetUUID")
		api.utilization = optionalProc(dll, "nvmlDeviceGetUtilizationRates")
		api.memoryInfo = optionalProc(dll, "nvmlDeviceGetMemoryInfo")
		api.powerUsage = optionalProc(dll, "nvmlDeviceGetPowerUsage")
		api.powerLimit = optionalProc(dll, "nvmlDeviceGetPowerManagementLimit")
		api.clockInfo = optionalProc(dll, "nvmlDeviceGetClockInfo")
		api.perfState = optionalProc(dll, "nvmlDeviceGetPerformanceState")
		api.fanSpeed = optionalProc(dll, "nvmlDeviceGetFanSpeed")

		if rc, _, _ := api.init.Call(); uint32(rc) != nvmlSuccess {
			nvmlErr = fmt.Errorf("nvmlInit failed with code %d", uint32(rc))
			return
		}
		nvmlInst = api
	})
	return nvmlInst, nvmlErr
}

func directCUDA() (*cudaDriverAPI, error) {
	cudaOnce.Do(func() {
		dll, err := loadTrustedDLL("nvcuda.dll", trustedWindowsDLLCandidates("nvcuda.dll")...)
		if err != nil {
			cudaErr = err
			return
		}
		api := &cudaDriverAPI{dll: dll}
		if api.init, err = procAny(dll, "cuInit"); err != nil {
			cudaErr = err
			return
		}
		if api.deviceCount, err = procAny(dll, "cuDeviceGetCount"); err != nil {
			cudaErr = err
			return
		}
		if api.deviceGet, err = procAny(dll, "cuDeviceGet"); err != nil {
			cudaErr = err
			return
		}
		if api.deviceGetUUID, err = procAny(dll, "cuDeviceGetUuid_v2", "cuDeviceGetUuid"); err != nil {
			cudaErr = err
			return
		}
		if rc, _, _ := api.init.Call(0); uint32(rc) != 0 {
			cudaErr = fmt.Errorf("cuInit failed with code %d", uint32(rc))
			return
		}
		cudaInst = api
	})
	return cudaInst, cudaErr
}

func callNVML(proc *syscall.Proc, args ...uintptr) error {
	if proc == nil {
		return errors.New("NVML procedure unavailable")
	}
	rc, _, _ := proc.Call(args...)
	if uint32(rc) != nvmlSuccess {
		return fmt.Errorf("NVML error %d", uint32(rc))
	}
	return nil
}

func cString(buf []byte) string {
	if i := strings.IndexByte(string(buf), 0); i >= 0 {
		buf = buf[:i]
	}
	return strings.TrimSpace(string(buf))
}

func parseGPUUUID(raw string) ([16]byte, error) {
	var out [16]byte
	s := strings.TrimSpace(raw)
	s = strings.TrimPrefix(strings.ToUpper(s), "GPU-")
	s = strings.ReplaceAll(s, "-", "")
	if len(s) != 32 {
		return out, fmt.Errorf("unexpected NVIDIA UUID %q", raw)
	}
	b, err := hex.DecodeString(s)
	if err != nil || len(b) != len(out) {
		return out, fmt.Errorf("invalid NVIDIA UUID %q", raw)
	}
	copy(out[:], b)
	return out, nil
}

func cudaOrdinalMap() (map[[16]byte]int, error) {
	api, err := directCUDA()
	if err != nil {
		return nil, err
	}
	var count int32
	if rc, _, _ := api.deviceCount.Call(uintptr(unsafe.Pointer(&count))); uint32(rc) != 0 {
		return nil, fmt.Errorf("cuDeviceGetCount failed with code %d", uint32(rc))
	}
	out := make(map[[16]byte]int, int(count))
	for ordinal := int32(0); ordinal < count; ordinal++ {
		var dev int32
		if rc, _, _ := api.deviceGet.Call(uintptr(unsafe.Pointer(&dev)), uintptr(ordinal)); uint32(rc) != 0 {
			return nil, fmt.Errorf("cuDeviceGet(%d) failed with code %d", ordinal, uint32(rc))
		}
		var uuid [16]byte
		if rc, _, _ := api.deviceGetUUID.Call(uintptr(unsafe.Pointer(&uuid[0])), uintptr(uint32(dev))); uint32(rc) != 0 {
			return nil, fmt.Errorf("cuDeviceGetUuid(%d) failed with code %d", ordinal, uint32(rc))
		}
		out[uuid] = int(ordinal)
	}
	return out, nil
}

func queryDirectNVIDIAGPUs() ([]gpuInfo, error) {
	api, err := directNVML()
	if err != nil {
		return nil, fmt.Errorf("direct NVIDIA hardware telemetry unavailable: %w", err)
	}

	var count uint32
	if err := callNVML(api.deviceCount, uintptr(unsafe.Pointer(&count))); err != nil {
		return nil, fmt.Errorf("NVML device count: %w", err)
	}
	if count == 0 {
		return nil, errors.New("NVML reports no NVIDIA GPU")
	}

	ordinals, mapErr := cudaOrdinalMap()
	if mapErr != nil && count > 1 {
		return nil, fmt.Errorf("cannot safely map NVML sensors to CUDA devices: %w", mapErr)
	}

	sampleMS := time.Now().UnixMilli()
	gpus := make([]gpuInfo, 0, count)
	for nvmlIndex := uint32(0); nvmlIndex < count; nvmlIndex++ {
		var device uintptr
		if err := callNVML(api.deviceHandle, uintptr(nvmlIndex), uintptr(unsafe.Pointer(&device))); err != nil {
			return nil, fmt.Errorf("NVML GPU %d handle: %w", nvmlIndex, err)
		}

		var temp uint32
		if err := callNVML(api.temperature, device, 0, uintptr(unsafe.Pointer(&temp))); err != nil {
			return nil, fmt.Errorf("NVML GPU %d temperature: %w", nvmlIndex, err)
		}

		name := fmt.Sprintf("NVIDIA GPU %d", nvmlIndex)
		if api.deviceName != nil {
			buf := make([]byte, 128)
			if callNVML(api.deviceName, device, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf))) == nil {
				if v := cString(buf); v != "" {
					name = v
				}
			}
		}

		cudaIndex := int(nvmlIndex)
		if api.deviceUUID != nil {
			buf := make([]byte, 96)
			if callNVML(api.deviceUUID, device, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf))) == nil {
				if uuid, e := parseGPUUUID(cString(buf)); e == nil {
					if ordinal, ok := ordinals[uuid]; ok {
						cudaIndex = ordinal
					} else if count > 1 {
						return nil, fmt.Errorf("NVML GPU %d UUID does not match any CUDA device; refusing unsafe sensor mapping", nvmlIndex)
					}
				}
			}
		}

		info := gpuInfo{
			Index:          cudaIndex,
			Name:           name,
			TemperatureC:   int(temp),
			FanPercent:     -1,
			UtilPercent:    -1,
			MemoryUsedMiB:  -1,
			MemoryTotalMiB: -1,
			PowerW:         -1,
			PowerLimitW:    -1,
			CoreClockMHz:   -1,
			MemoryClockMHz: -1,
			PState:         "",
			TelemetrySource: "NVML_DIRECT",
			SampleUnixMS:    sampleMS,
		}

		if api.utilization != nil {
			var u nvmlUtilization
			if callNVML(api.utilization, device, uintptr(unsafe.Pointer(&u))) == nil {
				info.UtilPercent = int(u.GPU)
			}
		}
		if api.memoryInfo != nil {
			var m nvmlMemory
			if callNVML(api.memoryInfo, device, uintptr(unsafe.Pointer(&m))) == nil {
				info.MemoryUsedMiB = int(m.Used / (1024 * 1024))
				info.MemoryTotalMiB = int(m.Total / (1024 * 1024))
			}
		}
		if api.powerUsage != nil {
			var mw uint32
			if callNVML(api.powerUsage, device, uintptr(unsafe.Pointer(&mw))) == nil {
				info.PowerW = float64(mw) / 1000.0
			}
		}
		if api.powerLimit != nil {
			var mw uint32
			if callNVML(api.powerLimit, device, uintptr(unsafe.Pointer(&mw))) == nil {
				info.PowerLimitW = float64(mw) / 1000.0
			}
		}
		if api.clockInfo != nil {
			var mhz uint32
			if callNVML(api.clockInfo, device, 0, uintptr(unsafe.Pointer(&mhz))) == nil {
				info.CoreClockMHz = int(mhz)
			}
			mhz = 0
			if callNVML(api.clockInfo, device, 2, uintptr(unsafe.Pointer(&mhz))) == nil {
				info.MemoryClockMHz = int(mhz)
			}
		}
		if api.perfState != nil {
			var p uint32
			if callNVML(api.perfState, device, uintptr(unsafe.Pointer(&p))) == nil && p <= 15 {
				info.PState = fmt.Sprintf("P%d", p)
			}
		}
		if api.fanSpeed != nil {
			var fan uint32
			if callNVML(api.fanSpeed, device, uintptr(unsafe.Pointer(&fan))) == nil {
				info.FanPercent = int(fan)
			}
		}
		gpus = append(gpus, info)
	}
	return gpus, nil
}
