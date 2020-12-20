package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/easystepper"
)

func main() {
	machine.P13.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.P15.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.P14.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.P16.Configure(machine.PinConfig{Mode: machine.PinOutput})

	motor := easystepper.New(machine.P13, machine.P15, machine.P14, machine.P16, 200, 75)
	motor.Configure()

	for {
		println("CLOCKWISE")
		motor.Move(2050)
		time.Sleep(time.Millisecond * 1000)

		println("COUNTERCLOCKWISE")
		motor.Move(-2050)
		time.Sleep(time.Millisecond * 1000)
	}
}
