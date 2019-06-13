package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/examples/pcd8544/setbuffer/data"
	"tinygo.org/x/drivers/pcd8544"
)

func main() {
	dcPin := machine.P3
	dcPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	rstPin := machine.P4
	rstPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	scePin := machine.P5
	scePin.Configure(machine.PinConfig{Mode: machine.PinOutput})

	machine.SPI0.Configure(machine.SPIConfig{})

	lcd := pcd8544.New(machine.SPI0, dcPin, rstPin, scePin)
	lcd.Configure(pcd8544.Config{})

	i := 0
	for {
		err := lcd.SetBuffer(data.Images[i])
		if err != nil {
			println(err.Error())
		}
		lcd.Display()
		i = (i + 1) % 2

		time.Sleep(800 * time.Millisecond)
	}
}
