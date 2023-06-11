/*
Package tm1638 provides a driver for TM1638 IC manufactured by Titan Microelectronics.
It integrates MCU digital interface, data latch, LED drive, and keypad scanning circuit.
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

// Config contains IC configuration
type Config struct {
	/* Brightness level from 0 to 7 */
	Brightness uint8
}

const (
	/* Address increasing mode: automatic address increased */
	cmdAddressAutoIncrement = 0x40
	/* Read key scan data */
	cmdReadKeyScan = 0x42
	/* Display off */
	cmdZeroBrightness = 0x80
	/* Display on command. Bits 0-2 may contain brightness value */
	cmdSetBrightness = 0x88
	/* Address Setting Command is used to set the address of the display memory. Bits 0-3 used for address value */
	cmdSetAddress = 0xC0
	// MaxAddress is max valid address of display memory
	MaxAddress = 0x0F
	// MaxBrightness is max display brightness level
	MaxBrightness = 0x07
)

// New Create new TM1638 device
func New(strobe machine.Pin, clock machine.Pin, data machine.Pin) Device {
	strobe.Configure(machine.PinConfig{Mode: machine.PinOutput})
	clock.Configure(machine.PinConfig{Mode: machine.PinOutput})
	return Device{strobe: strobe, clock: clock, data: data}
}

// Configure TM1638
func (d *Device) Configure(config Config) {
	d.Clear()
	d.SetBrightness(config.Brightness)
}

// Clear display memory
func (d *Device) Clear() {
	d.sendCommand(cmdAddressAutoIncrement)
	d.data.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.strobe.Low()
	d.transmissionDelay()
	d.write(cmdSetAddress)
	for i := 0; i < 16; i++ {
		d.write(0)
	}
	d.strobe.High()
}

// SetBrightness changes display brightness
func (d *Device) SetBrightness(value uint8) {
	if value == 0 {
		d.sendCommand(cmdZeroBrightness)
	} else {
		if value > MaxBrightness {
			value = MaxBrightness
		}
		d.sendCommand(cmdSetBrightness | value)
	}
}

// WriteAt writes array to display memory
func (d *Device) WriteAt(data []byte, offset int64) (n int, err error) {
	d.sendCommand(cmdAddressAutoIncrement)
	d.strobe.Low()
	d.transmissionDelay()
	d.data.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.write(cmdSetAddress | uint8(offset))
	for _, element := range data {
		d.write(element)
	}
	d.strobe.High()
	return len(data), nil
}

/*
	ScanKeyboard fill buffer with keyboard scan data.

Keyboard scan matrix has 3 rows (K1-K3) and 8 (KS1-KS8) columns.
Each button connected to one K and KS line of IC.

|---------|---------------------|---------------------|
|  Mapping table                                      |
|---------|---------------------|---------------------|
|Columns: |- K3 - K2 - K1 - X  -|- K3 - K2 - K1 - X  -|
|---------|---------------------|---------------------|
|Bit:     |- B0 - B1 - B2 - B3 -|- B4 - B5 - B6 - B7 -|
|---------|---------------------|---------------------|
|BYTE-1:  |         KS1         |         KS2         |
|BYTE-2:  |         KS3         |         KS4         |
|BYTE-3:  |         KS5         |         KS6         |
|BYTE-4:  |         KS7         |         KS8         |
|---------|---------------------|---------------------|
*/
func (d *Device) ScanKeyboard(buffer *[4]uint8) {
	d.strobe.Low()
	d.transmissionDelay()
	d.data.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.write(cmdReadKeyScan)
	d.data.Configure(machine.PinConfig{Mode: machine.PinInput})
	d.transmissionDelay()
	for index := range buffer {
		var element uint8 = 0
		for bitIndex := 0; bitIndex < 8; bitIndex++ {
			d.clock.Low()
			d.transmissionDelay()
			if d.data.Get() {
				element |= 1 << bitIndex
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
