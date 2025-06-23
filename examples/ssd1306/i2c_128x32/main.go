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
		Address: ssd1306.Address_128_32,
		Width:   128,
		Height:  32,
	})

	common.Loop(display)
}
