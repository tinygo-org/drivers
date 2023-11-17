package main

// Example for the SGP30 to be used on a Raspberry Pi pico.
// Connect the sensor I2C pins to GP26 and GP27 to test.

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/sgp30"
)

func main() {
	time.Sleep(time.Second)
	println("start")

	// Configure the I2C bus.
	bus := machine.I2C1
	err := bus.Configure(machine.I2CConfig{
		SDA:       machine.GP26,
		SCL:       machine.GP27,
		Frequency: 400 * machine.KHz,
	})
	if err != nil {
		println("could not configure I2C:", bus)
		return
	}

	// Configure the sensor.
	sensor := sgp30.New(bus)
	if !sensor.Connected() {
		println("sensor not connected")
		return
	}
	err = sensor.Configure(sgp30.Config{})
	if err != nil {
		println("sensor could not be configured:", err.Error())
		return
	}

	// Measure every second, as recommended by the datasheet.
	for {
		time.Sleep(time.Second)

		err := sensor.Update(0)
		if err != nil {
			println("could not read sensor:", err.Error())
			continue
		}
		println("CO₂ equivalent:", sensor.CO2())
		println("TVOC           ", sensor.TVOC())
	}
}
