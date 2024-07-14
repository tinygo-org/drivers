// Package dht20 implements a driver for the DHT20 temperature and humidity sensor.
//
// Datasheet: https://cdn-shop.adafruit.com/product-files/5183/5193_DHT20.pdf

package dht20

import (
	"errors"
	"time"

	"tinygo.org/x/drivers"
)

var (
	errUpdateCalledTooSoon = errors.New("Update() called within 80ms is invalid")
)

// Device wraps an I2C connection to a DHT20 device.
type Device struct {
	bus            drivers.I2C
	Address        uint16
	data           [8]uint8
	temperature    float32
	humidity       float32
	prevAccessTime time.Time
}

// New creates a new DHT20 connection. The I2C bus must already be
// configured.
//
// This function only creates the Device object, it does not touch the device.
func New(bus drivers.I2C) Device {
	return Device{
		bus:     bus,
		Address: defaultAddress, // Using the address defined in registers.go
	}
}

// Configure sets up the device for communication and initializes the registers if needed.
func (d *Device) Configure() {
	// Get the status word
	d.data[0] = 0x71
	d.bus.Tx(d.Address, d.data[:1], d.data[:1])
	if d.data[0] != 0x18 {
		// Initialize registers
		d.initRegisters()
	}
	// Set the previous access time to the current time
	d.prevAccessTime = time.Now()
}

// initRegisters initializes the registers 0x1B, 0x1C, and 0x1E to 0x00.
func (d *Device) initRegisters() {
	// Initialize register 0x1B
	d.data[0] = 0x1B
	d.data[1] = 0x00
	d.bus.Tx(d.Address, d.data[:2], nil)

	// Initialize register 0x1C
	d.data[0] = 0x1C
	d.data[1] = 0x00
	d.bus.Tx(d.Address, d.data[:2], nil)

	// Initialize register 0x1E
	d.data[0] = 0x1E
	d.data[1] = 0x00
	d.bus.Tx(d.Address, d.data[:2], nil)
}

// Update reads data from the sensor and updates the temperature and humidity values.
// Note that the values obtained by this function are from the previous call to Update.
// If you want to use the most recent values, shorten the interval at which Update is called.
func (d *Device) Update(which drivers.Measurement) error {
	// Check if 80ms have passed since the last access
	if time.Since(d.prevAccessTime) < 80*time.Millisecond {
		return errUpdateCalledTooSoon
	}

	// Check the status word Bit[7]
	d.data[0] = 0x71
	d.bus.Tx(d.Address, d.data[:1], d.data[:1])
	if (d.data[0] & 0x80) == 0 {
		// Read 7 bytes of data from the sensor
		d.bus.Tx(d.Address, nil, d.data[:7])
		rawHumidity := uint32(d.data[1])<<12 | uint32(d.data[2])<<4 | uint32(d.data[3])>>4
		rawTemperature := uint32(d.data[3]&0x0F)<<16 | uint32(d.data[4])<<8 | uint32(d.data[5])

		// Convert raw values to human-readable values
		d.humidity = float32(rawHumidity) / 1048576.0 * 100
		d.temperature = float32(rawTemperature)/1048576.0*200 - 50

		// Trigger the next measurement
		d.data[0] = 0xAC
		d.data[1] = 0x33
		d.data[2] = 0x00
		d.bus.Tx(d.Address, d.data[:3], nil)

		// Update the previous access time to the current time
		d.prevAccessTime = time.Now()
	}
	return nil
}

// Temperature returns the last measured temperature.
func (d *Device) Temperature() float32 {
	return d.temperature
}

// Humidity returns the last measured humidity.
func (d *Device) Humidity() float32 {
	return d.humidity
}
