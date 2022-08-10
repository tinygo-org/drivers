package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/bts7960"
)

// Configuration for the Arduino Uno.
// Please change the PWM and pin if you want to try this example on a different
// board.
var (
	pwm = machine.Timer0

	rEn = machine.D2
	lEn = machine.D3

	rPwm = machine.PD5
	lPwm = machine.PD6
)

func main() {
	if err := pwm.Configure(machine.PWMConfig{}); err != nil {
		println(err.Error())
		return
	}

	bts := bts7960.New(lEn, rEn, lPwm, rPwm, pwm)
	err := bts.Configure()
	if err != nil {
		println("cannot configure bts: " + err.Error())
		return
	}

	println("rotating left")
	bts.Left(50)

	println("rotating left")
	bts.Right(50)

	for {
		time.Sleep(time.Second)
	}
}
