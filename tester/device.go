package tester

// MaxRegisters is the maximum number of registers supported for a Device.
const MaxRegisters = 200

// I2CDevice represents a mock I2C device on a mock I2C bus.
type I2CDevice struct {
	c Failer
	// addr is the i2c device address.
	addr uint8
	// Registers holds the device registers. It can be inspected
	// or changed as desired for testing.
	registers [MaxRegisters]uint8
	// If Err is non-nil, it will be returned as the error from the
	// I2C methods.
	Err error
}

// NewI2CDevice returns a new mock I2C device.
func NewI2CDevice(c Failer, addr uint8) *I2CDevice {
	return &I2CDevice{
		c:    c,
		addr: addr,
	}
}

// Addr returns the Device address.
func (d *I2CDevice) Addr() uint8 {
	return d.addr
}

// SetupRegisters sets all of the Device registers.
// It is intended to be used when setting up a fake device
// for testing expected vs. actual values.
func (d *I2CDevice) SetupRegisters(regs []uint8) {
	if len(regs) > MaxRegisters {
		panic("exceeded maximum number of registers for fake device")
	}
	for k, v := range regs {
		d.registers[k] = v
	}
}

// SetupRegister sets one of the Device registers.
// It is intended to be used when setting up a fake device
// for testing expected vs. actual values.
func (d *I2CDevice) SetupRegister(r, v uint8) {
	if r > MaxRegisters {
		panic("exceeded maximum number of registers for fake device")
	}
	d.registers[r] = v
}

// ReadRegister implements I2C.ReadRegister.
func (d *I2CDevice) ReadRegister(r uint8, buf []byte) error {
	if d.Err != nil {
		return d.Err
	}
	d.AssertRegisterRange(r, buf)
	copy(buf, d.registers[r:])
	return nil
}

// WriteRegister implements I2C.WriteRegister.
func (d *I2CDevice) WriteRegister(r uint8, buf []byte) error {
	if d.Err != nil {
		return d.Err
	}
	d.AssertRegisterRange(r, buf)
	copy(d.registers[r:], buf)
	return nil
}

// AssertRegisterRange asserts that reading or writing the given
// register and subsequent registers is in range of the available registers.
func (d *I2CDevice) AssertRegisterRange(r uint8, buf []byte) {
	if int(r) >= len(d.registers) {
		d.c.Fatalf("register read/write [%#x, %#x] start out of range", r, int(r)+len(buf))
	}
	if int(r)+len(buf) > len(d.registers) {
		d.c.Fatalf("register read/write [%#x, %#x] end out of range", r, int(r)+len(buf))
	}
}
