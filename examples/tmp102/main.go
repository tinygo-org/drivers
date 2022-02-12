package main

import (
	"fmt"
	"machine"
	"time"

	"tinygo.org/x/drivers/tmp102"
)

func main() {
	machine.I2C0.Configure(machine.I2CConfig{
		Frequency: machine.TWI_FREQ_400KHZ,
	})

	thermo := tmp102.New(machine.I2C0)
	thermo.Configure(tmp102.Config{})

	for {

		temp, _ := thermo.ReadTemperature()

		print(fmt.Sprintf("%.2f°C\r\n", temp.Celsius()))

		time.Sleep(time.Millisecond * 1000)
	}

}
