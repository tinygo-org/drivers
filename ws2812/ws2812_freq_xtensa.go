//go:build xtensa && !esp32s3

package ws2812

import "machine"

// cpuFrequency returns the current CPU frequency in Hz on Xtensa targets that
// expose the constant-style machine.CPUFrequency() (e.g. the original ESP32).
func cpuFrequency() uint32 {
	return machine.CPUFrequency()
}
