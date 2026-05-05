//go:build !rp2040 && !rp2350

package ws2812

import "machine"

// newWS2812Device creates a WS2812 device using the bit-bang driver.
func newWS2812Device(pin machine.Pin) Device {
	return Device{Pin: pin, brightness: 255, writeColorFunc: writeColorsRGB}
}
