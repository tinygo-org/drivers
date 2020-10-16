// +build avr

package ws2812

// This file implements the WS2812 protocol for AVR microcontrollers.

import (
	"device/avr"
	"machine"
)

// Send a single byte using the WS2812 protocol.
func (d Device) WriteByte(c byte) error {
	// On AVR, the port is always the same for setting and clearing a register
	// so use only one. This avoids the following error:
	//     error: inline assembly requires more registers than available
	// Probably this is about pointer registers, which are very limited on AVR.
	port, maskSet := d.Pin.PortMaskSet()
	_, maskClear := d.Pin.PortMaskClear()

	switch machine.CPUFrequency() {
	case 16e6: // 16MHz
		// See:
		// https://wp.josh.com/2014/05/13/ws2812-neopixels-are-not-so-finicky-once-you-get-to-know-them/
		// T0H: 4  cycles or 250ns
		// T0L: 14 cycles or 875ns -> together 18 cycles or 1125ns
		// T1H: 9  cycles or 562ns
		// T1L: 8  cycles or 500ns -> together 17 cycles or 1062ns
		avr.AsmFull(`
	send_bit:
		st    {portSet}, {maskSet}     ; [2]   set output high
		lsl   {value}                  ; [1]   shift off the next bit, store it in C
		brcs  skip_store               ; [1/2] branch if this bit is high (long pulse)
		st    {portClear}, {maskClear} ; [2]   set output low (short pulse)
	skip_store:
		nop                            ; [4]   wait before changing the output again
		nop
		nop
		nop
		st    {portClear}, {maskClear} ; [2]   set output low (end of pulse)
		nop                            ; [3]
		nop
		nop
		subi  {i}, 1                   ; [1]   subtract one (for the loop)
		brne  send_bit                 ; [1/2] send the next bit, if not at the end of the loop
	`, map[string]interface{}{
			"value":     c,
			"i":         byte(8),
			"maskSet":   maskSet,
			"portSet":   port,
			"maskClear": maskClear,
			"portClear": port,
		})
		return nil
	default:
		return errUnknownClockSpeed
	}
}
