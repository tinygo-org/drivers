//go:build rp2040
// +build rp2040

package i2csoft

import (
	"device"
)

// wait waits for half the time of the SCL operation interval. It is set to
// about 100 kHz.
func (i2c *I2C) wait() {
	wait := 50
	for i := 0; i < wait; i++ {
		device.Asm(`nop`)
	}
}
