package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/adxl345"
)

func main() {
	machine.I2C0.Configure(machine.I2CConfig{})
	sensor := adxl345.New(machine.I2C0)
	sensor.Configure()

	println("ADXL345 starts")
	for {
		x, y, z, _ := sensor.ReadAcceleration()
		println("X:", x, "Y:", y, "Z:", z)

		rx, ry, rz := sensor.ReadRawAcceleration()
		println("X (raw):", rx, "Y (raw):", ry, "Z (raw):", rz)

		time.Sleep(100 * time.Millisecond)
	}
}
