// This example is designed to implement the button shifter for a PyBadge.
package main

import (
	"time"

	"tinygo.org/x/drivers/shifter"
)

func main() {
	buttons := shifter.NewButtons()
	buttons.Configure()

	for {
		// Update the pins state, to later be returned by .Get()
		buttons.ReadInput()

		if buttons.Pins[shifter.BUTTON_LEFT].Get() {
			println("Button LEFT pressed")
		}
		if buttons.Pins[shifter.BUTTON_UP].Get() {
			println("Button UP pressed")
		}
		if buttons.Pins[shifter.BUTTON_DOWN].Get() {
			println("Button DOWN pressed")
		}
		if buttons.Pins[shifter.BUTTON_RIGHT].Get() {
			println("Button RIGHT pressed")
		}
		if buttons.Pins[shifter.BUTTON_SELECT].Get() {
			println("Button SELECT pressed")
		}
		if buttons.Pins[shifter.BUTTON_START].Get() {
			println("Button START pressed")
		}
		if buttons.Pins[shifter.BUTTON_A].Get() {
			println("Button A pressed")
		}
		if buttons.Pins[shifter.BUTTON_B].Get() {
			println("Button B pressed")
		}
		time.Sleep(100 * time.Millisecond)
	}
}
