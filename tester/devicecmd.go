package tester

// Cmd represents a command sent via I2C to a device.
//
// A command matches when (Command & Mask) == (Data & Mask).  If
// a command is recognized, Response bytes is returned.
type Cmd struct {
	Command     []byte
	Mask        []byte
	Response    []byte
	Invocations int
}

// I2CDeviceCmd represents a mock I2C device that does not
// have 'registers', but has a command/response model.
//
// Commands and canned responses are pre-loaded into the
// Commands member.  For each command the mock receives it
// will lookup the command and return the corresponding
// canned response.
type I2CDeviceCmd struct {
	c Failer

	// addr is the i2c device address.
	addr uint8

	// Commands are the commands the device recognizes and responds to.
	Commands map[uint8]*Cmd

	// Command response that is pending (used with a command is split over)
	// two transactions
	pendingResponse []byte

	// If Err is non-nil, it will be returned as the error from the
	// I2C methods.
	Err error
}

// NewI2CDeviceCmd returns a new mock I2C device.
func NewI2CDeviceCmd(c Failer, addr uint8) *I2CDeviceCmd {
	return &I2CDeviceCmd{
		c:    c,
		addr: addr,
	}
}

// Addr returns the Device address.
func (d *I2CDeviceCmd) Addr() uint8 {
	return d.addr
}

// ReadRegister implements I2C.ReadRegister.
func (d *I2CDeviceCmd) readRegister(r uint8, buf []byte) error {
	if d.Err != nil {
		return d.Err
	}

	return nil
}

// WriteRegister implements I2C.WriteRegister.
func (d *I2CDeviceCmd) writeRegister(r uint8, buf []byte) error {
	if d.Err != nil {
		return d.Err
	}

	return nil
}

// Tx implements I2C.Tx.
func (d *I2CDeviceCmd) Tx(w, r []byte) error {
	if d.Err != nil {
		return d.Err
	}

	if len(w) == 0 && len(d.pendingResponse) != 0 {
		return d.respond(r)
	}

	cmd := d.FindCommand(w)
	if cmd == nil {
		d.c.Fatalf("command [%#x] not identified", w)
		return nil
	}

	cmd.Invocations++
	d.pendingResponse = cmd.Response
	return d.respond(r)
}

func (d *I2CDeviceCmd) FindCommand(command []byte) *Cmd {
	for _, c := range d.Commands {
		if len(c.Command) > len(command) {
			continue
		}

		match := true
		for i := 0; i < len(c.Command); i++ {
			mask := c.Mask[i]
			if (c.Command[i] & mask) != (command[i] & mask) {
				match = false
				break
			}
		}

		if match {
			return c
		}
	}

	return nil
}

func (d *I2CDeviceCmd) respond(r []byte) error {
	if len(r) > len(d.pendingResponse) {
		d.c.Fatalf("read too large (expected: <= %#x, got: %#x)",
			len(d.pendingResponse), len(r))
	}

	if len(r) > 0 {
		copy(r, d.pendingResponse[:len(r)])
		d.pendingResponse = nil
	}

	return nil
}
