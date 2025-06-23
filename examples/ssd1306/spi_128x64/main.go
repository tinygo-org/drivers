package main

import (
	"machine"

	"tinygo.org/x/drivers/examples/ssd1306/common"
	"tinygo.org/x/drivers/ssd1306"
)

func main() {
	machine.SPI0.Configure(machine.SPIConfig{
		Frequency: 8000000,
	})
	display := ssd1306.NewSPI(machine.SPI0, machine.P8, machine.P7, machine.P9)
	display.Configure(ssd1306.Config{
		Width:  128,
		Height: 64,
	})

	common.Loop(display)
}
