package bmi270

// AccelRange is the full scale range of the accelerometer.
type AccelRange uint8

// The zero value selects AccelRange8G.
const (
	AccelRange2G AccelRange = iota + 1
	AccelRange4G
	AccelRange8G
	AccelRange16G
)

// GyroRange is the full scale range of the gyroscope.
type GyroRange uint8

// The zero value selects GyroRange2000.
const (
	GyroRange2000 GyroRange = iota + 1
	GyroRange1000
	GyroRange500
	GyroRange250
	GyroRange125
)

// AccelRate is the output data rate of the accelerometer.
// The zero value selects AccelRate100Hz.
type AccelRate uint8

// Values match acc_odr, see datasheet section 5.2.41.
const (
	AccelRate12_5Hz AccelRate = iota + 0x05
	AccelRate25Hz
	AccelRate50Hz
	AccelRate100Hz
	AccelRate200Hz
	AccelRate400Hz
	AccelRate800Hz
	AccelRate1600Hz
)

// GyroRate is the output data rate of the gyroscope.
// The zero value selects GyroRate200Hz.
type GyroRate uint8

// Values match gyr_odr, see datasheet section 5.2.43.
const (
	GyroRate25Hz GyroRate = iota + 0x06
	GyroRate50Hz
	GyroRate100Hz
	GyroRate200Hz
	GyroRate400Hz
	GyroRate800Hz
	GyroRate1600Hz
	GyroRate3200Hz
)

// Config holds the sensor settings. The zero value uses the reset defaults.
type Config struct {
	AccelRange AccelRange
	AccelRate  AccelRate
	GyroRange  GyroRange
	GyroRate   GyroRate
}

// registers returns the ACC_RANGE, ACC_CONF, GYR_RANGE and GYR_CONF values.
func (cfg Config) registers() (accelRange, accelConf, gyroRange, gyroConf uint8, err error) {
	if cfg.AccelRange == 0 {
		cfg.AccelRange = AccelRange8G
	}
	if cfg.AccelRate == 0 {
		cfg.AccelRate = AccelRate100Hz
	}
	if cfg.GyroRange == 0 {
		cfg.GyroRange = GyroRange2000
	}
	if cfg.GyroRate == 0 {
		cfg.GyroRate = GyroRate200Hz
	}
	if cfg.AccelRange > AccelRange16G ||
		cfg.AccelRate < AccelRate12_5Hz || cfg.AccelRate > AccelRate1600Hz ||
		cfg.GyroRange > GyroRange125 ||
		cfg.GyroRate < GyroRate25Hz || cfg.GyroRate > GyroRate3200Hz {
		return 0, 0, 0, 0, ErrInvalidConfig
	}
	return uint8(cfg.AccelRange - 1), accelConfPerf | uint8(cfg.AccelRate),
		uint8(cfg.GyroRange - 1), gyroConfPerf | uint8(cfg.GyroRate), nil
}
