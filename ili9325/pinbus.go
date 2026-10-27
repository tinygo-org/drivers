package ili9325

import "tinygo.org/x/drivers/internal/pin"

// PinBus is a portable Bus that sets each data pin in turn. The pins must
// already be configured as outputs.
type PinBus struct {
	data [8]pin.OutputFunc
	wr   pin.OutputFunc
}

// NewPinBus returns a Bus using data pins d0 to d7 and the WR strobe.
func NewPinBus(d0, d1, d2, d3, d4, d5, d6, d7, wr pin.Output) *PinBus {
	return &PinBus{
		data: [8]pin.OutputFunc{d0.Set, d1.Set, d2.Set, d3.Set, d4.Set, d5.Set, d6.Set, d7.Set},
		wr:   wr.Set,
	}
}

func (b *PinBus) write(v byte) {
	for i, set := range b.data {
		set(v&(1<<i) != 0)
	}
	b.wr.Low()
	b.wr.High()
}

// WriteBytes writes each byte in turn, pulsing WR after each one.
func (b *PinBus) WriteBytes(buf []byte) {
	for _, v := range buf {
		b.write(v)
	}
}

// WriteRepeat writes v high byte first, n times.
func (b *PinBus) WriteRepeat(v uint16, n int) {
	for ; n > 0; n-- {
		b.write(byte(v >> 8))
		b.write(byte(v))
	}
}
