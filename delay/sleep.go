package delay

import (
	"machine"
	"time"
)

/*
#include <stdint.h>
#include <stdbool.h>
bool tinygo_drivers_sleep(uint32_t ticks);
*/
import "C"

// Sleep for a very precise short duration by busy-waiting for the given time.
// This is not an efficient way to sleep: it will needlessly burn cycles while
// sleeping. But it is useful for sleeping for a very short duration, for
// example for bit-banged protocols.
//
// Longer durations (longer than a few milliseconds) will be handled by calling
// time.Sleep instead.
//
// This function should be called with a constant duration value, in which case
// the call will typically be fully inlined and only take up around nine
// instructions for the entire loop.
//
//go:inline
func Sleep(duration time.Duration) {
	if time.Duration(uint32(duration)&0xff_ffff) != duration {
		// This is a long duration (more than 16ms) which shouldn't be done by
		// busy-waiting.
		time.Sleep(duration)
		return
	}

	// Calculate the number of cycles we should sleep:
	//   cycles = duration * freq / 1e9
	// Avoiding a 64-bit division:
	//   cycles = duration * (freq/1000_000) / 1000
	//
	// This assumes:
	//   * The CPU frequency is a constant and can trivially be
	//     const-propagated, therefore the divide by 1000_000 is done at compile
	//     time.
	//   * The CPU frequency is a multiple of 1000_000, which is true for most
	//     chips (examples: 16MHz, 48MHz, 120MHz, etc).
	//   * The division by 1000 can be done efficiently (Cortex-M3 and up), or
	//     can be fully const-propagated.
	//   * The CPU frequency is lower than 256MHz. If it is higher, long sleep
	//     times (1-16ms) may not work correctly.
	cycles := uint32(duration) * (machine.CPUFrequency() / 1000_000) / 1000
	slept := C.tinygo_drivers_sleep(cycles)
	if !slept {
		// Fallback for platforms without inline assembly support.
		time.Sleep(duration)
	}
}
