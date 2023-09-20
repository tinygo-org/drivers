package as7262

import (
	"time"
	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/internal/legacy"
)

type Error uint8

const (
	ErrInvalidID Error = 0x1
	TxValid      byte  = 0x02
	RxValid      byte  = 0x01
)

func (e Error) Error() string {
	switch e {
	case ErrInvalidID:
		return "Invalid chip ID"
	default:
		return "Unknown error"
	}
}

type Device struct {
	bus     drivers.I2C
	buf     []byte
	Address uint8
}

// New returns pointer of new as7262 device
func New(i2c drivers.I2C) *Device {
	return &Device{
		bus:     i2c,
		buf:     []byte{0},
		Address: DefaultAddress,
	}
}

/*
	Internal Functions
*/

// deviceStatus returns StatusReg of as7262
func (d *Device) deviceStatus() byte {
	d.buf[0] = 0
	legacy.ReadRegister(d.bus, DefaultAddress, StatusReg, d.buf)
	return d.buf[0]
}

// writeReady returns true if as7262 is ready to write write-register
func (d *Device) writeReady() bool {
	return d.deviceStatus()&TxValid == 0
}

// readReady return true if as7262 is ready to read read-register
func (d *Device) readReady() bool {
	return d.deviceStatus()&RxValid != 0
}

func (d *Device) readByte(reg byte) byte {
	for {
		if d.writeReady() {
			break
		}
	}

	legacy.WriteRegister(d.bus, d.Address, WriteReg, []byte{reg})

	for {
		if d.readReady() {
			break
		}
	}

	legacy.ReadRegister(d.bus, d.Address, ReadReg, d.buf)
	return d.buf[0]
}

func (d *Device) writeByte(reg byte, value byte) {
	for {
		if d.writeReady() {
			break
		}
	}

	legacy.WriteRegister(d.bus, d.Address, WriteReg, []byte{reg | 0x80})

	for {
		if d.writeReady() {
			break
		}
	}

	d.buf[0] = value
	legacy.WriteRegister(d.bus, d.Address, WriteReg, d.buf)
}

/*
	Official as7262 functions (exported)
*/

// Configure as7262 behaviour
func (d *Device) Configure(reset bool, gain float32, integrationTime float32, mode int) {
	cr := newVControlReg()
	cr.setReset(reset)
	cr.setGain(gain)
	cr.setMode(mode)
	crEncoded := cr.encode()

	// write ControlReg and read full ControlReg
	d.writeByte(ControlReg, crEncoded)
	time.Sleep(time.Second * 2)
	d.readByte(ControlReg)
	cr.decode(d.buf[0])

	// set integrationTime: float32 as ms
	t := byte(int(integrationTime*2.8) & 0xff)
	d.writeByte(IntegrationTimeReg, t)
}
