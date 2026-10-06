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
		{75, 0},
		{76, 25 * time.Millisecond},
		{77, 50 * time.Millisecond},
		{78, 100 * time.Millisecond},
		{79, 175 * time.Millisecond},
		{80, 300 * time.Millisecond},
		{81, 0}, // emergency cooldown path
	}
	for _, tc := range cases {
		if got := externalThermalPause(tc.temp, target, limit); got != tc.want {
			t.Fatalf("temp %d: got %s, want %s", tc.temp, got, tc.want)
		}
	}
}

func TestRampThermalPause(t *testing.T) {
	if got := rampThermalPause(0, 300*time.Millisecond); got != 150*time.Millisecond {
		t.Fatalf("fast throttle ramp: got %s", got)
	}
	if got := rampThermalPause(150*time.Millisecond, 300*time.Millisecond); got != 300*time.Millisecond {
		t.Fatalf("second throttle ramp: got %s", got)
	}
	if got := rampThermalPause(300*time.Millisecond, 0); got != 275*time.Millisecond {
		t.Fatalf("slow release: got %s", got)
	}
	if got := rampThermalPause(50*time.Millisecond, 25*time.Millisecond); got != 25*time.Millisecond {
		t.Fatalf("release floor: got %s", got)
	}
}

func TestThermalDutyPercent(t *testing.T) {
	cases := []struct {
		pause time.Duration
		want  int
	}{
		{0, 100},
		{25 * time.Millisecond, 95},
		{50 * time.Millisecond, 90},
		{100 * time.Millisecond, 80},
		{175 * time.Millisecond, 65},
		{300 * time.Millisecond, 40},
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
