package adxl345

const AddressLow = 0x53
const AddressHigh = 0x1D

const (
	// Data rate
	RATE_3200HZ Rate = 0x0F // 3200 Hz
	RATE_1600HZ Rate = 0x0E // 1600 Hz
	RATE_800HZ  Rate = 0x0D // 800 Hz
	RATE_400HZ  Rate = 0x0C // 400 Hz
	RATE_200HZ  Rate = 0x0B // 200 Hz
	RATE_100HZ  Rate = 0x0A // 100 Hz
	RATE_50HZ   Rate = 0x09 // 50 Hz
	RATE_25HZ   Rate = 0x08 // 25 Hz
	RATE_12_5HZ Rate = 0x07 // 12.5 Hz
	RATE_6_25HZ Rate = 0x06 // 6.25 Hz
	RATE_3_13HZ Rate = 0x05 // 3.13 Hz
	RATE_1_56HZ Rate = 0x04 // 1.56 Hz
	RATE_0_78HZ Rate = 0x03 // 0.78 Hz
	RATE_0_39HZ Rate = 0x02 // 0.39 Hz
	RATE_0_20HZ Rate = 0x01 // 0.20 Hz
	RATE_0_10HZ Rate = 0x00 // 0.10 Hz

	// Data range
	RANGE_2G  Range = 0x00 // +-2 g
	RANGE_4G  Range = 0x01 // +-4 g
	RANGE_8G  Range = 0x02 // +-8 g
	RANGE_16G Range = 0x03 // +-16 g)

	REG_DEVID          = 0x00 // R,     11100101,   Device ID
	REG_THRESH_TAP     = 0x1D // R/W,   00000000,   Tap threshold
	REG_OFSX           = 0x1E // R/W,   00000000,   X-axis offset
	REG_OFSY           = 0x1F // R/W,   00000000,   Y-axis offset
	REG_OFSZ           = 0x20 // R/W,   00000000,   Z-axis offset
	REG_DUR            = 0x21 // R/W,   00000000,   Tap duration
	REG_LATENT         = 0x22 // R/W,   00000000,   Tap latency
	REG_WINDOW         = 0x23 // R/W,   00000000,   Tap window
	REG_THRESH_ACT     = 0x24 // R/W,   00000000,   Activity threshold
	REG_THRESH_INACT   = 0x25 // R/W,   00000000,   Inactivity threshold
	REG_TIME_INACT     = 0x26 // R/W,   00000000,   Inactivity time
	REG_ACT_INACT_CTL  = 0x27 // R/W,   00000000,   Axis enable control for activity and inactiv ity detection
	REG_THRESH_FF      = 0x28 // R/W,   00000000,   Free-fall threshold
	REG_TIME_FF        = 0x29 // R/W,   00000000,   Free-fall time
	REG_TAP_AXES       = 0x2A // R/W,   00000000,   Axis control for single tap/double tap
	REG_ACT_TAP_STATUS = 0x2B // R,     00000000,   Source of single tap/double tap
	REG_BW_RATE        = 0x2C // R/W,   00001010,   Data rate and power mode control
	REG_POWER_CTL      = 0x2D // R/W,   00000000,   Power-saving features control
	REG_INT_ENABLE     = 0x2E // R/W,   00000000,   Interrupt enable control
	REG_INT_MAP        = 0x2F // R/W,   00000000,   Interrupt mapping control
	REG_INT_SOUCE      = 0x30 // R,     00000010,   Source of interrupts
	REG_DATA_FORMAT    = 0x31 // R/W,   00000000,   Data format control
	REG_DATAX0         = 0x32 // R,     00000000,   X-Axis Data 0
	REG_DATAX1         = 0x33 // R,     00000000,   X-Axis Data 1
	REG_DATAY0         = 0x34 // R,     00000000,   Y-Axis Data 0
	REG_DATAY1         = 0x35 // R,     00000000,   Y-Axis Data 1
	REG_DATAZ0         = 0x36 // R,     00000000,   Z-Axis Data 0
	REG_DATAZ1         = 0x37 // R,     00000000,   Z-Axis Data 1
	REG_FIFO_CTL       = 0x38 // R/W,   00000000,   FIFO control
	REG_FIFO_STATUS    = 0x39 // R,     00000000,   FIFO status
)
