package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/makeybutton"
)

var (
	led    machine.Pin = machine.LED
	button machine.Pin = machine.D10
	key    *makeybutton.Button
)

func main() {
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	key = makeybutton.NewButton(button)
	key.Configure()

	for {
		switch key.Get() {
		case makeybutton.Pressed:
			led.High()
		case makeybutton.Released:
			led.Low()
		}
		// the more frequent the more responsive
		time.Sleep(50 * time.Millisecond)
	}
}
