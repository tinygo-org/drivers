package main

import (
	"machine"
	"strconv"
	"time"

	"tinygo.org/x/drivers/bme688"
)

func main() {
	machine.I2C0.Configure(machine.I2CConfig{})

	sensor := bme688.New(machine.I2C0)
	// Uncomment if default does not work
	//sensor.Address = bme688.AddressAlt
	sensor.Configure()

	connected := sensor.Connected()
	if !connected {
		println("BME688 not detected. Exiting...")
	}

	println("BME688 detected")

	for {
		m, err := sensor.Read()
		if err != nil {
			println("read error")
			time.Sleep(2 * time.Second)
			continue
		}

		println("Temperature:", strconv.FormatFloat(float64(m.Temperature)/1000, 'f', 2, 64), "°C")
		println("Pressure:   ", strconv.FormatFloat(float64(m.Pressure)/100, 'f', 2, 64), "hPa")
		println("Humidity:   ", strconv.FormatFloat(float64(m.Humidity)/1000, 'f', 2, 64), "%")

		if m.GasValid && m.HeaterStable {
			println("Gas resist: ", strconv.FormatFloat(float64(m.GasResistance)/1000, 'f', 2, 64), "kΩ")
		} else {
			println("Gas resist:  (not ready)")
		}

		println("---")
		time.Sleep(5 * time.Second)
	}
}
