package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/servo"
)

// Configuration for the Arduino Uno.
// Please change the PWM and pin if you want to try this example on a different
// board.
var (
	pwm = machine.Timer1
	pin = machine.D9
)

func main() {
	s, err := servo.New(pwm, pin)
	if err != nil {
		for {
			println("could not configure servo")
			time.Sleep(time.Second)
		}
		return
	}

	for {
		println("setting to 0°")
		s.SetAngle(0)
		time.Sleep(3 * time.Second)

		println("setting to 45°")
		s.SetAngle(45)
		time.Sleep(3 * time.Second)

		println("setting to 90°")
		s.SetAngle(90)
		time.Sleep(3 * time.Second)

		println("setting to 135°")
		s.SetAngle(135)
		time.Sleep(3 * time.Second)

		println("setting to 180°")
		s.SetAngle(180)
		time.Sleep(3 * time.Second)
	}
}
