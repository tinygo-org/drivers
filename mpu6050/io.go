package mpu6050

// Read reads a single register with minimal heap allocations
func (d *DeviceStored) read(reg uint8) (byte, error) {
	d.rreg[0] = reg
	err := d.bus.Tx(uint16(d.Address), d.rreg[:1], d.buf[:1])
	return d.buf[0], err
}

// Write writes a single register.
func (d *DeviceStored) write(reg uint8, data byte) (err error) {
	d.buf[0] = reg
	d.buf[1] = data
	return d.bus.Tx(d.Address, d.buf[:2], nil)
}
