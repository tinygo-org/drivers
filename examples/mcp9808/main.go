package main

import (
	"machine"

	"tinygo.org/x/drivers/mcp9808"
)

func main() {
	println("hello")
	machine.I2C0.Configure(machine.I2CConfig{
		SCL: machine.GP1,
		SDA: machine.GP0,
	})

	sensor := mcp9808.New(machine.I2C0)

	if !sensor.Connected() {
		print("not connected!")
		return
	} else {
		print("connected")
	}
}
