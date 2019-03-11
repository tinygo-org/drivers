package mma8653

// Constants/addresses used for I2C.

// The I2C address which this device listens to.
const Address = 0x1D

// Registers. Names, addresses and comments copied from the datasheet.
const (
	STATUS       = 0x00 // Real time status
	OUT_X_MSB    = 0x01 // top 8 bits of 10-bit sample
	OUT_X_LSB    = 0x02 // bottom 2 bits of 10-bit sample
	OUT_Y_MSB    = 0x03 // top 8 bits of 10-bit sample
	OUT_Y_LSB    = 0x04 // bottom 2 bits of 10-bit sample
	OUT_Z_MSB    = 0x05 // top 8 bits of 10-bit sample
	OUT_Z_LSB    = 0x06 // bottom 2 bits of 10-bit sample
	SYSMOD       = 0x0B // Current System Mode
	INT_SOURCE   = 0x0C // Interrupt status
	WHO_AM_I     = 0x0D // Device ID (0x5A)
	XYZ_DATA_CFG = 0x0E // Dynamic Range Settings
	PL_STATUS    = 0x10 // Landscape/Portrait orientation status
	PL_CFG       = 0x11 // Landscape/Portrait configuration
	PL_COUNT     = 0x12 // Landscape/Portrait debouncer counter
	PL_BF_ZCOMP  = 0x13 // Back/Front, Z-Lock Trip threshold
	PL_THS_REG   = 0x14 // Portrait to Landscape Trip angle
	FF_MT_CFG    = 0x15 // Freefall/Motion functional block configuration
	FF_MT_SRC    = 0x16 // Freefall/Motion event source register
	FF_MT_THS    = 0x17 // Freefall/Motion threshold register
	FF_MT_COUNT  = 0x18 // Freefall/Motion debounce counter
	ASLP_COUNT   = 0x29 // Counter setting for Auto-SLEEP/WAKE
	CTRL_REG1    = 0x2A // Data Rates, ACTIVE Mode
	CTRL_REG2    = 0x2B // Sleep Enable, OS Modes, RST, ST
	CTRL_REG3    = 0x2C // Wake from Sleep, IPOL, PP_OD
	CTRL_REG4    = 0x2D // Interrupt enable register
	CTRL_REG5    = 0x2E // Interrupt pin (INT1/INT2) map
	OFF_X        = 0x2F // X-axis offset adjust
	OFF_Y        = 0x30 // Y-axis offset adjust
	OFF_Z        = 0x31 // Z-axis offset adjust
)

type DataRate uint8

// Data rate constants.
const (
	DataRate800Hz DataRate = iota // 800Hz,  1.25ms interval
	DataRate400Hz                 // 400Hz,  2.5ms  interval
	DataRate200Hz                 // 200Hz,  5ms    interval
	DataRate100Hz                 // 100Hz,  10ms   interval
	DataRate50Hz                  // 50Hz,   20ms   interval
	DataRate12Hz                  // 12.5Hz, 80ms   interval
	DataRate6Hz                   // 6.25Hz, 160ms  interval
	DataRate2Hz                   // 1.56Hz, 640ms  interval
)

type Sensitivity uint8

// Sensitivity constants.
const (
	Sensitivity2G Sensitivity = iota
	Sensitivity4G
	Sensitivity8G
)
