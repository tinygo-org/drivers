// This example uses the BMM150 of an M5Stack StampFly.
package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/bmm150"
)

func main() {
	time.Sleep(2 * time.Second)

	machine.I2C0.Configure(machine.I2CConfig{
		SDA:       machine.GPIO3,
		SCL:       machine.GPIO4,
		Frequency: 400 * machine.KHz,
	})

	sensor := bmm150.New(machine.I2C0)
	if err := sensor.Configure(bmm150.Config{}); err != nil {
		println("BMM150 configuration failed:", err.Error())
		return
	}

	for {
		if err := sensor.Update(drivers.MagneticField); err != nil {
			println("Error reading:", err.Error())
		} else {
			x, y, z := sensor.MagneticField()
			println("Magnetic field (nT):", x, y, z)
		}
		time.Sleep(100 * time.Millisecond)
	}
}
