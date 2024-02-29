//go:build !esp32c3

package espradio

import "errors"

// This is the value used for the ESP32-C3.
const ticksPerSecond = 16_000_000

var errUnsupportedChip = errors.New("espradio: unsupported chip")

func initHardware() error {
	return errUnsupportedChip
}
