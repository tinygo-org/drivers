package mpu6050

import (
	"encoding/binary"
	"fmt"
	"math"

	"tinygo.org/x/drivers"
)

const DefaultAddress = 0x68

type Config struct {
	// Use ACCEL_RANGE_2 through ACCEL_RANGE_16.
	AccRange byte
	// Use GYRO_RANGE_250 through GYRO_RANGE_2000
	GyroRange   byte
	sampleRatio byte // TODO(soypat): expose these as configurable.
	clkSel      byte
}

// Device contains MPU board abstraction for usage
type Device struct {
	conn    drivers.I2C
	aRange  int32 //Gyroscope FSR acording to SetAccelRange input
	gRange  int32 //Gyroscope FSR acording to SetGyroRange input
	data    [14]byte
	address byte
}

// New instantiates and initializes a MPU6050 struct without writing/reading
// i2c bus. Typical I2C MPU6050 address is 0x68.
func New(bus drivers.I2C, addr uint16) *Device {
	p := &Device{}
	p.address = uint8(addr)
	p.conn = bus
	return p
}

// Init configures the necessary registers for using the
// MPU6050. It sets the range of both the accelerometer
// and the gyroscope, the sample rate, the clock source
// and wakes up the peripheral.
func (p *Device) Configure(data Config) (err error) {

	// setSleep
	if err = p.SetSleep(false); err != nil {
		return fmt.Errorf("set sleep: %w", err)
	}
	// setClockSource
	if err = p.SetClockSource(data.clkSel); err != nil {
		return fmt.Errorf("set clksrc: %w", err)
	}
	// setSampleRate
	if err = p.SetSampleRate(data.sampleRatio); err != nil {
		return fmt.Errorf("set sample: %w", err)
	}
	// setFullScaleGyroRange
	if err = p.SetRangeGyro(data.GyroRange); err != nil {
		return fmt.Errorf("set gyro: %w", err)
	}
	// setFullScaleAccelRange
	if err = p.SetRangeAccel(data.AccRange); err != nil {
		return fmt.Errorf("set accelrange: %w", err)
	}
	return nil
}

// Update fetches the latest data from the MPU6050
func (p *Device) Update() (err error) {
	if err = p.read(_ACCEL_XOUT_H, p.data[:]); err != nil {
		return err
	}
	return nil
}

// Get treats the MPU6050 like an ADC and returns the raw reading of the channel.
// Channels 0-2 are acceleration, 3-5 are gyroscope and 6 is the temperature.
func (d *Device) RawReading(channel int) int16 {
	if channel > 6 {
		// Bad value.
		return 0
	}
	return convertWord(d.data[channel*2:])
}

func convertWord(buf []byte) int16 {
	return int16(binary.BigEndian.Uint16(buf))
}

func (d *Device) ChannelLen() int { return 7 }

// Acceleration returns last read acceleration in µg (micro-gravity).
// When one of the axes is pointing straight to Earth
// and the sensor is not moving the returned value will be around 1000000 or
// -1000000.
func (d *Device) Acceleration() (ax, ay, az int32) {
	const accelOffset = 0
	ax = int32(convertWord(d.data[accelOffset+0:])) * 15625 / 512 * d.aRange
	ay = int32(convertWord(d.data[accelOffset+2:])) * 15625 / 512 * d.aRange
	az = int32(convertWord(d.data[accelOffset+4:])) * 15625 / 512 * d.aRange
	return ax, ay, az
}

// Rotations reads the current rotation from the device and returns it in
// µ°/rad (micro-radians/sec). This means that if you were to do a complete
// rotation along one axis and while doing so integrate all values over time,
// you would get a value close to 6.3 radians (360 degrees).
func (d *Device) AngularVelocity() (gx, gy, gz int32) {
	const a = math.MaxInt16*8192 > math.MaxInt32
	const angvelOffset = 8
	gx = int32(convertWord(d.data[angvelOffset+0:])) * 4363 / 8192 * d.gRange
	gy = int32(convertWord(d.data[angvelOffset+2:])) * 4363 / 8192 * d.gRange
	gz = int32(convertWord(d.data[angvelOffset+4:])) * 4363 / 8192 * d.gRange
	return gx, gy, gz
}

// Temperature returns the temperature of the device in milli-centigrade.
func (d *Device) Temperature() (Celsius int32) {
	const tempOffset = 6
	return 1506*int32(convertWord(d.data[tempOffset:]))/512 + 37*1000
}

// SetSampleRate sets the sample rate for the FIFO,
// register ouput and DMP. The sample rate is determined
// by:
//
//	SR = Gyroscope Output Rate / (1 + srDiv)
//
// The Gyroscope Output Rate is 8kHz when the DLPF is
// disabled and 1kHz otherwise. The maximum sample rate
// for the accelerometer is 1kHz, if a higher sample rate
// is chosen, the same accelerometer sample will be output.
func (p *Device) SetSampleRate(srDiv byte) (err error) {
	// setSampleRate
	var sr [1]byte
	sr[0] = srDiv
	if err = p.write8(_SMPRT_DIV, sr[0]); err != nil {
		return err
	}
	return nil
}

// SetClockSource configures the source of the clock
// for the peripheral.
func (p *Device) SetClockSource(clkSel byte) (err error) {
	// setClockSource
	var pwrMgt [1]byte

	if err = p.read(_PWR_MGMT_1, pwrMgt[:]); err != nil {
		return err
	}
	pwrMgt[0] = (pwrMgt[0] & (^_CLK_SEL_MASK)) | clkSel // Escribo solo el campo de clk_sel
	if err = p.write8(_PWR_MGMT_1, pwrMgt[0]); err != nil {
		return err
	}
	return nil
}

// SetRangeGyro configures the full scale range of the gyroscope.
// It has four possible values +- 250°/s, 500°/s, 1000°/s, 2000°/s.
// The function takes values of gyroRange from 0-3 where 0 means the
// lowest FSR (250°/s) and 3 is the highest FSR (2000°/s).
func (p *Device) SetRangeGyro(gyroRange byte) (err error) {
	switch gyroRange {
	case GYRO_RANGE_250:
		p.gRange = 250
	case GYRO_RANGE_500:
		p.gRange = 500
	case GYRO_RANGE_1000:
		p.gRange = 1000
	case GYRO_RANGE_2000:
		p.gRange = 2000
	default:
		return fmt.Errorf("invalid gyroscope FSR input")
	}
	// setFullScaleGyroRange
	var gConfig [1]byte

	if err = p.read(_GYRO_CONFIG, gConfig[:]); err != nil {
		return err
	}
	gConfig[0] = (gConfig[0] & (^_G_FS_SEL)) | (gyroRange << _G_FS_SHIFT) // Escribo solo el campo de FS_sel

	if err = p.write8(_GYRO_CONFIG, gConfig[0]); err != nil {
		return err
	}
	return nil
}

// SetRangeAccel configures the full scale range of the accelerometer.
// It has four possible values +- 2g, 4g, 8g, 16g.
// The function takes values of accRange from 0-3 where 0 means the
// lowest FSR (2g) and 3 is the highest FSR (16g)
func (p *Device) SetRangeAccel(accRange byte) (err error) {
	switch accRange {
	case ACCEL_RANGE_2:
		p.aRange = 2
	case ACCEL_RANGE_4:
		p.aRange = 4
	case ACCEL_RANGE_8:
		p.aRange = 8
	case ACCEL_RANGE_16:
		p.aRange = 16
	default:
		return fmt.Errorf("invalid accelerometer FSR input")
	}
	// setFullScaleAccelRange
	var aConfig [1]byte

	if err = p.read(_ACCEL_CONFIG, aConfig[:]); err != nil {
		return err
	}
	aConfig[0] = (aConfig[0] & (^_AFS_SEL)) | (accRange << _AFS_SHIFT)

	if err = p.write8(_ACCEL_CONFIG, aConfig[0]); err != nil {
		return err
	}
	return nil
}

// SetSleep sets the sleep bit on the power managment 1 field.
// When the recieved bool is true, it sets the bit to 1 thus putting
// the peripheral in sleep mode.
// When false is recieved the bit is set to 0 and the peripheral wakes
// up.
func (p *Device) SetSleep(sleep bool) (err error) {
	// setSleepBit
	var pwrMgt [1]byte

	if err = p.read(_PWR_MGMT_1, pwrMgt[:]); err != nil {
		return err
	}
	if sleep {
		pwrMgt[0] = (pwrMgt[0] & (^_SLEEP_MASK)) | (1 << _SLEEP_SHIFT) // Overwrite only Sleep
	} else {
		pwrMgt[0] = (pwrMgt[0] & (^_SLEEP_MASK))
	}
	//Envio el registro modificado al periferico
	if err = p.write8(_PWR_MGMT_1, pwrMgt[0]); err != nil {
		return err
	}
	return nil
}

func DefaultConfig() Config {
	return Config{
		AccRange:    ACCEL_RANGE_16,
		GyroRange:   GYRO_RANGE_2000,
		sampleRatio: 0, // TODO add const values.
		clkSel:      0,
	}
}
