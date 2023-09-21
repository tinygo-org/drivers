package as7262

type vLedControlReg struct {
	illuminationCurrentLimit byte
	illuminationStatus       byte
	indicatorCurrentLimit    byte
	indicatorStatus          byte
}

func newVLedControlReg() *vLedControlReg {
	return &vLedControlReg{
		illuminationCurrentLimit: 0b00110000,
		illuminationStatus:       0b00001000,
		indicatorCurrentLimit:    0b00000110,
		indicatorStatus:          0b00000001,
	}
}

// encode register to a complete byte for writing
func (r *vLedControlReg) encode() byte {
	return r.illuminationCurrentLimit |
		r.illuminationStatus |
		r.indicatorCurrentLimit |
		r.indicatorStatus
}

// decode register to represent as7262 internal state
func (r *vLedControlReg) decode(encoded byte) {
	r.illuminationCurrentLimit = encoded & 0b00110000
	r.illuminationStatus = encoded & 0b00001000
	r.indicatorCurrentLimit = encoded & 0b00000110
	r.indicatorStatus = encoded & 0b00000001
}

// setIlCur
func (r *vLedControlReg) setIlCr(ilCurLim float32) {
	// values: 12.5, 25, 50, 100 (defaults to 50)
	var cr byte
	switch ilCurLim {
	case 12.5:
		cr = 0b00
	case 25:
		cr = 0b01
	case 100:
		cr = 0b11
	default:
		cr = 0b10
	}

	//bitwise clear operation & setting bit 4:5
	r.illuminationCurrentLimit &= 0b11001111
	r.illuminationCurrentLimit |= cr << 4
}

// setIlOn
func (r *vLedControlReg) setIlOn(ilOn bool) {
	if ilOn {
		r.illuminationStatus |= 0b00001000
	} else {
		r.illuminationStatus &= 0b11110111
	}
}

// setInCur
func (r *vLedControlReg) setInCur(inCurLim float32) {
	// values: 1, 2, 4, 8 (defaults to 8)
	var cr byte
	switch inCurLim {
	case 1:
		cr = 0b00
	case 2:
		cr = 0b01
	case 4:
		cr = 0b10
	default:
		cr = 0b11
	}

	//bitwise clear operation & setting bit 4:5
	r.indicatorCurrentLimit &= 0b11111001
	r.indicatorCurrentLimit |= cr << 1
}

// setInOn
func (r *vLedControlReg) setInOn(inOn bool) {
	if inOn {
		r.indicatorStatus |= 0b00000001
	} else {
		r.indicatorStatus &= 0b11111110
	}
}

// ConfigureLed with all possible configurations
func (d *Device) ConfigureLed(ilCurLim float32, ilOn bool, inCurLim float32, inOn bool) {
	lr := newVLedControlReg()
	lr.setIlCr(ilCurLim)
	lr.setIlOn(ilOn)
	lr.setInCur(inCurLim)
	lr.setInOn(inOn)
	lrEncoded := lr.encode()

	// write ControlReg and read full ControlReg
	d.writeByte(LedRegister, lrEncoded)
	d.readByte(LedRegister)
	lr.decode(d.buf[0])
}
