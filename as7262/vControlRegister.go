package as7262

import "time"

type vControlReg struct {
	reset     byte
	interrupt byte
	gain      byte
	bank      byte
	dataReady byte
}

func newVControlReg() *vControlReg {
	return &vControlReg{
		reset:     0b10000000,
		interrupt: 0b01000000,
		gain:      0b00110000,
		bank:      0b00001100, // measurement mode
		dataReady: 0b00000010,
	}
}

// encodeCReg register to a complete byte for writing
func (d *Device) encodeCReg() byte {
	return d.vControlReg.reset |
		d.vControlReg.interrupt |
		d.vControlReg.gain |
		d.vControlReg.bank |
		d.vControlReg.dataReady
}

// decodeCReg register to represent as7262 internal state
func (d *Device) decodeCReg(encoded byte) {
	d.vControlReg.reset = encoded & 0b10000000
	d.vControlReg.interrupt = encoded & 0b01000000
	d.vControlReg.gain = encoded & 0b00110000
	d.vControlReg.bank = encoded & 0b00001100
	d.vControlReg.dataReady = encoded & 0b00000010
}

// setReset bit which will soft reset the as7262 sensor
func (d *Device) setReset(reset bool) {
	if reset {
		d.vControlReg.reset |= 0b10000000
	} else {
		d.vControlReg.reset &= 0b01111111
	}
}

// setGain sets bit 4:5 of VControlReg for gain
func (d *Device) setGain(gain float32) {
	// set gain (defaults to 64)
	// values: 1, 3.7, 16, 64
	var g byte
	switch gain {
	case 1:
		g = 0b00
	case 3.7:
		g = 0b01
	case 16:
		g = 0b10
	default:
		g = 0b11
	}

	// bitwise clear operation & setting bit 4:5
	d.vControlReg.gain &= 0b11001111
	d.vControlReg.gain |= g << 4
}

// setMode sets bit 2:3 of VControlReg for mode
func (d *Device) setMode(mode int) {
	// set mode: 0, 1, 2, 3
	m := byte(mode)

	// bitwise clear operation & setting bit 4:5
	d.vControlReg.bank &= 0b11110011
	d.vControlReg.bank |= m << 2
}

// Configure as7262 behaviour
func (d *Device) Configure(reset bool, gain float32, integrationTime float32, mode int) {
	d.setReset(reset)
	d.setGain(gain)
	d.setMode(mode)
	crEncoded := d.encodeCReg()

	// write ControlReg and read full ControlReg
	d.writeByte(ControlReg, crEncoded)
	time.Sleep(time.Second)
	d.readByte(ControlReg)
	d.decodeCReg(d.buf[0])

	// set integrationTime: float32 as ms
	t := byte(int(integrationTime*2.8) & 0xff)
	d.writeByte(IntegrationTimeReg, t)
}
