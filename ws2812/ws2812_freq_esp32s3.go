//go:build esp32s3

package ws2812

import "machine"

// cpuFrequency returns the current CPU frequency in Hz on the ESP32-S3, which
// supports runtime frequency scaling and exposes GetCPUFrequency() instead of
// the constant-style CPUFrequency().
func cpuFrequency() uint32 {
	f, _ := machine.GetCPUFrequency()
	return f
}
