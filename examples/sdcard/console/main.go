package main

import (
	"fmt"
	"machine"
	"time"

	"tinygo.org/x/drivers/sdcard"
)

var (
	spi    *machine.SPI
	sckPin machine.Pin
	sdoPin machine.Pin
	sdiPin machine.Pin
	csPin  machine.Pin
	ledPin machine.Pin
)

func main() {
	fmt.Printf("sdcard console\r\n")

	led := ledPin
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})

	sd := sdcard.New(spi, sckPin, sdoPin, sdiPin, csPin)
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
