// +build nrf51

package ws2812

// This file implements the WS2812 protocol for 16MHz Cortex-M0
// microcontrollers.

import (
	"device/arm"
)

// Send a single byte using the WS2812 protocol.
func (p WS2812) WriteByte(c byte) {
	// For the Cortex-M0 at 16MHz
	portSet, maskSet := p.Pin.PortMaskSet()
	portClear, maskClear := p.Pin.PortMaskClear()

	value := uint32(c) << 24
	arm.AsmFull(`
	send_bit:
		str   {maskSet}, {portSet}
		lsls  {value}, #1
		bcs.n skip_store
		str   {maskClear}, {portClear}
	skip_store:
		nop
		nop
		nop
		nop
		nop
		nop
		str   {maskClear}, {portClear}
		subs  {i}, #1
		bne.n send_bit
	`, map[string]interface{}{
		"value":     value,
		"i":         8,
		"maskSet":   maskSet,
		"portSet":   portSet,
		"maskClear": maskClear,
		"portClear": portClear,
	})
}
