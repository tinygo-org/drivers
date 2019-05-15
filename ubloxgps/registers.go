package ubloxgps

// Constants/addresses used for I2C.

// The I2C address which this device listens to.
const (
	Address = 0x42
)

const (
	BYTES_AVAIL_REG = 0xfd
	DATA_STREAM_REG = 0xff
)

const (
	bufferSize = 32
)
