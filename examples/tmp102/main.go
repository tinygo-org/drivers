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
	thermo.Configure(tmp102.Config{
		Address: 0x48,
		Unit:    tmp102.UNIT_CELSIUS,
	})

	for {
		print(fmt.Sprintf("%.2f°C\r\n", thermo.ReadTemperature()))

		time.Sleep(time.Millisecond * 1000)
	}

}
