//go:build tinygo

package dht // import "tinygo.org/x/drivers/dht"

import (
	"time"

	"tinygo.org/x/drivers/internal/pin"
)

// Check if the pin is disabled
func powerUp(set pin.OutputFn, get pin.InputFn) bool {
	state := get()
	if !state {
		set.High()
		time.Sleep(startTimeout)
	}
	return state
}

func expectChange(get pin.InputFn, oldState bool) counter {
	cnt := counter(0)
	for ; get() == oldState && cnt != timeout; cnt++ {
	}
	return cnt
}

func checksum(buf []uint8) uint8 {
	return buf[4]
}
func computeChecksum(buf []uint8) uint8 {
	return buf[0] + buf[1] + buf[2] + buf[3]
}

func isValid(buf []uint8) bool {
	return checksum(buf) == computeChecksum(buf)
}
