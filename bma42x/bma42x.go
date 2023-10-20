// Package bma42x provides a driver for the BMA421 and BMA425 accelerometer
// chips.
//
// Here is a reasonably good datasheet:
// https://datasheet.lcsc.com/lcsc/1912111437_Bosch-Sensortec-BMA425_C437656.pdf
//
// This driver was originally written for the PineTime, using the datasheet as a
// guide. There is an open source C driver provided by Bosch, but unfortunately
// it needs some small modifications to work with other chips (most importantly,
// the "config file").
// The InfiniTime and Wasp-OS drivers for this accelerometer have also been used
// to figure out some driver details (especially step counting).
package bma42x

import (
	_ "embed"
	"errors"
	"reflect"
	"time"
	"unsafe"

	"tinygo.org/x/drivers"
)

// Driver for BMA421 and BMA425:
// BMA421: https://files.pine64.org/doc/datasheet/pinetime/BST-BMA421-FL000.pdf
// BMA425: https://datasheet.lcsc.com/lcsc/1912111437_Bosch-Sensortec-BMA425_C437656.pdf

// This is the BMA421 firmware from the Wasp-OS project.
// It is identical to the so-called BMA423 firmware in InfiniTime, which I
// suspect to be actually a BMA421 firmware. I don't know where this firmware
// comes from or what the licensing status is.
// It has the FEATURES_IN command prepended, so that it can be written directly
// using I2C.Tx.
// Source: https://github.com/wasp-os/bma42x-upy/blob/master/BMA42X-Sensor-API/bma421.h
//
//go:embed bma421-config-waspos.bin
var bma421Firmware string

// Same as the BMA421 firmware, but for the BMA425.
// Source: https://github.com/wasp-os/bma42x-upy/blob/master/BMA42X-Sensor-API/bma425.h
//
//go:embed bma425-config-waspos.bin
var bma425Firmware string

var (
	errUnknownDevice     = errors.New("bma42x: unknown device")
	errUnsupportedDevice = errors.New("bma42x: device not part of config")
	errConfigMismatch    = errors.New("bma42x: config mismatch")
	errTimeout           = errors.New("bma42x: timeout")
	errInitFailed        = errors.New("bma42x: failed to initialize")
)

const Address = 0x18 // BMA421/BMA425 address

type DeviceType uint8

const (
	DeviceBMA421 DeviceType = 1 << iota
	DeviceBMA425

	AnyDevice            = DeviceBMA421 | DeviceBMA425
	noDevice  DeviceType = 0
)

// Features to enable while configuring the accelerometer.
type Features uint8

const (
	FeatureStepCounting = 1 << iota
)

type Config struct {
	// Which devices to support (OR the device types together as needed).
	Device DeviceType

	// Which features to enable. With Features == 0, only the accelerometer will
	// be enabled.
	Features Features
}

type Device struct {
	bus               drivers.I2C
	address           uint8
	accelData         [6]byte
	combinedTempSteps [5]uint8 // [0:3] steps, [4] temperature
	dataBuf           [2]byte
}

func NewI2C(i2c drivers.I2C, address uint8) *Device {
	return &Device{
		bus:     i2c,
		address: address,
	}
}

func (d *Device) Connected() bool {
	val, err := d.read1(_CHIP_ID)
	return err == nil && identifyChip(val) != noDevice
}

func (d *Device) Configure(config Config) error {
	if config.Device == 0 {
		config.Device = AnyDevice
	}

	// Check chip ID, to check the connection and to determine which BMA42x
	// device we're dealing with.
	chipID, err := d.read1(_CHIP_ID)
	if err != nil {
		return err
	}

	// Determine which firmware (config file?) we'll be using.
	// There is an extra check for the device before using the given firmware.
	// This check will typically be optimized away if the given device is not
	// configured, so that the firmware (which is 6kB in size!) won't be linked
	// into the binary.
	var firmware string
	switch identifyChip(chipID) {
	case DeviceBMA421:
		if config.Device&DeviceBMA421 == 0 {
			return errUnsupportedDevice
		}
		firmware = bma421Firmware
	case DeviceBMA425:
		if config.Device&DeviceBMA425 == 0 {
			return errUnsupportedDevice
		}
		firmware = bma425Firmware
	default:
		return errUnknownDevice
	}

	// Reset the chip, to be able to initialize it properly.
	// The datasheet says a delay is needed after a SoftReset, but it doesn't
	// say how long this delay should be. The bma423 driver however uses a 200ms
	// delay, so that's what we'll be using.
	err = d.write1(_CMD, cmdSoftReset)
	if err != nil {
		return err
	}
	time.Sleep(200 * time.Millisecond)

	// Disable power saving.
	err = d.write1(_PWR_CONF, 0x00)
	if err != nil {
		return err
	}
	time.Sleep(450 * time.Microsecond)

	// Start initialization (because the datasheet says so).
	err = d.write1(_INIT_CTRL, 0x00)
	if err != nil {
		return err
	}

	// Write "config file" (actually a firmware, I think) to the chip.
	// To do this, unsafely cast the string to a byte slice to avoid putting it
	// in RAM. This is safe in this case because Tx won't write to the 'w'
	// slice.
	err = d.bus.Tx(uint16(d.address), unsafeStringToSlice(firmware), nil)
	if err != nil {
		return err
	}

	// Read the config data back.
	// We don't do that, as it slows down configuration and it probably isn't
	// _really_ necessary with a reasonably stable I2C bus.
	if false {
		data := make([]byte, len(firmware)-1)
		err = d.readn(_FEATURES_IN, data)
		if err != nil {
			return err
		}
		for i, c := range data {
			if firmware[i+1] != c {
				return errConfigMismatch
			}
		}
	}

	// Enable sensors.
	err = d.write1(_INIT_CTRL, 0x01)
	if err != nil {
		return err
	}

	// Wait until the device is initialized.
	start := time.Now()
	status := uint8(0) // busy
	for status == 0 {
		status, err = d.read1(_INTERNAL_STATUS)
		if err != nil {
			return err // I2C bus error.
		}
		if status > 1 {
			// Expected either 0 ("not_init") or 1 ("init_ok").
			return errInitFailed
		}
		if time.Since(start) >= 150*time.Millisecond {
			// The datasheet says initialization should not take longer than
			return errTimeout
		}
		// Don't bother the chip all the time while it's initializing.
		time.Sleep(50 * time.Microsecond)
	}

	if config.Features&FeatureStepCounting != 0 {
		// Enable step counter.
		// TODO: support step counter parameters.
		var buf [71]byte
		buf[0] = _FEATURES_IN // prefix buf with the command
		data := buf[1:]
		err = d.readn(_FEATURES_IN, data)
		if err != nil {
			return err
		}
		data[0x3A+1] |= 0x10 // enable step counting by setting a magical bit
		err = d.bus.Tx(uint16(d.address), buf[:], nil)
		if err != nil {
			return err
		}
	}

	// Enable the accelerometer.
	err = d.write1(_PWR_CTRL, 0x04)
	if err != nil {
		return err
	}

	// Configure accelerometer for low power usage:
	//   acc_perf_mode=0   (power saving enabled)
	//   acc_bwp=osr4_avg1 (no averaging)
	//   acc_odr=50Hz      (50Hz sampling interval, enough for the step counter)
	const accelConf = 0x00<<7 | 0x00<<4 | 0x07<<0
	err = d.write1(_ACC_CONF, accelConf)
	if err != nil {
		return err
	}

	// Reduce current consumption.
	// With power saving enabled (and the above ACC_CONF) the chip consumes only
	// 14µA.
	err = d.write1(_PWR_CONF, 0x03)
	if err != nil {
		return err
	}

	return nil
}

func (d *Device) Update(which drivers.Measurement) error {
	// TODO: combine temperature and step counter into a single read.
	if which&drivers.Temperature != 0 {
		val, err := d.read1(_TEMPERATURE)
		if err != nil {
			return err
		}
		d.combinedTempSteps[4] = val
	}
	if which&drivers.Acceleration != 0 {
		// The acceleration data is stored in DATA8 through DATA13 as 3 12-bit
		// values.
		err := d.readn(_DATA_8, d.accelData[:]) // ACC_X(LSB)
		if err != nil {
			return err
		}
		err = d.readn(_STEP_COUNTER_0, d.combinedTempSteps[:4])
		if err != nil {
			return err
		}
	}
	return nil
}

// Temperature returns the last read temperature in celsius milli degrees (1°C
// is 1000).
func (d *Device) Temperature() int32 {
	// The temperature value is a two's complement number (meaning: signed) in
	// units of 1 kelvin, with 0 being 23°C.
	return (int32(int8(d.combinedTempSteps[4])) + 23) * 1000
}

// Acceleration returns the last read acceleration in µg (micro-gravity).
// When one of the axes is pointing straight to Earth and the sensor is not
// moving the returned value will be around 1000000 or -1000000.
func (d *Device) Acceleration() (x, y, z int32) {
	// Combine raw data from d.accelData (stored as 12-bit signed values) into a
	// number (0..4095):
	x = int32(d.accelData[0])>>4 | int32(d.accelData[1])<<4
	y = int32(d.accelData[2])>>4 | int32(d.accelData[3])<<4
	z = int32(d.accelData[4])>>4 | int32(d.accelData[5])<<4
	// Sign extend this number to -2048..2047:
	x = (x << 20) >> 20
	y = (y << 20) >> 20
	z = (z << 20) >> 20
	// Scale from -512..511 to -1000_000..998_046.
	// Or, at the maximum range (4g), from -2048..2047 to -2000_000..3998_046.
	// The formula derived as follows (where 512 is the expected value at 1g):
	//   x = x * 1000_000      / 512
	//   x = x * (1000_000/64) / (512/64)
	//   x = x * 15625         / 8
	x = x * 15625 / 8
	y = y * 15625 / 8
	z = z * 15625 / 8
	return
}

// Steps returns the number of steps counted since the BMA42x sensor was
// initialized.
func (d *Device) Steps() (steps uint32) {
	steps |= uint32(d.combinedTempSteps[0]) << 0
	steps |= uint32(d.combinedTempSteps[1]) << 8
	steps |= uint32(d.combinedTempSteps[2]) << 16
	steps |= uint32(d.combinedTempSteps[3]) << 24
	return
}

func (d *Device) read1(register uint8) (uint8, error) {
	d.dataBuf[0] = register
	err := d.bus.Tx(uint16(d.address), d.dataBuf[:1], d.dataBuf[1:2])
	return d.dataBuf[1], err
}

func (d *Device) readn(register uint8, data []byte) error {
	d.dataBuf[0] = register
	return d.bus.Tx(uint16(d.address), d.dataBuf[:1], data)
}

func (d *Device) write1(register uint8, data uint8) error {
	d.dataBuf[0] = register
	d.dataBuf[1] = data
	return d.bus.Tx(uint16(d.address), d.dataBuf[:2], nil)
}

func unsafeStringToSlice(s string) []byte {
	// TODO: use unsafe.Slice(unsafe.StringData(...)) once we require Go 1.20.
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	return unsafe.Slice((*byte)(unsafe.Pointer(sh.Data)), len(s))
}

func identifyChip(chipID uint8) DeviceType {
	switch chipID {
	case 0x11:
		return DeviceBMA421
	case 0x13:
		return DeviceBMA425
	default:
		return noDevice
	}
}
