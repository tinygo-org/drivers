package main

import (
	"machine"
	"time"

	"github.com/tinygo-org/drivers/examples/pcd8544/setbuffer/data"
	"github.com/tinygo-org/drivers/pcd8544"
)

func main() {
	dcPin := machine.GPIO{machine.P3}
	dcPin.Configure(machine.GPIOConfig{Mode: machine.GPIO_OUTPUT})
	rstPin := machine.GPIO{machine.P4}
	rstPin.Configure(machine.GPIOConfig{Mode: machine.GPIO_OUTPUT})
	scePin := machine.GPIO{machine.P5}
	scePin.Configure(machine.GPIOConfig{Mode: machine.GPIO_OUTPUT})

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
