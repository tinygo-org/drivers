// Connects to an MPU6886 I2C accelerometer/gyroscope.
package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/mpu6886"
)

func main() {
	machine.I2C0.Configure(machine.I2CConfig{})

	accel := mpu6886.New(machine.I2C0)
	accel.Configure(mpu6886.Config{})

	for {
		x, y, z, _ := accel.ReadAcceleration()
		println(x, y, z)
		time.Sleep(time.Millisecond * 100)
	}
}
