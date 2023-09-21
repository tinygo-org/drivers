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

// encodeLedReg register to a complete byte for writing
func (d *Device) encodeLedReg() byte {
	return d.vLedControlReg.illuminationCurrentLimit |
		d.vLedControlReg.illuminationStatus |
		d.vLedControlReg.indicatorCurrentLimit |
		d.vLedControlReg.indicatorStatus
}

// decodeLedReg register to represent as7262 internal state
func (d *Device) decodeLedReg(encoded byte) {
	d.vLedControlReg.illuminationCurrentLimit = encoded & 0b00110000
	d.vLedControlReg.illuminationStatus = encoded & 0b00001000
	d.vLedControlReg.indicatorCurrentLimit = encoded & 0b00000110
	d.vLedControlReg.indicatorStatus = encoded & 0b00000001
}

// setIlCr
func (d *Device) setIlCr(ilCurLim float32) {
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
	d.vLedControlReg.illuminationCurrentLimit &= 0b11001111
	d.vLedControlReg.illuminationCurrentLimit |= cr << 4
}

// setIlOn
func (d *Device) setIlOn(ilOn bool) {
	if ilOn {
		d.vLedControlReg.illuminationStatus |= 0b00001000
	} else {
		d.vLedControlReg.illuminationStatus &= 0b11110111
	}
}

// setInCur
func (d *Device) setInCur(inCurLim float32) {
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
	d.vLedControlReg.indicatorCurrentLimit &= 0b11111001
	d.vLedControlReg.indicatorCurrentLimit |= cr << 1
}

// setInOn
func (d *Device) setInOn(inOn bool) {
	if inOn {
		d.vLedControlReg.indicatorStatus |= 0b00000001
	} else {
		d.vLedControlReg.indicatorStatus &= 0b11111110
	}
}

// ConfigureLed with all possible configurations
func (d *Device) ConfigureLed(ilCurLim float32, ilOn bool, inCurLim float32, inOn bool) {
	d.setIlCr(ilCurLim)
	d.setIlOn(ilOn)
	d.setInCur(inCurLim)
	d.setInOn(inOn)
	lrEncoded := d.encodeLedReg()

	// write ControlReg and read full ControlReg
	d.writeByte(LedRegister, lrEncoded)
	d.readByte(LedRegister)
	d.decodeLedReg(d.buf[0])
}
