package pca9685

// Software reset addresses for generic i2c implementation
// and PCA specific SWRST. See section 7.6
const (
	I2CSWRSTADR      = 0x0
	PCA9685SWRSTBYTE = 0b0000_0110 // 0x06
)

// Registries with nomenclature as seen in manual
const (
	SUBADR1 uint8 = 0x2
	SUBADR2 uint8 = 0x3
	SUBADR3 uint8 = 0x4

	MODE1      uint8 = 0x0
	MODE2      uint8 = 0x1
	ALLCALLADR uint8 = 0x05

	SWRESET uint8 = 0b0000_0011

	// Start of LED registries. corresponds to LED0_ON_L register.
	// Use LED function to get registries of a specific PWM channel register.
	LEDSTART = 0x06
)

// MODE1
const (
	RESET             byte = 0b1000_0000
	EXTCLK            byte = 0b0100_0000
	AI                byte = 0b0010_0000
	SLEEP             byte = 0b0001_0000
	defaultMODE1Value byte = 0b0001_0001
)

// MODE2
const (
	OUTDRV = 1 << 2
	INVRT  = 1 << 4
)

// LED channels from 0-15. Returns 4 registries associated with
// the channel PWM signal. Channel 250 (0xFA) gives ALL_LED registers.
// The L suffix represents the 8 LSB of 12 bit timing, the H suffix represents
// the 4 MSB of the timing. This way you have 0-4095 timing options. ON or OFF
// will be a number between these two or will be simply ON or OFF (fifth bit of H registry).
//
// The SetPWM implementation in this library does not do phase-shifting so OFF will always
// happen at time stamp 0. ON will solely decide duty cycle.
func LED(ch uint8) (ONL, ONH, OFFL, OFFH uint8) {
	switch {
	case ch == ALLLED:
		return ALLLED, ALLLED + 1, ALLLED + 2, ALLLED + 3

	case ch > 15:
		panic("PWM channel out of range [0-15]")
	default:
	}
	// 4 registries per channel. Starts at 6
	onLReg := LEDSTART + 4*ch
	return onLReg, onLReg + 1, onLReg + 2, onLReg + 3
}

const (
	// ALLLED is channel that selects registries to control all leds. Use with function LED
	ALLLED = 0xfa
	// PRESCALE Prescaling byte
	PRESCALE byte = 0xFE
)
