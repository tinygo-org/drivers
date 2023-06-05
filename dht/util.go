//go:build tinygo

package dht // import "tinygo.org/x/drivers/dht"

import (
	"machine"
	"time"
)

// Check if the pin is disabled
func powerUp(p machine.Pin) bool {
	state := p.Get()
	if !state {
		p.High()
		time.Sleep(startTimeout)
	}
	return state
}

func expectChange(p machine.Pin, oldState bool) counter {
	cnt := counter(0)
	for ; p.Get() == oldState && cnt != timeout; cnt++ {
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
