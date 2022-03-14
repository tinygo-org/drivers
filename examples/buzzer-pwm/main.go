package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/buzzer"
)

type note struct {
	tone     float64
	duration float64
}

func main() {
	bzrPin := machine.WIO_BUZZER
	bzrPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	pwm := machine.TCC0
	bzr := buzzer.NewPWM(bzrPin, pwm)

	song := []note{
		{buzzer.C3, buzzer.Quarter},
		{buzzer.D3, buzzer.Quarter},
		{buzzer.E3, buzzer.Quarter},
		{buzzer.F3, buzzer.Quarter},
		{buzzer.G3, buzzer.Quarter},
		{buzzer.A3, buzzer.Quarter},
		{buzzer.B3, buzzer.Quarter},
		{buzzer.C3, buzzer.Quarter},
	}

	for _, val := range song {
		bzr.Tone(val.tone, val.duration)
		time.Sleep(10 * time.Millisecond)
	}
}
