// Package qmi8658c provides a driver for the QMI8658C accelerometer and gyroscope
// made by QST Solutions.
//
// Datasheet:
// https://www.qstcorp.com/upload/pdf/202202/%EF%BC%88%E5%B7%B2%E4%BC%A0%EF%BC%89QMI8658C%20datasheet%20rev%200.9.pdf
package qmi8656c

import "tinygo.org/x/drivers"

// Device wraps the I2C connection to the QMIC8658 sensor
type Device struct {
	bus        drivers.I2C
	Address    uint16
	AccLsbDiv  uint16
	GyroLsbDiv uint16
}

type Config struct {
	// SPI Config
	SPIMode    byte // One of SPI_X_WIRE
	SPIEndian  byte // One of SPI_XXX_ENDIAN
	SPIAutoInc byte // One of SPI_NOT_AUTO_INC or SPI_AUTO_INC
	// Accelerometer
	AccEnable  byte // One of ACC_ENABLE or ACC_DISABLE
	AccScale   byte // One of ACC_XG
	AccRate    byte // One of ACC_XX_YYHZ
	AccLowPass byte // One of ACC_LOW_PASS_X
	// Gyro
	GyroEnable  byte // One of GYRO_X_ENABLE or GYRO_DISABLE
	GyroScale   byte // One of GYRO_XDPS
	GyroRate    byte // One of GYRO_X_YHZ
	GyroLowPass byte // One of GYRO_LOW_PASS_X
}

// Create a new device with the I2C passed, correct address and nil values for
// AccLsbDiv and GyroLsbDiv, which will be corrected based on the config.
func New(bus drivers.I2C) Device {
	return Device{
		bus,
		Address,
		1,
		1,
	}
}

// Check if the device is connected by calling WHO_AM_I and checking the
// default identifier.
func (d *Device) Connected() bool {
	data := []byte{0}
	d.ReadRegister(WHO_AM_I, data)
	return data[0] == IDENTIFIER
}

// Create a basic default configuration that works with the "WaveShare RP2040
// Round LCD 1.28in".
func DefaultConfig() (cfg Config) {
	return Config{
		SPIMode:     SPI_4_WIRE,
		SPIEndian:   SPI_BIG_ENDIAN,
		SPIAutoInc:  SPI_AUTO_INC,
		AccEnable:   ACC_ENABLE,
		AccScale:    ACC_8G,
		AccRate:     ACC_NORMAL_1000HZ,
		AccLowPass:  ACC_LOW_PASS_2_62,
		GyroEnable:  GYRO_FULL_ENABLE,
		GyroScale:   GYRO_512DPS,
		GyroRate:    GYRO_1000HZ,
		GyroLowPass: GYRO_LOW_PASS_2_62,
	}
}

// Check if the user has defined a desired configuration, if not uses the
// DefaultConfig, then defines the AccLsbDiv and GyroLsbDiv based on the
// configurations and, finally, send the commands and configure the IMU.
func (d *Device) Configure(cfg Config) {
	if cfg == (Config{}) {
		cfg = DefaultConfig()
	}
	var val uint16
	// Setting accelerometer LSB
	switch cfg.AccScale {
	case ACC_2G:
		d.AccLsbDiv = 1 << 14
	case ACC_4G:
		d.AccLsbDiv = 1 << 13
	case ACC_8G:
		d.AccLsbDiv = 1 << 12
	case ACC_16G:
		d.AccLsbDiv = 1 << 11
	default:
		d.AccLsbDiv = 1 << 12
	}
	// Setting gyro LSB
	switch cfg.GyroScale {
	case GYRO_16DPS:
		d.GyroLsbDiv = 2048
	case GYRO_32DPS:
		d.GyroLsbDiv = 1024
	case GYRO_64DPS:
		d.GyroLsbDiv = 512
	case GYRO_128DPS:
		d.GyroLsbDiv = 256
	case GYRO_256DPS:
		d.GyroLsbDiv = 128
	case GYRO_512DPS:
		d.GyroLsbDiv = 64
	case GYRO_1024DPS:
		d.GyroLsbDiv = 32
	case GYRO_2048DPS:
		d.GyroLsbDiv = 16
	default:
		d.GyroLsbDiv = 64
	}
	// SPI Modes
	val = uint16((cfg.SPIMode | cfg.SPIEndian | cfg.SPIAutoInc))
	d.WriteRegister(CTRL1, val)
	// Accelerometer config
	val = uint16(cfg.AccScale | cfg.AccRate)
	d.WriteRegister(CTRL2, val)
	// Gyro config
	val = uint16(cfg.GyroScale | cfg.GyroRate)
	d.WriteRegister(CTRL3, val)
	// Sensor DSP config
	val = uint16(cfg.GyroLowPass | cfg.AccLowPass)
	d.WriteRegister(CTRL5, val)
	// Sensors config
	val = uint16(cfg.GyroEnable | cfg.AccEnable)
	d.WriteRegister(CTRL7, val)
}

// Read the acceleration from the sensor, the values returned are in mg
// (milli gravity), which means that 1000 = 1g.
func (d *Device) ReadAcceleration() (x int32, y int32, z int32) {
	data := make([]byte, 6)
	raw := make([]int32, 3)
	d.ReadRegister(ACC_XOUT_L, data)
	for i := range raw {
		raw[i] = int32(uint16(data[(2*i+1)])<<8 | uint16(data[i]))
		if raw[i] >= 32767 {
			raw[i] = raw[i] - 65535
		}
	}
	x = -raw[0] * 1000 / int32(d.AccLsbDiv)
	y = -raw[1] * 1000 / int32(d.AccLsbDiv)
	z = -raw[2] * 1000 / int32(d.AccLsbDiv)
	return x, y, z
}

// Read the rotation from the sensor, the values returned are in mdeg/sec
// (milli degress/second), which means that a full rotation is 360000.
func (d *Device) ReadRotation() (x int32, y int32, z int32) {
	data := make([]byte, 6)
	raw := make([]int32, 3)
	d.ReadRegister(GYRO_XOUT_L, data)
	for i := range raw {
		raw[i] = int32(uint16(data[(2*i+1)])<<8 | uint16(data[i]))
		if raw[i] >= 32767 {
			raw[i] = raw[i] - 65535
		}
	}
	x = raw[0] * 1000 / int32(d.GyroLsbDiv)
	y = raw[1] * 1000 / int32(d.GyroLsbDiv)
	z = raw[2] * 1000 / int32(d.GyroLsbDiv)
	return x, y, z
}

// Read the temperature from the sensor, the values returned are in
// millidegrees Celsius.
func (d *Device) ReadTemperature() (int32, error) {
	data := make([]byte, 2)
	err := d.ReadRegister(TEMP_OUT_L, data)
	if err != nil {
		return 0, err
	}
	raw := uint16(data[1])<<8 | uint16(data[0])
	t := int32(raw) * 1000 / 256
	return t, err
}

// Convenience method to read the register and avoid repetition.
func (d *Device) ReadRegister(reg uint8, buf []byte) error {
	return d.bus.ReadRegister(uint8(d.Address), reg, buf)
}

// Convenience method to write the register and avoid repetition.
func (d *Device) WriteRegister(reg uint8, v uint16) error {
	data := []byte{byte(v)}
	err := d.bus.WriteRegister(uint8(d.Address), reg, data)
	return err
}
