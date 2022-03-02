package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/easystepper"
)

func main() {
	motor := easystepper.New(machine.P13, machine.P15, machine.P14, machine.P16, 200, 75)
	motor.Configure()

	for {
		println("CLOCKWISE")
		motor.Move(2050) // Legacy API - uses easystepper.FourStepMode
		time.Sleep(time.Millisecond * 1000)

		println("COUNTERCLOCKWISE")
		motor.MoveStepsMode(-2050, easystepper.FourStepMode)
		time.Sleep(time.Millisecond * 1000)
	}
}
