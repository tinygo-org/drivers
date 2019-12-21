package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/shifter"
)

func main() {
	buttons := shifter.New(machine.BUTTON_LATCH, machine.BUTTON_CLK, machine.BUTTON_OUT)
	buttons.Configure()

	for {
		pressed := buttons.Read8Input()
		if pressed&machine.BUTTON_LEFT_MASK > 0 {
			println("Button LEFT pressed")
		}
		if pressed&machine.BUTTON_UP_MASK > 0 {
			println("Button UP pressed")
		}
		if pressed&machine.BUTTON_DOWN_MASK > 0 {
			println("Button DOWN pressed")
		}
		if pressed&machine.BUTTON_RIGHT_MASK > 0 {
			println("Button RIGHT pressed")
		}
		if pressed&machine.BUTTON_SELECT_MASK > 0 {
			println("Button SELECT pressed")
		}
		if pressed&machine.BUTTON_START_MASK > 0 {
			println("Button START pressed")
		}
		if pressed&machine.BUTTON_A_MASK > 0 {
			println("Button A pressed")
		}
		if pressed&machine.BUTTON_B_MASK > 0 {
			println("Button B pressed")
		}
		time.Sleep(100 * time.Millisecond)
	}
}
