package drivers

import "testing"

func TestTemperature(t *testing.T) {
	tests := []struct {
		t Temperature
		c float32 // Celsius
		f float32 // Fahrenheit
	}{
		{-40000, -40, -40}, // -40°C
		{0, 0, 32},         // 0°C
		{20000, 20, 68},    // 20°C
		{25000, 25, 77},    // 25°C
	}
	for _, tc := range tests {
		c := tc.t.Celsius()
		f := tc.t.Fahrenheit()
		if c != tc.c {
			t.Errorf("expected value %d to be %f°C, but got %f°C", tc.t, tc.c, c)
		}
		if f != tc.f {
			t.Errorf("expected value %d to be %f°F, but got %f°F", tc.t, tc.f, f)
		}
	}
}
