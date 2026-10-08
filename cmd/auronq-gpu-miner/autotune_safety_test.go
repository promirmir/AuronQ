package main

import "testing"

func TestValidateAutotuneTemperature(t *testing.T) {
	tests := []struct {
		name string
		temp int
		limit int
		reject bool
	}{
		{"cool", 60, 81, false},
		{"edge-safe", 77, 81, false},
		{"guard-threshold", 78, 81, true},
		{"limit", 81, 81, true},
		{"too-hot", 90, 81, true},
		{"invalid-negative", -50, 81, true},
		{"missing-sensor", -1, 81, true},
		{"invalid-positive", 140, 81, true},
		{"invalid-limit-low", 50, 50, true},
		{"invalid-limit-high", 50, 100, true},
		{"minimum-limit-safe", 56, 60, false},
		{"minimum-limit-guard", 57, 60, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateAutotuneTemperature(tc.temp, tc.limit)
			if (err != nil) != tc.reject {
				t.Fatalf("temp=%d limit=%d: error=%v, wantReject=%t", tc.temp, tc.limit, err, tc.reject)
			}
		})
	}
}
