package as560x // import tinygo.org/x/drivers/ams560x

// DefaultAddress is the default I2C address of the AMS AS560x sensors (0x36).
const DefaultAddress uint8 = 0x36

// AS560x common device registers
const (
	// ZMCO contains the number of times a BURN_ANGLE command has been executed (max 3 burns)
	ZMCO = 0x00
	// ZPOS is the zero (start) position in RAW_ANGLE terms.
	ZPOS = 0x01
	// CONF supports custom config. Raw 14-bit register. See datasheet for mapping or use 'virtual registers' below.
	CONF = 0x07
	// STATUS indicates magnet position. Encapsulates MD, ML & MH. See also 'virtual registers' below.
	STATUS = 0x0b
	// RAW_ANGLE is the raw unscaled & unadjusted angle (12 bit: 0-4095/0xfff)
	RAW_ANGLE = 0x0c
	// ANGLE is RAW_ANGLE scaled & adjusted according to ZPOS (and MPOS/MANG on AS5600). (12 bit: 0-4095/0xfff)
	ANGLE = 0x0e
	// AGC is the Automatic Gain Control based on temp, airgap etc. 0-255 @ 5V, 0-128 @ 3.3V.
	AGC = 0x1a
	// MAGNITUDE indicates the magnitude value of the internal CORDIC output. See datasheet for more info.
	MAGNITUDE = 0x1b
	// BURN performs permanent programming of some registers. See BURN_XYZ cmd constants below for commands.
	BURN = 0xff
)

// AS5600 specific registers
const (
	// MPOS is the maximum position in RAW_ANGLE terms. With ZPOS, defines a 'narrower angle' for higher resolution.
	MPOS = 0x03
	// MANG is the maximum angle. With ZPOS, defines a 'narrower angle' for higher resolution.
	MANG = 0x05
)

// AS5601 specific registers
const (
	// ABN. See datasheet for mapping
	ABN = 0x09
	// PUSHTHR. Configures push-button function. See datasheet and AGC
	PUSHTHR = 0x0a
)

// 'Virtual Registers' (VRs) are bitfields within the registers above.
// These are not real register addresses recognized by the chip,
// but they are recognized by the driver for convenience.

// virtualRegisterStartAddress defines the start of the virtual register address range.
const virtualRegisterStartAddress = 0xa0

const (
	// VRs for CONF

	// WD is a Virtual Register for the Watchdog timer. See WATCHDOG_TIMER consts.
	WD = iota + virtualRegisterStartAddress
	// FTH is a Virtual Register for the Fast Filter Threshold. See FAST_FILTER_THRESHOLD consts.
	FTH
	// SF is a Virtual Register for the Slow Filter. See SLOW_FILTER_RESPONSE consts.
	SF
	// PWMF is a Virtual Register for PWM Frequency (AS5600 ONLY). See PWM_FREQUENCY consts.
	PWMF
	// OUTS is a Virtual Register for the Output Stage (AS5600 ONLY). See OUTPUT_STAGE consts.
	OUTS
	// HYST is a Virtual Register for Hysteresis. See HYSTERESIS consts.
	HYST
	// PM is a Virtual Register for the Power Mode. See POWER_MODE consts.
	PM

	// VRs for STATUS (0 = unset, 1 = set)

	// MD is a Virtual Register for the 'Magnet was detected' flag.
	MD
	// ML is a Virtual Register for the 'AGC maximum gain overflow' a.k.a 'magnet too weak' flag.
	ML
	// MH is a Virtual Register for the 'AGC minimum gain overflow' a.k.a 'magnet too strong' flag.
	MH
)

// POWER_MODE values for the PM component of CONF (and the PM VR)
const (
	// PM_NOM is the normal 'always on' power mode. No polling, max 6.5mA current
	PM_NOM = iota
	// PM_LPM1 is Low Power Mode 1. 5ms polling, max 3.4mA current
	PM_LPM1
	// PM_LPM2 is Low Power Mode 2. 20ms polling, max 1.8mA current
	PM_LPM2
	// PM_LPM3 is Low Power Mode 3. 100ms polling, max 1.5mA current
	PM_LPM3
)

// HYSTERESIS values for the HYST component of CONF (and the HYST VR)
const (
	// HYST_OFF disables any hysteresis of the output
	HYST_OFF = iota
	// HYST_1LSB enables output hysteresis using 1 LSB
	HYST_1LSB
	// HYST_2LSB enables output hysteresis using 2 LSBs
	HYST_2LSB
	// HYST_3LSB enables output hysteresis using 3 LSBs
	HYST_3LSB
)

// OUTPUT_STAGE values for the OUTS component of CONF (and the OUTS VR - AS5600 ONLY)
const (
	// OS_ANALOG_FULL_RANGE enables analog output with full range (0%-100% VDD)
	OS_ANALOG_FULL_RANGE = iota
	// OS_ANALOG_REDUCED_RANGE enables analog output with reduced range (10%-90% VDD)
	OS_ANALOG_REDUCED_RANGE
	// OS_DIGITAL_PWM enables digital PWM output. Frequency determined by PWMF
	OS_DIGITAL_PWM
)

// PWM_FREQUENCY values for the PWMF component of CONF (and the PWMF VR - ASS5600 ONLY)
const (
	// PWMF_115_HZ enables PWM at 115 Hz
	PWMF_115_HZ = iota
	// PWMF_230_HZ enables PWM at 230 Hz
	PWMF_230_HZ
	// PWMF_460_HZ enables PWM at 460 Hz
	PWMF_460_HZ
	// PWMF_920_HZ enables PWM at 920 Hz
	PWMF_920_HZ
)

// SLOW_FILTER_RESPONSE values for the SF (slow filter) component of CONF (and the SF VR)
const (
	// SF_16X enables a 16x Slow Filter step response
	SF_16X = iota
	// SF_8X enables a 8x Slow Filter step response
	SF_8X
	// SF_4X enables a 4x Slow Filter step response
	SF_4X
	// SF_2X enables a 2x Slow Filter step response
	SF_2X
)

// FAST_FILTER_THRESHOLD values for the FTH (fast filter threshold) component of CONF (and the FTH VR)
const (
	// FTH_NONE disables the fast filter (slow filter only)
	FTH_NONE = iota
	// FTH_6LSB enables a fast filter threshold with 6 LSBs
	FTH_6LSB
	// FTH_7LSB enables a fast filter threshold with 7 LSBs
	FTH_7LSB
	// FTH_9LSB enables a fast filter threshold with 9 LSBs
	FTH_9LSB
	// FTH_18LSB enables a fast filter threshold with 18 LSBs
	FTH_18LSB
	// FTH_21LSB enables a fast filter threshold with 21 LSBs
	FTH_21LSB
	// FTH_24LSB enables a fast filter threshold with 24 LSBs
	FTH_24SB
	// FTH_10LSB enables a fast filter threshold with 10 LSBs
	FTH_10LSB
)

// WATCHDOG_TIMER values for the WD component of CONF (and the WD VR)
const (
	// WD_OFF disables the Watchdog Timer
	WD_OFF = iota
	// WD_ON enables the Watchdog Timer (automatic entry into LPM3 low-power mode enabled)
	WD_ON
)

// constants for the raw STATUS register bitfield value.
const (
	// STATUS_MH is set in STATUS when the magnet field is too strong (AGC minimum gain overflow)
	STATUS_MH = 1 << (iota + 3)
	// STATUS_ML is set in STATUS when the magnet field is too weak (AGC maximum gain overflow)
	STATUS_ML
	// STATUS_MD is set n STATUS when the magnet is detected. Doesn't seem to work with some units.
	STATUS_MD
)

// ABN_MAPPING values for the ABN register (AS5601 ONLY)
const (
	// ABN_8 configures 8 output positions (61 Hz)
	ABN_8 = iota
	// ABN_16 configures 16 output positions (122 Hz)
	ABN_16
	// ABN_32 configures 32 output positions (244 Hz)
	ABN_32
	// ABN_64 configures 64 output positions (488 Hz)
	ABN_64
	// ABN_128 configures 128 output positions (976 Hz)
	ABN_128
	// ABN_256 configures 256 output positions (1.95 KHz)
	ABN_256
	// ABN_512 configures 512 output positions (3.9 KHz)
	ABN_512
	// ABN_1024 configures 1024 output positions (7.8 KHz)
	ABN_1024
	// ABN_2048 configures 2048 output positions (15.6 KHz)
	ABN_2048
)

// BURN_CMD is a command to write to the BURN register.
type BURN_CMD uint16

const (
	// BURN_ANGLE is the value to write to BURN to permanently program ZPOS & MPOS (Max 3 times!)
	BURN_ANGLE BURN_CMD = 0x80
	// BURN_SETTING is the value to write to BURN to permanently program MANG & CONF (ONCE ONLY!)
	BURN_SETTING BURN_CMD = 0x40
)

// BURN_ANGLE_COUNT_MAX is a constant for the maximum number of times a BURN_ANGLE command can be executed. Compare this with ZMCO
const BURN_ANGLE_COUNT_MAX uint16 = 3
