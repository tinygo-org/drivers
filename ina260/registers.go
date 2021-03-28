package ina260

// The default I2C address for this device.
//
// The actual address is configurable by connecting address pins.
const Address = 0x40

// Registers
const (
	REG_CONFIG     = 0x00
	REG_CURRENT    = 0x01
	REG_BUSVOLTAGE = 0x02
	REG_POWER      = 0x03
	REG_MASKENABLE = 0x06
	REG_ALERTLIMIT = 0x07
	REG_MANF_ID    = 0xFE
	REG_DIE_ID     = 0xFF
)

// Well-Known Values
const (
	MANF_ID        = 0x5449 // TI
	DEVICE_ID      = 0x2270 // 227h
	DEVICE_ID_MASK = 0xFFF0

	AVGMODE_1    = 0
	AVGMODE_4    = 1
	AVGMODE_16   = 2
	AVGMODE_64   = 3
	AVGMODE_128  = 4
	AVGMODE_256  = 5
	AVGMODE_512  = 6
	AVGMODE_1024 = 7

	CONVTIME_140USEC  = 0
	CONVTIME_204USEC  = 1
	CONVTIME_332USEC  = 2
	CONVTIME_588USEC  = 3
	CONVTIME_1100USEC = 4 // 1.1 ms
	CONVTIME_2116USEC = 5 // 2.1 ms
	CONVTIME_4156USEC = 6 // 4.2 ms
	CONVTIME_8244USEC = 7 // 8.2 ms

	MODE_CONTINUOUS = 0x4
	MODE_TRIGGERED  = 0x0
	MODE_VOLTAGE    = 0x2
	MODE_NO_VOLTAGE = 0x0
	MODE_CURRENT    = 0x1
	MODE_NO_CURRENT = 0x0
)
