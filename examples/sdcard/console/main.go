package main

import (
	"fmt"
	"machine"
	"time"

	"tinygo.org/x/drivers/sdcard"
)

var (
	spi    machine.SPI
	csPin  machine.Pin
	ledPin machine.Pin
)

func main() {
	waitSerial()
	fmt.Printf("sdcard console\r\n")

	led := ledPin
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})

	sd := sdcard.New(spi, csPin)
	err := sd.Configure()
	if err != nil {
		fmt.Printf("%s\r\n", err.Error())
		for {
			time.Sleep(time.Hour)
		}
	}

	go RunFor(&sd)

	for {
		led.High()
		time.Sleep(200 * time.Millisecond)
		led.Low()
		time.Sleep(200 * time.Millisecond)
	}
}

// Wait for user to open serial console
func waitSerial() {
	for !machine.Serial.DTR() {
		time.Sleep(100 * time.Millisecond)
	}
}
