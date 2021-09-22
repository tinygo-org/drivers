//go:build !esp32 && !atsamd51 && !atsame5x
// +build !esp32,!atsamd51,!atsame5x

package i2csoft

import (
	"device"
)

func (i2c *I2C) wait() {
	// atsamd21 @ 48MHz
	//  1 : about 360kHz
	// 24 : about 96kHz
	wait := 1
	if i2c.baudrate < 400*1e3 {
		wait = 19
	}

	for i := 0; i < wait; i++ {
		device.Asm(`nop`)
	}
}
