package si5351

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/internal/regmap"
)

// Device wraps an I2C connection to a SI5351 device.
type Device struct {
	bus     drivers.I2C
	Address uint8

	rw             regmap.Device8I2C
	initialised    bool
	crystalFreq    uint32
	crystalLoad    uint8
	pllaConfigured bool
	pllaFreq       uint32
	pllbConfigured bool
	pllbFreq       uint32
	lastRdivValue  [3]uint8
}

var ErrNotInitialised = errors.New("Si5351 not initialised")
var ErrInvalidParameter = errors.New("Si5351 invalid parameter")

// New creates a new SI5351 connection. The I2C bus must already be configured.
//
// This function only creates the Device object, it does not touch the device.
func New(bus drivers.I2C) Device {
	rw := regmap.Device8I2C{}
	rw.SetBus(bus, AddressDefault, binary.BigEndian)

	return Device{
		bus:         bus,
		rw:          rw,
		Address:     AddressDefault,
		crystalFreq: CRYSTAL_FREQ_25MHZ,
		crystalLoad: CRYSTAL_LOAD_10PF,
	}
}

// Configure sets up the device for communication
// TODO error handling
func (d *Device) Configure() error {
	// // Disable all outputs setting CLKx_DIS high
	d.rw.Write8(OUTPUT_ENABLE_CONTROL, 0xFF)

	// Set the load capacitance for the XTAL
	d.rw.Write8(CRYSTAL_INTERNAL_LOAD_CAPACITANCE, d.crystalLoad)

	// Power down all output drivers
	buf := []byte{CLK0_CONTROL, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80}
	d.bus.Tx(uint16(d.Address), buf, nil)

	// Disable spread spectrum output.
	if err := d.DisableSpreadSpectrum(); err != nil {
		return err
	}

	d.initialised = true

	return nil
}

// Connected returns whether a device at SI5351 address has been found.
func (d *Device) Connected() (bool, error) {
	if err := d.bus.Tx(uint16(d.Address), []byte{}, []byte{0}); err != nil {
		return false, err
	}
	return true, nil
}

// EnableSpreadSpectrum enables spread spectrum modulation to reduce EMI.
func (d *Device) EnableSpreadSpectrum() error {
	data, err := d.rw.Read8(SPREAD_SPECTRUM_PARAMETERS)
	if err != nil {
		return err
	}

	data |= 0x80
	return d.rw.Write8(SPREAD_SPECTRUM_PARAMETERS, data)
}

func (d *Device) DisableSpreadSpectrum() error {
	data, err := d.rw.Read8(SPREAD_SPECTRUM_PARAMETERS)
	if err != nil {
		return err
	}

	data &^= 0x80
	return d.rw.Write8(SPREAD_SPECTRUM_PARAMETERS, data)
}

func (d *Device) OutputEnable(output uint8, enable bool) error {
	if !d.initialised {
		return ErrNotInitialised
	}

	// Read the current value of the OUTPUT_ENABLE_CONTROL register
	regVal, err := d.rw.Read8(OUTPUT_ENABLE_CONTROL)
	if err != nil {
		return err
	}

	// Modify regVal based on clk and enable
	if enable {
		regVal &= ^(1 << output)
	} else {
		regVal |= (1 << output)
	}

	// Write the modified value back to the OUTPUT_ENABLE_CONTROL register
	return d.rw.Write8(OUTPUT_ENABLE_CONTROL, regVal)
}

func (d *Device) EnableOutputs() error {
	if !d.initialised {
		return ErrNotInitialised
	}

	return d.rw.Write8(OUTPUT_ENABLE_CONTROL, 0x00)
}

func (d *Device) DisableOutputs() error {
	if !d.initialised {
		return ErrNotInitialised
	}
	return d.rw.Write8(OUTPUT_ENABLE_CONTROL, 0xFF)
}

// packRegSet packs P1, P2, P3 values into the 8-byte register format
// used by both PLL and Multisynth configuration.
// For multisynth, rDivBits should contain the R divider value shifted left by 4.
// For PLL, rDivBits should be 0.
func packRegSet(p1, p2, p3 uint32, rDivBits uint8) [8]byte {
	var data [8]byte
	data[0] = uint8((p3 & 0xFF00) >> 8)
	data[1] = uint8(p3 & 0xFF)
	data[2] = uint8((p1&0x30000)>>16) | rDivBits
	data[3] = uint8((p1 & 0xFF00) >> 8)
	data[4] = uint8(p1 & 0xFF)
	data[5] = uint8(((p3 & 0xF0000) >> 12) | ((p2 & 0xF0000) >> 16))
	data[6] = uint8((p2 & 0xFF00) >> 8)
	data[7] = uint8(p2 & 0xFF)
	return data
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
func (d *Device) ConfigurePLL(pll uint8, mult uint8, num uint32, denom uint32) error {
	// Basic validation
	switch {
	case !d.initialised:
		return ErrNotInitialised
	// mult = 15..90
	case !((mult > 14) && (mult < 91)):
		return ErrInvalidParameter
	// Avoid divide by zero
	case !(denom > 0):
		return ErrInvalidParameter
	// 20-bit limit
	case !(num <= 0xFFFFF):
		return ErrInvalidParameter
	// 20-bit limit
	case !(denom <= 0xFFFFF):
		return ErrInvalidParameter
	}

	// Calculate PLL register values
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

	// Pack and write registers
	data := packRegSet(p1, p2, p3, 0)
	if err := d.bus.Tx(uint16(baseaddr), data[:], nil); err != nil {
		return err
	}

	// Reset both PLLs
	if err := d.rw.Write8(PLL_RESET, (1<<7)|(1<<5)); err != nil {
		return err
	}

	// Store the frequency settings for use with the Multisynth helper
	fvco := float64(d.crystalFreq) * (float64(mult) + (float64(num) / float64(denom)))
	if pll == PLL_A {
		d.pllaConfigured = true
		d.pllaFreq = uint32(math.Floor(fvco))
	} else {
		d.pllbConfigured = true
		d.pllbFreq = uint32(math.Floor(fvco))
	}
	return nil
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
func (d *Device) ConfigureMultisynth(output uint8, pll uint8, div uint32, num uint32, denom uint32) error {
	// Basic validation
	switch {
	case !d.initialised:
		return ErrNotInitialised
	// Channel range
	case !(output < 3):
		return fmt.Errorf("output channel must be between 0 and 2")
	// Divider integer value
	case !((div > 3) && (div < 2049)):
		return ErrInvalidParameter
	// Avoid divide by zero
	case !(denom > 0):
		return ErrInvalidParameter
	// 20-bit limit
	case !(num <= 0xFFFFF):
		return ErrInvalidParameter
	// 20-bit limit
	case !(denom <= 0xFFFFF):
		return ErrInvalidParameter
		// Make sure the requested PLL has been initialised
	case pll == PLL_A && !d.pllaConfigured:
		return ErrInvalidParameter
	case pll == PLL_B && !d.pllbConfigured:
		return ErrInvalidParameter
	}

	// Calculate register values
	var reg si5351RegSet
	switch {
	case num == 0:
		// Integer mode
		reg.p1 = 128*div - 512
		reg.p2 = 0
		reg.p3 = denom
	case denom == 1:
		// Fractional mode, simplified calculations
		reg.p1 = 128*div + 128*num - 512
		reg.p2 = 128*num - 128
		reg.p3 = 1
	default:
		// Fractional mode
		reg.p1 = uint32(128*float64(div) + math.Floor(128*(float64(num)/float64(denom))) - 512)
		reg.p2 = uint32(128*float64(num) - float64(denom)*math.Floor(128*(float64(num)/float64(denom))))
		reg.p3 = denom
	}

	// Determine if we should use integer mode
	intMode := num == 0

	// Use existing R divider value (0 if not previously set)
	rDiv := d.lastRdivValue[output] >> 4

	return d.setMS(output, reg, intMode, rDiv, pll)
}

func (d *Device) ConfigureRdiv(output uint8, div uint8) error {
	// Channel range
	if !(output < 3) {
		return ErrInvalidParameter
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

	data, err := d.rw.Read8(register)
	if err != nil {
		return err
	}

	d.lastRdivValue[output] = (div & 0x07) << 4
	data = (data & 0x0F) | d.lastRdivValue[output]
	return d.rw.Write8(register, data)
}

// si5351RegSet holds the register values for multisynth configuration
type si5351RegSet struct {
	p1 uint32
	p2 uint32
	p3 uint32
}

var ErrFrequencyOutOfRange = errors.New("Si5351 frequency out of range")
var ErrClockConflict = errors.New("Si5351 clock conflict with existing configuration")

// SetFrequency sets the clock frequency of the specified CLK output.
// Frequency range is 8 kHz to 150 MHz.
//
// freq - Output frequency in Hz
// output - Clock output (0, 1, or 2 for this driver)
// pll - The PLL to use (PLL_A or PLL_B)
func (d *Device) SetFrequency(freq uint64, output uint8, pll uint8) error {
	switch {
	case !d.initialised:
		return ErrNotInitialised
	case output > 2:
		return ErrInvalidParameter
	}

	switch {
	// Lower bounds check
	case freq < CLKOUT_MIN_FREQ:
		freq = CLKOUT_MIN_FREQ
	// Upper bounds check
	case freq > MULTISYNTH_MAX_FREQ:
		freq = MULTISYNTH_MAX_FREQ
	}

	// Select the proper R divider value for low frequencies
	rDiv := d.selectRDiv(&freq)

	// Calculate PLL and multisynth parameters
	var pllFreq uint64
	switch {
	case pll == PLL_A && d.pllaConfigured:
		pllFreq = uint64(d.pllaFreq)
	case pll == PLL_B && d.pllbConfigured:
		pllFreq = uint64(d.pllbFreq)
	default:
		// PLL not configured, calculate optimal PLL frequency
		pllFreq = d.calculatePLLFreq(freq)
	}

	// Calculate multisynth divider parameters
	msReg := d.multisynthCalc(freq, pllFreq)

	// Determine if we should use integer mode
	intMode := msReg.p2 == 0

	// Configure PLL if not already configured or if we need a new frequency
	if (pll == PLL_A && !d.pllaConfigured) || (pll == PLL_B && !d.pllbConfigured) {
		if err := d.setPLL(pllFreq, pll); err != nil {
			return err
		}
	}

	// Set multisynth registers
	if err := d.setMS(output, msReg, intMode, rDiv, pll); err != nil {
		return err
	}

	// Enable output
	return d.OutputEnable(output, true)
}

// selectRDiv selects the appropriate R divider for low frequencies
// and modifies the frequency accordingly
func (d *Device) selectRDiv(freq *uint64) uint8 {
	var rDiv uint8 = 0

	if *freq >= CLKOUT_MIN_FREQ && *freq < CLKOUT_MIN_FREQ*2 {
		rDiv = R_DIV_128
		*freq *= 128
	} else if *freq >= CLKOUT_MIN_FREQ*2 && *freq < CLKOUT_MIN_FREQ*4 {
		rDiv = R_DIV_64
		*freq *= 64
	} else if *freq >= CLKOUT_MIN_FREQ*4 && *freq < CLKOUT_MIN_FREQ*8 {
		rDiv = R_DIV_32
		*freq *= 32
	} else if *freq >= CLKOUT_MIN_FREQ*8 && *freq < CLKOUT_MIN_FREQ*16 {
		rDiv = R_DIV_16
		*freq *= 16
	} else if *freq >= CLKOUT_MIN_FREQ*16 && *freq < CLKOUT_MIN_FREQ*32 {
		rDiv = R_DIV_8
		*freq *= 8
	} else if *freq >= CLKOUT_MIN_FREQ*32 && *freq < CLKOUT_MIN_FREQ*64 {
		rDiv = R_DIV_4
		*freq *= 4
	} else if *freq >= CLKOUT_MIN_FREQ*64 && *freq < CLKOUT_MIN_FREQ*128 {
		rDiv = R_DIV_2
		*freq *= 2
	}

	return rDiv
}

// calculatePLLFreq calculates an optimal PLL frequency for the given output frequency
func (d *Device) calculatePLLFreq(freq uint64) uint64 {
	// Try to find an integer divider that puts PLL in valid range (600-900 MHz)
	// Start with a divider that gives us a PLL freq near 750 MHz (middle of range)
	targetPLL := uint64(750000000)
	divider := targetPLL / freq

	// Ensure divider is in valid range (8-900 for fractional, 4/6/8 for integer)

	switch {
	case divider < 8:
		divider = 8
	case divider > 900:
		divider = 900
	}

	pllFreq := freq * divider

	// Ensure PLL frequency is in valid range
	switch {
	case pllFreq < PLL_VCO_MIN:
		pllFreq = PLL_VCO_MIN
	case pllFreq > PLL_VCO_MAX:
		pllFreq = PLL_VCO_MAX
	}

	return pllFreq
}

// multisynthCalc calculates the multisynth register values
func (d *Device) multisynthCalc(freq, pllFreq uint64) si5351RegSet {
	var reg si5351RegSet

	// Calculate the division ratio
	// divider = pllFreq / freq
	a := uint32(pllFreq / freq)
	remainder := pllFreq % freq

	// Calculate b and c for fractional part
	// We use c = SI5351_PLL_C_MAX (max 20-bit value) for best resolution
	c := uint32(SI5351_PLL_C_MAX)
	b := uint32((uint64(remainder) * uint64(c)) / freq)

	// Calculate P1, P2, P3
	// P1 = 128 * a + floor(128 * b / c) - 512
	// P2 = 128 * b - c * floor(128 * b / c)
	// P3 = c
	floor128bc := uint32((128 * uint64(b)) / uint64(c))

	reg.p1 = 128*a + floor128bc - 512
	reg.p2 = 128*b - c*floor128bc
	reg.p3 = c

	return reg
}

// setPLL configures the PLL with the specified frequency
func (d *Device) setPLL(pllFreq uint64, pll uint8) error {
	// Calculate PLL multiplier from crystal frequency
	// pllFreq = crystalFreq * (a + b/c)
	xtalFreq := uint64(d.crystalFreq)

	a := uint32(pllFreq / xtalFreq)
	remainder := pllFreq % xtalFreq

	// Use max denominator for best resolution
	c := uint32(SI5351_PLL_C_MAX)
	b := uint32((remainder * uint64(c)) / xtalFreq)

	return d.ConfigurePLL(pll, uint8(a), b, c)
}

// setMS sets the multisynth registers for the specified output
func (d *Device) setMS(output uint8, reg si5351RegSet, intMode bool, rDiv uint8, pll uint8) error {
	// Get the appropriate starting point for the registers
	var baseaddr uint8
	switch output {
	case 0:
		baseaddr = MULTISYNTH0_PARAMETERS_1
	case 1:
		baseaddr = MULTISYNTH1_PARAMETERS_1
	case 2:
		baseaddr = MULTISYNTH2_PARAMETERS_1
	default:
		return ErrInvalidParameter
	}

	// Store R divider value
	d.lastRdivValue[output] = (rDiv & 0x07) << 4

	// Pack and write registers
	data := packRegSet(reg.p1, reg.p2, reg.p3, d.lastRdivValue[output])
	if err := d.bus.Tx(uint16(baseaddr), data[:], nil); err != nil {
		return err
	}

	// Configure the clk control register
	clkControlReg := uint8(0x0F) // 8mA drive strength, powered up
	if pll == PLL_B {
		clkControlReg |= (1 << 5) // Use PLLB
	}
	if intMode {
		clkControlReg |= (1 << 6) // Integer mode
	}

	var clkReg uint8
	switch output {
	case 0:
		clkReg = CLK0_CONTROL
	case 1:
		clkReg = CLK1_CONTROL
	case 2:
		clkReg = CLK2_CONTROL
	}

	return d.rw.Write8(clkReg, clkControlReg)
}

// GetFreqStep returns the frequency step size of the radio in Hz.
// This is the smallest frequency increment that can be achieved,
// determined by the PLL frequency and denominator resolution.
// If pll is PLL_A, uses PLLA settings; if PLL_B, uses PLLB settings.
// Returns 0 if the specified PLL is not configured.
func (d *Device) GetFreqStep(pll uint8) uint64 {
	// The frequency step at the output is:
	// step = pllFreq / (SI5351_PLL_C_MAX * multisynth_divider)
	//
	// However, since multisynth divider varies per output, we return
	// the base step from the PLL, which is:
	// step = pllFreq / SI5351_PLL_C_MAX

	var pllFreq uint64

	switch pll {
	case PLL_A:
		if !d.pllaConfigured {
			return 0
		}
		pllFreq = uint64(d.pllaFreq)
	case PLL_B:
		if !d.pllbConfigured {
			return 0
		}
		pllFreq = uint64(d.pllbFreq)
	default:
		return 0
	}

	return pllFreq / SI5351_PLL_C_MAX
}
