package tester

// I2CDevice represents a mock I2C device on a mock I2C bus with 16-bit registers.
type I2CDevice16 struct {
	c Failer
	// addr is the i2c device address.
	addr uint8
	// Registers holds the device registers. It can be inspected
	// or changed as desired for testing.
	Registers map[uint8]uint16
	// If Err is non-nil, it will be returned as the error from the
	// I2C methods.
	Err error
}

// NewI2CDevice returns a new mock I2C device.
//
// To use this mock, populate the Registers map with known / expected
// registers.  Attempts by the code under test to write to a register
// that has not been populated into the map will be treated as an
// error.
func NewI2CDevice16(c Failer, addr uint8) *I2CDevice16 {
	return &I2CDevice16{
		c:         c,
		addr:      addr,
		Registers: map[uint8]uint16{},
	}
}

// Addr returns the Device address.
func (d *I2CDevice16) Addr() uint8 {
	return d.addr
}

// ReadRegister implements I2C.ReadRegister.
func (d *I2CDevice16) readRegister(r uint8, buf []byte) error {
	if d.Err != nil {
		return d.Err
	}

	if len(buf) > 2 {
		d.c.Fatalf("register read [%#x, %#x] oversized buffer", r, len(buf))
	}

	val, ok := d.Registers[r]
	if !ok {
		d.c.Fatalf("register read [%#x] unknown register", r)
	}

	buf[0] = byte(val >> 8)
	buf[1] = byte(val & 0xff)

	return nil
}

// WriteRegister implements I2C.WriteRegister.
func (d *I2CDevice16) writeRegister(r uint8, buf []byte) error {
	if d.Err != nil {
		return d.Err
	}

	if len(buf) != 2 {
		d.c.Fatalf("register write [%#x, %#x] mis-sized write", r, len(buf))
	}

	_, ok := d.Registers[r]
	if !ok {
		d.c.Fatalf("register write [%#x] unknown register", r)
	}

	d.Registers[r] = uint16(buf[0])<<8 | uint16(buf[1])

	return nil
}

// Tx implements I2C.Tx.
func (bus *I2CDevice16) Tx(w, r []byte) error {
	switch len(w) {
	case 0:
		bus.c.Fatalf("i2c mock: need a write byte")
		return nil
	case 1:
		return bus.readRegister(w[0], r)
	default:
		if len(r) > 0 || len(w) == 1 {
			bus.c.Fatalf("i2c mock: unsupported lengths in Tx(%d, %d)", len(w), len(r))
		}
		return bus.writeRegister(w[0], w[1:])
	}
}
