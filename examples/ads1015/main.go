package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/ads1015"
)

var adc = ads1015.New(machine.I2C1)

func main() {
	if err := machine.I2C1.Configure(machine.I2CConfig{
		SDA:       machine.P0_17,
		SCL:       machine.P0_20,
		Frequency: 2.0 * machine.MHz,
	}); err != nil {
		println("could not configure I2C:", err.Error())
		return
	}

	if !adc.Connected() {
		println("ADS1015 not detected")
		return
	}

	config := ads1015.DefaultConfig
	config.Gain = ads1015.Gain4096mV
	if err := adc.Configure(config); err != nil {
		println("could not configure ADS1015:", err.Error())
		return
	}

	for {
		for channel := uint8(0); channel < 4; channel++ {
			raw, err := adc.ReadADC(channel)
			if err != nil {
				println("could not read channel:", err.Error())
				continue
			}
			voltage := adc.ToVoltage(raw)

			print("AIN", channel, ": raw=", raw, " voltage=")
			println(voltage, "mV")
		}

		println()
		time.Sleep(500 * time.Millisecond)
	}
}
