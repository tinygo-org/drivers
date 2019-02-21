// +build circuitplay_express itsybitsy_m0

package ws2812

// This file implements the WS2812 protocol for 48MHz Cortex-M0
// microcontrollers.

import (
	"device/arm"
)

// Send a single byte using the WS2812 protocol.
func (d Device) WriteByte(c byte) error {
	// For the Cortex-M0 at 48MHz
	portSet, maskSet := d.Pin.PortMaskSet()
	portClear, maskClear := d.Pin.PortMaskClear()

	// See:
	// https://wp.josh.com/2014/05/13/ws2812-neopixels-are-not-so-finicky-once-you-get-to-know-them/
	// T0H: 10 cycles or 208.3ns
	// T0L: 39 cycles or 812.5ns -> together 49 cycles or 1020.8ns
	// T1H: 27 cycles or 562.5ns
	// T1L: 22 cycles or 458.3ns -> together 49 cycles or 1020.8ns
	value := uint32(c) << 24
	arm.AsmFull(`
	send_bit:
		str   {maskSet}, {portSet}     @ [2]   T0H and T0L start here
		lsls  {value}, #1              @ [1]
		nop                            @ [6]
		nop
		nop
		nop
		nop
		nop
		bcs.n skip_store               @ [1/3]
		str   {maskClear}, {portClear} @ [2]   T0H -> T0L transition
	skip_store:
		nop                            @ [15]
		nop
		nop
		nop
		nop
		nop
		nop
		nop
		nop
		nop
		nop
		nop
		nop
		nop
		nop
		str   {maskClear}, {portClear} @ [2]   T1H -> T1L transition
		nop                            @ [16]
		nop
		nop
		nop
		nop
		nop
		nop
		nop
		nop
		nop
		nop
		nop
		nop
		nop
		nop
		nop
		subs  {i}, #1                  @ [1]
		bne.n send_bit                 @ [1/3]
	`, map[string]interface{}{
		"value":     value,
		"i":         8,
		"maskSet":   maskSet,
		"portSet":   portSet,
		"maskClear": maskClear,
		"portClear": portClear,
	})
	return nil
}
