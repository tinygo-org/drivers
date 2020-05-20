// +build atsamd51

package ili9341

import (
	"machine"
)

type spiDriver struct {
	bus machine.SPI
	dc  machine.Pin
	rst machine.Pin
	cs  machine.Pin
	rd  machine.Pin
}

func NewSpi(bus machine.SPI, dc, cs, rst machine.Pin) *Device {
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

//go:inline
func (pd *spiDriver) write8(b byte) {
	pd.bus.Tx([]byte{b}, nil)
}

//go:inline
func (pd *spiDriver) write8n(b byte, n int) {
	for i := 0; i < n; i++ {
		pd.write8(b)
	}
}

//go:inline
func (pd *spiDriver) write8sl(b []byte) {
	pd.bus.Tx(b, nil)
}

//go:inline
func (pd *spiDriver) write16(data uint16) {
	pd.bus.Transfer2((data << 8) | (data >> 8))
}

//go:inline
func (pd *spiDriver) write16n(data uint16, n int) {
	for i := 0; i < n; i++ {
		pd.write16(data)
	}
}

//go:inline
func (pd *spiDriver) write16sl(data []uint16) {
	for i, c := 0, len(data)-2; i < c; i += 2 {
		d := uint32((data[i+1]<<8)|(data[i+1]>>8)) << 16
		d |= uint32((data[i] << 8) | (data[i] >> 8))
		pd.bus.Transfer4(d)
	}

	for i, c := len(data)-2, len(data); i < c; i++ {
		pd.bus.Transfer2((data[i] << 8) | (data[i] >> 8))
	}
}
