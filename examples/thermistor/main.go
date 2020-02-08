// This example uses the settings for the thermistor that is built in to the
// Adafruit Circuit Playground Express.
package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/thermistor"
)

const ADC_PIN = machine.TEMPSENSOR

func main() {
	machine.InitADC()

	sensor := thermistor.New(ADC_PIN)
	sensor.Configure()

	for {
		temp, _ := sensor.ReadTemperature()
		println("Temperature:", temp/1000, "°C")

		time.Sleep(2 * time.Second)
	}
}
