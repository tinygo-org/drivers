//go:build esp32c3

package ws2812

import "machine"

// Send a single byte using the WS2812 protocol.
func (d Device) WriteByte(c byte) error {
	switch machine.CPUFrequency() {
	case 160_000_000: // 160MHz
		d.writeByte160(c)
		return nil
	default:
		return errUnknownClockSpeed
	}
}
