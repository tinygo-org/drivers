//go:build !esp32 && !atsamd51 && !atsame5x && !stm32f4 && !rp2040 && !nrf52840
// +build !esp32,!atsamd51,!atsame5x,!stm32f4,!rp2040,!nrf52840

package i2csoft

import (
	"device"
)

// wait waits for half the time of the SCL operation interval.
func (i2c *I2C) wait() {
	wait := 20
	for i := 0; i < wait; i++ {
		device.Asm(`nop`)
	}
}
