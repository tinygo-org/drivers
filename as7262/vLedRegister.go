package as7262

func (d *Device) Led(status bool) {
	var led byte
	if status {
		led = 0b00000111
	} else {
		led = 0b00000110
	}
	d.writeByte(LedRegister, led)
}
