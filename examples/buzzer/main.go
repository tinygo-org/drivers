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
	speaker := machine.PA30
	speaker.Configure(machine.PinConfig{Mode: machine.PinOutput})
	speaker.Set(true)

	bzrPin := machine.A0
	bzrPin.Configure(machine.PinConfig{Mode: machine.PinOutput})

	bzr := buzzer.New(machine.A0)

	song := []note{
		{buzzer.C4, buzzer.Quarter},
		{buzzer.D4, buzzer.Quarter},
		{buzzer.E4, buzzer.Quarter},
		{buzzer.F4, buzzer.Quarter},
		{buzzer.G4, buzzer.Quarter},
		{buzzer.A4, buzzer.Quarter},
		{buzzer.B4, buzzer.Quarter},
		{buzzer.C5, buzzer.Quarter},
	}

	for _, val := range song {
		bzr.Tone(val.tone, val.duration)
		time.Sleep(10 * time.Millisecond)
	}
}
