package tester

// I2CDeviceAddr16 represents a mock I2C device with 16-bit register
// addresses and 8-bit registers. Reads and writes auto increment the address.
type I2CDeviceAddr16 struct {
	c Failer
	// addr is the i2c device address.
	addr uint8
	// Registers holds the device registers. It can be inspected
	// or changed as desired for testing.
	Registers [0x10000]uint8
	// If Err is non-nil, it will be returned as the error from the
	// I2C methods.
	Err error
}

// NewI2CDeviceAddr16 returns a new mock I2C device with 16-bit register addresses.
func NewI2CDeviceAddr16(c Failer, addr uint8) *I2CDeviceAddr16 {
	return &I2CDeviceAddr16{
		c:    c,
		addr: addr,
	}
}

// Addr returns the Device address.
func (d *I2CDeviceAddr16) Addr() uint8 {
	return d.addr
}

func (d *I2CDeviceAddr16) readRegister(r uint8, buf []byte) error {
	return d.read(uint16(r), buf)
}

func (d *I2CDeviceAddr16) writeRegister(r uint8, buf []byte) error {
	return d.write(uint16(r), buf)
}

func (d *I2CDeviceAddr16) read(r uint16, buf []byte) error {
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

func (d *I2CDeviceAddr16) write(r uint16, buf []byte) error {
	if d.Err != nil {
		return d.Err
	}
	d.assertRegisterRange(r, buf)
	copy(d.Registers[r:], buf)
	return nil
}

// Tx implements I2C.Tx. The first two bytes written are the big endian
// register address.
func (d *I2CDeviceAddr16) Tx(w, r []byte) error {
	if len(w) < 2 {
		d.c.Fatalf("i2c mock: need a 16-bit register address")
		return nil
	}
	reg := uint16(w[0])<<8 | uint16(w[1])
	if len(w) == 2 {
		return d.read(reg, r)
	}
	if len(r) > 0 {
		d.c.Fatalf("i2c mock: unsupported lengths in Tx(%d, %d)", len(w), len(r))
	}
	return d.write(reg, w[2:])
}

func (d *I2CDeviceAddr16) assertRegisterRange(r uint16, buf []byte) {
	if int(r)+len(buf) > len(d.Registers) {
		d.c.Fatalf("register read/write [%#x, %#x] end out of range", r, int(r)+len(buf))
	}
}
