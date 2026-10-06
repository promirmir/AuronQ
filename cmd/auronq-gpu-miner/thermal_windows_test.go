//go:build windows

package main

import (
	"testing"
	"time"
)

func TestThermalCooldownPolicy(t *testing.T) {
	c := newThermalController(0, 76, 81, 60)
	cases := []struct {
		temp int
		want time.Duration
	}{
		{70, 0},
		{75, 0},
		{76, 80 * time.Millisecond},
		{77, 160 * time.Millisecond},
		{78, 300 * time.Millisecond},
		{79, 500 * time.Millisecond},
		{80, 900 * time.Millisecond},
	}
	for _, tc := range cases {
		if got := c.cooldownFor(tc.temp); got != tc.want {
			t.Fatalf("temp %d: got %s want %s", tc.temp, got, tc.want)
		}
	}
}
