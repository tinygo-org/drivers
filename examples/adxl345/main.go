package main

import (
	"machine"
	"time"

	"github.com/tinygo-org/drivers/adxl345"
)

func main() {
	machine.I2C0.Configure(machine.I2CConfig{})
	sensor := adxl345.New(machine.I2C0, adxl345.AddressLow)
	sensor.Configure()

	println("ADXL345 starts")
	for {
		sensor.Update()
		x, y, z := sensor.Acceleration()
		println("X:", x, "Y:", y, "Z:", z)

		rx, ry, rz := sensor.RawXYZ()
		println("X (raw):", rx, "Y (raw):", ry, "Z (raw):", rz)

		time.Sleep(100 * time.Millisecond)
	}
}
