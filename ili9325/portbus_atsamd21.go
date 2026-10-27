//go:build atsamd21

package ili9325

import (
	"device/arm"
	"device/sam"
	"errors"
	"machine"
	"runtime/volatile"
)

// PortBus is a Bus for the SAMD21 that writes each byte with single cycle
// IOBUS stores. The data pins and WR must be on the same port and already be
// configured as outputs. WR is held low and high for at least 3 cycles, 62 ns
// at 48 MHz, to meet the ILI9325 datasheet V0.43 13.4.1 PWLW and PWHW of 50 ns.
type PortBus struct {
	outset   *volatile.Register32
	outclr   *volatile.Register32
	dataMask uint32
	wrMask   uint32
	setMask  [256]uint32
}

// NewPortBus returns a Bus using data pins d0 to d7 and the WR strobe.
func NewPortBus(data [8]machine.Pin, wr machine.Pin) (*PortBus, error) {
	group := uint8(wr) / 32
	b := &PortBus{wrMask: 1 << (uint8(wr) % 32)}
	for i, p := range data {
		if uint8(p)/32 != group {
			return nil, errors.New("ili9325: data and WR pins must be on the same port")
		}
		bit := uint32(1) << (uint8(p) % 32)
		b.dataMask |= bit
		for v := range b.setMask {
			if v&(1<<i) != 0 {
				b.setMask[v] |= bit
			}
		}
	}
	if b.dataMask&b.wrMask != 0 {
		return nil, errors.New("ili9325: WR pin must not be a data pin")
	}
	if group == 0 {
		b.outset, b.outclr = &sam.PORT_IOBUS.OUTSET0, &sam.PORT_IOBUS.OUTCLR0
	} else {
		b.outset, b.outclr = &sam.PORT_IOBUS.OUTSET1, &sam.PORT_IOBUS.OUTCLR1
	}
	return b, nil
}

// WriteBytes writes each byte in turn, pulsing WR after each one.
func (b *PortBus) WriteBytes(buf []byte) {
	outset, outclr, wr := b.outset, b.outclr, b.wrMask
	for _, v := range buf {
		set := b.setMask[v]
		outclr.Set(b.dataMask&^set | wr)
		outset.Set(set)
		arm.Asm("nop")
		outset.Set(wr)
		arm.Asm("nop\nnop")
	}
}

// WriteRepeat writes v high byte first, n times.
func (b *PortBus) WriteRepeat(v uint16, n int) {
	outset, outclr, wr := b.outset, b.outclr, b.wrMask
	hiSet, loSet := b.setMask[v>>8], b.setMask[v&0xff]
	hiClr, loClr := b.dataMask&^hiSet|wr, b.dataMask&^loSet|wr
	for ; n > 0; n-- {
		outclr.Set(hiClr)
		outset.Set(hiSet)
		arm.Asm("nop")
		outset.Set(wr)
		arm.Asm("nop\nnop")
		outclr.Set(loClr)
		outset.Set(loSet)
		arm.Asm("nop")
		outset.Set(wr)
		arm.Asm("nop\nnop")
	}
}
