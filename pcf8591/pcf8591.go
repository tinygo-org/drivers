// Package pcf8591 implements a driver for the PCF8591 Analog to Digital/Digital to Analog Converter.
//
// Datasheet: https://www.nxp.com/docs/en/data-sheet/PCF8591.pdf
package pcf8591 // import "tinygo.org/x/drivers/pcf8591"

import (
	"machine"

	"errors"

	"tinygo.org/x/drivers"
)

// Device wraps PCF8591 ADC functions.
type Device struct {
	bus     drivers.I2C
	Address uint16
	CH0     ADCPin
	CH1     ADCPin
	CH2     ADCPin
	CH3     ADCPin
}

// ADCPin is the implementation of the ADConverter interface.
type ADCPin struct {
	machine.Pin
	d *Device
}

// New returns a new PCF8591 driver. Pass in a fully configured I2C bus.
func New(b drivers.I2C) *Device {
	d := &Device{
		bus:     b,
		Address: defaultAddress,
	}

	// setup all channels
	d.CH0 = d.GetADC(0)
	d.CH1 = d.GetADC(1)
	d.CH2 = d.GetADC(2)
	d.CH3 = d.GetADC(3)

	return d
}

// Configure here just for interface compatibility.
func (d *Device) Configure() {
}

// Read analog data from channel
func (d *Device) Read(ch int) (uint16, error) {
	if ch < 0 || ch > 3 {
		return 0, errors.New("invalid channel for pcf8591 Read")
	}

	return d.GetADC(ch).Get(), nil
}

// GetADC returns an ADC for a specific channel.
func (d *Device) GetADC(ch int) ADCPin {
	return ADCPin{machine.Pin(ch), d}
}

// Get the current reading for a specific ADCPin.
func (p ADCPin) Get() uint16 {
	// TODO: also implement DAC
	tx := make([]byte, 2)
	tx[0] = byte(p.Pin)

	rx := make([]byte, 2)

	// The result from the measurement triggered by the first write,
	// however, the second write is required to get the result.
	// See section 8.4 "A/D Conversion" in the datasheet for more info
	p.d.bus.Tx(p.d.Address, tx, rx)
	p.d.bus.Tx(p.d.Address, tx, rx)

	// scale result to 16bit value like other ADCs
	return uint16(rx[1]) << 8
}

// Configure here just for interface compatibility.
func (p ADCPin) Configure() {
}
