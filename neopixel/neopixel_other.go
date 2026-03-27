//go:build !rp2040 && !rp2350

package neopixel

import (
	"image/color"
	"machine"

	"tinygo.org/x/drivers/ws2812"
)

type otherBackend struct {
	ws ws2812.Device
}

func newBackend() backend {
	return &otherBackend{}
}

func (b *otherBackend) init(pin machine.Pin) error {
	pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	b.ws = ws2812.NewWS2812(pin)
	return nil
}

func (b *otherBackend) writePixels(pixels []color.RGBA, brightness uint8) error {
	adjusted := make([]color.RGBA, len(pixels))
	for i, c := range pixels {
		r, g, bl := applyBrightness(c, brightness)
		adjusted[i] = color.RGBA{R: r, G: g, B: bl}
	}
	return b.ws.WriteColors(adjusted)
}
