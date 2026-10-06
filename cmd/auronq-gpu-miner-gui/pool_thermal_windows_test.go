//go:build windows

package main

import (
	"testing"
	"time"
)

func TestExternalThermalPauseCurve(t *testing.T) {
	target, limit := 76, 81
	cases := []struct {
		temp int
		want time.Duration
	}{
		{74, 0},
		{75, 50 * time.Millisecond},
		{76, 100 * time.Millisecond},
		{77, 150 * time.Millisecond},
		{78, 225 * time.Millisecond},
		{79, 300 * time.Millisecond},
		{80, 350 * time.Millisecond},
		{81, 0}, // emergency cooldown path
	}
	for _, tc := range cases {
		if got := externalThermalPause(tc.temp, target, limit); got != tc.want {
			t.Fatalf("temp %d: got %s, want %s", tc.temp, got, tc.want)
		}
	}
}

func TestRampThermalPause(t *testing.T) {
	if got := rampThermalPause(0, 300*time.Millisecond); got != 50*time.Millisecond {
		t.Fatalf("first throttle ramp: got %s", got)
	}
	if got := rampThermalPause(50*time.Millisecond, 300*time.Millisecond); got != 100*time.Millisecond {
		t.Fatalf("second throttle ramp: got %s", got)
	}
	if got := rampThermalPause(300*time.Millisecond, 0); got != 290*time.Millisecond {
		t.Fatalf("slow release: got %s", got)
	}
	if got := rampThermalPause(50*time.Millisecond, 25*time.Millisecond); got != 40*time.Millisecond {
		t.Fatalf("release step: got %s", got)
	}
}

func TestThermalDutyPercent(t *testing.T) {
	cases := []struct {
		pause time.Duration
		want  int
	}{
		{0, 100},
		{50 * time.Millisecond, 90},
		{100 * time.Millisecond, 80},
		{150 * time.Millisecond, 70},
		{225 * time.Millisecond, 55},
		{300 * time.Millisecond, 40},
		{350 * time.Millisecond, 30},
	}
	for _, tc := range cases {
		if got := thermalDutyPercent(tc.pause); got != tc.want {
			t.Fatalf("pause %s: got %d%%, want %d%%", tc.pause, got, tc.want)
		}
	}
}

func TestThermalBand(t *testing.T) {
	target, limit := 76, 81
	cases := map[int]string{
		74: "full",
		76: "regulating",
		78: "warm",
		79: "hot",
		80: "critical",
		81: "emergency",
	}
	for temp, want := range cases {
		if got := thermalBand(temp, target, limit); got != want {
			t.Fatalf("temp %d: got %q, want %q", temp, got, want)
		}
	}
}


func TestStableThermalDesiredDoesNotReleaseWhileHeating(t *testing.T) {
	target, limit := 70, 75
	current := 300 * time.Millisecond

	if got := stableThermalDesired(current, 71, 70, target, limit, false); got != current {
		t.Fatalf("rising temperature released throttle: got %s want %s", got, current)
	}
	if got := stableThermalDesired(current, 70, 70, target, limit, false); got != current {
		t.Fatalf("flat temperature near target released throttle: got %s want %s", got, current)
	}
	if got := stableThermalDesired(current, 67, 68, target, limit, false); got != 0 {
		t.Fatalf("cooling below target should allow gradual release: got %s", got)
	}
}

func TestStableThermalDesiredPostCooldownHold(t *testing.T) {
	target, limit := 70, 75
	if got := stableThermalDesired(300*time.Millisecond, 67, 67, target, limit, true); got != poolPostCooldownMinPause {
		t.Fatalf("post-cooldown hold: got %s want %s", got, poolPostCooldownMinPause)
	}
}

func TestStableThermalSequenceAvoidsLargeDutyJumps(t *testing.T) {
	target, limit := 70, 75
	temps := []int{68, 69, 70, 71, 72, 73, 74, 73, 72, 71, 70, 69}
	current := time.Duration(0)
	prev := -1
	lastDuty := thermalDutyPercent(current)
	for _, temp := range temps {
		desired := stableThermalDesired(current, temp, prev, target, limit, false)
		current = rampThermalPause(current, desired)
		duty := thermalDutyPercent(current)
		delta := duty - lastDuty
		if delta < 0 {
			delta = -delta
		}
		if delta > 10 {
			t.Fatalf("duty jumped by %d points at %d C (from %d%% to %d%%)", delta, temp, lastDuty, duty)
		}
		lastDuty = duty
		prev = temp
	}
}
