package main

import (
	"time"

	"machine"

	"tinygo.org/x/drivers/bh1750"
)

func main() {
	machine.I2C0.Configure(machine.I2CConfig{})
	sensor := bh1750.New(machine.I2C0)
	sensor.Configure()

	for {
		lux := sensor.Illuminance()
		println("Illuminance:", lux, "lx")

		time.Sleep(500 * time.Millisecond)
	}
}
