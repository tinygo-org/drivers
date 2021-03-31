package tester

// I2CDevice represents a mock I2C device on a mock I2C bus.
type I2CDevice struct {
	c Failer
	// addr is the i2c device address.
	addr uint8
	// Registers holds the device registers. It can be inspected
	// or changed as desired for testing.
	Registers []uint8
	// Stride holds the multiplier for the register number (size
	// of each register)
	Stride uint8
	// If Err is non-nil, it will be returned as the error from the
	// I2C methods.
	Err error
}

// I2CConfig holds the configuration of the mock device.
type I2CConfig struct {
	Stride       uint8
	MaxRegisters uint8
}

// NewI2CDevice returns a new mock I2C device.
func NewI2CDevice(c Failer, addr uint8) *I2CDevice {
	return &I2CDevice{
		c:    c,
		addr: addr,
	}
}

// Configure the mock I2C device. (mandatory)
func (d *I2CDevice) Configure(config I2CConfig) {
	stride := config.Stride
	if stride == 0 {
		stride = 1
	}

	maxreg := config.MaxRegisters
	if maxreg == 0 {
		maxreg = 255
	}

	d.Stride = stride
	d.Registers = make([]byte, maxreg*stride)
}

// Addr returns the Device address.
func (d *I2CDevice) Addr() uint8 {
	return d.addr
}

// ReadRegister implements I2C.ReadRegister.
func (d *I2CDevice) ReadRegister(r uint8, buf []byte) error {
	if d.Err != nil {
		return d.Err
	}

	d.assertRegisterRange(r, buf)

	offset := int(r) * int(d.Stride)
	copy(buf, d.Registers[offset:])

	return nil
}

// WriteRegister implements I2C.WriteRegister.
func (d *I2CDevice) WriteRegister(r uint8, buf []byte) error {
	if d.Err != nil {
		return d.Err
	}

	d.assertRegisterRange(r, buf)

	offset := int(r) * int(d.Stride)
	copy(d.Registers[offset:], buf)

	return nil
}

// assertRegisterRange asserts that reading or writing the given
// register and subsequent registers is in range of the available registers.
func (d *I2CDevice) assertRegisterRange(r uint8, buf []byte) {
	if int(r)*int(d.Stride) >= len(d.Registers) {
		d.c.Fatalf("register read/write [%#x, %#x] start out of range", r, int(r)+len(buf))
	}
	if int(r)*int(d.Stride)+len(buf) > len(d.Registers) {
		d.c.Fatalf("register read/write [%#x, %#x] end out of range (stride=%#x)", r, int(r)+len(buf), d.Stride)
	}
}
