// Package bh1745 provides a driver for the BH1745NUC RGBC colour sensor by ROHM.
//
// Datasheet: https://fscdn.rohm.com/en/products/databook/datasheet/ic/sensor/light/bh1745nuc-e.pdf
package bh1745

// I2C addresses (set by ADDR pin: low → 0x38, high → 0x39).
const (
	Address    uint16 = 0x38
	AddressAlt uint16 = 0x39
)

// Registers
const (
	REG_SYSTEM_CONTROL = 0x40 // sw_reset (bit 7), int_reset (bit 6), part_id (bits 5:0)
	REG_MODE_CONTROL1  = 0x41 // measurement_time (bits 2:0)
	REG_MODE_CONTROL2  = 0x42 // valid (bit 7 RO), rgbc_en (bit 4), adc_gain (bits 1:0)
	REG_MODE_CONTROL3  = 0x44 // write 0x02 to activate (undocumented; 0x00 = off)
	REG_RED_DATA_LSB   = 0x50 // red channel, little-endian uint16 at 0x50–0x51
	REG_GREEN_DATA_LSB = 0x52
	REG_BLUE_DATA_LSB  = 0x54
	REG_CLEAR_DATA_LSB = 0x56
	REG_DINT_DATA_LSB  = 0x58
	REG_INTERRUPT      = 0x60 // int_status (bit 7 RO), int_latch (bit 4), int_source (bits 3:2), int_en (bit 0)
	REG_PERSISTENCE    = 0x61 // persistence mode (bits 1:0)
	REG_THRESHOLD_LSB  = 0x62 // interrupt low threshold, little-endian uint16
	REG_THRESHOLD_MSB  = 0x64 // interrupt high threshold, little-endian uint16
	REG_MANUFACTURER   = 0x92 // manufacturer ID
)

// SYSTEM_CONTROL bits
const (
	SW_RESET     = 0x80 // software reset; self-clears when reset is complete
	INT_RESET    = 0x40 // interrupt reset
	PART_ID_MASK = 0x3F // read-only part ID field
	PART_ID      = 0x0B // expected part ID value
)

// MODE_CONTROL2 bits
const (
	VALID   = 0x80 // read-only; 1 when RGBC data is valid after first measurement
	RGBC_EN = 0x10 // enable RGBC measurement
)

// Chip constants
const (
	MANUFACTURER_ID = 0xE0
)

// MeasurementTime represents the RGBC integration time.
type MeasurementTime byte

const (
	MeasTime160ms  MeasurementTime = 0b000 // 160 ms
	MeasTime320ms  MeasurementTime = 0b001 // 320 ms
	MeasTime640ms  MeasurementTime = 0b010 // 640 ms
	MeasTime1280ms MeasurementTime = 0b011 // 1280 ms
	MeasTime2560ms MeasurementTime = 0b100 // 2560 ms
	MeasTime5120ms MeasurementTime = 0b101 // 5120 ms
)

// ADCGain represents the ADC gain multiplier.
type ADCGain byte

const (
	Gain1X  ADCGain = 0b00 // 1×
	Gain2X  ADCGain = 0b01 // 2×
	Gain16X ADCGain = 0b10 // 16×
)
