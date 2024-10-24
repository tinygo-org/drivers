package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/ina219"
)

func main() {
	machine.I2C0.Configure(machine.I2CConfig{})

	dev := ina219.New(machine.I2C0)
	dev.Configure()

	for {
		busVoltage, shuntVoltage, current, power, err := dev.Measurements()
		if err != nil {
			println("Error reading measurements", err)
		}

		println("Bus Voltage:", busVoltage, "V")
		println("Shunt Voltage:", shuntVoltage/100, "mV")
		println("Current:", current, "mA")
		println("Power:", power, "mW")

		time.Sleep(10 * time.Millisecond)
	}
}
