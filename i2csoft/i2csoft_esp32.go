//go:build esp32
// +build esp32

package i2csoft

import (
	"device"
)

func (i2c *I2C) wait() {
	// atsamd51
	// 10 : about 387kHz
	// 56 : about 99kHz

	wait := 10
	if i2c.baudrate < 400*1e3 {
		wait = 56
	}

	for i := 0; i < wait; i++ {
		device.Asm(`nop`)
	}
}
