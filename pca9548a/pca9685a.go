// Package pca9548a provides a driver for the PCA9548A 8-channel I2C-bus.
//
// Datasheet:
// https://www.nxp.com/docs/en/data-sheet/PCA9548A.pdf
package pca9548a // import "tinygo.org/x/drivers/pca9548a"

import (
	"tinygo.org/x/drivers"
)

// Device is a handle to the PCA9548A device given an address
type Device struct {
	addr uint8
	bus  drivers.I2C
	buf  [4]byte
}

// New creates a new instance of a PCA9548A device. It performs
// no IO on the i2c bus.
func New(bus drivers.I2C, addr uint8) Device {
	return Device{
		bus:  bus,
		addr: addr,
	}
}

// Connected returns whether a PCA9548A has been found.
// It does a "who am I" request and checks the response.
func (d *Device) IsConnected() bool {
	d.SetPortState(0xA5)
	response := d.GetPortState()
	d.SetPortState(0x00)
	return response == 0xA5
}

// SetPort enables the given port to send data.
func (d *Device) SetPort(portNumber byte) {
	portValue := uint8(0)
	if portNumber <= 7 {
		portValue = 1 << portNumber
	}
	d.bus.Tx(uint16(d.addr), []byte{portValue}, nil)
}

// GetPort gets the first port enabled.
func (d *Device) GetPort() byte {
	portBits := d.GetPortState()
	for i := uint8(0); i < 8; i++ {
		if (portBits & (1 << i)) != 0x00 {
			return i
		}
	}
	return InvalidPort
}

// SetPortState set the states of all the ports at the same time, this could cause some issues if
// you enable two (or more) ports with the same devices at the same time.
func (d *Device) SetPortState(portBits byte) {
	d.bus.Tx(uint16(d.addr), []byte{portBits}, nil)
}

// GetPortState get the state of all the ports.
func (d *Device) GetPortState() byte {
	portBits := make([]byte, 1)
	d.bus.Tx(uint16(d.addr), nil, portBits)
	return portBits[0]
}

// EnablePort enables the given port without modifying any other, this could cause some issues if
// you enable two (or more) ports with the same devices at the same time.
func (d *Device) EnablePort(portNumber byte) {
	if portNumber > 7 {
		portNumber = 7
	}

	settings := d.GetPortState()
	settings |= 1 << portNumber

	d.SetPortState(settings)
}

// DisablePort disables the given port without modifying any other.
func (d *Device) DisablePort() byte {
	portBits := make([]byte, 1)
	d.bus.Tx(uint16(d.addr), nil, portBits)
	return portBits[0]
}
