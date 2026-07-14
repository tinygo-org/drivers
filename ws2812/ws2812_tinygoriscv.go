//go:build tinygo.riscv32 && !esp32c3

package ws2812

import "machine"

// Send a single byte using the WS2812 protocol.
func (d Device) WriteByte(c byte) error {
	switch machine.CPUFrequency() {
	case 320_000_000: // 320MHz, e.g. fe310
		d.writeByte320(c)
		return nil
	default:
		return errUnknownClockSpeed
	}
}
