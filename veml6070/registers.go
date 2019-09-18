package veml6070

// I2C addresses and other constants

const (
	ADDR_L = 0x38 // 7bit address of the VEML6070 (write, read)
	ADDR_H = 0x39 // 7bit address of the VEML6070 (read)
)

// Some possible values for resistance value (in ohm) of VEML6070 calibration resistor
const (
	RSET_240K = 240000
	RSET_270K = 270000
	RSET_300K = 300000
	RSET_600K = 600000
)

// Possible values for integration time of VEML6070
// (internally represents the config register bit mask)
const (
	IT_HALF = 0x00
	IT_1    = 0x04
	IT_2    = 0x08
	IT_4    = 0x0C
)

// Possible values for UVI (UV index) risk level estimations - the VEML6070 can
// only estimate UVI risk levels since it can only sense UVA rays but the vendor
// tried to come up with some coarse thresholds, from application notes
const (
	UVI_RISK_LOW = iota
	UVI_RISK_MODERATE
	UVI_RISK_HIGH
	UVI_RISK_VERY_HIGH
	UVI_RISK_EXTREME
)

// Scale factor in milliseconds / ohm to determine refresh time
// (aka sampling time) without IT_FACTOR for any given RSET, from datasheet.
// Note: 100.0 milliseconds are applicable for RSET=240 kOhm and IT_FACTOR=1
const RSET_TO_REFRESHTIME_SCALE = 100.0 / RSET_240K

// The refresh time in milliseconds for which NORMALIZED_UVA_SENSITIVITY
// is applicable to a step count
const NORMALIZED_REFRESHTIME = 100.0

// The UVA sensitivity in mW/(m*m)/step which is applicable to a step count
// normalized to the NORMALIZED_REFRESHTIME, from datasheet for RSET=240 kOhm
// and IT_FACTOR=1
const NORMALIZED_UVA_SENSITIVITY = 50.0

// Config register

// Possible values for shutdown
const (
	CONFIG_SD_DISABLE = 0x00
	CONFIG_SD_ENABLE  = 0x01
)

// Enable / disable
const (
	CONFIG_DEFAULTS = 0x02
	CONFIG_ENABLE   = CONFIG_SD_DISABLE | CONFIG_DEFAULTS
	CONFIG_DISABLE  = CONFIG_SD_ENABLE | CONFIG_DEFAULTS
)
