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
