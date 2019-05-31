// Connects to an MMA8653 I2C accelerometer.
package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/mma8653"
)

func main() {
	machine.I2C0.Configure(machine.I2CConfig{})

	accel := mma8653.New(machine.I2C0)
	accel.Configure(mma8653.DataRate200Hz, mma8653.Sensitivity2G)

	for {
		x, y, z, _ := accel.ReadAcceleration()
		println(x, y, z)
		time.Sleep(time.Millisecond * 100)
	}
}
