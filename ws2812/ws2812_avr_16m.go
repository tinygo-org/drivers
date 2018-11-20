// +build arduino

package ws2812

// This file implements the WS2812 protocol for 16MHz AVR microcontrollers.

import (
	"device/avr"
)

// Send a single byte using the WS2812 protocol.
func (d Device) WriteByte(c byte) {
	// For the AVR at 16MHz
	portSet, maskSet := d.Pin.PortMaskSet()
	portClear, maskClear := d.Pin.PortMaskClear()

	avr.AsmFull(`
	send_bit:
		st    {portSet}, {maskSet}     ; [2]   set output high
		lsl   {value}                  ; [1]   shift off the next bit, store it in C
		brcs  skip_store               ; [1/2] branch if this bit is high (long pulse)
		st    {portClear}, {maskClear} ; [2]   set output low (short pulse)
	skip_store:
		nop                            ; [6]   wait before changing the output again
		nop
		nop
		nop
		nop
		nop
		st    {portClear}, {maskClear} ; [2]   set output low (end of pulse)
		subi  {i}, 1                   ; [1]   subtract one (for the loop)
		brne  send_bit                 ; [1/2] send the next bit, if not at the end of the loop
	`, map[string]interface{}{
		"value":     c,
		"i":         8,
		"maskSet":   maskSet,
		"portSet":   portSet,
		"maskClear": maskClear,
		"portClear": portClear,
	})
}
