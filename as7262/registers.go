package as7262

const (
	// DefaultAddress Address is default I2C address of AS7262
	DefaultAddress = 0x49

	// StatusReg
	StatusReg = 0x00
	// WriteReg
	WriteReg = 0x01
	// ReadReg
	ReadReg = 0x02

	// ControlReg
	ControlReg = 0x04
	// IntegrationTimeReg
	IntegrationTimeReg = 0x05
	// TempRegister
	TempRegister = 0x06
	// LedRegister
	LedRegister = 0x07

	/*
		Sensor Raw Data Registers
	*/
	// VHighRawReg Channel V High Data Byte
	VHighRawReg = 0x08
	// VLowRawReg Channel V Low Data Byte
	VLowRawReg = 0x09
	// BHighRawReg Channel B High Data Byte
	BHighRawReg = 0x0A
	// BLowRawReg Channel B Low Data Byte
	BLowRawReg = 0x0B
	// GHighRawReg Channel G High Data Byte
	GHighRawReg = 0x0C
	// GLowRawReg Channel G Low Data Byte
	GLowRawReg = 0x0D
	// YHighRawReg Channel Y High Data Byte
	YHighRawReg = 0x0E
	// YLowRawReg Channel Y Low Data Byte
	YLowRawReg = 0x0F
	// OHighRawReg Channel O High Data Byte
	OHighRawReg = 0x10
	// OLowRawReg Channel O Low Data Byte
	OLowRawReg = 0x11
	// RHighRawReg Channel R High Data Byte
	RHighRawReg = 0x12
	// RLowRawReg Channel R Low Data Byte
	RLowRawReg = 0x13

	/*
		Sensor Calibrated Data Registers
	*/
	// VCalReg  address for Channel V Calibrated Data
	VCalReg = 0x14
	// BCalReg  address for Channel B Calibrated Data
	BCalReg = 0x18
	// GCalReg  address for Channel G Calibrated Data
	GCalReg = 0x1C
	// YCalReg  address for Channel Y Calibrated Data
	YCalReg = 0x20
	// OCalReg  address for Channel O Calibrated Data
	OCalReg = 0x24
	// RCalReg  address for Channel R Calibrated Data
	RCalReg = 0x28
)
