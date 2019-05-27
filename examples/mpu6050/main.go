// Connects to an MPU6050 I2C accelerometer/gyroscope.
package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/mpu6050"
)

func main() {
	machine.I2C0.Configure(machine.I2CConfig{})

	accel := mpu6050.New(machine.I2C0)
	accel.Configure()

	for {
		x, y, z := accel.ReadAcceleration()
		println(x, y, z)
		time.Sleep(time.Millisecond * 100)
	}
}
