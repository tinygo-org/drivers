package lps22hb

import (
	"machine"
	"time"

	"tinygo.org/x/drivers"
)

// Device wraps an I2C connection to a HTS221 device.
type Device struct {
	bus         drivers.I2C
	Address     uint8
	OnNano33BLE bool
}

// New creates a new LPS22HB connection. The I2C bus must already be
// configured.
//
// This function only creates the Device object, it does not touch the device.
//
// For Nano 33 BLE Sense, use machine.P0_15 (SCL1) and machine.P0_14 (SDA1),
// and set onNano33BLE as hts221.ON_NANO_33_BLE.
func New(bus drivers.I2C, deviceType uint8) Device {
	return Device{bus: bus, Address: LPS22HB_ADDRESS, OnNano33BLE: deviceType == 1}
}

// Connected returns whether LPS22HB has been found.
// It does a "who am I" request and checks the response.
func (d *Device) Connected() bool {

	// if the LPS22HB is on Nano 33 BLE Sense,
	// turn on power pin (machine.P0_22) and I2C1 pullups power pin (machine.P1_00)
	// and wait a moment.
	if d.OnNano33BLE {
		ENV := machine.Pin(22)
		ENV.Configure(machine.PinConfig{Mode: machine.PinOutput})
		ENV.High()
		R := machine.Pin(32)
		R.Configure(machine.PinConfig{Mode: machine.PinOutput})
		R.High()
		time.Sleep(time.Millisecond * 10)
	}

	data := []byte{0}
	d.bus.ReadRegister(d.Address, LPS22HB_WHO_AM_I_REG, data)
	return data[0] == 0xB1
}

// Configure sets up the LPS22HB device for communication.
func (d *Device) Configure() {
	// set to block update mode
	d.bus.WriteRegister(d.Address, LPS22HB_CTRL1_REG, []byte{0x02})
}

// ReadPressure returns the pressure in milli pascals (mPa).
func (d *Device) ReadPressure() (pressure int32, err error) {
	d.waitForOneShot()

	// read data
	data := []byte{0, 0, 0}
	d.bus.ReadRegister(d.Address, LPS22HB_PRESS_OUT_REG, data[:1])
	d.bus.ReadRegister(d.Address, LPS22HB_PRESS_OUT_REG+1, data[1:2])
	d.bus.ReadRegister(d.Address, LPS22HB_PRESS_OUT_REG+2, data[2:])
	pValue := float32(uint32(data[2])<<16|uint32(data[1])<<8|uint32(data[0])) / 4096.0

	return int32(pValue * 1000000), nil
}

// ReadTemperature returns the temperature in celsius milli degrees (°C/1000).
func (d *Device) ReadTemperature() (temperature int32, err error) {
	d.waitForOneShot()

	// read data
	data := []byte{0, 0}
	d.bus.ReadRegister(d.Address, LPS22HB_TEMP_OUT_REG, data[:1])
	d.bus.ReadRegister(d.Address, LPS22HB_TEMP_OUT_REG+1, data[1:])
	tValue := float32(uint16(data[1])<<8|uint16(data[0])) / 100.0

	return int32(tValue * 1000000), nil
}

// private functions

// wait and trigger one shot in block update
func (d *Device) waitForOneShot() {
	// wait until one shot is cleared
	data := []byte{1}
	for {
		d.bus.ReadRegister(d.Address, LPS22HB_CTRL2_REG, data)
		if data[0]&0x01 == 0 {
			break
		}
	}

	// trigger one shot
	d.bus.WriteRegister(d.Address, LPS22HB_CTRL2_REG, []byte{0x01})

	// wait until one shot is cleared
	data[0] = 1
	for {
		d.bus.ReadRegister(d.Address, LPS22HB_CTRL2_REG, data)
		if data[0]&0x01 == 0 {
			break
		}
	}
}
