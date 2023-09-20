package as7262

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

// encode register to a complete byte for writing
func (r *vControlReg) encode() byte {
	return r.reset | r.interrupt | r.gain | r.bank | r.dataReady
}

// decode register to represent as7262 internal state
func (r *vControlReg) decode(encoded byte) {
	r.reset = encoded & 0b10000000
	r.interrupt = encoded & 0b01000000
	r.gain = encoded & 0b00110000
	r.bank = encoded & 0b00001100
	r.dataReady = encoded & 0b00000010
}

// setReset bit which will soft reset the as7262 sensor
func (r *vControlReg) setReset(reset bool) {
	if reset {
		r.reset |= 0b10000000
	} else {
		r.reset &= 0b01111111
	}
}

// setGain sets bit 4:5 of VControlReg for gain
func (r *vControlReg) setGain(gain float32) {
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
	r.gain &= 0b11001111
	r.gain |= g << 4
}

// setMode sets bit 2:3 of VControlReg for mode
func (r *vControlReg) setMode(mode int) {
	// set mode: 0, 1, 2, 3
	m := byte(mode)

	// bitwise clear operation & setting bit 4:5
	r.bank &= 0b11110011
	r.bank |= m << 2
}
