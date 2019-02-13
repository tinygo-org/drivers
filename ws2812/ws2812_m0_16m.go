// +build nrf51

package ws2812

// This file implements the WS2812 protocol for 16MHz Cortex-M0
// microcontrollers.

import (
	"device/arm"
)

// Send a single byte using the WS2812 protocol.
func (d Device) WriteByte(c byte) error {
	// For the Cortex-M0 at 16MHz
	portSet, maskSet := d.Pin.PortMaskSet()
	portClear, maskClear := d.Pin.PortMaskClear()

	// See:
	// https://wp.josh.com/2014/05/13/ws2812-neopixels-are-not-so-finicky-once-you-get-to-know-them/
	// T0H: 4  cycles or 250ns
	// T0L: 14 cycles or 875ns -> together 18 cycles or 1125ns
	// T1H: 10 cycles or 625ns
	// T1H: 8  cycles or 500ns -> together 18 cycles or 1125ns
	value := uint32(c) << 24
	arm.AsmFull(`
	send_bit:
		str   {maskSet}, {portSet}     @ [2]   T0H and T0L start here
		lsls  {value}, #1              @ [1]
		bcs.n skip_store               @ [1/3]
		str   {maskClear}, {portClear} @ [2]   T0H -> T0L transition
	skip_store:
		nop                            @ [4]
		nop
		nop
		nop
		str   {maskClear}, {portClear} @ [2]   T1H -> T1L transition
		nop                            @ [2]
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
