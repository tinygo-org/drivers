// Connects to a LIS3DH I2C accelerometer on the Adafruit Circuit Playground Express.
package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/lis3dh"
)

var i2c = machine.I2C1

func main() {
	i2c.Configure(machine.I2CConfig{SCL: machine.SCL1_PIN, SDA: machine.SDA1_PIN})

	accel := lis3dh.New(i2c)
	accel.Address = lis3dh.Address1 // address on the Circuit Playground Express
	accel.Configure()
	accel.SetRange(lis3dh.RANGE_2_G)

	println(accel.Connected())

	for {
		x, y, z, _ := accel.ReadAcceleration()
		println("X:", x, "Y:", y, "Z:", z)

		rx, ry, rz := accel.ReadRawAcceleration()
		println("X (raw):", rx, "Y (raw):", ry, "Z (raw):", rz)

		time.Sleep(time.Millisecond * 100)
	}
}
