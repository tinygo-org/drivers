//go:build rp2040 || rp2350

package neopixel

import (
	"image/color"
	"machine"

	pio "github.com/tinygo-org/pio/rp2-pio"
	"github.com/tinygo-org/pio/rp2-pio/piolib"
)

type rp2Backend struct {
	ws *piolib.WS2812B
}

func newBackend() backend {
	return &rp2Backend{}
}

func (b *rp2Backend) init(pin machine.Pin) error {
	sm, err := pio.PIO0.ClaimStateMachine()
	if err != nil {
		return err
	}
	ws, err := piolib.NewWS2812B(sm, pin)
	if err != nil {
		return err
	}
	b.ws = ws
	return nil
}

func (b *rp2Backend) writePixels(pixels []color.RGBA, brightness uint8) error {
	for _, c := range pixels {
		r, g, bl := applyBrightness(c, brightness)
		b.ws.PutRGB(r, g, bl)
	}
	return nil
}
