//go:build xtensa

package ws2812

import (
	"device"
	"machine"
	"runtime/interrupt"
	"unsafe"
)

func (d Device) WriteByte(c byte) error {
	portSet, maskSet := d.Pin.PortMaskSet()
	portClear, maskClear := d.Pin.PortMaskClear()
	mask := interrupt.Disable()

	switch machine.CPUFrequency() {
	case 160e6: // 160MHz
		// See:
		// https://wp.josh.com/2014/05/13/ws2812-neopixels-are-not-so-finicky-once-you-get-to-know-them/
		// Because I do not know the exact instruction timings, I'm going to
		// assume that every instruction executes in one cycle. Branches and
		// load/stores will probably be slower than that, but as long as all
		// timings are only increased a little bit this should not be a problem
		// (see above post).
		// T0H: 40  cycles or  333.3ns
		// T0L: 131 cycles or 1091.7ns
		//   +: 171 cycles or 1425.0ns
		// T1H: 95  cycles or  791.7ns
		// T1L: 75  cycles or  625.0ns
		//   +: 170 cycles or 1416.7ns
		// Some documentation:
		// http://cholla.mmto.org/esp8266/xtensa.html
		// https://0x04.net/~mwk/doc/xtensa.pdf
		device.AsmFull(`
		1: // send_bit
			s32i  {maskSet}, {portSet}, 0     // [1]  T0H and T1H start here
			nop                               // [37]
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
			slli  {value}, {value}, 1         // [1]  shift {value} to the left by 1
			bbsi  {value}, 8, 2f              // [1]  branch to skip_store if bit 8 is set
			s32i  {maskClear}, {portClear}, 0 // [1]  T0H -> T0L transition
		2: // skip_store
			nop                               // [55]
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
			nop
			s32i  {maskClear}, {portClear}, 0 // [1]  T1H -> T1L transition
			nop                               // [72]
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
			addi  {i}, {i}, -1                // [1]
			bnez {i}, 1b                      // [1]  send_bit, T1H and T1L end here

			// Restore original values after modifying them in the inline
			// assembly. Not doing that would result in undefined behavior as
			// the compiler doesn't know we're modifying these values.
			movi.n {i}, 8
			slli  {value}, {value}, 8
		`, map[string]interface{}{
			// Note: casting pointers to uintptr here because of what might be
			// an Xtensa backend bug with inline assembly.
			"value":     uint32(c),
			"i":         8,
			"maskSet":   maskSet,
			"portSet":   uintptr(unsafe.Pointer(portSet)),
			"maskClear": maskClear,
			"portClear": uintptr(unsafe.Pointer(portClear)),
		})
		interrupt.Restore(mask)
		return nil
	case 80e6: // 80MHz
		// See docs for 160MHz.
		// T0H: 21 cycles or  262.5ns
		// T0L: 67 cycles or  837.5ns
		//   +: 88 cycles or 1100.0ns
		// T1H: 47 cycles or  587.5ns
		// T1L: 39 cycles or  487.5ns
		//   +: 86 cycles or 1075.0ns
		device.AsmFull(`
		1: // send_bit
			s32i  {maskSet}, {portSet}, 0     // [1]  T0H and T1H start here
			nop                               // [18]
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
			slli  {value}, {value}, 1         // [1]  shift {value} to the left by 1
			bbsi  {value}, 8, 2f              // [1]  branch to skip_store if bit 8 is set
			s32i  {maskClear}, {portClear}, 0 // [1]  T0H -> T0L transition
		2: // skip_store
			nop                               // [27]
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
			s32i  {maskClear}, {portClear}, 0 // [1]  T1H -> T1L transition
			nop                               // [36]
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
			addi  {i}, {i}, -1                // [1]
			bnez {i}, 1b                      // [1]  send_bit, T1H and T1L end here

			// Restore original values after modifying them in the inline
			// assembly. Not doing that would result in undefined behavior as
			// the compiler doesn't know we're modifying these values.
			movi.n {i}, 8
			slli  {value}, {value}, 8
		`, map[string]interface{}{
			// Note: casting pointers to uintptr here because of what might be
			// an Xtensa backend bug with inline assembly.
			"value":     uint32(c),
			"i":         8,
			"maskSet":   maskSet,
			"portSet":   uintptr(unsafe.Pointer(portSet)),
			"maskClear": maskClear,
			"portClear": uintptr(unsafe.Pointer(portClear)),
		})
		interrupt.Restore(mask)
		return nil
	default:
		interrupt.Restore(mask)
		return errUnknownClockSpeed
	}
}
