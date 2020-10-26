// +build atsamd51

package ili9341

import (
	"machine"

	"tinygo.org/x/drivers"
)

type spiDriver struct {
	bus drivers.SPI
}

var txBuf [2]uint8

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
	pd.bus.Transfer(b)
}

func (pd *spiDriver) write8n(b byte, n int) {
	panic("not impl")
}

func (pd *spiDriver) write8sl(b []byte) {
	pd.bus.Tx(b, nil)
}

func (pd *spiDriver) write16(data uint16) {
	txBuf[0] = uint8(data >> 8)
	txBuf[1] = uint8(data)
	pd.bus.Tx(txBuf[:], nil)
}

func (pd *spiDriver) write16n(data uint16, n int) {
	for i, c := 0, n; i < c; i++ {
		pd.write16(data)
	}
}

func (pd *spiDriver) write16sl(data []uint16) {
	for i, c := 0, len(data); i < c; i++ {
		pd.write16(data[i])
	}
}
