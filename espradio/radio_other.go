//go:build !esp32c3

package espradio

import "errors"

var errUnsupportedChip = errors.New("espradio: unsupported chip")

func initHardware() error {
	return errUnsupportedChip
}
