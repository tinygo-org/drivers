package i2csoft

import (
	"errors"
	"machine"
	"time"

	"tinygo.org/x/drivers/delay"
)

// I2C is an I2C implementation by Software. Since it is implemented by
// software, it can be used with microcontrollers that do not have I2C
// function. This is not efficient but works around broken or missing drivers.
type I2C struct {
	scl      machine.Pin
	sda      machine.Pin
	nack     bool
	baudrate uint32
}

// I2CConfig is used to store config info for I2C.
type I2CConfig struct {
	Frequency uint32
	SCL       machine.Pin
	SDA       machine.Pin
}

var (
	errSI2CAckExpected = errors.New("I2C error: expected ACK not NACK")
)

// New returns the i2csoft driver. For the arguments, specify the pins to be
// used as SCL and SDA. As I2C is implemented in software, any GPIO pin can be
// specified.
func New(sclPin, sdaPin machine.Pin) *I2C {
	return &I2C{
		scl:      sclPin,
		sda:      sdaPin,
		baudrate: 100e3,
	}
}

// Configure is intended to setup the I2C interface.
func (i2c *I2C) Configure(config I2CConfig) error {
	// Default I2C bus speed is 100 kHz.
	if config.Frequency != 0 {
		i2c.SetBaudRate(config.Frequency)
	}

	// This exists for compatibility with machine.I2CConfig. SCL and SDA must
	// be set at the same time. Because Pin(0) is sometimes set, it is not
	// checked for 0.
	if config.SCL != config.SDA {
		i2c.scl = config.SCL
		i2c.sda = config.SDA
	}

	// enable pins
	i2c.sda.Configure(machine.PinConfig{Mode: machine.PinOutput})
	i2c.sda.High()
	i2c.scl.Configure(machine.PinConfig{Mode: machine.PinOutput})
	i2c.scl.High()

	return nil
}

// SetBaudRate sets the communication speed for the I2C.
func (i2c *I2C) SetBaudRate(br uint32) {
	// At this time, the value of i2c.baudrate is ignored because it is fixed
	// at 100 kHz. SetBaudrate() is exist for compatibility with machine.I2C.
	i2c.baudrate = br
}

// Tx does a single I2C transaction at the specified address.
// It clocks out the given address, writes the bytes in w, reads back len(r)
// bytes and stores them in r, and generates a stop condition on the bus.
func (i2c *I2C) Tx(addr uint16, w, r []byte) error {
	i2c.nack = false
	if len(w) != 0 {
		// send start/address for write
		i2c.sendAddress(addr, true)

		// wait until transmission complete

		// ACK received (0: ACK, 1: NACK)
		if i2c.nack {
			i2c.signalStop()
			return errSI2CAckExpected
		}

		// write data
		for _, b := range w {
			i2c.writeByte(b)
		}

		i2c.signalStop()
	}
	if len(r) != 0 {
		// send start/address for read
		i2c.sendAddress(addr, false)

		// wait transmission complete

		// ACK received (0: ACK, 1: NACK)
		if i2c.nack {
			i2c.signalStop()
			return errSI2CAckExpected
		}

		// read first byte
		r[0] = i2c.readByte()
		for i := 1; i < len(r); i++ {
			// Send an ACK

			i2c.signalRead()

			// Read data and send the ACK
			r[i] = i2c.readByte()
		}

		// Send NACK to end transmission
		i2c.sendNack()

		i2c.signalStop()
	}

	return nil
}

// writeByte writes a single byte to the I2C bus.
func (i2c *I2C) writeByte(data byte) {
	// Send data byte
	i2c.scl.Low()
	i2c.sda.High()
	i2c.sda.Configure(machine.PinConfig{Mode: machine.PinOutput})
	i2c.wait()

	for i := 0; i < 8; i++ {
		i2c.scl.Low()
		if ((data >> (7 - i)) & 1) == 1 {
			i2c.sda.High()
		} else {
			i2c.sda.Low()
		}
		i2c.wait()
		i2c.wait()
		i2c.scl.High()
		i2c.wait()
		i2c.wait()
	}

	i2c.scl.Low()
	i2c.wait()
	i2c.wait()
	i2c.sda.Configure(machine.PinConfig{Mode: machine.PinInput})
	i2c.scl.High()
	i2c.wait()

	i2c.nack = i2c.sda.Get()

	i2c.wait()

	// wait until transmission successful
}

// sendAddress sends the address and start signal
func (i2c *I2C) sendAddress(address uint16, write bool) {
	data := (address << 1)
	if !write {
		data |= 1 // set read flag
	}

	i2c.scl.High()
	i2c.sda.Low()
	i2c.wait()
	i2c.wait()
	for i := 0; i < 8; i++ {
		i2c.scl.Low()
		if ((data >> (7 - i)) & 1) == 1 {
			i2c.sda.High()
		} else {
			i2c.sda.Low()
		}
		i2c.wait()
		i2c.wait()
		i2c.scl.High()
		i2c.wait()
		i2c.wait()
	}

	i2c.scl.Low()
	i2c.wait()
	i2c.wait()
	i2c.sda.Configure(machine.PinConfig{Mode: machine.PinInput})
	i2c.scl.High()
	i2c.wait()

	i2c.nack = i2c.sda.Get()

	i2c.wait()

	// wait until bus ready
}

func (i2c *I2C) signalStop() {
	i2c.scl.Low()
	i2c.sda.Low()
	i2c.sda.Configure(machine.PinConfig{Mode: machine.PinOutput})
	i2c.wait()
	i2c.wait()
	i2c.scl.High()
	i2c.wait()
	i2c.wait()
	i2c.sda.High()
	i2c.wait()
	i2c.wait()
}

func (i2c *I2C) signalRead() {
	i2c.wait()
	i2c.wait()
	i2c.scl.Low()
	i2c.sda.Low()
	i2c.sda.Configure(machine.PinConfig{Mode: machine.PinOutput})
	i2c.wait()
	i2c.wait()
	i2c.scl.High()
	i2c.wait()
	i2c.wait()
}

func (i2c *I2C) readByte() byte {
	var data byte
	for i := 0; i < 8; i++ {
		i2c.scl.Low()
		i2c.sda.Configure(machine.PinConfig{Mode: machine.PinInput})
		i2c.wait()
		i2c.wait()
		i2c.scl.High()
		if i2c.sda.Get() {
			data |= 1 << (7 - i)
		}
		i2c.wait()
		i2c.wait()
	}
	return data
}

func (i2c *I2C) sendNack() {
	i2c.wait()
	i2c.wait()
	i2c.scl.Low()
	i2c.sda.High()
	i2c.sda.Configure(machine.PinConfig{Mode: machine.PinOutput})
	i2c.wait()
	i2c.wait()
	i2c.scl.High()
	i2c.wait()
	i2c.wait()
}

// WriteRegister transmits first the register and then the data to the
// peripheral device.
//
// Many I2C-compatible devices are organized in terms of registers. This method
// is a shortcut to easily write to such registers. Also, it only works for
// devices with 7-bit addresses, which is the vast majority.
func (i2c *I2C) WriteRegister(address uint8, register uint8, data []byte) error {
	buf := make([]uint8, len(data)+1)
	buf[0] = register
	copy(buf[1:], data)
	return i2c.Tx(uint16(address), buf, nil)
}

// ReadRegister transmits the register, restarts the connection as a read
// operation, and reads the response.
//
// Many I2C-compatible devices are organized in terms of registers. This method
// is a shortcut to easily read such registers. Also, it only works for devices
// with 7-bit addresses, which is the vast majority.
func (i2c *I2C) ReadRegister(address uint8, register uint8, data []byte) error {
	return i2c.Tx(uint16(address), []byte{register}, data)
}

// wait waits for half the time of the SCL operation interval.
func (i2c *I2C) wait() {
	delay.Sleep(50 * time.Microsecond) // half of a 100kHz cycle (50µs)
}
