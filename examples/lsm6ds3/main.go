// Connects to an LSM6DS3 I2C a 6 axis Inertial Measurement Unit (IMU)
package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/lsm6ds3"
)

func main() {
	machine.I2C0.Configure(machine.I2CConfig{})

	accel := lsm6ds3.New(machine.I2C0)
	accel.Configure(lsm6ds3.Configuration{})
	if !accel.Connected() {
		println("LSM6DS3 not connected")
		return
	}

	for {
		x, y, z := accel.ReadAcceleration()
		println("Acceleration:", float32(x)/1000000, float32(y)/1000000, float32(z)/1000000)
		x, y, z = accel.ReadRotation()
		println("Gyroscope:", float32(x)/1000000, float32(y)/1000000, float32(z)/1000000)
		x, _ = accel.ReadTemperature()
		println("Degrees C", float32(x)/1000, "\n\n")
		time.Sleep(time.Millisecond * 1000)
	}
}
