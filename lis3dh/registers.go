package lis3dh

// Constants/addresses used for I2C.

// The I2C addresses which this device listens to.
const (
	Address0 = 0x18 // SA0 is low
	Address1 = 0x19 // SA0 is high
)

// Registers. Names, addresses and comments copied from the datasheet.
const (
	WHO_AM_I      = 0x0F
	REG_STATUS1   = 0x07
	REG_OUTADC1_L = 0x08
	REG_OUTADC1_H = 0x09
	REG_OUTADC2_L = 0x0A
	REG_OUTADC2_H = 0x0B
	REG_OUTADC3_L = 0x0C
	REG_OUTADC3_H = 0x0D
	REG_INTCOUNT  = 0x0E
	REG_WHOAMI    = 0x0F
	REG_TEMPCFG   = 0x1F
	REG_CTRL1     = 0x20
	REG_CTRL2     = 0x21
	REG_CTRL3     = 0x22
	REG_CTRL4     = 0x23
	REG_CTRL5     = 0x24
	REG_CTRL6     = 0x25
	REG_REFERENCE = 0x26
	REG_STATUS2   = 0x27
	REG_OUT_X_L   = 0x28
	REG_OUT_X_H   = 0x29
	REG_OUT_Y_L   = 0x2A
	REG_OUT_Y_H   = 0x2B
	REG_OUT_Z_L   = 0x2C
	REG_OUT_Z_H   = 0x2D
	REG_FIFOCTRL  = 0x2E
	REG_FIFOSRC   = 0x2F
	REG_INT1CFG   = 0x30
	REG_INT1SRC   = 0x31
	REG_INT1THS   = 0x32
	REG_INT1DUR   = 0x33
	REG_CLICKCFG  = 0x38
	REG_CLICKSRC  = 0x39
	REG_CLICKTHS  = 0x3A
	REG_TIMELIMIT = 0x3B
	REG_TIMELATEN = 0x3C
	REG_TIMEWINDO = 0x3D
	REG_ACTTHS    = 0x3E
	REG_ACTDUR    = 0x3F
)

type Range uint8

const (
	RANGE_16_G Range = 3 // +/- 16g
	RANGE_8_G        = 2 // +/- 8g
	RANGE_4_G        = 1 // +/- 4g
	RANGE_2_G        = 0 // +/- 2g (default value)
)

type DataRate uint8

// Data rate constants.
const (
	DATARATE_400_HZ         DataRate = 7 //  400Hz
	DATARATE_200_HZ                  = 6 //  200Hz
	DATARATE_100_HZ                  = 5 //  100Hz
	DATARATE_50_HZ                   = 4 //   50Hz
	DATARATE_25_HZ                   = 3 //   25Hz
	DATARATE_10_HZ                   = 2 // 10 Hz
	DATARATE_1_HZ                    = 1 // 1 Hz
	DATARATE_POWERDOWN               = 0
	DATARATE_LOWPOWER_1K6HZ          = 8
	DATARATE_LOWPOWER_5KHZ           = 9
)
