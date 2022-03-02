package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/easystepper"
)

func main() {
	config := easystepper.DeviceConfig{
		Pin1: machine.P13, Pin2: machine.P15, Pin3: machine.P14, Pin4: machine.P16,
		StepCount: 200, RPM: 75, Mode: easystepper.ModeFour,
	}
	motor, _ := easystepper.New(config)
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
