// This example demostrates how to control the "Dotstar" (APA102) LED included
// on the Adafruit Itsy Bitsy M0 board.  It implements a "rainbow effect" based
// on the following example:
// https://github.com/adafruit/Adafruit_Learning_System_Guides/blob/master/CircuitPython_Essentials/CircuitPython_Internal_RGB_LED_rainbow.py
package main

import (
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers/apa102"
)

var (
	apa *apa102.Device

	pwm   = machine.TCC0
	leds  = make([]color.RGBA, 1)
	wheel = &Wheel{Brightness: 0x10}
)

func init() {

	// APA102 on Itsy Bitsy is connected to pins that require a software-based
	// SPI implementation.
	apa = apa102.NewSoftwareSPI(machine.PA00, machine.PA01, 1)

	// Configure the regular on-board LED for PWM fading
	err := pwm.Configure(machine.PWMConfig{})
	if err != nil {
		println("failed to configure PWM")
		return
	}
}

func main() {
	channelLED, err := pwm.Channel(machine.LED)
	if err != nil {
		println("failed to configure LED PWM channel")
		return
	}

	// We'll fade the on-board LED in a goroutine to show/ensure that the APA102
	// works fine with the scheduler enabled.  Comment this out to test this code
	// with the scheduler disabled.
	go func() {
		for i, brightening := uint8(0), false; ; i++ {
			if i == 0 {
				brightening = !brightening
				continue
			}
			var brightness uint32 = uint32(i)
			if !brightening {
				brightness = 256 - brightness
			}
			pwm.Set(channelLED, pwm.Top()*brightness/256)
			time.Sleep(5 * time.Millisecond)
		}
	}()

	// Use the "wheel" function from Adafruit's example to cycle the APA102
	for {
		leds[0] = wheel.Next()
		apa.WriteColors(leds)
		time.Sleep(25 * time.Millisecond)
	}

}

// Wheel is a port of Adafruit's Circuit Python example referenced above.
type Wheel struct {
	Brightness uint8
	pos        uint8
}

// Next increments the internal state of the color and returns the new RGBA
func (w *Wheel) Next() (c color.RGBA) {
	pos := w.pos
	if w.pos < 85 {
		c = color.RGBA{R: 0xFF - pos*3, G: pos * 3, B: 0x0, A: w.Brightness}
	} else if w.pos < 170 {
		pos -= 85
		c = color.RGBA{R: 0x0, G: 0xFF - pos*3, B: pos * 3, A: w.Brightness}
	} else {
		pos -= 170
		c = color.RGBA{R: pos * 3, G: 0x0, B: 0xFF - pos*3, A: w.Brightness}
	}
	w.pos++
	return
}
