//go:build xtensa

package ws2812

import (
	"machine"
)

func (d Device) WriteByte(c byte) error {
	switch machine.CPUFrequency() {
	case 160e6: // 160MHz
		d.writeByte160(c)
		return nil
	case 80e6: // 80MHz
		d.writeByte80(c)
		return nil
	default:
		return errUnknownClockSpeed
	}
}
