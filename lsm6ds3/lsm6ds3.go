// Package lsm6ds3 implements a driver for the LSM6DS3 a 6 axis Inertial
// Measurement Unit (IMU)
//
// Datasheet: https://www.st.com/resource/en/datasheet/lsm6ds3.pdf
//
package lsm6ds3 // import "tinygo.org/x/drivers/lsm6ds3"

import "tinygo.org/x/drivers"

type AccelRange uint8
type AccelSampleRate uint8
type AccelBandwidth uint8

type GyroRange uint8
type GyroSampleRate uint8

// Device wraps an I2C connection to a LSM6DS3 device.
type Device struct {
	bus             drivers.I2C
	Address         uint16
	accelRange      AccelRange
	accelSampleRate AccelSampleRate
	accelBandWidth  AccelBandwidth
	gyroRange       GyroRange
	gyroSampleRate  GyroSampleRate
	dataBufferSix   []uint8
	dataBufferTwo   []uint8
}

// Configuration for LSM6DS3 device.
type Configuration struct {
	AccelRange       AccelRange
	AccelSampleRate  AccelSampleRate
	AccelBandWidth   AccelBandwidth
	GyroRange        GyroRange
	GyroSampleRate   GyroSampleRate
	IsPedometer      bool
	ResetStepCounter bool
}

// New creates a new LSM6DS3 connection. The I2C bus must already be
// configured.
//
// This function only creates the Device object, it does not touch the device.
func New(bus drivers.I2C) Device {
	return Device{bus: bus, Address: Address}
}

// Configure sets up the device for communication.
func (d *Device) Configure(cfg Configuration) {
	if cfg.AccelRange != 0 {
		d.accelRange = cfg.AccelRange
	} else {
		d.accelRange = ACCEL_2G
	}

	if cfg.AccelSampleRate != 0 {
		d.accelSampleRate = cfg.AccelSampleRate
	} else {
		d.accelSampleRate = ACCEL_SR_104
	}

	if cfg.AccelBandWidth != 0 {
		d.accelBandWidth = cfg.AccelBandWidth
	} else {
		d.accelBandWidth = ACCEL_BW_100
	}

	if cfg.GyroRange != 0 {
		d.gyroRange = cfg.GyroRange
	} else {
		d.gyroRange = GYRO_2000DPS
	}

	if cfg.GyroSampleRate != 0 {
		d.gyroSampleRate = cfg.GyroSampleRate
	} else {
		d.gyroSampleRate = GYRO_SR_104
	}

	d.dataBufferSix = make([]uint8, 6)
	d.dataBufferTwo = make([]uint8, 2)

	if cfg.IsPedometer { // CONFIGURE AS PEDOMETER
		// Configure accelerometer: 2G + 26Hz
		d.bus.WriteRegister(uint8(d.Address), CTRL1_XL, []byte{uint8(ACCEL_2G) | uint8(ACCEL_SR_26)})

		// Configure Zen_G, Yen_G, Xen_G, reset steps
		if cfg.ResetStepCounter {
			d.bus.WriteRegister(uint8(d.Address), CTRL10_C, []byte{0x3E})
		} else {
			d.bus.WriteRegister(uint8(d.Address), CTRL10_C, []byte{0x3C})
		}

		// Enable pedometer
		d.bus.WriteRegister(uint8(d.Address), TAP_CFG, []byte{0x40})
	} else { // NORMAL USE
		// Configure accelerometer
		data := make([]uint8, 1)
		data[0] = uint8(d.accelRange) | uint8(d.accelSampleRate) | uint8(d.accelBandWidth)
		d.bus.WriteRegister(uint8(d.Address), CTRL1_XL, data)

		// Set ODR bit
		d.bus.ReadRegister(uint8(d.Address), CTRL4_C, data)
		data[0] = data[0] &^ BW_SCAL_ODR_ENABLED
		data[0] |= BW_SCAL_ODR_ENABLED
		d.bus.WriteRegister(uint8(d.Address), CTRL4_C, data)

		// Configure gyroscope
		data[0] = uint8(d.gyroRange) | uint8(d.gyroSampleRate)
		d.bus.WriteRegister(uint8(d.Address), CTRL2_G, data)
	}
}

// Connected returns whether a LSM6DS3 has been found.
// It does a "who am I" request and checks the response.
func (d *Device) Connected() bool {
	data := []byte{0}
	d.bus.ReadRegister(uint8(d.Address), WHO_AM_I, data)
	return data[0] == 0x69
}

// ReadAcceleration reads the current acceleration from the device and returns
// it in µg (micro-gravity). When one of the axes is pointing straight to Earth
// and the sensor is not moving the returned value will be around 1000000 or
// -1000000.
func (d *Device) ReadAcceleration() (x int32, y int32, z int32) {
	d.bus.ReadRegister(uint8(d.Address), OUTX_L_XL, d.dataBufferSix)
	// k comes from "Table 3. Mechanical characteristics" 3 of the datasheet * 1000
	k := int32(61) // 2G
	if d.accelRange == ACCEL_4G {
		k = 122
	} else if d.accelRange == ACCEL_8G {
		k = 244
	} else if d.accelRange == ACCEL_16G {
		k = 488
	}
	x = int32(int16((uint16(d.dataBufferSix[1])<<8)|uint16(d.dataBufferSix[0]))) * k
	y = int32(int16((uint16(d.dataBufferSix[3])<<8)|uint16(d.dataBufferSix[2]))) * k
	z = int32(int16((uint16(d.dataBufferSix[5])<<8)|uint16(d.dataBufferSix[4]))) * k
	return
}

// ReadRotation reads the current rotation from the device and returns it in
// µ°/s (micro-degrees/sec). This means that if you were to do a complete
// rotation along one axis and while doing so integrate all values over time,
// you would get a value close to 360000000.
func (d *Device) ReadRotation() (x int32, y int32, z int32) {
	d.bus.ReadRegister(uint8(d.Address), OUTX_L_G, d.dataBufferSix)
	// k comes from "Table 3. Mechanical characteristics" 3 of the datasheet * 1000
	k := int32(4375) // 125DPS
	if d.gyroRange == GYRO_250DPS {
		k = 8750
	} else if d.gyroRange == GYRO_500DPS {
		k = 17500
	} else if d.gyroRange == GYRO_1000DPS {
		k = 35000
	} else if d.gyroRange == GYRO_2000DPS {
		k = 70000
	}
	x = int32(int16((uint16(d.dataBufferSix[1])<<8)|uint16(d.dataBufferSix[0]))) * k
	y = int32(int16((uint16(d.dataBufferSix[3])<<8)|uint16(d.dataBufferSix[2]))) * k
	z = int32(int16((uint16(d.dataBufferSix[5])<<8)|uint16(d.dataBufferSix[4]))) * k
	return
}

// ReadTemperature returns the temperature in celsius milli degrees (°C/1000)
func (d *Device) ReadTemperature() (int32, error) {
	d.bus.ReadRegister(uint8(d.Address), OUT_TEMP_L, d.dataBufferTwo)

	// From "Table 5. Temperature sensor characteristics"
	// temp = value/16 + 25
	t := 25000 + (int32(int16((int16(d.dataBufferTwo[1])<<8)|int16(d.dataBufferTwo[0])))*125)/2
	return t, nil
}

// ReadSteps returns the steps of the pedometer
func (d *Device) ReadSteps() int32 {
	d.bus.ReadRegister(uint8(d.Address), STEP_COUNTER_L, d.dataBufferTwo)
	return int32(int16((uint16(d.dataBufferTwo[1]) << 8) | uint16(d.dataBufferTwo[0])))
}
