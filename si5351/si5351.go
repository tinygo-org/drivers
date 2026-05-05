package si5351

import (
	"encoding/binary"
	"errors"
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/internal/regmap"
)

// Device wraps an I2C connection to a SI5351 device.
type Device struct {
	bus     drivers.I2C
	Address uint8

	rw            regmap.Device8I2C
	initialized   bool
	crystalFreq   [2]CrystalFrequency
	pllaRefOsc    PLLReferenceOscillator
	pllbRefOsc    PLLReferenceOscillator
	clkinDiv      uint8
	pllaFreq      Frequency
	pllbFreq      Frequency
	pllAssignment [8]PLLType
	clkFreq       [8]Frequency
	clkFirstSet   [8]bool
	refCorrection [2]int32
}

var (
	ErrInitTimeout            = errors.New("si5351: init timeout")
	ErrNotInitialized         = errors.New("si5351: not initialized")
	ErrInvalidParameter       = errors.New("si5351: invalid parameter")
	ErrDeviceNotFound         = errors.New("si5351: device not found")
	ErrInvalidPLLClockSetting = errors.New("si5351: cannot set >100MHz with other >100MHz on same PLL")
	ErrInvalidPLLDivision     = errors.New("si5351: CLK6/7 requires integer division ratio")
)

// Frequency in Hz
type Frequency uint64

// CrystalFrequency in Hz
type CrystalFrequency uint32

// CrystalLoad options
type CrystalLoad uint8

const (
	CrystalLoad0PF CrystalLoad = iota
	CrystalLoad6PF
	CrystalLoad8PF
	CrystalLoad10PF
)

// PLL identifiers
type PLLType uint8

const (
	PLL_A PLLType = iota
	PLL_B
)

// Reference oscillator identifiers
type PLLReferenceOscillator uint8

const (
	PLLInputXO PLLReferenceOscillator = iota
	PLLInputClockIn
)

// Clock output identifiers
type Clock uint8

const (
	Clock0 Clock = iota
	Clock1
	Clock2
	Clock3
	Clock4
	Clock5
	Clock6
	Clock7
)

const rfracDenominator = Frequency(PLL_C_MAX)

// RegisterSet holds PLL/multisynth register values
type RegisterSet struct {
	p1 uint32
	p2 uint32
	p3 uint32
}

// New creates a new SI5351 connection. The I2C bus must already be configured.
func New(bus drivers.I2C) *Device {
	rw := regmap.Device8I2C{}
	rw.SetBus(bus, AddressDefault, binary.BigEndian)

	d := Device{
		bus:        bus,
		rw:         rw,
		Address:    AddressDefault,
		pllaRefOsc: PLLInputXO,
		pllbRefOsc: PLLInputXO,
		clkinDiv:   CLKIN_DIV_1,
	}
	d.crystalFreq[0] = XTAL_FREQ

	return &d
}

// Config holds configuration parameters for the SI5351.
type Config struct {
	Capacitance   CrystalLoad
	CrystalOutput CrystalFrequency
	Correction    int32
}

// Configure initializes the SI5351 with the specified crystal load capacitance,
// reference oscillator frequency, and frequency correction.
func (d *Device) Configure(cfg Config) error {
	// Check for device on bus
	if err := d.bus.Tx(uint16(d.Address), []byte{}, []byte{0}); err != nil {
		return ErrDeviceNotFound
	}

	// Wait for SYS_INIT flag to clear
	timeout := time.Now().Add(100 * time.Millisecond)
	for {
		status, err := d.rw.Read8(DEVICE_STATUS)
		if err != nil {
			return err
		}
		if (status >> 7) == 0 {
			break
		}
		if time.Now().After(timeout) {
			return ErrInitTimeout
		}
		time.Sleep(time.Millisecond)
	}

	// Set crystal load capacitance
	var xtalLoadC uint8
	switch cfg.Capacitance {
	case CrystalLoad0PF:
		xtalLoadC = CRYSTAL_LOAD_0PF
	case CrystalLoad6PF:
		xtalLoadC = CRYSTAL_LOAD_6PF
	case CrystalLoad8PF:
		xtalLoadC = CRYSTAL_LOAD_8PF
	case CrystalLoad10PF:
		xtalLoadC = CRYSTAL_LOAD_10PF
	default:
		xtalLoadC = CRYSTAL_LOAD_10PF
	}
	if err := d.rw.Write8(CRYSTAL_LOAD, uint8(xtalLoadC&CRYSTAL_LOAD_MASK)|0x12); err != nil {
		return err
	}

	// Set up the XO reference frequency
	if cfg.CrystalOutput == 0 {
		cfg.CrystalOutput = XTAL_FREQ
	}
	d.SetReferenceFrequency(PLLInputXO, cfg.CrystalOutput)

	// Set frequency calibration for XO
	if err := d.SetCorrection(PLLInputXO, cfg.Correction); err != nil {
		return err
	}

	// Reset device
	if err := d.Reset(); err != nil {
		return err
	}

	d.initialized = true
	return nil
}

// Reset resets the Si5351.
func (d *Device) Reset() error {
	// Power down all outputs
	for i := range uint8(8) {
		if err := d.rw.Write8(CLK0_CTRL+i, 0x80); err != nil {
			return err
		}
	}

	time.Sleep(100 * time.Millisecond)

	// Turn clocks back on with default settings
	for i := range uint8(8) {
		if err := d.rw.Write8(CLK0_CTRL+i, 0x0C); err != nil {
			return err
		}
	}

	time.Sleep(100 * time.Millisecond)

	// Set PLLA and PLLB to 800 MHz
	if err := d.SetPLL(PLL_A, PLL_FIXED); err != nil {
		return err
	}
	if err := d.SetPLL(PLL_B, PLL_FIXED); err != nil {
		return err
	}

	// Make PLL to CLK assignments
	for i := range 6 {
		d.pllAssignment[i] = PLL_A
		d.SetMultisynthSource(Clock(i), PLL_A)
	}
	d.pllAssignment[6] = PLL_B
	d.pllAssignment[7] = PLL_B
	d.SetMultisynthSource(Clock(6), PLL_B)
	d.SetMultisynthSource(Clock(7), PLL_B)

	// Reset VCXO parameters
	d.rw.Write8(VXCO_PARAMETERS_LOW, 0)
	d.rw.Write8(VXCO_PARAMETERS_MID, 0)
	d.rw.Write8(VXCO_PARAMETERS_HIGH, 0)

	// Reset PLLs
	d.PLLReset(PLL_A)
	d.PLLReset(PLL_B)

	// Initialize clock state
	for i := range 8 {
		d.clkFreq[i] = 0
		d.EnableOutput(Clock(i), false)
		d.clkFirstSet[i] = false
	}

	return nil
}

// SetPLL programs the specified PLL with the given frequency.
func (d *Device) SetPLL(pll PLLType, pllFreq Frequency) error {
	var refOsc PLLReferenceOscillator
	var baseAddr uint8

	switch pll {
	case PLL_A:
		refOsc = d.pllaRefOsc
		baseAddr = PLLA_PARAMETERS
		d.pllaFreq = pllFreq
	case PLL_B:
		refOsc = d.pllbRefOsc
		baseAddr = PLLB_PARAMETERS
		d.pllbFreq = pllFreq
	default:
		return ErrInvalidParameter
	}

	_, reg := d.CalculatePLL(pll, pllFreq, d.refCorrection[refOsc], false)

	params := make([]byte, 8)
	params[0] = byte((reg.p3 >> 8) & 0xFF)
	params[1] = byte(reg.p3 & 0xFF)
	params[2] = byte((reg.p1 >> 16) & 0x03)
	params[3] = byte((reg.p1 >> 8) & 0xFF)
	params[4] = byte(reg.p1 & 0xFF)
	params[5] = byte(((reg.p3 >> 12) & 0xF0) | ((reg.p2 >> 16) & 0x0F))
	params[6] = byte((reg.p2 >> 8) & 0xFF)
	params[7] = byte(reg.p2 & 0xFF)

	for i := range params {
		if err := d.rw.Write8(baseAddr+uint8(i), params[i]); err != nil {
			return err
		}
	}

	return nil
}

// SetFrequency sets the clock frequency of the specified CLK output.
// Frequency range is 8 kHz to 150 MHz for CLK0-5, up to 150 MHz for CLK6-7.
func (d *Device) SetFrequency(clk Clock, freq Frequency) error {
	if !d.initialized {
		return ErrNotInitialized
	}

	freqMult := freq * FREQ_MULT

	switch {
	case clk <= 5:
		return d.setFreqCLK0to5(clk, freqMult)
	case clk <= 7:
		return d.setFreqCLK6to7(clk, freqMult)
	default:
		return ErrInvalidParameter
	}
}

// SetRawFrequency sets the clock frequency of the specified CLK output without
// applying the frequency multiplier.
// Frequency range is 8 kHz to 150 MHz for CLK0-5, up to 150 MHz for CLK6-7.
func (d *Device) SetRawFrequency(clk Clock, freq Frequency) error {
	if !d.initialized {
		return ErrNotInitialized
	}

	switch {
	case clk <= 5:
		return d.setFreqCLK0to5(clk, freq)
	case clk <= 7:
		return d.setFreqCLK6to7(clk, freq)
	default:
		return ErrInvalidParameter
	}
}

// SetMultisynthSource sets the PLL source for a multisynth.
func (d *Device) SetMultisynthSource(clk Clock, pll PLLType) error {
	regVal, err := d.rw.Read8(CLK0_CTRL + uint8(clk))
	if err != nil {
		return err
	}

	switch pll {
	case PLL_A:
		regVal &^= CLK_PLL_SELECT
	case PLL_B:
		regVal |= CLK_PLL_SELECT
	default:
		return ErrInvalidParameter
	}

	if err := d.rw.Write8(CLK0_CTRL+uint8(clk), regVal); err != nil {
		return err
	}

	d.pllAssignment[clk] = pll
	return nil
}

// SetCorrection sets the oscillator correction factor in parts-per-billion.
func (d *Device) SetCorrection(refOsc PLLReferenceOscillator, corr int32) error {
	d.refCorrection[refOsc] = corr

	if err := d.SetPLL(PLL_A, d.pllaFreq); err != nil {
		return err
	}
	if err := d.SetPLL(PLL_B, d.pllbFreq); err != nil {
		return err
	}
	return nil
}

// GetCorrection returns the oscillator correction factor in parts-per-billion.
func (d *Device) GetCorrection(refOsc PLLReferenceOscillator) int32 {
	return d.refCorrection[refOsc]
}

// PLLReset applies a reset to the indicated PLL.
func (d *Device) PLLReset(pll PLLType) error {
	switch pll {
	case PLL_A:
		return d.rw.Write8(PLL_RESET, PLL_RESET_A)
	case PLL_B:
		return d.rw.Write8(PLL_RESET, PLL_RESET_B)
	}
	return ErrInvalidParameter
}

// SetReferenceFrequency sets the reference frequency for the specified reference oscillator.
func (d *Device) SetReferenceFrequency(refOsc PLLReferenceOscillator, refFreq CrystalFrequency) {
	switch {
	case refFreq <= 30_000_000:
		d.crystalFreq[refOsc] = refFreq
		if refOsc == PLLInputClockIn {
			d.clkinDiv = CLKIN_DIV_1
		}
	case refFreq <= 60_000_000:
		d.crystalFreq[refOsc] = refFreq / 2
		if refOsc == PLLInputClockIn {
			d.clkinDiv = CLKIN_DIV_2
		}
	case refFreq <= 100_000_000:
		d.crystalFreq[refOsc] = refFreq / 4
		if refOsc == PLLInputClockIn {
			d.clkinDiv = CLKIN_DIV_4
		}
	}
}

// EnableOutput enables or disables a clock output.
func (d *Device) EnableOutput(clk Clock, enable bool) error {
	if clk > Clock7 {
		return ErrInvalidParameter
	}

	regVal, err := d.rw.Read8(OUTPUT_ENABLE_CTRL)
	if err != nil {
		return err
	}

	if enable {
		regVal &^= (1 << clk)
	} else {
		regVal |= (1 << clk)
	}

	return d.rw.Write8(OUTPUT_ENABLE_CTRL, regVal)
}

type DriveStrength uint8

const (
	DriveStrength2MA DriveStrength = iota
	DriveStrength4MA
	DriveStrength6MA
	DriveStrength8MA
)

// SetDriveStrength sets the drive strength of the specified clock output.
func (d *Device) SetDriveStrength(clk Clock, drive DriveStrength) error {
	if clk > Clock7 {
		return ErrInvalidParameter
	}

	regVal, err := d.rw.Read8(CLK0_CTRL + uint8(clk))
	if err != nil {
		return err
	}

	regVal &^= 0x03

	switch drive {
	case DriveStrength2MA: // 2mA
		regVal |= CLK_DRIVE_STRENGTH_2MA
	case DriveStrength4MA: // 4mA
		regVal |= CLK_DRIVE_STRENGTH_4MA
	case DriveStrength6MA: // 6mA
		regVal |= CLK_DRIVE_STRENGTH_6MA
	case DriveStrength8MA: // 8mA
		regVal |= CLK_DRIVE_STRENGTH_8MA
	default:
		return ErrInvalidParameter
	}

	return d.rw.Write8(CLK0_CTRL+uint8(clk), regVal)
}

// SetPhase sets the 7-bit phase register for the specified clock.
func (d *Device) SetPhase(clk Clock, phase uint8) error {
	phase &= 0x7F // Mask upper bit
	return d.rw.Write8(CLK0_PHASE_OFFSET+uint8(clk), phase)
}

// Fanout options for clock signals
type Fanout uint8

const (
	FanoutClockIn Fanout = iota
	FanoutXO
	FanoutMultisynth
)

// SetClockFanout enables or disables the clock fanout options for individual clock outputs.
// If you intend to output the XO or CLKIN on the clock outputs, enable this first.
// By default, only the Multisynth fanout is enabled at startup.
func (d *Device) SetClockFanout(fanout Fanout, enable bool) error {
	regVal, err := d.rw.Read8(FANOUT_ENABLE)
	if err != nil {
		return err
	}

	switch fanout {
	case FanoutClockIn:
		if enable {
			regVal |= CLKIN_ENABLE
		} else {
			regVal &^= CLKIN_ENABLE
		}
	case FanoutXO:
		if enable {
			regVal |= XTAL_ENABLE
		} else {
			regVal &^= XTAL_ENABLE
		}
	case FanoutMultisynth:
		if enable {
			regVal |= MULTISYNTH_ENABLE
		} else {
			regVal &^= MULTISYNTH_ENABLE
		}
	default:
		return ErrInvalidParameter
	}

	return d.rw.Write8(FANOUT_ENABLE, regVal)
}

// Clock source options
type ClockSource uint8

const (
	ClockSourceXTAL ClockSource = iota
	ClockSourceClockIn
	ClockSourceMS0
	ClockSourceMS
)

// SetClockSource sets the clock source for a multisynth (based on the options
// presented for Registers 16-23 in the Silicon Labs AN619 document).
// Choices are XTAL, CLKIN, MS0, or the multisynth associated with the clock output.
func (d *Device) SetClockSource(clk Clock, src ClockSource) error {
	if clk > Clock7 {
		return ErrInvalidParameter
	}

	regVal, err := d.rw.Read8(CLK0_CTRL + uint8(clk))
	if err != nil {
		return err
	}

	// Clear the input mask bits first
	regVal &^= CLK_INPUT_MASK

	switch src {
	case ClockSourceXTAL:
		regVal |= CLK_INPUT_XTAL
	case ClockSourceClockIn:
		regVal |= CLK_INPUT_CLKIN
	case ClockSourceMS0:
		if clk == Clock0 {
			return ErrInvalidParameter
		}
		regVal |= CLK_INPUT_MULTISYNTH_0_4
	case ClockSourceMS:
		regVal |= CLK_INPUT_MULTISYNTH_N
	default:
		return ErrInvalidParameter
	}

	return d.rw.Write8(CLK0_CTRL+uint8(clk), regVal)
}

// SetClockPower enables or disables power to a clock output (a power saving feature).
func (d *Device) SetClockPower(clk Clock, enable bool) error {
	if clk > Clock7 {
		return ErrInvalidParameter
	}

	regVal, err := d.rw.Read8(CLK0_CTRL + uint8(clk))
	if err != nil {
		return err
	}

	if enable {
		regVal &= 0x7F // Clear bit 7 (power on)
	} else {
		regVal |= 0x80 // Set bit 7 (power off)
	}

	return d.rw.Write8(CLK0_CTRL+uint8(clk), regVal)
}

// SetClockInvert inverts the clock output waveform.
func (d *Device) SetClockInvert(clk Clock, invert bool) error {
	if clk > Clock7 {
		return ErrInvalidParameter
	}

	regVal, err := d.rw.Read8(CLK0_CTRL + uint8(clk))
	if err != nil {
		return err
	}

	if invert {
		regVal |= CLK_INVERT
	} else {
		regVal &^= CLK_INVERT
	}

	return d.rw.Write8(CLK0_CTRL+uint8(clk), regVal)
}

// CalculatePLL calculates the PLL register values for the specified frequency
func (d *Device) CalculatePLL(pll PLLType, freq Frequency, correction int32, vcxo bool) (Frequency, RegisterSet) {
	var refFreq Frequency
	if pll == PLL_A {
		refFreq = Frequency(d.crystalFreq[d.pllaRefOsc]) * FREQ_MULT
	} else {
		refFreq = Frequency(d.crystalFreq[d.pllbRefOsc]) * FREQ_MULT
	}

	// Apply correction
	refFreq = refFreq + Frequency(((int64(correction)<<31)/1000000000)*int64(refFreq)>>31)

	// Bounds checking
	switch {
	case freq < PLL_VCO_MIN*FREQ_MULT:
		freq = PLL_VCO_MIN * FREQ_MULT
	case freq > PLL_VCO_MAX*FREQ_MULT:
		freq = PLL_VCO_MAX * FREQ_MULT
	}

	a := uint32(freq / refFreq)

	switch {
	case a < PLL_A_MIN:
		freq = refFreq * PLL_A_MIN
	case a > PLL_A_MAX:
		freq = refFreq * PLL_A_MAX
	}

	var b, c uint32
	if vcxo {
		b = uint32(((freq % refFreq) * 1000000) / refFreq)
		c = 1000000
	} else {
		b = uint32(((freq % refFreq) * rfracDenominator) / refFreq)
		if b != 0 {
			c = uint32(rfracDenominator)
		} else {
			c = 1
		}
	}

	p1 := 128*a + ((128 * b) / c) - 512
	p2 := 128*b - c*((128*b)/c)
	p3 := c

	lltmp := (refFreq * Frequency(b)) / Frequency(c)
	freqOut := lltmp + refFreq*Frequency(a)

	reg := RegisterSet{p1: p1, p2: p2, p3: p3}

	if vcxo {
		return Frequency(128*a*1000000 + b), reg
	}
	return freqOut, reg
}

// CalculateMultisynth calculates the multisynth register values for the specified frequency
func (d *Device) CalculateMultisynth(freq, pllFreq Frequency) (Frequency, RegisterSet) {
	divby4 := false
	retVal := uint8(0)

	// Bounds checking
	switch {
	case freq > MULTISYNTH_MAX_FREQ*FREQ_MULT:
		freq = MULTISYNTH_MAX_FREQ * FREQ_MULT
	case freq < MULTISYNTH_MIN_FREQ*FREQ_MULT:
		freq = MULTISYNTH_MIN_FREQ * FREQ_MULT
	}

	if freq >= MULTISYNTH_DIVBY4_FREQ*FREQ_MULT {
		divby4 = true
	}

	var a, b, c uint32

	if pllFreq == 0 {
		if !divby4 {
			lltmp := Frequency(PLL_VCO_MAX * FREQ_MULT)
			lltmp = lltmp / freq
			switch lltmp {
			case 5:
				lltmp = 4
			case 7:
				lltmp = 6
			}
			a = uint32(lltmp)
		} else {
			a = 4
		}
		b = 0
		c = 1
		pllFreq = Frequency(a) * freq
	} else {
		retVal = 1
		a = uint32(pllFreq / freq)

		switch {
		case a < MULTISYNTH_A_MIN:
			freq = pllFreq / MULTISYNTH_A_MIN
			a = MULTISYNTH_A_MIN
		case a > MULTISYNTH_A_MAX:
			freq = pllFreq / MULTISYNTH_A_MAX
			a = MULTISYNTH_A_MAX
		}

		b = uint32(((pllFreq % freq) * rfracDenominator) / freq)
		if b != 0 {
			c = uint32(rfracDenominator)
		} else {
			c = 1
		}
	}

	var p1, p2, p3 uint32
	if divby4 {
		p3 = 1
		p2 = 0
		p1 = 0
	} else {
		p1 = 128*a + ((128 * b) / c) - 512
		p2 = 128*b - c*((128*b)/c)
		p3 = c
	}

	reg := RegisterSet{p1: p1, p2: p2, p3: p3}

	if retVal == 0 {
		return pllFreq, reg
	}
	return freq, reg
}

// SetMultisynth programs the multisynth registers for the specified clock.
// For CLK0-5, reg contains p1, p2, p3 values. For CLK6/7, only p1 is used.
func (d *Device) SetMultisynth(clk Clock, reg RegisterSet, intMode, rDiv, divBy4 uint8) error {
	switch {
	case clk <= 5:
		params := make([]byte, 8)
		params[0] = byte((reg.p3 >> 8) & 0xFF)
		params[1] = byte(reg.p3 & 0xFF)

		regVal, err := d.rw.Read8(CLK0_PARAMETERS + 2 + uint8(clk)*8)
		if err != nil {
			return err
		}
		regVal &^= 0x03
		params[2] = regVal | byte((reg.p1>>16)&0x03)

		params[3] = byte((reg.p1 >> 8) & 0xFF)
		params[4] = byte(reg.p1 & 0xFF)
		params[5] = byte(((reg.p3 >> 12) & 0xF0) | ((reg.p2 >> 16) & 0x0F))
		params[6] = byte((reg.p2 >> 8) & 0xFF)
		params[7] = byte(reg.p2 & 0xFF)

		baseAddr := CLK0_PARAMETERS + uint8(clk)*8
		for i := range params {
			if err := d.rw.Write8(baseAddr+uint8(i), params[i]); err != nil {
				return err
			}
		}

		d.setInt(clk, intMode)
		return d.msDiv(clk, rDiv, divBy4)
	case clk <= 7:
		// CLK6/7
		baseAddr := CLK6_PARAMETERS
		if clk == 7 {
			baseAddr = CLK7_PARAMETERS
		}
		if err := d.rw.Write8(uint8(baseAddr), byte(reg.p1)); err != nil {
			return err
		}
		return d.msDiv(clk, rDiv, divBy4)
	default:
		return ErrInvalidParameter
	}
}

func (d *Device) setFreqCLK0to5(clk Clock, freq Frequency) error {
	var rDiv uint8
	var divBy4 uint8
	var intMode uint8

	// Bounds checking
	switch {
	case freq < CLKOUT_MIN_FREQ*FREQ_MULT:
		freq = CLKOUT_MIN_FREQ * FREQ_MULT
	case freq > MULTISYNTH_MAX_FREQ*FREQ_MULT:
		freq = MULTISYNTH_MAX_FREQ * FREQ_MULT
	}

	// Check if frequency requires PLL recalculation
	if freq > MULTISYNTH_SHARE_MAX*FREQ_MULT {
		// Check other clocks on same PLL
		for i := range Clock(6) {
			if d.clkFreq[i] > MULTISYNTH_SHARE_MAX*FREQ_MULT {
				if i != clk && d.pllAssignment[i] == d.pllAssignment[clk] {
					return ErrInvalidPLLClockSetting
				}
			}
		}

		// Enable output on first set
		if !d.clkFirstSet[clk] {
			d.EnableOutput(clk, true)
			d.clkFirstSet[clk] = true
		}

		d.clkFreq[clk] = freq

		// Calculate PLL frequency
		pllFreq, _ := d.CalculateMultisynth(freq, 0)
		d.SetPLL(d.pllAssignment[clk], pllFreq)

		// Recalculate other synths on same PLL
		for i := range Clock(6) {
			if d.clkFreq[i] != 0 && d.pllAssignment[i] == d.pllAssignment[clk] {
				tempFreq := d.clkFreq[i]
				tempFreq, rDiv = d.selectRDiv(tempFreq)

				_, tempReg := d.CalculateMultisynth(tempFreq, pllFreq)

				if tempFreq >= MULTISYNTH_DIVBY4_FREQ*FREQ_MULT {
					divBy4 = 1
					intMode = 1
				} else {
					divBy4 = 0
					intMode = 0
				}

				d.SetMultisynth(i, tempReg, intMode, rDiv, divBy4)
			}
		}

		d.PLLReset(d.pllAssignment[clk])
	} else {
		d.clkFreq[clk] = freq

		if !d.clkFirstSet[clk] {
			d.EnableOutput(clk, true)
			d.clkFirstSet[clk] = true
		}

		freq, rDiv = d.selectRDiv(freq)

		var pllFreq Frequency
		if d.pllAssignment[clk] == PLL_A {
			pllFreq = d.pllaFreq
		} else {
			pllFreq = d.pllbFreq
		}

		_, msReg := d.CalculateMultisynth(freq, pllFreq)
		d.SetMultisynth(clk, msReg, intMode, rDiv, divBy4)
	}

	return nil
}

func (d *Device) setFreqCLK6to7(clk Clock, freq Frequency) error {
	var rDiv uint8
	var divBy4 uint8
	var intMode uint8

	// Bounds checking for CLK6/7
	if freq > 0 && freq < CLKOUT67_MIN_FREQ*FREQ_MULT {
		freq = CLKOUT_MIN_FREQ * FREQ_MULT
	}
	if freq >= MULTISYNTH_DIVBY4_FREQ*FREQ_MULT {
		freq = MULTISYNTH_DIVBY4_FREQ*FREQ_MULT - 1
	}

	var msReg RegisterSet
	var pllFreq Frequency

	otherClk := uint8(7)
	if clk == 7 {
		otherClk = 6
	}

	if d.clkFreq[otherClk] != 0 {
		// Other CLK6/7 already set, must use integer division
		if d.pllbFreq%freq != 0 || (d.pllbFreq/freq)%2 != 0 {
			return ErrInvalidPLLDivision
		}

		d.clkFreq[clk] = freq
		freq, rDiv = d.selectRDivMS67(freq)
		_, msReg = d.multisynth67Calc(freq, d.pllbFreq)
	} else {
		// Set PLLB based on this clock
		d.clkFreq[clk] = freq
		freq, rDiv = d.selectRDivMS67(freq)
		pllFreq, msReg = d.multisynth67Calc(freq, 0)

		d.SetPLL(d.pllAssignment[clk], pllFreq)
	}

	divBy4 = 0
	intMode = 0

	return d.SetMultisynth(clk, msReg, intMode, rDiv, divBy4)
}

func (d *Device) setInt(clk Clock, enable uint8) error {
	regVal, err := d.rw.Read8(CLK0_CTRL + uint8(clk))
	if err != nil {
		return err
	}

	if enable == 1 {
		regVal |= CLK_INTEGER_MODE
	} else {
		regVal &^= CLK_INTEGER_MODE
	}

	return d.rw.Write8(CLK0_CTRL+uint8(clk), regVal)
}

func (d *Device) msDiv(clk Clock, rDiv, divBy4 uint8) error {
	var regAddr uint8

	switch clk {
	case 0:
		regAddr = CLK0_PARAMETERS + 2
	case 1:
		regAddr = CLK1_PARAMETERS + 2
	case 2:
		regAddr = CLK2_PARAMETERS + 2
	case 3:
		regAddr = CLK3_PARAMETERS + 2
	case 4:
		regAddr = CLK4_PARAMETERS + 2
	case 5:
		regAddr = CLK5_PARAMETERS + 2
	case 6, 7:
		regAddr = CLK6_7_OUTPUT_DIVIDER
	default:
		return ErrInvalidParameter
	}

	regVal, err := d.rw.Read8(regAddr)
	if err != nil {
		return err
	}

	switch {
	case clk <= 5:
		regVal &^= 0x7C

		if divBy4 == 0 {
			regVal &^= OUTPUT_CLK_DIVBY4
		} else {
			regVal |= OUTPUT_CLK_DIVBY4
		}

		regVal |= (rDiv << OUTPUT_CLK_DIV_SHIFT)
	case clk == 6:
		regVal &^= 0x07
		regVal |= rDiv
	case clk == 7:
		regVal &^= 0x70
		regVal |= (rDiv << OUTPUT_CLK_DIV_SHIFT)
	}

	return d.rw.Write8(regAddr, regVal)
}

func (d *Device) selectRDiv(freq Frequency) (Frequency, uint8) {
	rDiv := OUTPUT_CLK_DIV_1

	switch {
	case freq >= CLKOUT_MIN_FREQ*FREQ_MULT && freq < CLKOUT_MIN_FREQ*FREQ_MULT*2:
		rDiv = OUTPUT_CLK_DIV_128
		freq *= 128
	case freq >= CLKOUT_MIN_FREQ*FREQ_MULT*2 && freq < CLKOUT_MIN_FREQ*FREQ_MULT*4:
		rDiv = OUTPUT_CLK_DIV_64
		freq *= 64
	case freq >= CLKOUT_MIN_FREQ*FREQ_MULT*4 && freq < CLKOUT_MIN_FREQ*FREQ_MULT*8:
		rDiv = OUTPUT_CLK_DIV_32
		freq *= 32
	case freq >= CLKOUT_MIN_FREQ*FREQ_MULT*8 && freq < CLKOUT_MIN_FREQ*FREQ_MULT*16:
		rDiv = OUTPUT_CLK_DIV_16
		freq *= 16
	case freq >= CLKOUT_MIN_FREQ*FREQ_MULT*16 && freq < CLKOUT_MIN_FREQ*FREQ_MULT*32:
		rDiv = OUTPUT_CLK_DIV_8
		freq *= 8
	case freq >= CLKOUT_MIN_FREQ*FREQ_MULT*32 && freq < CLKOUT_MIN_FREQ*FREQ_MULT*64:
		rDiv = OUTPUT_CLK_DIV_4
		freq *= 4
	case freq >= CLKOUT_MIN_FREQ*FREQ_MULT*64 && freq < CLKOUT_MIN_FREQ*FREQ_MULT*128:
		rDiv = OUTPUT_CLK_DIV_2
		freq *= 2
	}

	return freq, uint8(rDiv)
}

func (d *Device) selectRDivMS67(freq Frequency) (Frequency, uint8) {
	rDiv := OUTPUT_CLK_DIV_1

	// The minimum frequency for MS67 with max divider is lower than the calculated constant
	// We use the same ranges as selectRDiv for consistency
	minFreq := Frequency(CLKOUT_MIN_FREQ * FREQ_MULT)

	switch {
	case freq >= minFreq && freq < minFreq*2:
		rDiv = OUTPUT_CLK_DIV_128
		freq *= 128
	case freq >= minFreq*2 && freq < minFreq*4:
		rDiv = OUTPUT_CLK_DIV_64
		freq *= 64
	case freq >= minFreq*4 && freq < minFreq*8:
		rDiv = OUTPUT_CLK_DIV_32
		freq *= 32
	case freq >= minFreq*8 && freq < minFreq*16:
		rDiv = OUTPUT_CLK_DIV_16
		freq *= 16
	case freq >= minFreq*16 && freq < minFreq*32:
		rDiv = OUTPUT_CLK_DIV_8
		freq *= 8
	case freq >= minFreq*32 && freq < minFreq*64:
		rDiv = OUTPUT_CLK_DIV_4
		freq *= 4
	case freq >= minFreq*64 && freq < minFreq*128:
		rDiv = OUTPUT_CLK_DIV_2
		freq *= 2
	}

	return freq, uint8(rDiv)
}

func (d *Device) multisynth67Calc(freq, pllFreq Frequency) (Frequency, RegisterSet) {
	// Bounds checking
	if freq > MULTISYNTH67_MAX_FREQ*FREQ_MULT {
		freq = MULTISYNTH67_MAX_FREQ * FREQ_MULT
	}
	if freq < MULTISYNTH_MIN_FREQ*FREQ_MULT {
		freq = MULTISYNTH_MIN_FREQ * FREQ_MULT
	}

	var a uint32

	if pllFreq == 0 {
		lltmp := Frequency(PLL_VCO_MAX*FREQ_MULT - MULTISYNTH_SHARE_MAX)
		lltmp = lltmp / freq
		a = uint32(lltmp)

		// Must be even
		if a%2 != 0 {
			a++
		}

		// Bounds check
		if a < MULTISYNTH_A_MIN {
			a = MULTISYNTH_A_MIN
		}
		if a > MULTISYNTH67_A_MAX {
			a = MULTISYNTH67_A_MAX
		}

		pllFreq = Frequency(a) * freq

		// PLL bounds
		if pllFreq > PLL_VCO_MAX*FREQ_MULT {
			a -= 2
			pllFreq = Frequency(a) * freq
		} else if pllFreq < PLL_VCO_MIN*FREQ_MULT {
			a += 2
			pllFreq = Frequency(a) * freq
		}

		return pllFreq, RegisterSet{p1: a, p2: 0, p3: 0}
	} else {
		if pllFreq%freq != 0 {
			return 0, RegisterSet{}
		}

		a = uint32(pllFreq / freq)

		if a < MULTISYNTH_A_MIN || a > MULTISYNTH67_A_MAX {
			return 0, RegisterSet{}
		}

		return 1, RegisterSet{p1: a, p2: 0, p3: 0}
	}
}
