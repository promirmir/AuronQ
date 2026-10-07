//go:build windows

package main

import (
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"unsafe"
)

type cpuTelemetry struct {
	Name              string  `json:"name"`
	UsagePercent      float64 `json:"usage_percent"`
	LogicalProcessors int     `json:"logical_processors"`
	MiningThreads     int     `json:"mining_threads"`
	NominalClockMHz   int     `json:"nominal_clock_mhz"`
	AQM64MemoryMiB    int     `json:"aqm64_memory_mib"`
	TemperatureC      float64 `json:"temperature_c"`
	TelemetrySource   string  `json:"telemetry_source"`
}

type cpuSampler struct {
	mu sync.Mutex

	name string
	mhz  int

	havePrevious bool
	lastIdle     uint64
	lastKernel   uint64
	lastUser     uint64
}

type winFiletime struct {
	LowDateTime  uint32
	HighDateTime uint32
}

var (
	cpuKernel32       = syscall.NewLazyDLL("kernel32.dll")
	cpuGetSystemTimes = cpuKernel32.NewProc("GetSystemTimes")
)

func filetimeValue(v winFiletime) uint64 {
	return uint64(v.HighDateTime)<<32 | uint64(v.LowDateTime)
}

func readSystemTimes() (idle, kernel, user uint64, ok bool) {
	var i, k, u winFiletime
	r1, _, _ := cpuGetSystemTimes.Call(
		uintptr(unsafe.Pointer(&i)),
		uintptr(unsafe.Pointer(&k)),
		uintptr(unsafe.Pointer(&u)),
	)
	if r1 == 0 {
		return 0, 0, 0, false
	}
	return filetimeValue(i), filetimeValue(k), filetimeValue(u), true
}

func regCPUValue(name string) string {
	out, err := exec.Command("reg.exe", "query",
		`HKLM\HARDWARE\DESCRIPTION\System\CentralProcessor\0`,
		"/v", name,
	).CombinedOutput()
	if err != nil {
		return ""
	}
	for _, raw := range strings.Split(string(out), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || !strings.Contains(strings.ToLower(line), strings.ToLower(name)) {
			continue
		}
		fields := strings.Fields(line)
		for i, f := range fields {
			if strings.HasPrefix(strings.ToUpper(f), "REG_") && i+1 < len(fields) {
				return strings.TrimSpace(strings.Join(fields[i+1:], " "))
			}
		}
	}
	return ""
}

func detectCPUIdentity() (string, int) {
	name := strings.TrimSpace(regCPUValue("ProcessorNameString"))
	mhz := 0
	if raw := strings.TrimSpace(regCPUValue("~MHz")); raw != "" {
		if v, err := strconv.ParseInt(raw, 0, 64); err == nil && v > 0 {
			mhz = int(v)
		}
	}
	if name == "" {
		name = "CPU"
	}
	return name, mhz
}

func newCPUSampler() *cpuSampler {
	name, mhz := detectCPUIdentity()
	return &cpuSampler{name: name, mhz: mhz}
}

func (s *cpuSampler) Sample(logicalProcessors, miningThreads int) cpuTelemetry {
	if logicalProcessors < 1 {
		logicalProcessors = 1
	}
	if miningThreads < 1 {
		miningThreads = 1
	}

	out := cpuTelemetry{
		Name:              s.name,
		UsagePercent:      -1,
		LogicalProcessors: logicalProcessors,
		MiningThreads:     miningThreads,
		NominalClockMHz:   s.mhz,
		AQM64MemoryMiB:    miningThreads * 64,
		TemperatureC:      -1, // no universal trustworthy Windows package-temperature API
		TelemetrySource:   "WINDOWS_SYSTEM_TIMES",
	}

	idle, kernel, user, ok := readSystemTimes()
	if !ok {
		return out
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.havePrevious {
		dIdle := idle - s.lastIdle
		dKernel := kernel - s.lastKernel
		dUser := user - s.lastUser
		total := dKernel + dUser
		if total > 0 && total >= dIdle {
			busy := total - dIdle
			out.UsagePercent = float64(busy) * 100 / float64(total)
			if out.UsagePercent < 0 {
				out.UsagePercent = 0
			}
			if out.UsagePercent > 100 {
				out.UsagePercent = 100
			}
		}
	}

	s.lastIdle = idle
	s.lastKernel = kernel
	s.lastUser = user
	s.havePrevious = true
	return out
}
