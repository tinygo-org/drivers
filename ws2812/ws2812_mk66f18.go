//go:build mk66f18

package ws2812

// portMasks returns the bit-band aliases of the pin's PSOR and PCOR bits,
// so writing 1 sets or clears only this pin.
func (d Device) portMasks() (portSet *uint32, maskSet uint32, portClear *uint32, maskClear uint32) {
	p := d.Pin.Fast()
	return &p.PSOR.Reg, 1, &p.PCOR.Reg, 1
}
