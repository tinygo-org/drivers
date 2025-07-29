// Capacitive touch sensing example.
//
// This capacitive touch sensor works by charging a normal GPIO pin, then slowly
// discharging it through a 1MΩ resistor and seeing how long it takes to go from
// high to low.
//
// Use as follows:
//   - Change touchPin below as needed.
//   - Connect this pin to some metal surface, like a piece of aluminimum foil.
//     Make sure this surface is covered (using paper, Scotch tape, etc).
//   - Also connect this same pin to ground through a 1MΩ resistor.
//
// This sensor is very sensitive to noise on the power source, so you should
// probably try to limit it by running from a battery for example. Especially
// phone chargers can produce a lot of noise.
package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/touch/capacitive"
)

const touchPin = machine.GP16 // Raspberry Pi Pico

func main() {
	time.Sleep(time.Second * 2)
	println("start")

	led := machine.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	led.Low()

	// Configure the array of GPIO pins used for capacitive touch sensing.
	// We're using only one pin.
	array := capacitive.NewArray([]machine.Pin{touchPin})

	// Use a dynamic threshold, meaning the GPIO pin is automatically calibrated
	// and re-calibrated to adjust for varying environments (e.g. changing
	// humidity).
	array.SetDynamicThreshold(100)

	wasTouching := false
	for i := uint32(0); ; i++ {
		// Update the GPIO pin. This must be called very often.
		array.Update()
		touching := array.Touching(0)

		// Indicate whether the pin is touched via the LED.
		led.Set(touching)

		// Print something when the touch state changed.
		if wasTouching != touching {
			wasTouching = touching
			if touching {
				println("  touch!")
			} else {
				println("  release!")
			}
		}

		// Print the current value, as a debugging aid. It's not really meant to
		// be used directly.
		if i%128 == 32 {
			println("touch value:", array.RawValue(0))
		}
	}
}
