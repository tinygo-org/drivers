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

	println("setting to 0°")
	s.SetMicroseconds(1000)
	time.Sleep(3 * time.Second)

	println("setting to 45°")
	s.SetMicroseconds(1500)
	time.Sleep(3 * time.Second)

	println("setting to 90°")
	s.SetMicroseconds(2000)
	time.Sleep(3 * time.Second)

	for {
		time.Sleep(time.Second)
	}
}
