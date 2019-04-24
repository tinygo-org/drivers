
package main

import (
	"machine"
	"time"

	"github.com/tinygo-org/drivers/buzzer"
)

type note struct {
	tone     float64
	duration float64
}

func main() {
	speaker := machine.GPIO{machine.PA30}
	speaker.Configure(machine.GPIOConfig{Mode: machine.GPIO_OUTPUT})
	speaker.Set(true)

	bzrPin := machine.GPIO{machine.A0}
	bzrPin.Configure(machine.GPIOConfig{Mode: machine.GPIO_OUTPUT})

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

