// Native interface abstraction layer for target-specific operations supporting
// GPIO-driven HUB75 RGB LED matrix panels.
package native

import "machine"

// Hub75 extends the TinyGo 'machine' package API to support machine-specific
// operations that are required to efficiently utilize a HUB75 interface found
// on RGB LED matrix panels.
//
// In particular, these methods provide the consumer an ability to:
//   - Set/clear multiple HUB75 data/control Pins on a single GPIO port
//     simultaneosly, which is necessary to implement bit-banged drivers with
//     sufficient performance.
//   - Hold the HUB75 clock/latch control lines high/low for an extended number
//     of CPU cycles to ensure detection by the shift registers or other ICs on
//     the RGB LED matrix panel.
//   - Initialize a timer peripheral used to regularly signal HUB75 row data
//     transmission via driver's interrupt service routine.
//
// Since there is typically only a single HUB75 connector (even when driving
// multiple RGB LED matrix panels), you will not need more than one instance of
// any object implementing this interface. Thus, a singleton class hub75 is
// declared below that is intended to be used by all machines implementing this
// interface. The exported variable HUB75 of type hub75 (also declared below) is
// then used by the driver package to obtain concrete references to each of
// these methods.
type Hub75 interface {

	// SetPins configures the pre-computed GPIO port bitmasks for the HUB75 data
	// and control pins.
	//
	// The 6-element Pin array contains the two sets (upper- and lower-half) of
	// RGB pins, ordered as upper-red, green, blue, lower-red, green, blue.
	//
	// The sole Pin following the RGB Pin array is for CLK. This Pin is only
	// necessary if the CLK pin is on the same GPIO port as the RGB data pins.
	// Otherwise, this can be NoPin.
	//
	// The remaining pins define the row address select pins (in order);
	// e.g., address-A, B, C, D would define a matrix with height 32 px.
	SetPins([6]machine.Pin, machine.Pin, ...machine.Pin)

	// SetRgb sets/clears each of the 6 RGB data pins.
	SetRgb(bool, bool, bool, bool, bool, bool) // R1, G1, B1, R2, G2, B2

	// SetRgbMask sets/clears each of the 6 RGB data pins from the given bitmask.
	SetRgbMask(uint32)

	// ClkRgb sets CLK, sets/clears each of the 6 RGB data pins, then clears CLK.
	// Note that this method is only permitted when RGB data pins and CLK pin are
	// all on the same GPIO port.
	ClkRgb(bool, bool, bool, bool, bool, bool) // R1, G1, B1, R2, G2, B2

	// ClkRgbMask sets CLK, sets/clears each of the 6 RGB data pins from the given
	// bitmask, then clears CLK.
	// Note that this method is only permitted when RGB data pins and CLK pin are
	// all on the same GPIO port.
	ClkRgbMask(uint32)

	// SetRow sets the active pair of data rows with the given index.
	SetRow(int)

	// GetPinGroupAlignment returns true if and only if all given Pins are on the
	// same GPIO port, and returns the minimum size of the group to which all pins
	// belong (8, 16, or 32 if true, otherwise 0). Returns (true, 0) if no Pins
	// are provided.
	GetPinGroupAlignment(...machine.Pin) (bool, uint8)

	// InitTimer is used to initialize a timer service that fires an interrupt at
	// regular frequency, which is used to signal row data transmission with the
	// given interrupt service routine (ISR). The timer does not begin raising
	// interrupts until StartTimer is called.
	InitTimer(func())

	// ResumeTimer resumes the timer service, with given current value, that
	// signals row data transmission for HUB75 by raising interrupts with given
	// periodicity.
	ResumeTimer(uint32, uint32)

	// PauseTimer pauses the timer service that signals row data transmission for
	// HUB75 and returns the current value of the timer.
	PauseTimer() uint32
}

// hub75 is a singleton, for implementing interface type Hub75, and is realized
// by the exported variable HUB75 below.
//
// The actual Hub75 interface elaborations are hardware-dependent and are
// implemented in build-contrained (per target arch) source files.
type hub75 struct {
	maskR1, maskG1, maskB1, maskR2, maskG2, maskB2 uint32
	maskRGB, groupRGB                              uint32
	maskCLK                                        uint32
	maskA, maskB, maskC, maskD, maskE              uint32
	maskAddr, groupAddr                            uint32

	handleRow func()
}

// HUB75 represents the physical HUB75 connector for RGB LED matrices.
var HUB75 = &hub75{}
