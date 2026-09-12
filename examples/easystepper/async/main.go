package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/easystepper"
)

const stepsPerTurn = 2048

func main() {
	motor, err := easystepper.New(easystepper.DeviceConfig{
		Pin1: machine.P13, Pin2: machine.P15, Pin3: machine.P14, Pin4: machine.P16,
		StepCount: stepsPerTurn, RPM: 4, Mode: easystepper.ModeFour,
	})
	if err != nil {
		println("motor init failed:", err.Error())
		return
	}
	motor.Configure()

	for {
		// MoveAsync returns at once. Update does the steps, so the loop is
		// free for other work.
		println("one turn, faster every 700ms")
		motor.MoveAsync(stepsPerTurn)

		rpm := uint(4)
		loops := 0
		change := time.Now().Add(700 * time.Millisecond)

		for motor.IsMoving() {
			motor.Update()

			// Your own work goes here. The motor does not stop it.
			loops++

			// SetRPM works while the motor turns. Move cannot do this.
			if rpm < 16 && time.Now().After(change) {
				rpm += 2
				motor.SetRPM(rpm)
				println("rpm:", rpm)
				change = time.Now().Add(700 * time.Millisecond)
			}
		}
		println("loop ran", loops, "times while the motor turned")
		motor.Off()
		time.Sleep(time.Second)

		// Start turns until Stop. Update still does the steps.
		println("continuous for 3s, then stop")
		motor.SetRPM(10)
		motor.Start(false)

		end := time.Now().Add(3 * time.Second)
		for time.Now().Before(end) {
			motor.Update()
		}
		motor.Stop()
		motor.Off()
		time.Sleep(time.Second)
	}
}
