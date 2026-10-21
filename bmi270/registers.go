package bmi270

// Register map, see datasheet section 5.2.
const (
	RegChipID         = 0x00
	RegErr            = 0x02
	RegStatus         = 0x03
	RegAccelData      = 0x0C
	RegGyroData       = 0x12
	RegSensorTime     = 0x18
	RegInternalStatus = 0x21
	RegTemperature    = 0x22
	RegFeatPage       = 0x2F
	RegFeatures       = 0x30
	RegAccelConf      = 0x40
	RegAccelRange     = 0x41
	RegGyroConf       = 0x42
	RegGyroRange      = 0x43
	RegInitCtrl       = 0x59
	RegInitAddr0      = 0x5B
	RegInitAddr1      = 0x5C
	RegInitData       = 0x5E
	RegPwrConf        = 0x7C
	RegPwrCtrl        = 0x7D
	RegCmd            = 0x7E
)

const (
	// ChipID is the value of RegChipID.
	ChipID = 0x24

	// Address is the I2C address with SDO connected to GND.
	Address = 0x68
	// AddressAlt is the I2C address with SDO connected to VDDIO.
	AddressAlt = 0x69
)

const (
	cmdSoftReset = 0xB6

	spiRead = 0x80

	statusAccelReady = 0x80
	statusGyroReady  = 0x40

	internalStatusMask   = 0x0F
	internalStatusInitOK = 0x01

	pwrCtrlGyro  = 0x02
	pwrCtrlAccel = 0x04
	pwrCtrlTemp  = 0x08

	pwrConfAdvPowerSave = 0x01
	pwrConfFIFOWakeup   = 0x02

	// Filter performance mode with normal bandwidth.
	accelConfPerf = 0xA0
	// Filter and noise performance mode with normal bandwidth.
	gyroConfPerf = 0xE0

	temperatureInvalid = -0x8000

	featGyroCAS = 0x0C
)
