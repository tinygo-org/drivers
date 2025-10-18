package honeyhsc

import (
	"errors"
	"math"

	"tinygo.org/x/drivers"
)

var (
	errSensorMissing = errors.New("hsc: not connected")
	errDiagnostic    = errors.New("hsc: diagnostic error")
)

const (
	measuremask  = drivers.Pressure | drivers.Temperature
	statusMask   = 0b1100_0000
	statusOffset = 6
)

// DevI2C is the TruStability® High Accuracy Silicon Ceramic (HSC) Series is a piezoresistive silicon pressure sensor offering a ratiometric
// analog or digital output for reading pressure over the specified full scale pressure span and temperature range.
type DevI2C struct {
	bus drivers.I2C
	dev
	addr uint8
	buf  [6]byte
}

// NewDevI2C creates and returns a new DevI2C that communicates with an HSC device over the provided I2C bus.
// Parameters:
//   - bus: the I2C bus to use.
//   - addr: the 7-bit I2C address of the sensor.
//   - outMin, outMax: raw output code range (counts) corresponding to the pressure span. Depends on sensor model.
//   - pMin, pMax: pressure range endpoints in millipascals (mPa). Depends on sensor model.
//
// The returned DevI2C will use these calibration parameters to convert raw bridge counts to pressure.
func NewDevI2C(bus drivers.I2C, addr, outMin, outMax uint16, pMin, pMax int32) *DevI2C {
	h := &DevI2C{
		bus:  bus,
		addr: uint8(addr),
		dev: dev{
			cmin: outMin,
			cmax: outMax,
			pmin: pMin,
			pmax: pMax,
		},
	}
	return h
}

// ReadTemperature reads and returns the temperature in milliKelvin (mC) from the I2C-attached HSC device.
// It performs an Update internally to get the latest temperature value.
func (h *DevI2C) ReadTemperature() (int32, error) {
	err := h.Update(drivers.Temperature)
	if err != nil {
		return 0, err
	}
	return h.Temperature(), nil
}

// Update reads both temperature and pressure data from the I2C-attached HSC device when
// the requested measurement mask includes pressure or temperature.
// If neither pressure nor temperature is requested, Update is a no-op.
func (d *DevI2C) Update(which drivers.Measurement) error {
	// Update performs an I2C transaction to read 4 bytes, parses the status bits, 14-bit bridge data and
	// temperature bits, and forwards them to the internal update routine. Any I2C transport error is returned,
	// as well as errors produced by the internal update (e.g. errSensorMissing, errDiagnostic).
	if which&measuremask == 0 {
		return nil
	}
	rbuf := d.buf[:4]
	wbuf := d.buf[4:6]
	const reg = 0
	value := (d.addr << 1) | 1
	wbuf[0] = reg
	wbuf[1] = value
	err := d.bus.Tx(uint16(d.addr), wbuf, rbuf)
	if err != nil {
		return err
	}
	status := (rbuf[0] & statusMask) >> statusOffset
	bridgeData := (uint16(rbuf[0]&^statusMask) << 8) | uint16(rbuf[1])
	tempData := uint16(rbuf[2])<<8 | uint16(rbuf[3]&0xe0)>>5
	return d.dev.update(status, bridgeData, tempData)
}

type pinout func(level bool)

// DevI2C is the TruStability® High Accuracy Silicon Ceramic (HSC) Series is a piezoresistive silicon pressure sensor offering a ratiometric
// analog or digital output for reading pressure over the specified full scale pressure span and temperature range.
type DevSPI struct {
	spi drivers.SPI
	cs  pinout
	dev
	buf [4]byte
}

// NewDevSPI creates and returns a new DevSPI that communicates with an HSC device over SPI.
// Parameters:
//   - conn: the SPI connection to use.
//   - cs: a chip-select function that drives the device select line low/high.
//   - outMin, outMax: raw output code range (counts) corresponding to the pressure span. Depends on sensor model.
//   - pMin, pMax: pressure range endpoints in millipascals (mPa). Depends on sensor model.
//
// The function returns the constructed DevSPI and an error value (currently always nil).
func NewDevSPI(conn drivers.SPI, cs pinout, outMin, outMax uint16, pMin, pMax int32) (*DevSPI, error) {
	h := &DevSPI{
		spi: conn,
		cs:  cs,
		dev: dev{
			cmin: outMin,
			cmax: outMax,
			pmin: pMin,
			pmax: pMax,
		},
	}
	return h, nil
}

// ReadTemperature reads and returns the temperature in milliKelvin (mC) from the SPI-attached HSC device.
// It performs an Update internally to get the latest temperature value.
func (h *DevSPI) ReadTemperature() (int32, error) {
	err := h.Update(drivers.Temperature)
	if err != nil {
		return 0, err
	}
	return h.Temperature(), nil
}

// Update reads pressure and temperature data from the SPI-attached HSC device when the requested measurement mask includes
// pressure or temperature. If neither pressure nor temperature is requested, Update is a no-op.
func (h *DevSPI) Update(which drivers.Measurement) error {
	// It toggles the provided chip-select, performs an SPI transfer to read 4 bytes, parses the status bits,
	// 14-bit bridge data and temperature bits, and forwards them to the internal update routine. Any SPI
	// transport error is returned, as well as errors produced by the internal update (e.g. errSensorMissing, errDiagnostic).
	if which&measuremask == 0 {
		return nil
	}
	buf := &h.buf
	h.cs(false)
	err := h.spi.Tx(nil, buf[:4])
	h.cs(true)
	if err != nil {
		return err
	}
	// First two bits are status bits.
	status := (buf[0] & statusMask) >> statusOffset
	bridgeData := (uint16(buf[0]&^statusMask) << 8) | uint16(buf[1])

	tempData := uint16(buf[2])<<8 | uint16(buf[3]&0xe0)>>5
	return h.dev.update(status, bridgeData, tempData)
}

type dev struct {
	pressure   int32
	temp       int32
	cmin, cmax uint16
	pmin, pmax int32
}

// Pressure returns the most recently computed pressure value in millipascals (mPa).
// The value is taken from the last successful Update.
func (d *dev) Pressure() int32 {
	return d.pressure
}

// Temperature returns the most recently read temperature value in milliKelvin (mC).
// The value is taken from the last successful Update.
func (d *dev) Temperature() int32 {
	return d.temp + 273_150
}

// update interprets raw sensor fields (status, bridgeData, tempData) and updates the dev's stored
// pressure and temperature. It returns errSensorMissing when the temperature raw value indicates no sensor
// (tempData == math.MaxUint16), errDiagnostic when the status indicates a device diagnostic condition
// (status == 3), or nil on success. Pressure is computed with integer arithmetic using the configured
// cmin/cmax -> pmin/pmax linear mapping in order to avoid overflows.
func (d *dev) update(status uint8, bridgeData, tempData uint16) error {
	if tempData == math.MaxUint16 {
		return errSensorMissing
	} else if status == 3 {
		return errDiagnostic
	}

	// Take care not to overflow here.
	p := (int32(bridgeData)-int32(d.cmin))*(d.pmax-d.pmin)/int32(d.cmax-d.cmin) + d.pmin
	d.temp = int32(tempData)
	d.pressure = p
	return nil
}
