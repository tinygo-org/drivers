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

var (
	dev   pmw3901.Device
	frame [pmw3901.FrameSize]byte
)

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
		if err := dev.ReadFrame(&frame); err != nil {
			println("Error reading frame:", err.Error())
		} else {
			printFrame(&frame)
		}
		time.Sleep(time.Second)
	}
}

const shades = " .:-=+*#%@"

func printFrame(frame *[pmw3901.FrameSize]byte) {
	for y := 0; y < pmw3901.FrameHeight; y++ {
		for x := 0; x < pmw3901.FrameWidth; x++ {
			p := frame[y*pmw3901.FrameWidth+x]
			c := shades[int(p)*len(shades)/256]
			print(string(c), string(c))
		}
		println()
	}
	println()
}
