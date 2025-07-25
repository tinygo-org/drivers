// This example demonstrates ENS160 usage.
//
// Wiring:
// - VCC to 3.3V, GND to ground
// - SDA to board SDA, SCL to board SCL

package main

import (
	"time"

	"machine"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/ens160"
)

func main() {
	err := machine.I2C0.Configure(machine.I2CConfig{
		Frequency: 400 * machine.KHz,
	})
	if err != nil {
		println("Failed to configure I2C:", err)
	}

	dev := ens160.New(machine.I2C0, ens160.DefaultAddress)

	connected := dev.Connected()
	if !connected {
		println("ENS160 not detected")
		return
	}
	println("ENS160 detected")

	if err := dev.Configure(); err != nil {
		println("Failed to configure ENS160:", err)
	}

	for {
		err := dev.Update(drivers.Concentration)
		if err != nil {
			println("Error reading ENS160: %v\n", err)
			time.Sleep(5 * time.Second)
			continue
		}

		println(
			"AQI:", dev.AQI(),
			"TVOC:", dev.TVOC(),
			"eCO2:", dev.ECO2(),
			"Validity:", dev.ValidityString(),
		)

		time.Sleep(2 * time.Second)
	}
}
