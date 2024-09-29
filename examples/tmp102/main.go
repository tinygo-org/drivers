package main

import (
	"fmt"
	"machine"
	"time"

	"tinygo.org/x/drivers/tmp102"
)

func main() {
	machine.I2C0.Configure(machine.I2CConfig{
		Frequency: 400e3,
	})

	thermo := tmp102.New(machine.I2C0)
	thermo.Configure(tmp102.Config{})

	for {

		temp, _ := thermo.ReadTemperature()

		print(fmt.Sprintf("%.2f°C\r\n", float32(temp)/1000.0))

		time.Sleep(time.Millisecond * 1000)
	}

}
