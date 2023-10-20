// Connects to an MPU9150 I2C accelerometer/gyroscope.
package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/mpu9150"
)

func main() {
	machine.I2C0.Configure(machine.I2CConfig{})

	accel := mpu9150.New(machine.I2C0)
	accel.Configure()

	for {
		x, y, z := accel.ReadAcceleration(mpu9150.ACCEL_XOUT_H)
		println(x, y, z)
		time.Sleep(time.Millisecond * 100)
	}
}
