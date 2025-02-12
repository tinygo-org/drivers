package main

import (
	"fmt"
	"machine"
	"time"

	"tinygo.org/x/drivers/max6675"
)

// example for reading temperature from a thermocouple
func main() {
	// Pins are for an Adafruit Feather nRF52840 Express
	machine.D5.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.D5.High()

	machine.SPI0.Configure(machine.SPIConfig{
		Frequency: 1_000_000,
		SCK:       machine.SPI0_SCK_PIN,
		SDI:       machine.SPI0_SDI_PIN,
	})

	thermocouple := max6675.NewDevice(machine.SPI0, machine.D5)

	for {
		temp, err := thermocouple.Read()
		if err != nil {
			println(err)
			return
		}
		fmt.Printf("%0.02f C : %0.02f F\n", temp, (temp*9/5)+32)
		time.Sleep(time.Second)
	}
}
