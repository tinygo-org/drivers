package tester

// I2CDevice represents a mock I2C device on a mock I2C bus with 8-bit registers.
type I2CDevice8 struct {
	c Failer
	// addr is the i2c device address.
	addr uint8
	// Registers holds the device registers. It can be inspected
	// or changed as desired for testing.
	Registers [MaxRegisters]uint8
	// If Err is non-nil, it will be returned as the error from the
	// I2C methods.
	Err error
}

// NewI2CDevice returns a new mock I2C device.
//
// For compatibility, this creates an instance of NewI2CDevice8
func NewI2CDevice(c Failer, addr uint8) *I2CDevice8 {
	return NewI2CDevice8(c, addr)
}

// NewI2CDevice8 returns a new mock I2C device.
func NewI2CDevice8(c Failer, addr uint8) *I2CDevice8 {
	return &I2CDevice8{
		c:    c,
		addr: addr,
	}
}

// Addr returns the Device address.
func (d *I2CDevice8) Addr() uint8 {
	return d.addr
}

// ReadRegister implements I2C.ReadRegister.
func (d *I2CDevice8) readRegister(r uint8, buf []byte) error {
	if d.Err != nil {
		return d.Err
	}
	if len(buf) == 0 {
		d.c.Fatalf("no register buffer to read into")
	}
	d.assertRegisterRange(r, buf)
	copy(buf, d.Registers[r:])
	return nil
}

// WriteRegister implements I2C.WriteRegister.
func (d *I2CDevice8) writeRegister(r uint8, buf []byte) error {
	if d.Err != nil {
		return d.Err
	}
	d.assertRegisterRange(r, buf)
	copy(d.Registers[r:], buf)
	return nil
}

// Tx implements I2C.Tx.
func (bus *I2CDevice8) Tx(w, r []byte) error {
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

// assertRegisterRange asserts that reading or writing the given
// register and subsequent registers is in range of the available registers.
func (d *I2CDevice8) assertRegisterRange(r uint8, buf []byte) {
	if int(r) >= len(d.Registers) {
		d.c.Fatalf("register read/write [%#x, %#x] start out of range", r, int(r)+len(buf))
	}
	if int(r)+len(buf) > len(d.Registers) {
		d.c.Fatalf("register read/write [%#x, %#x] end out of range", r, int(r)+len(buf))
	}
}
