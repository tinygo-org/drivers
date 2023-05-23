/*
TM1638 is an LED Controller
*/
package tm1638

import (
	"machine"
	"time"
)

// Device wraps the pins of the TM1638.
type Device struct {
	/* STB - When this pin is "LO" chip accepting transmission */
	strobe machine.Pin
	/* CLK - DIO pin reads data at the rising edge and outputs at the falling edge */
	clock machine.Pin
	/* DIO - This pin outputs/inputs serial data */
	data machine.Pin
}

const (
	addressAutoIncrement = 0x40
	fixedAddress         = 0x44
	baseAddress          = 0xC0
	maxAddress           = 0x0F
	readKeys             = 0x42
	brightness           = 0x88
	maxBrightness        = 0x07
	zeroBrightness       = 0x80
)

// Create new TM1638 device
func New(strobe machine.Pin, clock machine.Pin, data machine.Pin) Device {
	strobe.Configure(machine.PinConfig{Mode: machine.PinOutput})
	clock.Configure(machine.PinConfig{Mode: machine.PinOutput})
	return Device{strobe: strobe, clock: clock, data: data}
}

// Configure TM1638
func (d *Device) Configure() {
	d.Clear()
	d.Brightness(maxBrightness)
}

// Clear display memory
func (d *Device) Clear() {
	d.sendCommand(addressAutoIncrement)
	d.data.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.strobe.Low()
	d.transmissionDelay()
	d.write(baseAddress)
	for i := 0; i < 16; i++ {
		d.write(0)
	}
	d.strobe.High()
}

// Set display brightness
func (d *Device) Brightness(value uint8) {
	if value == 0 {
		d.sendCommand(zeroBrightness)
	} else {
		d.sendCommand(brightness | (value & maxBrightness))
	}
}

// Write one display memory element
func (d *Device) Write(offset uint8, data uint8) {
	d.sendCommand(fixedAddress)
	d.strobe.Low()
	d.transmissionDelay()
	d.data.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.write(baseAddress | (offset & maxAddress))
	d.write(data)
	d.strobe.High()
}

// Write array to display memory
func (d *Device) WriteArray(offset uint8, data []uint8) {
	d.sendCommand(addressAutoIncrement)
	d.strobe.Low()
	d.transmissionDelay()
	d.data.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.write(baseAddress | (offset & maxAddress))
	for _, element := range data {
		d.write(element)
	}
	d.strobe.High()
}

// Scan keyboard
func (d *Device) ScanKeyboard(buffer *[4]uint8) {
	d.strobe.Low()
	d.transmissionDelay()
	d.data.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.write(readKeys)
	d.data.Configure(machine.PinConfig{Mode: machine.PinInput})
	d.transmissionDelay()
	for index := range buffer {
		var element uint8 = 0
		for bitIndex := 0; bitIndex < 8; bitIndex++ {
			d.clock.Low()
			d.transmissionDelay()
			if d.data.Get() {
				element |= (1 << bitIndex)
			}
			d.transmissionDelay()
			d.clock.High()
			d.transmissionDelay()
		}
		buffer[index] = element
	}
	d.strobe.High()
}

func (d *Device) sendCommand(command uint8) {
	d.strobe.Low()
	d.transmissionDelay()
	d.data.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.write(command)
	d.strobe.High()
}

func (d *Device) write(value uint8) {
	for i := 0; i < 8; i++ {
		d.clock.Low()
		d.transmissionDelay()
		d.data.Set((value & (1 << i)) > 0)
		d.transmissionDelay()
		d.clock.High()
		d.transmissionDelay()
	}
}

func (d *Device) transmissionDelay() {
	time.Sleep(time.Microsecond * time.Duration(2))
}
