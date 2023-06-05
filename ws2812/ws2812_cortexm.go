//go:build cortexm

package ws2812

// This file implements the WS2812 protocol for various Cortex-M
// microcontrollers. It is intended to work with various variants: M0, M0+, M3,
// and M4. Because machine.CPUFrequency() is usually a constant, the value will
// usually be constant-propagated and the switch below will be a direct
// (inlinable function) - thus there is usually no code size penalty over build
// tags per CPU speed.

import (
	"machine"
)

// Send a single byte using the WS2812 protocol.
func (d Device) WriteByte(c byte) error {
	switch machine.CPUFrequency() {
	case 16_000_000: // 16MHz
		d.writeByte16(c)
		return nil
	case 48_000_000: // 48MHz
		d.writeByte48(c)
		return nil
	case 64_000_000: // 64MHz
		d.writeByte64(c)
		return nil
	case 120_000_000: // 120MHz
		d.writeByte120(c)
		return nil
	case 125_000_000: // 125 MHz e.g. rp2040
		d.writeByte125(c)
		return nil
	case 168_000_000: // 168MHz, e.g. stm32f405
		d.writeByte168(c)
		return nil
	default:
		return errUnknownClockSpeed
	}
}
