package mcp3008

import (
	"machine"

	"tinygo.org/x/drivers"
)

const (
	Chan0 = 1 << iota
	Chan1
	Chan2
	Chan3
	Chan4
	Chan5
	Chan6
	Chan7

	numChannels = iota
	AllChannels = 0xff
)

// Buffered implementation of the MCP3008.
type BufferedDev struct {
	active uint8
	bus    drivers.SPI
	cs     machine.Pin
	tx     [3]byte
	rx     [3]byte
	data   [numChannels]uint16
}

// New returns a new MCP3008 driver. Pass in a fully configured SPI bus.
// activeChannels informs which channels are to be measured.
//
// For example, to measure only on channels 0, 3 and 4 you'd call New as following:
//  d := mcp3008.NewBuffered(spi, cs, mcp3008.Chan0|mcp3008.Chan3|mcp3008.Chan4)
func NewBuffered(b drivers.SPI, csPin machine.Pin, activeChannels uint8) *BufferedDev {
	if activeChannels&AllChannels == 0 {
		panic("mcp3008: no channels selected")
	}
	d := &BufferedDev{bus: b,
		cs:     csPin,
		active: activeChannels,
	}
	csPin.High()
	d.tx[0] = 0x01
	d.tx[2] = 0x00
	return d
}

// Configure sets up the device for communication
func (d *BufferedDev) Configure() {
	d.cs.Configure(machine.PinConfig{Mode: machine.PinOutput})
}

// UNTESTED
func (d *BufferedDev) Update(which drivers.Measurement) error {
	if which|drivers.Voltage == 0 {
		return nil
	}
	for i := 0; i < numChannels; i++ {
		if 1<<i&d.active != 0 { // measure only desired channels.
			d.tx[1] = byte(8+i) << 4
			d.cs.Low()
			// This supposes the ADC supports autoincrement like for example the mcp3464 ADC (UNTESTED)
			err := d.bus.Tx(d.tx[:], d.rx[:])
			if err != nil {
				return err
			}
			d.cs.High()
			d.data[i] = uint16((d.rx[1]&0x3))<<(8+6) + uint16(d.rx[2])<<6
		}
	}
	return nil
}

// Conversion Constants
const (
	// Maximum value read by 10bit ADC after conversion to 16 bit value
	maxVal = 0x3ff << 6
	vRef   = 5 // input voltage

	convConst = 5 * 1000 / maxVal
	convDiv   = maxVal / 1000
	convMul   = 5
)

// Channel returns an ADC measurement for a specific channel. Must call Update Beforehand
func (d *BufferedDev) Channel(ch int) uint16 {
	return d.data[ch]
}

func (d *BufferedDev) ReadChannel(ch int) (uint16, error) {
	err := d.Update(drivers.Voltage)
	return d.Channel(ch), err
}

// Voltage returns voltage in microvolts supposing 5V Vref.
func (d *BufferedDev) Voltage(ch int) int32 {
	return int32(d.Channel(ch)) / convDiv * convMul
}
