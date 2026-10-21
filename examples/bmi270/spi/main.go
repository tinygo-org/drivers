// This example uses the pins of the M5Stack StampFly with a Stamp-S3A.
package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/bmi270"
)

const (
	sck   = machine.GPIO44
	sdo   = machine.GPIO14
	sdi   = machine.GPIO43
	csPin = machine.GPIO46

	// The PMW3901 optical flow sensor shares the SPI bus.
	flowCS = machine.GPIO12
)

func main() {
	flowCS.Configure(machine.PinConfig{Mode: machine.PinOutput})
	flowCS.High()

	csPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	csPin.High()
	machine.SPI0.Configure(machine.SPIConfig{
		Frequency: 8_000_000,
		SCK:       sck,
		SDO:       sdo,
		SDI:       sdi,
	})

	dev := bmi270.NewSPI(machine.SPI0, csPin)
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
