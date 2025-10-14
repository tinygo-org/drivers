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
	buf  [4]byte
}

func NewDevI2C(bus drivers.I2C, addr uint8, outMin, outMax uint16, pMin, pMax int32) *DevI2C {
	h := &DevI2C{
		bus:  bus,
		addr: addr,
		dev: dev{
			cmin: outMin,
			cmax: outMax,
			pmin: pMin,
			pmax: pMax,
		},
	}
	return h
}

func (d *DevI2C) Update(which drivers.Measurement) error {
	if which&measuremask == 0 {
		return nil
	}
	buf := &d.buf
	const reg = 0
	value := (d.addr << 1) | 1
	err := d.bus.Tx(uint16(d.addr), []byte{reg, value}, buf[:])
	if err != nil {
		return err
	}
	status := (buf[0] & statusMask) >> statusOffset
	bridgeData := (uint16(buf[0]&^statusMask) << 8) | uint16(buf[1])
	tempData := uint16(buf[2])<<8 | uint16(buf[3]&0xe0)>>5
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

// Update implements the sensor interface.
func (h *DevSPI) Update(which drivers.Measurement) error {
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

// Pressure returns pressure in millipascals [mPa].
func (d *dev) Pressure() int32 {
	return d.pressure
}

// Temperature returns temperature in milliKelvin [mK].
func (d *dev) Temperature() int32 {
	return d.temp
}

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
