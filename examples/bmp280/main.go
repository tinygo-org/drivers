package main

import (
	"fmt"
	"machine"
	"time"
	"tinygo.org/x/drivers/bmp280"
)

func main() {
	time.Sleep(5 * time.Second)

	machine.I2C0.Configure(machine.I2CConfig{})
	sensor := bmp280.New(machine.I2C0)
	sensor.Configure(bmp280.STANDBY_125MS, bmp280.FILTER_4X, bmp280.SAMPLING_16X, bmp280.SAMPLING_16X, bmp280.MODE_FORCED)

	connected := sensor.Connected()
	if !connected {
		println("\nBMP280 Sensor not detected\n")
		return
	}
	println("\nBMP280 Sensor detected\n")

	println("Calibration:")
	sensor.PrintCali()

	for {
		t, err := sensor.ReadTemperature()
		if err != nil {
			println("Error reading temperature")
		}
		// Temperature in degrees Celsius
		fmt.Printf("Temperature: %.2f °C\n", float32(t)/1000)

		p, err := sensor.ReadPressure()
		if err != nil {
			println("Error reading pressure")
		}
		// Pressure in hectoPascal
		fmt.Printf("Pressure: %.2f hPa\n", float32(p)/100000)

		time.Sleep(5 * time.Second)
	}
}
