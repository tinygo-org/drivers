package si5351

import (
	"errors"
	"math"

	"tinygo.org/x/drivers"
)

// Device wraps an I2C connection to a SI5351 device.
type Device struct {
	bus     drivers.I2C
	Address uint8

	buf            [8]byte
	initialised    bool
	crystalFreq    uint32
	crystalLoad    uint8
	pllaConfigured bool
	pllaFreq       uint32
	pllbConfigured bool
	pllbFreq       uint32
	lastRdivValue  [3]uint8
}

var errNotInitialised = errors.New("Si5351 not initialised")
var errInvalidParameter = errors.New("Si5351 invalid parameter")

// New creates a new SI5351 connection. The I2C bus must already be configured.
//
// This function only creates the Device object, it does not touch the device.
func New(bus drivers.I2C) Device {
	return Device{
		bus:         bus,
		Address:     AddressDefault,
		crystalFreq: CRYSTAL_FREQ_25MHZ,
		crystalLoad: CRYSTAL_LOAD_10PF,
	}
}

// Configure sets up the device for communication
// TODO error handling
func (d *Device) Configure() {

	data := d.buf[:1]

	// Disable all outputs setting CLKx_DIS high
	data[0] = 0xFF
	d.bus.WriteRegister(d.Address, OUTPUT_ENABLE_CONTROL, data)

	// Set the load capacitance for the XTAL
	data[0] = d.crystalLoad
	d.bus.WriteRegister(d.Address, CRYSTAL_INTERNAL_LOAD_CAPACITANCE, data)

	data = d.buf[:8]

	// Power down all output drivers
	for i := range data {
		data[i] = 0x80
	}
	d.bus.WriteRegister(d.Address, CLK0_CONTROL, data)

	// Disable spread spectrum output.
	d.DisableSpreadSpectrum()

	d.initialised = true

}

// Connected returns whether a device at SI5351 address has been found.
func (d *Device) Connected() bool {
	err := d.bus.Tx(uint16(d.Address), []byte{}, []byte{0})
	return err == nil
}

func (d *Device) EnableSpreadSpectrum() (err error) {
	data := d.buf[:1]
	err = d.bus.ReadRegister(d.Address, SPREAD_SPECTRUM_PARAMETERS, data)
	if err != nil {
		return
	}
	data[0] |= 0x80
	err = d.bus.WriteRegister(d.Address, SPREAD_SPECTRUM_PARAMETERS, data)
	return
}

func (d *Device) DisableSpreadSpectrum() (err error) {
	data := d.buf[:1]
	err = d.bus.ReadRegister(d.Address, SPREAD_SPECTRUM_PARAMETERS, data)
	if err != nil {
		return
	}
	data[0] &^= 0x80
	err = d.bus.WriteRegister(d.Address, SPREAD_SPECTRUM_PARAMETERS, data)
	return
}

func (d *Device) EnableOutputs() (err error) {
	if !d.initialised {
		return errNotInitialised
	}
	data := d.buf[:1]
	data[0] = 0x00
	err = d.bus.WriteRegister(d.Address, OUTPUT_ENABLE_CONTROL, data)
	return
}

func (d *Device) DisableOutputs() (err error) {
	if !d.initialised {
		return errNotInitialised
	}
	data := d.buf[:1]
	data[0] = 0xFF
	err = d.bus.WriteRegister(d.Address, OUTPUT_ENABLE_CONTROL, data)
	return
}

// ConfigurePLL sets the multiplier for the specified PLL
// pll   The PLL to configure, which must be one of the following:
// - PLL_A
// - PLL_B
//
// mult  The PLL integer multiplier (must be between 15 and 90)
//
// num   The 20-bit numerator for fractional output (0..1,048,575).
// Set this to '0' for integer output.
//
// denom The 20-bit denominator for fractional output (1..1,048,575).
// Set this to '1' or higher to avoid divider by zero errors.
//
// PLL Configuration
// fVCO is the PLL output, and must be between 600..900MHz, where:
//
// fVCO = fXTAL * (a+(b/c))
//
// fXTAL = the crystal input frequency
// a     = an integer between 15 and 90
// b     = the fractional numerator (0..1,048,575)
// c     = the fractional denominator (1..1,048,575)
//
// NOTE: Try to use integers whenever possible to avoid clock jitter
// (only use the a part, setting b to '0' and c to '1').
//
// See: http://www.silabs.com/Support%20Documents/TechnicalDocs/AN619.pdf
func (d *Device) ConfigurePLL(pll uint8, mult uint8, num uint32, denom uint32) (err error) {

	// Basic validation
	if !d.initialised {
		return errNotInitialised
	}
	// mult = 15..90
	if !((mult > 14) && (mult < 91)) {
		return errInvalidParameter
	}
	// Avoid divide by zero
	if !(denom > 0) {
		return errInvalidParameter
	}
	// 20-bit limit
	if !(num <= 0xFFFFF) {
		return errInvalidParameter
	}
	// 20-bit limit
	if !(denom <= 0xFFFFF) {
		return errInvalidParameter
	}

	// PLL Multiplier Equations
	//
	// P1 register is an 18-bit value using following formula:
	//
	// 	P1[17:0] = 128 * mult + floor(128*(num/denom)) - 512
	//
	// P2 register is a 20-bit value using the following formula:
	//
	// 	P2[19:0] = 128 * num - denom * floor(128*(num/denom))
	//
	// P3 register is a 20-bit value using the following formula:
	//
	// 	P3[19:0] = denom
	//

	// Set PLL config registers
	var p1, p2, p3 uint32
	if num == 0 {
		// Integer mode
		p1 = 128*uint32(mult) - 512
		p2 = num
		p3 = denom
	} else {
		// Fractional mode
		p1 = uint32(128*float64(mult) + math.Floor(128*(float64(num)/float64(denom))) - 512)
		p2 = uint32(128*float64(num) - float64(denom)*math.Floor(128*(float64(num)/float64(denom))))
		p3 = denom
	}

	// Get the appropriate starting point for the PLL registers
	baseaddr := uint8(26)
	if pll == PLL_B {
		baseaddr = 34
	}

	// The datasheet is a nightmare of typos and inconsistencies here!
	data := d.buf[:8]
	data[0] = uint8((p3 & 0x0000FF00) >> 8)
	data[1] = uint8(p3 & 0x000000FF)
	data[2] = uint8((p1 & 0x00030000) >> 16)
	data[3] = uint8((p1 & 0x0000FF00) >> 8)
	data[4] = uint8(p1 & 0x000000FF)
	data[5] = uint8(((p3 & 0x000F0000) >> 12) | ((p2 & 0x000F0000) >> 16))
	data[6] = uint8((p2 & 0x0000FF00) >> 8)
	data[7] = uint8(p2 & 0x000000FF)
	d.bus.WriteRegister(d.Address, baseaddr, data)

	// Reset both PLLs
	data = d.buf[:1]
	data[0] = (1 << 7) | (1 << 5)
	d.bus.WriteRegister(d.Address, PLL_RESET, data)

	// Store the frequency settings for use with the Multisynth helper
	fvco := float64(d.crystalFreq) * (float64(mult) + (float64(num) / float64(denom)))
	if pll == PLL_A {
		d.pllaConfigured = true
		d.pllaFreq = uint32(math.Floor(fvco))
	} else {
		d.pllbConfigured = true
		d.pllbFreq = uint32(math.Floor(fvco))
	}
	return
}

// ConfigureMultisynth divider, which determines the
// output clock frequency based on the specified PLL input.
//
// output    The output channel to use (0..2)
//
// pll       The PLL input source to use, which must be one of:
//   - PLL_A
//   - PLL_B
//
// div       The integer divider for the Multisynth output.
//
//	If pure integer values are used, this value must be one of:
//	- MULTISYNTH_DIV_4
//	- MULTISYNTH_DIV_6
//	- MULTISYNTH_DIV_8
//	If fractional output is used, this value must be between 8 and 900.
//
// num       The 20-bit numerator for fractional output (0..1,048,575).
//
//	Set this to '0' for integer output.
//
// denom     The 20-bit denominator for fractional output (1..1,048,575).
//
//	Set this to '1' or higher to avoid divide by zero errors.
//
// # Output Clock Configuration
//
// The multisynth dividers are applied to the specified PLL output,
// and are used to reduce the PLL output to a valid range (500kHz
// to 160MHz). The relationship can be seen in this formula, where
// fVCO is the PLL output frequency and MSx is the multisynth divider:
//
// fOUT = fVCO / MSx
//
// Valid multisynth dividers are 4, 6, or 8 when using integers,
// or any fractional values between 8 + 1/1,048,575 and 900 + 0/1
// The following formula is used for the fractional mode divider:
//
// a + b / c
//
// a = The integer value, which must be 4, 6 or 8 in integer mode (MSx_INT=1) or 8..900 in fractional mode (MSx_INT=0).
// b = The fractional numerator (0..1,048,575)
// c = The fractional denominator (1..1,048,575)
//
// NOTE: Try to use integers whenever possible to avoid clock jitter
// NOTE: For output frequencies > 150MHz, you must set the divider
//
//	to 4 and adjust to PLL to generate the frequency (for example
//	a PLL of 640 to generate a 160MHz output clock). This is not
//	yet supported in the driver, which limits frequencies to 500kHz .. 150MHz.
//
// NOTE: For frequencies below 500kHz (down to 8kHz) Rx_DIV must be
//
//	used, but this isn't currently implemented in the driver.
func (d *Device) ConfigureMultisynth(output uint8, pll uint8, div uint32, num uint32, denom uint32) (err error) {

	// Basic validation
	if !d.initialised {
		return errNotInitialised
	}
	// Channel range
	if !(output < 3) {
		return errInvalidParameter
	}
	// Divider integer value
	if !((div > 3) && (div < 2049)) {
		return errInvalidParameter
	}
	// Avoid divide by zero
	if !(denom > 0) {
		return errInvalidParameter
	}
	// 20-bit limit
	if !(num <= 0xFFFFF) {
		return errInvalidParameter
	}
	// 20-bit limit
	if !(denom <= 0xFFFFF) {
		return errInvalidParameter
	}

	// Make sure the requested PLL has been initialised
	if pll == PLL_A && !d.pllaConfigured {
		return errInvalidParameter
	}
	if pll == PLL_B && !d.pllbConfigured {
		return errInvalidParameter
	}

	// Output Multisynth Divider Equations
	//
	// where: a = div, b = num and c = denom
	//
	// P1 register is an 18-bit value using following formula:
	//
	//  P1[17:0] = 128 * a + floor(128*(b/c)) - 512
	//
	// P2 register is a 20-bit value using the following formula:
	//
	//  P2[19:0] = 128 * b - c * floor(128*(b/c))
	//
	// P3 register is a 20-bit value using the following formula:
	//
	//  P3[19:0] = c
	//

	// Set PLL config registers
	var p1, p2, p3 uint32
	if num == 0 {
		// Integer mode
		p1 = 128*div - 512
		p2 = 0
		p3 = denom
	} else if denom == 1 {
		// Fractional mode, simplified calculations
		p1 = 128*div + 128*num - 512
		p2 = 128*num - 128
		p3 = 1
	} else {
		// Fractional mode
		p1 = uint32(128*float64(div) + math.Floor(128*(float64(num)/float64(denom))) - 512)
		p2 = uint32(128*float64(num) - float64(denom)*math.Floor(128*(float64(num)/float64(denom))))
		p3 = denom
	}

	// Get the appropriate starting point for the PLL registers
	baseaddr := uint8(0)
	switch output {
	case 0:
		baseaddr = MULTISYNTH0_PARAMETERS_1
		break
	case 1:
		baseaddr = MULTISYNTH1_PARAMETERS_1
		break
	case 2:
		baseaddr = MULTISYNTH2_PARAMETERS_1
		break
	}

	// Set the MSx config registers
	data := d.buf[:8]
	data[0] = uint8((p3 & 0xFF00) >> 8)
	data[1] = uint8(p3 & 0xFF)
	data[2] = uint8(((p1 & 0x30000) >> 16)) | d.lastRdivValue[output]
	data[3] = uint8((p1 & 0xFF00) >> 8)
	data[4] = uint8(p1 & 0xFF)
	data[5] = uint8(((p3 & 0xF0000) >> 12) | ((p2 & 0xF0000) >> 16))
	data[6] = uint8((p2 & 0xFF00) >> 8)
	data[7] = uint8(p2 & 0xFF)
	err = d.bus.WriteRegister(d.Address, baseaddr, data)
	if err != nil {
		return
	}

	// Configure the clk control and enable the output
	// TODO: Check if the clk control byte needs to be updated.
	clkControlReg := uint8(0x0F) // 8mA drive strength, MS0 as CLK0 source, Clock not inverted, powered up
	if pll == PLL_B {
		clkControlReg |= (1 << 5) // Uses PLLB
	}
	if num == 0 {
		clkControlReg |= (1 << 6) // Integer mode
	}

	var register uint8
	switch output {
	case 0:
		register = CLK0_CONTROL
		break
	case 1:
		register = CLK1_CONTROL
		break
	case 2:
		register = CLK2_CONTROL
		break
	}

	data = d.buf[:1]
	data[0] = clkControlReg
	err = d.bus.WriteRegister(d.Address, register, data)

	return
}

func (d *Device) ConfigureRdiv(output uint8, div uint8) (err error) {
	// Channel range
	if !(output < 3) {
		return errInvalidParameter
	}

	var register uint8
	switch output {
	case 0:
		register = MULTISYNTH0_PARAMETERS_3
	case 1:
		register = MULTISYNTH1_PARAMETERS_3
	case 2:
		register = MULTISYNTH2_PARAMETERS_3
	}

	data := d.buf[:1]
	err = d.bus.ReadRegister(d.Address, register, data)
	if err != nil {
		return
	}

	d.lastRdivValue[output] = (div & 0x07) << 4
	data[0] = (data[0] & 0x0F) | d.lastRdivValue[output]
	err = d.bus.WriteRegister(d.Address, register, data)

	return
}
