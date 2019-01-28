package main

import (
	"time"

	"github.com/tinygo-org/drivers/bmp180"
	"github.com/aykevl/tinygo/src/machine"
)

func main() {
	machine.I2C0.Configure(machine.I2CConfig{})
	sensor := bmp180.New(machine.I2C0)
	sensor.Configure()

	connected := sensor.Connected()
	if !connected {
		println("BMP180 not detected")
		return
	}
	println("BMP180 detected")

	for {
		temp, _ := sensor.Temperature()
		println("Temperature:", temp, "ºmC")

		pressure, _ := sensor.Pressure()
		println("Pressure", pressure, "mPa")

		time.Sleep(2 * time.Second)
	}
}
