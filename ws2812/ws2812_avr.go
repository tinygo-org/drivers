//go:build avr

package ws2812

// This file implements the WS2812 protocol for AVR microcontrollers.

import (
	"machine"
	"runtime/interrupt"
	"unsafe"
)

/*
#include <stdint.h>

__attribute__((always_inline))
void ws2812_writeByte16(char c, uint8_t *port, uint8_t maskSet, uint8_t maskClear) {
	// See:
	// https://wp.josh.com/2014/05/13/ws2812-neopixels-are-not-so-finicky-once-you-get-to-know-them/
	// T0H: 4  cycles or 250ns
	// T0L: 14 cycles or 875ns -> together 18 cycles or 1125ns
	// T1H: 9  cycles or 562ns
	// T1L: 8  cycles or 500ns -> together 17 cycles or 1062ns
	char i = 8;
	__asm__ __volatile__(
		"1:\n"
		"\t  st    %[port], %[maskSet]   ; [2]   set output high\n"
		"\t  lsl   %[value]              ; [1]   shift off the next bit, store it in C\n"
		"\t  brcs  2f                    ; [1/2] branch if this bit is high (long pulse)\n"
		"\t  st    %[port], %[maskClear] ; [2]   set output low (short pulse)\n"
		"\t2:\n"
		"\t  nop                         ; [4]   wait before changing the output again\n"
		"\t  nop\n"
		"\t  nop\n"
		"\t  nop\n"
		"\t  st    %[port], %[maskClear] ; [2]   set output low (end of pulse)\n"
		"\t  nop                         ; [3]\n"
		"\t  nop\n"
		"\t  nop\n"
		"\t  subi  %[i], 1               ; [1]   subtract one (for the loop)\n"
		"\t  brne  1b                    ; [1/2] send the next bit, if not at the end of the loop\n"
	: [value]"+r"(c),
	  [i]"+r"(i)
	: [maskSet]"r"(maskSet),
	  [maskClear]"r"(maskClear),
	  [port]"m"(*port));
}
*/
import "C"

// Send a single byte using the WS2812 protocol.
func (d Device) WriteByte(c byte) error {
	// On AVR, the port is always the same for setting and clearing a register
	// so use only one. This avoids the following error:
	//     error: inline assembly requires more registers than available
	// Probably this is about pointer registers, which are very limited on AVR.
	port, maskSet := d.Pin.PortMaskSet()
	_, maskClear := d.Pin.PortMaskClear()
	mask := interrupt.Disable()

	switch machine.CPUFrequency() {
	case 16e6: // 16MHz
		C.ws2812_writeByte16(C.char(c), (*uint8)(unsafe.Pointer(port)), maskSet, maskClear)
		interrupt.Restore(mask)
		return nil
	default:
		interrupt.Restore(mask)
		return errUnknownClockSpeed
	}
}
