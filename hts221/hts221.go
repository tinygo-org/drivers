// Package hts221 implements a driver for HTS221,
// a capacitive digital sensor for relative humidity and temperature.
//
// Datasheet: https://www.st.com/resource/en/datasheet/hts221.pdf
package hts221

import (
	"errors"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/internal/legacy"
)

// Device wraps an I2C connection to a HTS221 device.
type Device struct {
	bus              drivers.I2C
	Address          uint8
	humiditySlope    float32
	humidityZero     float32
	temperatureSlope float32
	temperatureZero  float32
}

// New creates a new HTS221 connection. The I2C bus must already be
// configured.
//
// This function only creates the Device object, it does not touch the device.
func New(bus drivers.I2C) Device {
	return Device{bus: bus, Address: HTS221_ADDRESS}
}

// Connected returns whether HTS221 has been found.
// It does a "who am I" request and checks the response.
func (d *Device) Connected() bool {
	data := []byte{0}
	legacy.ReadRegister(d.bus, d.Address, HTS221_WHO_AM_I_REG, data)
	return data[0] == 0xBC
}

// Power is for turn on/off the HTS221 device
func (d *Device) Power(status bool) {
	data := []byte{0}
	if status {
		data[0] = 0x84
	}
	legacy.WriteRegister(d.bus, d.Address, HTS221_CTRL1_REG, data)
}

// ReadHumidity returns the relative humidity in percent * 100.
// Returns an error if the device is not turned on.
func (d *Device) ReadHumidity() (humidity int32, err error) {
	err = d.waitForOneShot(0x02)
	if err != nil {
		return
	}

	// read data and calibrate
	data := []byte{0, 0}
	legacy.ReadRegister(d.bus, d.Address, HTS221_HUMID_OUT_REG, data[:1])
	legacy.ReadRegister(d.bus, d.Address, HTS221_HUMID_OUT_REG+1, data[1:])
	hValue := readInt(data[1], data[0])
	hValueCalib := float32(hValue)*d.humiditySlope + d.humidityZero

	return int32(hValueCalib * 100), nil
}

// ReadTemperature returns the temperature in celsius milli degrees (°C/1000).
// Returns an error if the device is not turned on.
func (d *Device) ReadTemperature() (temperature int32, err error) {
	err = d.waitForOneShot(0x01)
	if err != nil {
		return
	}

	// read data and calibrate
	data := []byte{0, 0}
	legacy.ReadRegister(d.bus, d.Address, HTS221_TEMP_OUT_REG, data[:1])
	legacy.ReadRegister(d.bus, d.Address, HTS221_TEMP_OUT_REG+1, data[1:])
	tValue := readInt(data[1], data[0])
	tValueCalib := float32(tValue)*d.temperatureSlope + d.temperatureZero

	return int32(tValueCalib * 1000), nil
}

// Resolution sets the HTS221's resolution mode.
// The higher resolutions are more accurate but comsume more power (see datasheet).
// The number of averaged samples will be (h + 2) ^ 2, (t + 1) ^ 2
func (d *Device) Resolution(h uint8, t uint8) {
	if h > 7 {
		h = 3 // default
	}
	if t > 7 {
		t = 3 // default
	}
	legacy.WriteRegister(d.bus, d.Address, HTS221_AV_CONF_REG, []byte{h<<3 | t})
}

// private functions

// read factory calibration data
func (d *Device) calibration() {
	h0rH, h1rH := []byte{0}, []byte{0}
	t0degC, t1degC := []byte{0}, []byte{0}
	t1t0msb := []byte{0}
	h0t0Out, h1t0Out := []byte{0, 0}, []byte{0, 0}
	t0Out, t1Out := []byte{0, 0}, []byte{0, 0}

	legacy.ReadRegister(d.bus, d.Address, HTS221_H0_rH_x2_REG, h0rH)
	legacy.ReadRegister(d.bus, d.Address, HTS221_H1_rH_x2_REG, h1rH)
	legacy.ReadRegister(d.bus, d.Address, HTS221_T0_degC_x8_REG, t0degC)
	legacy.ReadRegister(d.bus, d.Address, HTS221_T1_degC_x8_REG, t1degC)
	legacy.ReadRegister(d.bus, d.Address, HTS221_T1_T0_MSB_REG, t1t0msb)
	legacy.ReadRegister(d.bus, d.Address, HTS221_H0_T0_OUT_REG, h0t0Out[:1])
	legacy.ReadRegister(d.bus, d.Address, HTS221_H0_T0_OUT_REG+1, h0t0Out[1:])
	legacy.ReadRegister(d.bus, d.Address, HTS221_H1_T0_OUT_REG, h1t0Out[:1])
	legacy.ReadRegister(d.bus, d.Address, HTS221_H1_T0_OUT_REG+1, h1t0Out[1:])
	legacy.ReadRegister(d.bus, d.Address, HTS221_T0_OUT_REG, t0Out[:1])
	legacy.ReadRegister(d.bus, d.Address, HTS221_T0_OUT_REG+1, t0Out[1:])
	legacy.ReadRegister(d.bus, d.Address, HTS221_T1_OUT_REG, t1Out[:1])
	legacy.ReadRegister(d.bus, d.Address, HTS221_T1_OUT_REG+1, t1Out[1:])

	h0rH_v := float32(h0rH[0]) / 2.0
	h1rH_v := float32(h1rH[0]) / 2.0
	t0degC_v := float32(readUint(t1t0msb[0]&0x03, t0degC[0])) / 8.0
	t1degC_v := float32(readUint(t1t0msb[0]&0x0C>>2, t1degC[0])) / 8.0
	h0t0Out_v := float32(readInt(h0t0Out[1], h0t0Out[0]))
	h1t0Out_v := float32(readInt(h1t0Out[1], h1t0Out[0]))
	t0Out_v := float32(readInt(t0Out[1], t0Out[0]))
	t1Out_v := float32(readInt(t1Out[1], t1Out[0]))

	d.humiditySlope = (h1rH_v - h0rH_v) / (h1t0Out_v - h0t0Out_v)
	d.humidityZero = h0rH_v - d.humiditySlope*h0t0Out_v
	d.temperatureSlope = (t1degC_v - t0degC_v) / (t1Out_v - t0Out_v)
	d.temperatureZero = t0degC_v - d.temperatureSlope*t0Out_v
}

// wait and trigger one shot in block update
func (d *Device) waitForOneShot(filter uint8) error {
	data := []byte{0}

	// check if the device is on
	legacy.ReadRegister(d.bus, d.Address, HTS221_CTRL1_REG, data)
	if data[0]&0x80 == 0 {
		return errors.New("device is off, unable to query")
	}

	// wait until one shot (one conversion) is ready to go
	data[0] = 1
	for {
		legacy.ReadRegister(d.bus, d.Address, HTS221_CTRL2_REG, data)
		if data[0]&0x01 == 0 {
			break
		}
	}

	// trigger one shot
	legacy.WriteRegister(d.bus, d.Address, HTS221_CTRL2_REG, []byte{0x01})

	// wait until conversion completed
	data[0] = 0
	for {
		legacy.ReadRegister(d.bus, d.Address, HTS221_STATUS_REG, data)
		if data[0]&filter == filter {
			break
		}
	}

	return nil
}

func readUint(msb byte, lsb byte) uint16 {
	return uint16(msb)<<8 | uint16(lsb)
}

func readInt(msb byte, lsb byte) int16 {
	return int16(uint16(msb)<<8 | uint16(lsb))
}
