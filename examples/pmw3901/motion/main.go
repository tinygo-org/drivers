// This example uses the pins of the M5Stack StampFly with a Stamp-S3A.
package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/pmw3901"
)

const (
	sck   = machine.GPIO44
	sdo   = machine.GPIO14
	sdi   = machine.GPIO43
	csPin = machine.GPIO12

	// The BMI270 IMU shares the SPI bus.
	imuCS = machine.GPIO46
)

var dev pmw3901.Device

func main() {
	imuCS.Configure(machine.PinConfig{Mode: machine.PinOutput})
	imuCS.High()

	csPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	csPin.High()
	machine.SPI0.Configure(machine.SPIConfig{
		Frequency: 2_000_000,
		SCK:       sck,
		SDO:       sdo,
		SDI:       sdi,
		Mode:      3,
	})

	dev = pmw3901.New(machine.SPI0, csPin)
	if err := dev.Configure(); err != nil {
		for {
			println("Error configuring device:", err.Error())
			time.Sleep(time.Second)
		}
	}

	for {
		m, err := dev.ReadMotionBurst()
		if err != nil {
			println("Error reading motion:", err.Error())
		} else {
			println("dx:", m.DeltaX, "dy:", m.DeltaY, "squal:", m.Squal, "shutter:", m.Shutter, "valid:", m.Valid())
		}
		time.Sleep(100 * time.Millisecond)
	}
}
