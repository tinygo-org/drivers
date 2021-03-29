// +build atsamd51

package ili9341

import (
	"machine"
)

type spiDriver struct {
	bus machine.SPI
}

func NewSPI(bus machine.SPI, dc, cs, rst machine.Pin) *Device {
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
	pd.bus.Tx([]byte{b}, nil)
}

func (pd *spiDriver) write8n(b byte, n int) {
	for i := 0; i < n; i++ {
		pd.bus.Tx([]byte{b}, nil)
	}
}

func (pd *spiDriver) write8sl(b []byte) {
	pd.bus.Tx(b, nil)
}

func (pd *spiDriver) write16(data uint16) {
	pd.bus.Tx([]byte{uint8(data >> 8), uint8(data)}, nil)
}

func (pd *spiDriver) write16n(data uint16, n int) {
	for i := 0; i < n; i++ {
		pd.bus.Tx([]byte{uint8(data >> 8), uint8(data)}, nil)
	}
}

func (pd *spiDriver) write16sl(data []uint16) {
	for i, c := 0, len(data); i < c; i++ {
		pd.bus.Tx([]byte{uint8(data[i] >> 8), uint8(data[i])}, nil)
	}
}
