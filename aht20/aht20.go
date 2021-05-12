package aht20

import (
	"time"

	"tinygo.org/x/drivers"
)

// Device wraps an I2C connection to an AHT20 device.
type Device struct {
	bus      drivers.I2C
	Address  uint16
	humidity uint32
	temp     uint32
}

// New creates a new AHT20 connection. The I2C bus must already be
// configured.
//
// This function only creates the Device object, it does not touch the device.
func New(bus drivers.I2C) Device {
	return Device{
		bus:     bus,
		Address: Address,
	}
}

// Configure the device
func (d *Device) Configure() {
	// Check initialization state
	status := d.Status()
	if status&0x08 == 1 {
		// Device is initialized
		return
	}

	// Force initialization
	d.bus.Tx(d.Address, []byte{CMD_INITIALIZE, 0x08, 0x00}, nil)
	time.Sleep(10 * time.Millisecond)
}

// Reset the device
func (d *Device) Reset() {
	d.bus.Tx(d.Address, []byte{CMD_SOFTRESET}, nil)
}

// Status of the device
func (d *Device) Status() byte {
	data := []byte{0}

	d.bus.Tx(d.Address, []byte{CMD_STATUS}, data)

	return data[0]
}

// Read the temperature and humidity
//
// The actual temperature and humidity are stored
// and can be accessed using `Temp` and `Humidity`.
func (d *Device) Read() error {
	d.bus.Tx(d.Address, []byte{CMD_TRIGGER, 0x33, 0x00}, nil)

	data := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	for retry := 0; retry < 3; retry++ {
		time.Sleep(80 * time.Millisecond)
		err := d.bus.Tx(d.Address, nil, data)
		if err != nil {
			return err
		}

		// If measurement complete, store values
		if data[0]&0x04 != 0 && data[0]&0x80 == 0 {
			d.humidity = uint32(data[1])<<12 | uint32(data[2])<<4 | uint32(data[3])>>4
			d.temp = (uint32(data[3])&0xF)<<16 | uint32(data[4])<<8 | uint32(data[5])
			return nil
		}
	}

	return ErrTimeout
}

func (d *Device) RawHumidity() uint32 {
	return d.humidity
}

func (d *Device) RawTemp() uint32 {
	return d.temp
}

func (d *Device) RelHumidity() float32 {
	return (float32(d.humidity) * 100) / 0x100000
}

func (d *Device) DeciRelHumidity() int32 {
	return (int32(d.humidity) * 1000) / 0x100000
}

// Temperature in degrees celsius
func (d *Device) Celsius() float32 {
	return (float32(d.temp*200.0) / 0x100000) - 50
}

// Temperature in mutiples of one tenth of a degree celsius
//
// Using this method avoids floating point calculations.
func (d *Device) DeciCelsius() int32 {
	return ((int32(d.temp) * 2000) / 0x100000) - 500
}
