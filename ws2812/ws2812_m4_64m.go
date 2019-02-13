// +build nrf52

package ws2812

// This file implements the WS2812 protocol for 64MHz Cortex-M4
// microcontrollers.

import (
	"device/arm"
)

// Send a single byte using the WS2812 protocol.
func (d Device) WriteByte(c byte) error {
	// For the Cortex-M4 at 64MHz
	portSet, maskSet := d.Pin.PortMaskSet()
	portClear, maskClear := d.Pin.PortMaskClear()

	// See:
	// https://wp.josh.com/2014/05/13/ws2812-neopixels-are-not-so-finicky-once-you-get-to-know-them/
	// T0H: 14-16 cycles or  218.75ns -  250.00ns
	// T0L: 54-56 cycles or  843.75ns -  875.00ns
	//   +: 68-72 cycles or 1062.50ns - 1125.00ns
	// T1H: 36-38 cycles or  562.50ns -  593.75ns
	// T1L: 30-32 cycles or  468.75ns -  500.0ns
	//   +: 66-70 cycles or 1031.25ns - 1093.75ns
	// A branch is treated here as 1-3 cycles, because apparently it might get
	// speculated. This is more of a guess than hard fact, because the only docs
	// by ARM that state this are now considered superseded (by what?).
	value := uint32(c) << 24
	arm.AsmFull(`
	send_bit:
		str   {maskSet}, {portSet}     @ [2]   T0H and T0L start here
		lsls  {value}, #1              @ [1]
		nop                            @ [10]
		nop
		nop
		nop
		nop
		nop
		nop
		nop
		nop
		nop
		bcs.n skip_store               @ [1-3]
		str   {maskClear}, {portClear} @ [2]   T0H -> T0L transition
	skip_store:
		nop                            @ [22]
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
		nop
		nop
		nop
		nop
		nop
		nop
		str   {maskClear}, {portClear} @ [2]   T1H -> T1L transition
		nop                            @ [26]
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
		bne.n send_bit                 @ [1-3]
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
