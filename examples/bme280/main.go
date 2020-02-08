package main

import (
	"machine"
	"strconv"
	"time"

	"tinygo.org/x/drivers/bme280"
)

func main() {

	machine.I2C0.Configure(machine.I2CConfig{})
	sensor := bme280.New(machine.I2C0)
	sensor.Configure()

	connected := sensor.Connected()
	if !connected {
		println("BME280 not detected")
	}
	println("BME280 detected")

	for {
		temp, _ := sensor.ReadTemperature()
		println("Temperature:", strconv.FormatFloat(float64(temp)/1000, 'f', 2, 64), "°C")
		press, _ := sensor.ReadPressure()
		println("Pressure:", strconv.FormatFloat(float64(press)/100000, 'f', 2, 64), "hPa")
		hum, _ := sensor.ReadHumidity()
		println("Humidity:", strconv.FormatFloat(float64(hum)/100, 'f', 2, 64), "%")
		alt, _ := sensor.ReadAltitude()
		println("Altitude:", alt, "m")

		time.Sleep(2 * time.Second)
	}
}
