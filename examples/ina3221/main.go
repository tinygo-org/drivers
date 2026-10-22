package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/ina3221"
)

func main() {
	machine.I2C0.Configure(machine.I2CConfig{})

	dev := ina3221.New(machine.I2C0)
	if err := dev.Configure(ina3221.DefaultConfig); err != nil {
		println("Error configuring device:", err.Error())
	}

	for {
		for ch := ina3221.Channel1; ch <= ina3221.Channel3; ch++ {
			bus, shunt, current, power, err := dev.Measurements(ch)
			if err != nil {
				println("Error reading channel", ch, err.Error())
				continue
			}
			println("Channel", ch)
			println("  Bus Voltage:", bus/1000, "mV")
			println("  Shunt Voltage:", shunt, "uV")
			println("  Current:", current/1000, "mA")
			println("  Power:", power/1000, "mW")
		}
		time.Sleep(time.Second)
	}
}
