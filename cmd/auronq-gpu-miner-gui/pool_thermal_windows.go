//go:build windows

package main

import (
	"fmt"
	"os/exec"
	"syscall"
	"time"
)

const (
	processSuspendResume       = 0x0800
	poolThermalPollInterval    = 500 * time.Millisecond
	poolThermalLogInterval     = 15 * time.Second
	poolThermalBandLogInterval = 5 * time.Second
	poolTelemetryFailLimit     = 6
	poolEmergencyTimeout       = 20 * time.Second
	poolEmergencyStableSamples = 4
)

var (
	kernel32PoolThermal      = syscall.NewLazyDLL("kernel32.dll")
	procOpenProcessPool      = kernel32PoolThermal.NewProc("OpenProcess")
	procCloseHandlePool      = kernel32PoolThermal.NewProc("CloseHandle")
	ntdllPoolThermal         = syscall.NewLazyDLL("ntdll.dll")
	procNtSuspendProcessPool = ntdllPoolThermal.NewProc("NtSuspendProcess")
	procNtResumeProcessPool  = ntdllPoolThermal.NewProc("NtResumeProcess")
)

// externalThermalPause returns the requested pulse length for one 500 ms
// control window. The curve is intentionally progressive: the external pool
// miner stays close to full speed near the target and is only strongly
// throttled immediately below the emergency ceiling.
func externalThermalPause(temp, target, limit int) time.Duration {
	if temp < 0 || limit <= 0 || target >= limit || temp < target {
		return 0
	}
	switch {
	case temp >= limit:
		return 0 // emergency cooldown is handled separately
	case temp >= limit-1:
		return 300 * time.Millisecond
	case temp >= target+3:
		return 175 * time.Millisecond
	case temp >= target+2:
		return 100 * time.Millisecond
	case temp >= target+1:
		return 50 * time.Millisecond
	default:
		return 25 * time.Millisecond
	}
}

// rampThermalPause raises throttling quickly but releases it gradually. This
// hysteresis prevents the old 75 C -> full speed -> 80 C -> long pause loop.
func rampThermalPause(current, desired time.Duration) time.Duration {
	if current < 0 {
		current = 0
	}
	if desired < 0 {
		desired = 0
	}
	if desired > current {
		const rise = 150 * time.Millisecond
		next := current + rise
		if next > desired {
			next = desired
		}
		return next
	}
	if desired < current {
		const fall = 25 * time.Millisecond
		next := current - fall
		if next < desired {
			next = desired
		}
		return next
	}
	return current
}

func thermalDutyPercent(pause time.Duration) int {
	if pause <= 0 {
		return 100
	}
	if pause >= poolThermalPollInterval {
		return 0
	}
	duty := 100 - int((pause*100+poolThermalPollInterval/2)/poolThermalPollInterval)
	if duty < 0 {
		return 0
	}
	if duty > 100 {
		return 100
	}
	return duty
}

func thermalBand(temp, target, limit int) string {
	switch {
	case temp >= limit:
		return "emergency"
	case temp >= limit-1:
		return "critical"
	case temp >= target+3:
		return "hot"
	case temp >= target+2:
		return "warm"
	case temp >= target:
		return "regulating"
	default:
		return "full"
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

// monitorPoolThermals wraps an external pool miner with a fully autonomous
// AuronQ-side governor. It samples faster than the old implementation, applies
// short adaptive duty-cycle pulses, releases throttling slowly after cooling,
// and performs an emergency suspend/cool/resume cycle at the configured hard
// limit. If telemetry or suspend/resume control becomes unavailable, the miner
// is stopped rather than allowed to run blind.
func (a *App) monitorPoolThermals(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}

	ticker := time.NewTicker(poolThermalPollInterval)
	defer ticker.Stop()

	currentPause := time.Duration(0)
	lastBand := ""
	lastLog := time.Time{}
	telemetryFailures := 0

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

		gpus, err := a.hardwareTelemetrySnapshot(1500 * time.Millisecond)
		if err != nil {
			telemetryFailures++
			if telemetryFailures >= poolTelemetryFailLimit {
				a.addLog(fmt.Sprintf("POOL THERMAL FAILSAFE: direct NVML hardware telemetry unavailable for %d consecutive samples (%v); stopping miner", telemetryFailures, err))
				a.stopWorker()
				return
			}
			continue
		}
		telemetryFailures = 0

		devices, err := selectedDeviceIndices(cfg, gpus)
		if err != nil {
			a.addLog("POOL THERMAL FAILSAFE: selected GPU set is no longer available; stopping miner")
			a.stopWorker()
			return
		}
		hottest, ok := hottestSelectedGPU(gpus, devices)
		if !ok {
			telemetryFailures++
			if telemetryFailures >= poolTelemetryFailLimit {
				a.addLog("POOL THERMAL FAILSAFE: no valid temperature for selected GPU(s); stopping miner")
				a.stopWorker()
				return
			}
			continue
		}

		limit := cfg.ThermalStopC
		target := limit - 5
		if target < 50 {
			target = 50
		}
		temp := hottest.TemperatureC

		if temp >= limit {
			a.addLog(fmt.Sprintf("POOL THERMAL EMERGENCY: GPU %d reached %d C (limit %d C); suspending mining for automatic cooldown", hottest.Index, temp, limit))
			if !a.emergencyPoolCooldown(cmd, devices, target, limit) {
				return
			}
			// Resume conservatively and let the hysteresis ramp back toward full
			// duty only if the card remains cool.
			currentPause = 150 * time.Millisecond
			lastBand = "cooldown"
			lastLog = time.Now()
			continue
		}

		desired := externalThermalPause(temp, target, limit)
		currentPause = rampThermalPause(currentPause, desired)
		band := thermalBand(temp, target, limit)
		now := time.Now()
		if lastLog.IsZero() ||
			now.Sub(lastLog) >= poolThermalLogInterval ||
			(band != lastBand && now.Sub(lastLog) >= poolThermalBandLogInterval) {
			a.addLog(fmt.Sprintf("POOL THERMAL AUTO: GPU %d %d C, target %d C, limit %d C; adaptive duty ~%d%%",
				hottest.Index, temp, target, limit, thermalDutyPercent(currentPause)))
			lastBand = band
			lastLog = now
		}

		if currentPause <= 0 {
			continue
		}
		if err := pulseSuspendProcess(cmd.Process.Pid, currentPause); err != nil {
			a.addLog("POOL THERMAL FAILSAFE: adaptive suspend/resume unavailable (" + err.Error() + "); stopping miner")
			a.stopWorker()
			return
		}
	}
}

// emergencyPoolCooldown fully suspends the pool miner at the hard temperature
// limit, waits until the GPU is safely below the target, then resumes the same
// process. This avoids a manual restart and also avoids reconnect/re-tune churn.
// Any abnormal condition falls back to a hard stop.
func (a *App) emergencyPoolCooldown(cmd *exec.Cmd, devices []int, target, limit int) bool {
	if cmd == nil || cmd.Process == nil {
		return false
	}

	handle, err := openPoolProcess(cmd.Process.Pid)
	if err != nil {
		a.addLog("POOL THERMAL FAILSAFE: cannot open miner process for emergency cooldown (" + err.Error() + "); stopping miner")
		a.stopWorker()
		return false
	}
	defer procCloseHandlePool.Call(handle)

	if err := suspendPoolProcess(handle, cmd.Process.Pid); err != nil {
		a.addLog("POOL THERMAL FAILSAFE: cannot suspend miner for emergency cooldown (" + err.Error() + "); stopping miner")
		a.stopWorker()
		return false
	}

	restartAt := target - 2
	if restartAt < 45 {
		restartAt = 45
	}
	deadline := time.Now().Add(poolEmergencyTimeout)
	stable := 0
	telemetryFailures := 0

	for time.Now().Before(deadline) {
		time.Sleep(poolThermalPollInterval)

		a.mu.RLock()
		active := a.minerCmd == cmd && a.miner.Running && a.miner.Mode == "pool"
		stopping := a.minerStopRequested
		a.mu.RUnlock()
		if !active || stopping {
			// User stop/taskkill owns the process from here. Do not resume it.
			return false
		}

		gpus, qerr := a.hardwareTelemetrySnapshot(1500 * time.Millisecond)
		if qerr != nil {
			telemetryFailures++
			if telemetryFailures >= poolTelemetryFailLimit {
				a.addLog("POOL THERMAL FAILSAFE: direct NVML telemetry lost during emergency cooldown; stopping miner")
				a.stopWorker()
				return false
			}
			continue
		}
		telemetryFailures = 0

		hottest, ok := hottestSelectedGPU(gpus, devices)
		if !ok {
			continue
		}
		temp := hottest.TemperatureC

		// A suspended miner should cool. Continued rise means another workload,
		// broken direct hardware telemetry or a cooling-system problem, so stop the miner.
		if temp >= limit+1 {
			a.addLog(fmt.Sprintf("POOL THERMAL FAILSAFE: GPU %d still at %d C while miner is suspended; stopping miner", hottest.Index, temp))
			a.stopWorker()
			return false
		}

		if temp <= restartAt {
			stable++
		} else {
			stable = 0
		}
		if stable >= poolEmergencyStableSamples {
			if err := resumePoolProcess(handle, cmd.Process.Pid); err != nil {
				a.addLog("POOL THERMAL FAILSAFE: cannot resume miner after cooldown (" + err.Error() + "); stopping miner")
				a.stopWorker()
				return false
			}
			a.addLog(fmt.Sprintf("POOL THERMAL AUTO: GPU %d cooled to %d C; mining resumed automatically with conservative duty", hottest.Index, temp))
			return true
		}
	}

	a.addLog(fmt.Sprintf("POOL THERMAL FAILSAFE: GPU did not cool below %d C within %s; stopping miner", restartAt, poolEmergencyTimeout))
	a.stopWorker()
	return false
}

func openPoolProcess(pid int) (uintptr, error) {
	if pid <= 0 {
		return 0, fmt.Errorf("invalid process id %d", pid)
	}
	handle, _, openErr := procOpenProcessPool.Call(processSuspendResume, 0, uintptr(uint32(pid)))
	if handle == 0 {
		return 0, fmt.Errorf("OpenProcess(%d) for suspend/resume failed: %v", pid, openErr)
	}
	return handle, nil
}

func suspendPoolProcess(handle uintptr, pid int) error {
	status, _, _ := procNtSuspendProcessPool.Call(handle)
	if status != 0 {
		return fmt.Errorf("NtSuspendProcess(%d) failed with NTSTATUS 0x%x", pid, status)
	}
	return nil
}

func resumePoolProcess(handle uintptr, pid int) error {
	status, _, _ := procNtResumeProcessPool.Call(handle)
	if status != 0 {
		return fmt.Errorf("NtResumeProcess(%d) failed with NTSTATUS 0x%x", pid, status)
	}
	return nil
}

func pulseSuspendProcess(pid int, pause time.Duration) error {
	if pid <= 0 || pause <= 0 {
		return nil
	}
	handle, err := openPoolProcess(pid)
	if err != nil {
		return err
	}
	defer procCloseHandlePool.Call(handle)

	if err := suspendPoolProcess(handle, pid); err != nil {
		return err
	}

	resumed := false
	defer func() {
		if !resumed {
			_ = resumePoolProcess(handle, pid)
		}
	}()

	time.Sleep(pause)

	if err := resumePoolProcess(handle, pid); err != nil {
		return err
	}
	resumed = true
	return nil
}
