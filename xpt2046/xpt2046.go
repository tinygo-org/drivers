// Package xpt2046 implements a driver for the XPT2046 resistive touch controller
// as packaged on the TFT_320QVT board or a Waveshare Pico-ResTouch-LCD-2.8 board.
//
// Datasheet: http://grobotronics.com/images/datasheets/xpt2046-datasheet.pdf
package xpt2046

import (
	"machine"
	"time"
	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/touch"
)

// This driver supports both GPIO and SPI interfaces for a given touch controller,
// but not at the same time.  GPIO works fine unless t_clk, t_din, and t_dout
// are shared with another device and that device is expecting to use SPI over
// those pins.
type Device struct {
	bus    drivers.SPI
	t_clk  machine.Pin
	t_cs   machine.Pin
	t_din  machine.Pin
	t_dout machine.Pin
	t_irq  machine.Pin

	precision uint8
}

// Simple configuration -- Precision indicates the number of samples
// averaged to produce X, Y, and Z (pressure) coordinates.
type Config struct {
	Precision uint8
}

// Create a new GPIO-based device.  SPI bus not used in this setup.
func New(t_clk, t_cs, t_din, t_dout, t_irq machine.Pin) Device {
	return Device{
		bus:       nil,
		precision: 10,
		t_clk:     t_clk,
		t_cs:      t_cs,
		t_din:     t_din,
		t_dout:    t_dout,
		t_irq:     t_irq,
	}
}

// Create a new SPI-based device.  GPIO not available for this instance
// when SPI is used.  
func NewSPI(bus drivers.SPI, t_cs, t_irq machine.Pin) Device {
	return Device{
		bus:       bus,
		precision: 10,
		t_cs:      t_cs,
		t_irq:     t_irq,
	}
}

// Configure a GPIO-based device.  Sets up the Precision of the device
// and initializes the GPIO pins.
func (d *Device) Configure(config *Config) error {

	if config.Precision == 0 {
		d.precision = 10
	} else {
		d.precision = config.Precision
	}

	d.t_clk.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.t_cs.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.t_din.Configure(machine.PinConfig{Mode: machine.PinOutput})

	d.t_dout.Configure(machine.PinConfig{Mode: machine.PinInput})
	d.t_irq.Configure(machine.PinConfig{Mode: machine.PinInput})

	d.t_clk.Low()
	d.t_cs.High()
	d.t_din.Low()

	d.readRaw() //Set Powerdown mode to enable T_IRQ

	return nil
}

// Configure a SPI-based device.  Sets up the Precision of the device.
// Also initializes t_cs and t_irq.  All of the other pins in Device are
// used by SPI.
func (d *Device) ConfigureSPI(config *Config) error {

	if config.Precision == 0 {
		d.precision = 10
	} else {
		d.precision = config.Precision
	}

	d.t_cs.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.t_irq.Configure(machine.PinConfig{Mode: machine.PinInput})

	d.t_cs.High()

	//S       = 1    --> Required Control bit
	//A2-A0   = 000  --> nothing
	//MODE    = 0    --> 12 bit conversion
	//SER/DFR = 0    --> Differential preferred for pressure
	//PD1-PD0 = 10   --> Powerdown and enable PEN_IRQ
	d.tx([]byte{0x80})  // make sure PD1 is cleared on start
	return nil
}

// tx and rx pulled from st7789, modified for xps2046

// tx sends data to the touchpad.
func (d *Device) tx(data []byte) {
	d.t_cs.Low()
	d.bus.Tx(data, nil)
	d.t_cs.High()
}

// rx reads data from the touchpad.
func (d *Device) rx(command uint8, data []byte) {
	cmd := make([]byte, len(data))
	cmd[0] = command
	d.t_cs.Low()
	d.bus.Tx(cmd, data)
	d.t_cs.High()
}

// Very short sleep for GPIO pulsing.
func busSleep() {
	time.Sleep(5 * time.Nanosecond)
}

// Pulse the given pin p high and then low.
func pulseHigh(p machine.Pin) {
	p.High()
	busSleep()
	p.Low()
	busSleep()
}

// Write a command to the touchscreen using GPIO.
func (d *Device) writeCommand(data uint8) {

	for count := uint8(0); count < 8; count++ {
		d.t_din.Set((data & 0x80) != 0)
		data <<= 1
		pulseHigh(d.t_clk)
	}

}

// Read whatever data is waiting on t_dout using GPIO.
func (d *Device) readData() uint16 {

	data := uint16(0)

	for count := uint8(0); count < 12; count++ {
		data <<= 1
		pulseHigh(d.t_clk)
		if d.t_dout.Get() {
			data |= 1
		}
	}
	pulseHigh(d.t_clk) //13
	pulseHigh(d.t_clk) //14
	pulseHigh(d.t_clk) //15
	pulseHigh(d.t_clk) //16

	return data
}

// Read the X, Y, and Z (pressure) coordinates of the point currently being
// touched on the screen.  Works for both GPIO and SPI by calling the associated
// raw read routines.  The device is queried at most d.precision times, with
// the resulting touch.Point having the average values for each of X, Y, and Z.
func (d *Device) ReadTouchPoint() touch.Point {

	tx := uint32(0)
	ty := uint32(0)
	tz := uint32(0)
	rx := int32(0)
	ry := int32(0)
	rz := int32(0)
	sampleCount := uint8(0)

	if d.bus == nil {
		d.t_cs.Low()
	}

	for ; sampleCount < d.precision && d.Touched(); sampleCount++ {
		if d.bus == nil {
			rx, ry, rz = d.readRaw()
		} else {
			rx, ry, rz = d.readRawSPI()
		}
		tx += uint32(rx)
		ty += uint32(ry)
		tz += uint32(rz)
		if d.bus != nil {
			time.Sleep(200 * time.Microsecond)
		}
	}
	if d.bus == nil {
		d.t_cs.High()
	}

	if sampleCount > 0 {
		x := int(tx / uint32(sampleCount))
		y := int(ty / uint32(sampleCount))
		z := int(tz / uint32(sampleCount))
		return touch.Point{
			X: x,
			Y: y,
			Z: z,
		}
	} else {
		return touch.Point{
			X: 0,
			Y: 0,
			Z: 0,
		}
	}
}

// Touched() is true if the touch device senses a touch at the current moment.
func (d *Device) Touched() bool {
	avail := !d.t_irq.Get()
	return avail
}

// Read the current X, Y, and Z values using the GPIO interface.
func (d *Device) readRaw() (int32, int32, int32) {

	d.t_cs.Low()

	//S       = 1    --> Required Control bit
	//A2-A0   = 001  --> Y-Position
	//MODE    = 0    --> 12 bit conversion
	//SER/DFR = 0    --> Differential preferred for X,Y position
	//PD1-PD0 = 00   --> Powerdown and enable PEN_IRQ
	d.writeCommand(0x90)
	ty := d.readData()

	//S       = 1    --> Required Control bit
	//A2-A0   = 101  --> X-Position
	//MODE    = 0    --> 12 bit conversion
	//SER/DFR = 0    --> Differential preferred for X,Y position
	//PD1-PD0 = 00   --> Powerdown and enable PEN_IRQ
	d.writeCommand(0xD0)
	tx := d.readData()

	//S       = 1    --> Required Control bit
	//A2-A0   = 011  --> Z1-position (pressure)
	//MODE    = 0    --> 12 bit conversion
	//SER/DFR = 0    --> Differential preferred for pressure
	//PD1-PD0 = 00   --> Powerdown and enable PEN_IRQ
	d.writeCommand(0xB0)
	tz1 := int32(d.readData())

	//S       = 1    --> Required Control bit
	//A2-A0   = 100  --> Z2-position (pressure)
	//MODE    = 0    --> 12 bit conversion
	//SER/DFR = 0    --> Differential preferred for pressure
	//PD1-PD0 = 00   --> Powerdown and enable PEN_IRQ
	d.writeCommand(0xC0)
	tz2 := int32(d.readData())

	tz := int32(0)
	if tz1 != 0 {
		//Touch pressure is proportional to the ratio of z2 to z1 and the x position.
		tz = int32(tx) * ((tz2 << 12) / (tz1 << 12))
	}

	d.t_cs.High()

	//Scale X&Y to 16 bit for consistency across touch drivers
	return int32(tx) << 4, int32(4096-ty) << 4, tz
}

// Read the current X, Y, and Z values using the SPI interface.
func (d *Device) readRawSPI() (int32, int32, int32) {

	data := make([]byte, 4)

	//S       = 1    --> Required Control bit
	//A2-A0   = 101  --> X-Position
	//MODE    = 0    --> 12 bit conversion
	//SER/DFR = 0    --> Differential preferred for X,Y position
	//PD1-PD0 = 00   --> Powerdown and enable PEN_IRQ

	d.rx(0xD0, data)
	tx := int32((uint16(data[1])<<8 | uint16(data[2])) >> 3) // 7 bits come from data[1], remaining 5 from top of data[2]

	//S       = 1    --> Required Control bit
	//A2-A0   = 001  --> Y-Position
	//MODE    = 0    --> 12 bit conversion
	//SER/DFR = 0    --> Differential preferred for X,Y position
	//PD1-PD0 = 00   --> Powerdown and enable PEN_IRQ

	d.rx(0x90, data)
	ty := int32((uint16(data[1])<<8 | uint16(data[2])) >> 3) // 7 bits come from data[0], remaining 5 from top of data[1]

	//S       = 1    --> Required Control bit
	//A2-A0   = 011  --> Z1-position (pressure)
	//MODE    = 0    --> 12 bit conversion
	//SER/DFR = 0    --> Differential preferred for pressure
	//PD1-PD0 = 00   --> Powerdown and enable PEN_IRQ

	d.rx(0xB0, data)
	tz1 := int32((uint16(data[1])<<8 | uint16(data[2])) >> 3) // 7 bits come from data[0], remaining 5 from top of data[1]

	//S       = 1    --> Required Control bit
	//A2-A0   = 100  --> Z2-position (pressure)
	//MODE    = 0    --> 12 bit conversion
	//SER/DFR = 0    --> Differential preferred for pressure
	//PD1-PD0 = 00   --> Powerdown and enable PEN_IRQ

	d.rx(0xC0, data)
	tz2 := int32((uint16(data[1])<<8 | uint16(data[2])) >> 3) // 7 bits come from data[0], remaining 5 from top of data[1]

	tz := int32(0)
	if tz1 != 0 {
		//Touch pressure is proportional to the ratio of z2 to z1 and the x position.
		tz = int32(tx) * ((tz2 << 12) / (tz1 << 12))
	}

	//Scale X&Y to 16 bit for consistency across touch drivers
	return int32(tx) << 4, int32(4096-ty) << 4, tz
}
