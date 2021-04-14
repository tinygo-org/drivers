package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/tone"
)

var (
	// Configuration for the Adafruit Circuit Playground Bluefruit.
	pwm = machine.PWM0
	pin = machine.D12
)

func main() {
	speaker, err := tone.New(pwm, pin)
	if err != nil {
		println("failed to configure PWM")
		return
	}

	// Two tone siren.
	for {
		println("nee")
		speaker.SetNote(tone.B5)
		time.Sleep(time.Second / 2)

		println("naw")
		speaker.SetNote(tone.A5)
		time.Sleep(time.Second / 2)
	}
}
