package main

import (
	"machine"
	"strconv"
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/dht20"
)

var (
	i2c = machine.I2C0
)

func main() {
	i2c.Configure(machine.I2CConfig{})
	sensor := dht20.New(i2c)
	sensor.Configure()

	// Trigger the first measurement
	sensor.Update(drivers.AllMeasurements)

	for {
		time.Sleep(1 * time.Second)

		// Update sensor dasta
		sensor.Update(drivers.AllMeasurements)
		temp := sensor.Temperature()
		hum := sensor.Humidity()

		// Note: The sensor values are from the previous measurement (1 second ago)
		println("Temperature:", strconv.FormatFloat(float64(temp), 'f', 2, 64), "°C")
		println("Humidity:", strconv.FormatFloat(float64(hum), 'f', 2, 64), "%")
	}
}
