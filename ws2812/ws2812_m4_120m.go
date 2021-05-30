// +build atsamd51

package ws2812

// This file implements the WS2812 protocol for 120MHz Cortex-M4
// microcontrollers.
// Note: This implementation does not work with tinygo 0.9.0 or older.

import (
	"device/arm"
	"runtime/interrupt"
)

// Send a single byte using the WS2812 protocol.
func (d Device) WriteByte(c byte) error {
	// For the Cortex-M4 at 120MHz
	portSet, maskSet := d.Pin.PortMaskSet()
	portClear, maskClear := d.Pin.PortMaskClear()
	mask := interrupt.Disable()

	// See:
	// https://wp.josh.com/2014/05/13/ws2812-neopixels-are-not-so-finicky-once-you-get-to-know-them/
	// T0H: 32-34   cycles or  266.67ns -  283.33ns
	// T0L: 101-103 cycles or  841.67ns -  858.33ns
	//   +: 133-137 cycles or 1108.33ns - 1141.67ns
	// T1H: 73-75   cycles or  608.33ns -  625.00ns
	// T1L: 58-60   cycles or  483.33ns -  500.00ns
	//   +: 131-135 cycles or 1091.67ns - 1125.00ns
	// A branch is treated here as 1-3 cycles, because apparently it might get
	// speculated. This is more of a guess than hard fact, because the only docs
	// by ARM that state this are now considered superseded (by what?).
	value := uint32(c) << 24
	arm.AsmFull(`
	1: @ send_bit
		str   {maskSet}, {portSet}     @ [2]   T0H and T0L start here
		lsls  {value}, #1              @ [1]
		nop                            @ [28]
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
		nop
		nop
		bcs.n 2f                       @ [1-3] skip_store
		str   {maskClear}, {portClear} @ [2]   T0H -> T0L transition
	2: @ skip_store
		nop                            @ [41]
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
		nop                            @ [54]
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
		nop
		nop
		nop
		subs  {i}, #1                  @ [1]
		bne.n 1b                       @ [1-3] send_bit
	`, map[string]interface{}{
		"value":     value,
		"i":         8,
		"maskSet":   maskSet,
		"portSet":   portSet,
		"maskClear": maskClear,
		"portClear": portClear,
	})
	interrupt.Restore(mask)
	return nil
}
