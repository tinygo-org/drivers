// Package xpt2046 implements a driver for the XPT2046 resistive touch controller as packaged on the TFT_320QVT board
//
// Datasheet: http://grobotronics.com/images/datasheets/xpt2046-datasheet.pdf
package xpt2046

import (
	"machine"
	"time"
	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/touch"
)

type Device struct {
	bus    drivers.SPI
	t_clk  machine.Pin
	t_cs   machine.Pin
	t_din  machine.Pin
	t_dout machine.Pin
	t_irq  machine.Pin

	precision uint8
}

type Config struct {
	Precision uint8
}

func New(t_clk, t_cs, t_din, t_dout, t_irq machine.Pin) Device {
	return Device{
		precision: 10,
		t_clk:     t_clk,
		t_cs:      t_cs,
		t_din:     t_din,
		t_dout:    t_dout,
		t_irq:     t_irq,
	}
}

func NewSPI(bus drivers.SPI, t_cs, t_irq machine.Pin) Device {
	t_cs.Configure(machine.PinConfig{Mode: machine.PinOutput})
	t_irq.Configure(machine.PinConfig{Mode: machine.PinInput})

	return Device{
		bus:       bus,
		precision: 10,
		t_cs:      t_cs,
		t_irq:     t_irq,
	}
}

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

func (d *Device) ConfigureSPI(config *Config) error {

	if config.Precision == 0 {
		d.precision = 10
	} else {
		d.precision = config.Precision
	}

	d.t_cs.High()

	//S       = 1    --> Required Control bit
	//A2-A0   = 000  --> nothing
	//MODE    = 0    --> 12 bit conversion
	//SER/DFR = 0    --> Differential preferred for pressure
	//PD1-PD0 = 10   --> Powerdown and enable PEN_IRQ
	d.Command(0x80) // make sure PD1 is cleared on start

	return nil
}


// Tx and Rx pulled from st7789, modified for xps2046

// Tx sends data to the touchpad
func (d *Device) Tx(data []byte, isCommand bool) {
	d.t_cs.Low()
	d.bus.Tx(data, nil)
	d.t_cs.High()
}

// Rx reads data from the touchpad
func (d *Device) Rx(command uint8, data []byte) {
	cmd := make([]byte,len(data))
	cmd[0] = command
	d.t_cs.Low()
	d.bus.Tx(cmd,data)
	d.t_cs.High()
}

// Command sends a command to the touch screen.
func (d *Device) Command(command uint8) {
	d.Tx([]byte{command}, true)
}

// Data sends data to the touch screen. // XXX needed?
func (d *Device) Data(data uint8) {
	d.Tx([]byte{data}, false)
}

func busSleep() {
	time.Sleep(5 * time.Nanosecond)
}

func pulseHigh(p machine.Pin) {
	p.High()
	busSleep()
	p.Low()
	busSleep()
}

func (d *Device) writeCommand(data uint8) {

	for count := uint8(0); count < 8; count++ {
		d.t_din.Set((data & 0x80) != 0)
		data <<= 1
		pulseHigh(d.t_clk)
	}

}

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

func (d *Device) ReadTouchPoint() touch.Point {

	tx := uint32(0)
	ty := uint32(0)
	tz := uint32(0)
	rx := uint32(0)
	ry := uint32(0)
	rz := uint32(0)
	sampleCount := uint8(0)

	d.t_cs.Low()

	for ; sampleCount < d.precision && d.Touched(); sampleCount++ {
		if d.bus != nil {
			rx, ry, rz = d.readRawSPI()
		} else {
			rx, ry, rz = d.readRaw()
		}
		tx += uint32(rx)
		ty += uint32(ry)
		tz += uint32(rz)
	}
	d.t_cs.High()

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

func (d *Device) Touched() bool {
	avail := !d.t_irq.Get()
	return avail
}

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


func (d *Device) readRawSPI() (int32, int32, int32) {

	data := make([]byte,4)

	//S       = 1    --> Required Control bit
	//A2-A0   = 101  --> X-Position
	//MODE    = 0    --> 12 bit conversion
	//SER/DFR = 0    --> Differential preferred for X,Y position
	//PD1-PD0 = 00   --> Powerdown and enable PEN_IRQ

	d.Rx(0xD0,data)
	tx := int32((uint16(data[1])<<8 | uint16(data[2])) >>3) // 7 bits come from data[1], remaining 5 from top of data[2]

	//S       = 1    --> Required Control bit
	//A2-A0   = 001  --> Y-Position
	//MODE    = 0    --> 12 bit conversion
	//SER/DFR = 0    --> Differential preferred for X,Y position
	//PD1-PD0 = 00   --> Powerdown and enable PEN_IRQ

	d.Rx(0x90,data)
	ty := int32((uint16(data[1])<<8 | uint16(data[2])) >>3) // 7 bits come from data[0], remaining 5 from top of data[1]

	//S       = 1    --> Required Control bit
	//A2-A0   = 011  --> Z1-position (pressure)
	//MODE    = 0    --> 12 bit conversion
	//SER/DFR = 0    --> Differential preferred for pressure
	//PD1-PD0 = 00   --> Powerdown and enable PEN_IRQ

	d.Rx(0xB0,data)
	tz1 := int32((uint16(data[1])<<8 | uint16(data[2])) >>3) // 7 bits come from data[0], remaining 5 from top of data[1]

	//S       = 1    --> Required Control bit
	//A2-A0   = 100  --> Z2-position (pressure)
	//MODE    = 0    --> 12 bit conversion
	//SER/DFR = 0    --> Differential preferred for pressure
	//PD1-PD0 = 00   --> Powerdown and enable PEN_IRQ

	d.Rx(0xC0,data)
	tz2 := int32((uint16(data[1])<<8 | uint16(data[2])) >>3) // 7 bits come from data[0], remaining 5 from top of data[1]

	tz := int32(0)
	if tz1 != 0 {
		//Touch pressure is proportional to the ratio of z2 to z1 and the x position.
		tz = int32(tx) * ((tz2 << 12) / (tz1 << 12))
	}

	//Scale X&Y to 16 bit for consistency across touch drivers
	return int32(tx) << 4, int32(4096-ty) << 4, tz
}
