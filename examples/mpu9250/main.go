package main

import (
	"time"

	"github.com/aykevl/tinygo-drivers/mpu9250"
	"github.com/aykevl/tinygo/src/machine"
)

func main() {
	machine.I2C0.Configure(machine.I2CConfig{})
	sensor := mpu9250.New(machine.I2C0)
	sensor.Configure()

	connected := sensor.Connected()
	if !connected {
		println("MPU9250 not detected")
		return
	}
	println("MPU9250 detected")

	for {
		sensor.AccelUpdate()
		sensor.MagUpdate()
		sensor.GyroUpdate()

		ax, ay, az := sensor.Acceleration()
		println("ACCELERATION", ax, ay, az)

		rx, ry, rz := sensor.Rotation()
		println("ROTATION", rx, ry, rz)

		mx, my, mz := sensor.Magnetometer()
		println("MAGNETOMETER", mx, my, mz)

		hd := sensor.MagHorizDirection()
		println("HORIZONTAL DIRECTION", hd)

		time.Sleep(100 * time.Millisecond)
	}
}
