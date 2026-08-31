package ads1015

// Address is the default I2C address of the ADS1015. The actual address
// depends on how the ADDR pin is wired; see the datasheet for the other
// three options (0x49, 0x4A, 0x4B).
const Address uint16 = 0x48
const Address2 uint16 = 0x49
const Address3 uint16 = 0x4A
const Address4 uint16 = 0x4B

// Registers, see the datasheet Table 8-2
const (
	regConversion    uint8 = 0b00
	regConfig        uint8 = 0b01
	regLowThreshold  uint8 = 0b10
	regHighThreshold uint8 = 0b11
)

// Config register bit masks, see the datasheet section 8.6.3.
const (
	// configOSStart, written to bit 15, starts a single conversion.
	configOSStart uint16 = 0x8000
	// configOSReady, read from bit 15, is set when no conversion is
	// in progress.
	configOSReady uint16 = 0x8000
)

// Mux selects the input(s) measured by a conversion.
type Mux uint16

// Table 8-4
const (
	MuxDiff01  Mux = 0b000 << 12 // AINP = AIN0, AINN = AIN1
	MuxDiff03  Mux = 0b001 << 12 // AINP = AIN0, AINN = AIN3
	MuxDiff13  Mux = 0b010 << 12 // AINP = AIN1, AINN = AIN3
	MuxDiff23  Mux = 0b011 << 12 // AINP = AIN2, AINN = AIN3
	MuxSingle0 Mux = 0b100 << 12 // AINP = AIN0, AINN = GND
	MuxSingle1 Mux = 0b101 << 12 // AINP = AIN1, AINN = GND
	MuxSingle2 Mux = 0b110 << 12 // AINP = AIN2, AINN = GND
	MuxSingle3 Mux = 0b111 << 12 // AINP = AIN3, AINN = GND
)

// muxSingleEnded maps a channel number (0-3) to its Mux setting.
var muxSingleEnded = [4]Mux{MuxSingle0, MuxSingle1, MuxSingle2, MuxSingle3}

// Gain selects the full-scale input range of the programmable gain
// amplifier (PGA).
type Gain uint16

const (
	Gain6144mV Gain = 0b000 << 9 // +/-6.144V
	Gain4096mV Gain = 0b001 << 9 // +/-4.096V
	Gain2048mV Gain = 0b010 << 9 // +/-2.048V, power-on default
	Gain1024mV Gain = 0b011 << 9 // +/-1.024V
	Gain0512mV Gain = 0b100 << 9 // +/-0.512V
	Gain0256mV Gain = 0b101 << 9 // +/-0.256V
	// Gain0256mV Gain = 0b110 << 9 // +/-0.256V
	// Gain0256mV Gain = 0b111 << 9 // +/-0.256V
)

// FullScaleVoltage returns the largest voltage magnitude, in milliVolts, that a
// conversion at this gain can represent.
func (g Gain) FullScaleVoltage() int32 {
	switch g {
	case Gain6144mV:
		return 6144
	case Gain4096mV:
		return 4096
	case Gain2048mV:
		return 2048
	case Gain1024mV:
		return 1024
	case Gain0512mV:
		return 512
	case Gain0256mV:
		return 256
	default:
		return 0
	}
}

// Mode selects between continuous and single-shot conversions.
type Mode uint16

const (
	ModeContinuous Mode = 0b0 << 8
	ModeSingle     Mode = 0b1 << 8
)

// DataRate sets the output data rate, in samples per second (SPS).
type DataRate uint16

const (
	DataRate128SPS  DataRate = 0b000 << 5
	DataRate250SPS  DataRate = 0b001 << 5
	DataRate490SPS  DataRate = 0b010 << 5
	DataRate920SPS  DataRate = 0b011 << 5
	DataRate1600SPS DataRate = 0b100 << 5 // power-on default
	DataRate2400SPS DataRate = 0b101 << 5
	DataRate3300SPS DataRate = 0b110 << 5
	// DataRate3300SPS DataRate = 0b111 << 5 // repeated value
)

// ComparatorMode selects between traditional and window comparator modes.
type ComparatorMode uint16

const (
	ComparatorModeTraditional ComparatorMode = 0b0 << 4
	ComparatorModeWindow      ComparatorMode = 0b1 << 4
)

// ComparatorPolarity sets the polarity of the ALERT/RDY pin when the
// comparator is active.
type ComparatorPolarity uint16

const (
	ComparatorPolarityActiveLow  ComparatorPolarity = 0b0 << 3
	ComparatorPolarityActiveHigh ComparatorPolarity = 0b1 << 3
)

// ComparatorLatch enables or disables the latching comparator.
type ComparatorLatch uint16

const (
	ComparatorNonLatching ComparatorLatch = 0b0 << 2
	ComparatorLatching    ComparatorLatch = 0b1 << 2
)

// ComparatorQueue sets how many successive conversions must lie beyond the
// threshold before ALERT/RDY is asserted, or disables the comparator.
type ComparatorQueue uint16

const (
	ComparatorQueueAfter1Conv ComparatorQueue = 0b00
	ComparatorQueueAfter2Conv ComparatorQueue = 0b01
	ComparatorQueueAfter4Conv ComparatorQueue = 0b10
	ComparatorQueueDisable    ComparatorQueue = 0b11 // power-on default
)
