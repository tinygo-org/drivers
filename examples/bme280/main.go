package main

import (
	"time"

	"machine"

	"github.com/tinygo-org/drivers/bme280"
)

func main() {
	machine.I2C0.Configure(machine.I2CConfig{})
	sensor := bme280.New(machine.I2C0)
	sensor.Configure()

	connected := sensor.Connected()
	if !connected {
		println("BME280 not detected")
		return
	}
	//println("BME280 detected")

	for {
		temp, _ := sensor.ReadTemperature()
		println("Temperature:", temp/10, "ºC")

		time.Sleep(2 * time.Second)
	}
}
