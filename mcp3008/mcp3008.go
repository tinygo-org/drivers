// Package mcp3008 implements a driver for the MCP3008 Analog to Digital Converter.
//
// Datasheet: http://ww1.microchip.com/downloads/en/DeviceDoc/21295d.pdf
//
package mcp3008 // import "tinygo.org/x/drivers/mcp3008"

import (
	"errors"
	"machine"

	"tinygo.org/x/drivers"
)

// Device wraps MCP3008 SPI ADC.
type Device struct {
	bus drivers.SPI
	cs  machine.Pin
	tx  []byte
	rx  []byte
	CH0 ADCPin
	CH1 ADCPin
	CH2 ADCPin
	CH3 ADCPin
	CH4 ADCPin
	CH5 ADCPin
	CH6 ADCPin
	CH7 ADCPin
}

// ADCPin is the implementation of the ADConverter interface.
type ADCPin struct {
	machine.Pin
	d *Device
}

// New returns a new MCP3008 driver. Pass in a fully configured SPI bus.
func New(b drivers.SPI, csPin machine.Pin) *Device {
	d := &Device{bus: b,
		cs: csPin,
		tx: make([]byte, 3),
		rx: make([]byte, 3),
	}

	// setup all channels
	d.CH0 = d.GetADC(0)
	d.CH1 = d.GetADC(1)
	d.CH2 = d.GetADC(2)
	d.CH3 = d.GetADC(3)
	d.CH4 = d.GetADC(4)
	d.CH5 = d.GetADC(5)
	d.CH6 = d.GetADC(6)
	d.CH7 = d.GetADC(7)

	return d
}

// Configure sets up the device for communication
func (d *Device) Configure() {
	d.cs.Configure(machine.PinConfig{Mode: machine.PinOutput})
}

// Read analog data from channel
func (d *Device) Read(ch int) (uint16, error) {
	if ch < 0 || ch > 7 {
		return 0, errors.New("invalid channel for MCP3008 Read")
	}

	return d.GetADC(ch).Get(), nil
}

// GetADC returns an ADC for a specific channel.
func (d *Device) GetADC(ch int) ADCPin {
	return ADCPin{machine.Pin(ch), d}
}

// Get the current reading for a specific ADCPin.
func (p ADCPin) Get() uint16 {
	p.d.tx[0] = 0x01
	p.d.tx[1] = byte(8+p.Pin) << 4
	p.d.tx[2] = 0x00

	p.d.cs.Low()
	p.d.bus.Tx(p.d.tx, p.d.rx)

	// scale result to 16bit value like other ADCs
	result := uint16((p.d.rx[1]&0x3))<<8 + uint16(p.d.rx[2])<<6
	p.d.cs.High()

	return result
}

// Configure here just for interface compatibility.
func (p ADCPin) Configure() {
}
