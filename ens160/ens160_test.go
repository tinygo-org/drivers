package ens160

import (
	"testing"
)

func TestCalculateTempRaw(t *testing.T) {
	testCases := []struct {
		name        string
		tempMilliC  int32
		expectedRaw uint16
	}{
		{"25°C", 25000, 19082},
		{"-10.5°C", -10500, 16810},
		{"Min temp", -40000, 14922},
		{"Below min", -50000, 14922},
		{"Max temp", 85000, 22922},
		{"Above max", 90000, 22922},
		{"Zero", 0, 17482},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			raw := calculateTempRaw(tc.tempMilliC)
			if raw != tc.expectedRaw {
				t.Errorf("expected %d, got %d", tc.expectedRaw, raw)
			}
		})
	}
}

func TestCalculateHumRaw(t *testing.T) {
	testCases := []struct {
		name        string
		rhMilliPct  int32
		expectedRaw uint16
	}{
		{"50%", 50000, 25600},
		{"0%", 0, 0},
		{"100%", 100000, 51200},
		{"Below 0%", -10000, 0},
		{"Above 100%", 110000, 51200},
		{"33.3%", 33300, 17050},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			raw := calculateHumRaw(tc.rhMilliPct)
			if raw != tc.expectedRaw {
				t.Errorf("expected %d, got %d", tc.expectedRaw, raw)
			}
		})
	}
}
