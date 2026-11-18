package bmp580

// I2C addresses. Datasheet section 5.6.
const (
	// Address is the address with SDO pulled high, the default of most boards.
	Address = 0x47
	// AddressAlt is the address with SDO pulled low.
	AddressAlt = 0x46
)

// Registers. Datasheet section 7.
const (
	regChipID    = 0x01
	regTempData  = 0x1D // TEMP_DATA_XLSB, followed by TEMP_DATA_LSB and _MSB, then PRESS_DATA_*.
	regIntStatus = 0x27
	regStatus    = 0x28
	regOSRConfig = 0x36
	regODRConfig = 0x37
	regCmd       = 0x7E
)

// Register values and fields. Datasheet sections 4.3 and 7.
const (
	chipIDBMP580 = 0x50

	cmdSoftReset = 0xB6

	intStatusPOR = 1 << 4 // Power-on or soft reset complete.

	statusNVMReady = 1 << 1
	statusNVMError = 1 << 2

	osrPressureEnable = 1 << 6

	// ODR_CONFIG: deep_dis keeps the sensor out of deep standby, odr stays at
	// its reset value of 1 Hz (it only matters in normal mode) and the power
	// mode is in the two low bits.
	odrDeepDisable = 1 << 7
	odrReset1Hz    = 0x1C << 2
	pwrModeMask    = 0x03
	pwrModeStandby = 0x00
	pwrModeForced  = 0x02
)

// Oversampling is the number of samples the sensor averages for a reading.
// More samples mean less noise and a longer measurement.
type Oversampling uint8

const (
	Sampling1X Oversampling = iota
	Sampling2X
	Sampling4X
	Sampling8X
	Sampling16X
	Sampling32X
	Sampling64X
	Sampling128X
)
