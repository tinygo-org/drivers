package main

import (
	"machine"

	"tinygo.org/x/drivers/examples/flash/console"
	"tinygo.org/x/drivers/flash"
)

func main() {
	console_example.RunFor(
		flash.NewSPI(
			&machine.SPI1,
			machine.SPI1_MOSI_PIN,
			machine.SPI1_MISO_PIN,
			machine.SPI1_SCK_PIN,
			machine.SPI1_CS_PIN,
		),
	)
}
