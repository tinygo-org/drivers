package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/bmi270"
)

func main() {
	machine.I2C0.Configure(machine.I2CConfig{})

	dev := bmi270.NewI2C(machine.I2C0)
	if err := dev.Configure(bmi270.Config{}); err != nil {
		for {
			println("Error configuring device:", err.Error())
			time.Sleep(time.Second)
		}
	}

	for {
		ax, ay, az, err := dev.ReadAcceleration()
		if err != nil {
			println("Error reading acceleration:", err.Error())
		}
		gx, gy, gz, err := dev.ReadRotation()
		if err != nil {
			println("Error reading rotation:", err.Error())
		}
		temp, err := dev.ReadTemperature()
		if err != nil {
			println("Error reading temperature:", err.Error())
		}
		println("Acceleration (µg):", ax, ay, az)
		println("Rotation (µ°/s):", gx, gy, gz)
		println("Temperature (m°C):", temp)
		time.Sleep(500 * time.Millisecond)
	}
}
