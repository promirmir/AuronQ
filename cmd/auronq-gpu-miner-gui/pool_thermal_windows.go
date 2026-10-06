//go:build windows

package main

import (
	"fmt"
	"os/exec"
	"syscall"
	"time"
)

const processSuspendResume = 0x0800

var (
	kernel32PoolThermal       = syscall.NewLazyDLL("kernel32.dll")
	procOpenProcessPool       = kernel32PoolThermal.NewProc("OpenProcess")
	procCloseHandlePool       = kernel32PoolThermal.NewProc("CloseHandle")
	ntdllPoolThermal          = syscall.NewLazyDLL("ntdll.dll")
	procNtSuspendProcessPool  = ntdllPoolThermal.NewProc("NtSuspendProcess")
	procNtResumeProcessPool   = ntdllPoolThermal.NewProc("NtResumeProcess")
)

// externalThermalPause returns a short duty-cycle pause for an external pool
// miner. The configured limit remains the emergency ceiling; normal operation
// targets roughly five degrees below it.
func externalThermalPause(temp, target, limit int) time.Duration {
	if temp < 0 || limit <= 0 || target >= limit || temp < target {
		return 0
	}
	switch {
	case temp >= limit:
		return 0
	case temp >= limit-1:
		return 850 * time.Millisecond
	case temp >= target+3:
		return 600 * time.Millisecond
	case temp >= target+2:
		return 400 * time.Millisecond
	case temp >= target+1:
		return 220 * time.Millisecond
	default:
		return 100 * time.Millisecond
	}
}

func hottestSelectedGPU(gpus []gpuInfo, devices []int) (gpuInfo, bool) {
	selected := make(map[int]bool, len(devices))
	for _, d := range devices {
		selected[d] = true
	}
	var hottest gpuInfo
	found := false
	for _, g := range gpus {
		if !selected[g.Index] || g.TemperatureC < 0 {
			continue
		}
		if !found || g.TemperatureC > hottest.TemperatureC {
			hottest = g
			found = true
		}
	}
	return hottest, found
}

// monitorPoolThermals adds an AuronQ-side thermal governor around an external
// GPU pool miner. It does not modify MeshMiner internals: it briefly suspends
// the process to reduce GPU duty cycle when temperature rises, then lets it run
// continuously again after cooling.
func (a *App) monitorPoolThermals(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	lastPause := time.Duration(-1)
	lastTemp := -1
	for range ticker.C {
		a.mu.RLock()
		active := a.minerCmd == cmd && a.miner.Running && a.miner.Mode == "pool"
		stopping := a.minerStopRequested
		cfg := a.cfg
		a.mu.RUnlock()
		if !active || stopping {
			return
		}
		if cfg.PoolBackend == "cpu" || cfg.ThermalStopC <= 0 {
			return
		}

		gpus, err := queryNVIDIAGPUs()
		if err != nil {
			continue
		}
		devices, err := selectedDeviceIndices(cfg, gpus)
		if err != nil {
			continue
		}
		hottest, ok := hottestSelectedGPU(gpus, devices)
		if !ok {
			continue
		}

		limit := cfg.ThermalStopC
		target := limit - 5
		if target < 50 {
			target = 50
		}
		temp := hottest.TemperatureC
		lastTemp = temp

		if temp >= limit {
			a.addLog(fmt.Sprintf("POOL THERMAL SAFETY: GPU %d reached %d C (hard limit %d C); stopping miner", hottest.Index, temp, limit))
			a.stopWorker()
			return
		}

		pause := externalThermalPause(temp, target, limit)
		if pause != lastPause {
			if pause > 0 {
				a.addLog(fmt.Sprintf("POOL THERMAL AUTO: GPU %d %d C, target %d C, hard limit %d C; applying %s duty-cycle pause", hottest.Index, temp, target, limit, pause))
			} else if lastPause > 0 {
				a.addLog(fmt.Sprintf("POOL THERMAL AUTO: GPU %d cooled to %d C; full duty cycle restored", hottest.Index, temp))
			}
			lastPause = pause
		}
		if pause <= 0 {
			continue
		}

		if err := pulseSuspendProcess(cmd.Process.Pid, pause); err != nil {
			a.addLog("POOL THERMAL AUTO unavailable: " + err.Error())
			// Do not spam the same failure forever. The existing hard-stop
			// monitor remains active as a final safety layer.
			return
		}
	}
	_ = lastTemp
}

func pulseSuspendProcess(pid int, pause time.Duration) error {
	if pid <= 0 || pause <= 0 {
		return nil
	}
	handle, _, openErr := procOpenProcessPool.Call(processSuspendResume, 0, uintptr(uint32(pid)))
	if handle == 0 {
		return fmt.Errorf("OpenProcess(%d) for suspend/resume failed: %v", pid, openErr)
	}
	defer procCloseHandlePool.Call(handle)

	status, _, _ := procNtSuspendProcessPool.Call(handle)
	if status != 0 {
		return fmt.Errorf("NtSuspendProcess(%d) failed with NTSTATUS 0x%x", pid, status)
	}

	resumed := false
	defer func() {
		if !resumed {
			_, _, _ = procNtResumeProcessPool.Call(handle)
		}
	}()

	time.Sleep(pause)

	status, _, _ = procNtResumeProcessPool.Call(handle)
	if status != 0 {
		return fmt.Errorf("NtResumeProcess(%d) failed with NTSTATUS 0x%x", pid, status)
	}
	resumed = true
	return nil
}
