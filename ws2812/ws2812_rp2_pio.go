//go:build rp2040 || rp2350

package ws2812

import (
	"image/color"
	"machine"

	pio "github.com/tinygo-org/pio/rp2-pio"
	"github.com/tinygo-org/pio/rp2-pio/piolib"
)

// newWS2812Device creates a WS2812 device using PIO for hardware-timed control.
// If PIO initialization fails, it falls back to the bit-bang driver.
func newWS2812Device(pin machine.Pin) Device {
	sm, err := pio.PIO0.ClaimStateMachine()
	if err != nil {
		return Device{Pin: pin, brightness: 255, writeColorFunc: writeColorsRGB}
	}
	ws, err := piolib.NewWS2812B(sm, pin)
	if err != nil {
		return Device{Pin: pin, brightness: 255, writeColorFunc: writeColorsRGB}
	}
	return Device{
		Pin:        pin,
		brightness: 255,
		writeColorFunc: func(_ Device, buf []color.RGBA, brightness uint8) error {
			for _, c := range buf {
				r, g, b := applyBrightness(c, brightness)
				ws.PutRGB(r, g, b)
			}
			return nil
		},
	}
}
