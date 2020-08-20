package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/lis2mdl"
)

func main() {
	machine.I2C0.Configure(machine.I2CConfig{})
	compass := lis2mdl.New(machine.I2C0)

	if !compass.Connected() {
		for {
			println("LIS2MDL not connected!")
			time.Sleep(1 * time.Second)
		}
	}

	compass.Configure(lis2mdl.Configuration{}) //default settings

	for {
		heading := compass.ReadCompass()
		println("Heading:", heading)

		time.Sleep(time.Millisecond * 100)
	}
}
