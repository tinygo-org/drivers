package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/shifter"
)

const (
	BUTTON_LEFT = iota
	BUTTON_UP
	BUTTON_DOWN
	BUTTON_RIGHT
	BUTTON_SELECT
	BUTTON_START
	BUTTON_A
	BUTTON_B
)

func main() {
	buttons := shifter.New(shifter.EIGHT_BITS, machine.BUTTON_LATCH, machine.BUTTON_CLK, machine.BUTTON_OUT)
	buttons.Configure()

	for {
		// Update the pins state, to later be returned by .Get()
		buttons.Read8Input()

		if buttons.Pins[BUTTON_LEFT].Get() {
			println("Button LEFT pressed")
		}
		if buttons.Pins[BUTTON_UP].Get() {
			println("Button UP pressed")
		}
		if buttons.Pins[BUTTON_DOWN].Get() {
			println("Button DOWN pressed")
		}
		if buttons.Pins[BUTTON_RIGHT].Get() {
			println("Button RIGHT pressed")
		}
		if buttons.Pins[BUTTON_SELECT].Get() {
			println("Button SELECT pressed")
		}
		if buttons.Pins[BUTTON_START].Get() {
			println("Button START pressed")
		}
		if buttons.Pins[BUTTON_A].Get() {
			println("Button A pressed")
		}
		if buttons.Pins[BUTTON_B].Get() {
			println("Button B pressed")
		}
		time.Sleep(100 * time.Millisecond)
	}
}
