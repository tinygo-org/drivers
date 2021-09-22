//go:build atsamd51 || atsame5x
// +build atsamd51 atsame5x

package i2csoft

import (
	"device/arm"
)

func (i2c *I2C) wait() {
	// atsamd51
	//  1 : about 388kHz
	// 17 : about 97kHz

	wait := 1
	if i2c.baudrate < 400*1e3 {
		wait = 17
	}

	for i := 0; i < wait; i++ {
		arm.Asm(`nop`)
	}
}
