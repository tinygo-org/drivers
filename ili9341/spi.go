//go:build !atsamd51 && !atsame5x && !atsamd21

package ili9341

import (
	"machine"

	"tinygo.org/x/drivers"
)

var buf [64]byte

type spiDriver struct {
	bus drivers.SPI
}

func NewSPI(bus drivers.SPI, dc, cs, rst machine.Pin) *Device {
	return &Device{
		dc:  dc,
		cs:  cs,
		rst: rst,
		rd:  machine.NoPin,
		driver: &spiDriver{
			bus: bus,
		},
	}
}

func (pd *spiDriver) configure(config *Config) {
}

func (pd *spiDriver) write8(b byte) {
	buf[0] = b
	pd.bus.Tx(buf[:1], nil)
}

func (pd *spiDriver) write8n(b byte, n int) {
	buf[0] = b
	for i := 0; i < n; i++ {
		pd.bus.Tx(buf[:1], nil)
	}
}

func (pd *spiDriver) write8sl(b []byte) {
	pd.bus.Tx(b, nil)
}

func (pd *spiDriver) write16(data uint16) {
	buf[0] = uint8(data >> 8)
	buf[1] = uint8(data)
	pd.bus.Tx(buf[:2], nil)
}

func (pd *spiDriver) write16n(data uint16, n int) {
	for i := 0; i < len(buf); i += 2 {
		buf[i] = uint8(data >> 8)
		buf[i+1] = uint8(data)
	}

	for i := 0; i < (n >> 5); i++ {
		pd.bus.Tx(buf[:], nil)
	}

	pd.bus.Tx(buf[:n%64], nil)
}

func (pd *spiDriver) write16sl(data []uint16) {
	for i, c := 0, len(data); i < c; i++ {
		buf[0] = uint8(data[i] >> 8)
		buf[1] = uint8(data[i])
		pd.bus.Tx(buf[:2], nil)
	}
}
