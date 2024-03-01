package mpu6050

import (
	"errors"
	"time"

	"tinygo.org/x/drivers"
)

var (
	ErrInvalidGyroFSR  = errors.New("mpu6050:invalid gyro range")
	ErrInvalidAccelFSR = errors.New("mpu6050:invalid accel range")
	ErrInvalidClkSel   = errors.New("mpu6050:invalid clock source")
	ErrInvalidDLPF     = errors.New("mpu6050:invalid DLPF")
)

// DeviceStored reads all IMU data in a single i2c transaction
// to reduce bus usage. It is a better alternative to a simple Device
// if getting gyroscopes+acceleration measurements together many times in
// a second.
type DeviceStored struct {
	Device
	gyroRange  int16
	accelRange int16
	// buf is buffer for small reads/writes
	buf [2]byte
	// rreg stores read register
	rreg [1]byte
	// data contains IMU readings
	data [14]byte
}

// NewStored creates a versatile instance of a mpu6050 device handle.
// Use Frequencies above 40kHz to avoid timeout issues.
func NewStored(bus drivers.I2C) DeviceStored {
	return DeviceStored{Device: New(bus), accelRange: 2, gyroRange: 250}
}

type Config struct {
	// Set to one of ACCEL_RANGE_X register values
	AccRange byte
	// Set to one of GYRO_RANGE_x register values
	GyroRange byte
	// SetSampleRate sets the sample rate for the FIFO,
	// register ouput and DMP. The sample rate is determined
	// by:
	//		SR = Gyroscope Output Rate / (1 + srDiv)
	SampleRatio byte
	ClkSel      byte
}

// Init configures the necessary registers for using the
// MPU6050. It sets the range of both the accelerometer
// and the gyroscope, the sample rate, the clock source
// and wakes up the peripheral.
func (d *DeviceStored) Configure(data Config) (err error) {
	// d.Reset()

	d.Device.Configure()
	// setClockSource
	if err = d.SetClockSource(data.ClkSel); err != nil {
		return err
	}
	// setSampleRate
	if err = d.SetSampleRate(data.SampleRatio); err != nil {
		return err
	}
	// setFullScaleGyroRange
	if err = d.SetGyroRange(data.GyroRange); err != nil {
		return err
	}
	// setFullScaleAccelRange
	if err = d.SetAccelRange(data.AccRange); err != nil {
		return err
	}
	// setSleep
	return d.SetSleep(false) // wake MPU6050 up
}

// Acceleration returns last read acceleration in µg (micro-gravity).
// When one of the axes is pointing straight to Earth
// and the sensor is not moving the returned value will be around 1000000 or
// -1000000.
func (d *DeviceStored) Acceleration() (ax, ay, az int32) {
	const accelOffset = 0
	return int32(int16((uint16(d.data[accelOffset])<<8)|uint16(d.data[accelOffset+1]))) * 15625 / 512 * int32(d.accelRange),
		int32(int16((uint16(d.data[accelOffset+2])<<8)|uint16(d.data[accelOffset+3]))) * 15625 / 512 * int32(d.accelRange),
		int32(int16((uint16(d.data[accelOffset+4])<<8)|uint16(d.data[accelOffset+5]))) * 15625 / 512 * int32(d.accelRange)
}

// Rotations reads the current rotation from the device and returns it in
// µ°/s (micro-degrees/sec). This means that if you were to do a complete
// rotation along one axis and while doing so integrate all values over time,
// you would get a value close to 360000000.
func (d *DeviceStored) Rotation() (gx, gy, gz int32) {
	const angvelOffset = 8
	return int32(int16((uint16(d.data[angvelOffset])<<8)|uint16(d.data[angvelOffset+1]))) * 15625 / 2048 * int32(d.gyroRange) * 4,
		int32(int16((uint16(d.data[angvelOffset+2])<<8)|uint16(d.data[angvelOffset+3]))) * 15625 / 2048 * int32(d.gyroRange) * 4,
		int32(int16((uint16(d.data[angvelOffset+4])<<8)|uint16(d.data[angvelOffset+5]))) * 15625 / 2048 * int32(d.gyroRange) * 4
}

// Temperature returns the temperature of the device in centigrade.
func (d *DeviceStored) Temperature() (Celsius int16) {
	const tempOffset = 6
	return ((int16(d.data[tempOffset])<<8)|int16(d.data[tempOffset+1]))/340 + 37 // float64(temp/340) + 36.53
}

// Get reads IMU data and stores it inside DeviceStored. The data can then be accessed through Rotation and Acceleration
// methods.
func (d *DeviceStored) Get() error {
	d.rreg[0] = ACCEL_XOUT_H
	return d.bus.Tx(d.Address, d.rreg[:1], d.data[:14])
}

// SetSampleRate sets the sample rate for the FIFO,
// register ouput and DMP. The sample rate is determined
// by:
//		SR = Gyroscope Output Rate / (1 + srDiv)
//
// The Gyroscope Output Rate is 8kHz when the DLPF is
// disabled and 1kHz otherwise. The maximum sample rate
// for the accelerometer is 1kHz, if a higher sample rate
// is chosen, the same accelerometer sample will be output.
func (d *DeviceStored) SetSampleRate(srDiv byte) (err error) {
	return d.write(SMPLRT_DIV, srDiv)
}

// SetClockSource configures the source of the clock
// for the peripheral. When the MPU6050 starts up it uses it's own
// clock until configured otherwise.
//
// 0: Internal 8MHz Oscillator
// 1: PLL with X axis gyroscope reference
// 2: PLL with Y axis gyroscope reference
// 3: PLL with Z axis gyroscope reference
// 4: PLL with external 32.768kHz reference
// 5: PLL with external 19.2MHz reference
// 6: reserved
// 7: Stops the clock and keeps the timing generator in reset
func (d *DeviceStored) SetClockSource(clkSel byte) error {
	if clkSel == 6 || clkSel > 7 {
		return ErrInvalidClkSel
	}
	regdata, err := d.read(PWR_MGMT_1)
	if err != nil {
		return err
	}
	regdata = (regdata &^ CLK_SEL_Msk) | clkSel // Write CLKSEL field
	return d.write(PWR_MGMT_1, regdata)
}

// SetGyroRange configures the full scale range of the gyroscope.
// It has four possible values +- 250°/s, 500°/s, 1000°/s, 2000°/s.
// The function takes values of gyroRange from 0-3 where 0 means the
// lowest FSR (250°/s) and 3 is the highest FSR (2000°/s).
func (d *DeviceStored) SetGyroRange(gyroRange byte) (err error) {
	switch gyroRange {
	case GYRO_RANGE_250:
		d.gyroRange = 250
	case GYRO_RANGE_500:
		d.gyroRange = 500
	case GYRO_RANGE_1000:
		d.gyroRange = 1000
	case GYRO_RANGE_2000:
		d.gyroRange = 2000
	default:
		return ErrInvalidGyroFSR
	}
	// setFullScaleGyroRange
	regdata, err := d.read(GYRO_CONFIG)
	if err != nil {
		return err
	}

	regdata = (regdata &^ G_FS_SEL) | (gyroRange << GFS_Pos) // Write FS_SEL field
	return d.write(GYRO_CONFIG, regdata)
}

// SetAccelRange configures the full scale range of the accelerometer.
// It has four possible values +- 2g, 4g, 8g, 16g.
// The function takes values of accRange from 0-3 where 0 means the
// lowest FSR (2g) and 3 is the highest FSR (16g)
func (d *DeviceStored) SetAccelRange(accRange byte) (err error) {
	switch accRange {
	case ACCEL_RANGE_2:
		d.accelRange = 2
	case ACCEL_RANGE_4:
		d.accelRange = 4
	case ACCEL_RANGE_8:
		d.accelRange = 8
	case ACCEL_RANGE_16:
		d.accelRange = 16
	default:
		return ErrInvalidAccelFSR
	}
	regdata, err := d.read(ACCEL_CONFIG)
	if err != nil {
		return err
	}
	regdata = (regdata &^ AFS_SEL) | (accRange << AFS_Pos) // Write only FS_SEL field
	return d.write(ACCEL_CONFIG, regdata)
}

// Set filter bandwidth. Has side effect of reducing the sample rate
// with higher low pass filter values.
func (d *DeviceStored) SetDigitalLowPass(dlpf byte) error {
	if dlpf >= 7 {
		return ErrInvalidDLPF
	}
	regdata, err := d.read(CONFIG)
	if err != nil {
		return err
	}
	return d.write(CONFIG, (regdata&^DLPF_Msk)|dlpf)
}

// SetSleep sets the sleep bit on the power managment 1 field.
// When the recieved bool is true, it sets the bit to 1 thus putting
// the peripheral in sleep mode.
// When false is recieved the bit is set to 0 and the peripheral wakes
// up.
func (d *DeviceStored) SetSleep(sleep bool) (err error) {
	regdata, err := d.read(PWR_MGMT_1)
	if err != nil {
		return err
	}
	regdata &^= SLEEP_Msk
	if sleep {
		regdata |= (1 << SLEEP_Pos) // Set CLK_SEL bits only
	}
	return d.write(PWR_MGMT_1, regdata)
}

func (d *DeviceStored) Reset() error {
	err := d.write(PWR_MGMT_1, RESET_Byte)
	time.Sleep(100 * time.Millisecond)
	return err
}

func DefaultConfig() Config {
	return Config{AccRange: ACCEL_RANGE_2, GyroRange: GYRO_RANGE_250}
}
