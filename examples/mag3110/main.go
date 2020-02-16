// Connects to an MAG3110 I2C magnetometer.
package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/mag3110"
)

func main() {
	machine.I2C0.Configure(machine.I2CConfig{})

	mag := mag3110.New(machine.I2C0)
	mag.Configure()

	for {
		x, y, z := mag.ReadMagnetic()
		println("Magnetic readings:", x, y, z)

		c, _ := mag.ReadTemperature()
		println("Temperature:", float32(c)/1000, "°C")

		time.Sleep(time.Millisecond * 100)
	}
}
