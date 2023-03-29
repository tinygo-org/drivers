package main

import (
	"time"

	"machine"

	"tinygo.org/x/drivers/bmp180"
)

func main() {
	machine.I2C0.Configure(machine.I2CConfig{})
	sensor := bmp180.New(machine.I2C0)
	sensor.Configure()

	connected := sensor.Connected()
	if !connected {
		println("BMP180 not detected")
		return
	}
	println("BMP180 detected")

	for {
		temp, _ := sensor.ReadTemperature()
		println("Temperature:", float32(temp)/1000, "°C")

		pressure, _ := sensor.ReadPressure()
		println("Pressure", float32(pressure)/100000, "hPa")

		altitude, _ := sensor.ReadAltitude()
		println("Altitude", altitude, "meters")

		time.Sleep(2 * time.Second)
	}
}
