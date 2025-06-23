// This example shows how to use 128x64 display over I2C
// Tested on Seeeduino XIAO Expansion Board https://wiki.seeedstudio.com/Seeeduino-XIAO-Expansion-Board/
//
// According to manual, I2C address of the display is 0x78, but that's 8-bit address.
// TinyGo operates on 7-bit addresses and respective 7-bit address would be 0x3C, which we use below.
//
// To learn more about different types of I2C addresses, please see following page
// https://www.totalphase.com/support/articles/200349176-7-bit-8-bit-and-10-bit-I2C-Slave-Addressing

package main

import (
	"machine"

	"tinygo.org/x/drivers/examples/ssd1306/common"

	"tinygo.org/x/drivers/ssd1306"
)

func main() {
	machine.I2C0.Configure(machine.I2CConfig{
		Frequency: 400 * machine.KHz,
	})

	display := ssd1306.NewI2C(machine.I2C0)
	display.Configure(ssd1306.Config{
		Address: 0x3C,
		Width:   128,
		Height:  64,
	})

	common.Loop(display)
}
