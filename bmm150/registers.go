package bmm150

// DefaultAddress is the I2C address with SDO and CSB2 pulled low.
const DefaultAddress = 0x10

const chipID = 0x32

// Registers. Datasheet section 5.
const (
	regChipID    = 0x40
	regDataX     = 0x42
	regPowerCtrl = 0x4B
	regOpMode    = 0x4C
	regRepXY     = 0x51
	regRepZ      = 0x52

	// Trim registers DIG_X1 to DIG_XY1.
	regTrim    = 0x5D
	trimLength = 21
)

const (
	powerOn   = 0x01
	softReset = 0x82

	opModeNormal = 0x00
)

// Output values that indicate an ADC overflow. Bosch BMM150 Sensor API.
const (
	overflowXY = -4096
	overflowZ  = -16384
)

// DataRate is the output data rate in normal mode.
type DataRate uint8

const (
	DataRate10Hz DataRate = 0
	DataRate2Hz  DataRate = 1
	DataRate6Hz  DataRate = 2
	DataRate8Hz  DataRate = 3
	DataRate15Hz DataRate = 4
	DataRate20Hz DataRate = 5
	DataRate25Hz DataRate = 6
	DataRate30Hz DataRate = 7
)

// Preset selects the number of repetitions per measurement. More repetitions
// reduce noise but increase power use and limit the maximum data rate.
// Datasheet section 4.2.4.
type Preset uint8

const (
	Regular Preset = iota
	LowPower
	Enhanced
	HighAccuracy
)
