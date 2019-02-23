// Connects to an MAG3110 I2C magnetometer.
package main

import (
	"machine"
	"time"

	"github.com/tinygo-org/drivers/mag3110"
)

func main() {
	machine.I2C0.Configure(machine.I2CConfig{})

	mag := mag3110.New(machine.I2C0)
	mag.Configure()

	for {
		x, y, z := mag.ReadMagnetic()
		println(x, y, z)
		time.Sleep(time.Millisecond * 100)
	}
}
