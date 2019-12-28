// +build atsamd51

package ili9341

import (
	"machine"
	"runtime/volatile"
)

type parallelDriver struct {
	d0 machine.Pin
	wr machine.Pin

	setPort *uint32
	setMask uint32

	clrPort *uint32
	clrMask uint32

	wrPortSet *uint32
	wrMaskSet uint32

	wrPortClr *uint32
	wrMaskClr uint32
}

func NewParallel(d0, wr, dc, cs, rst, rd machine.Pin) *Device {
	return &Device{
		dc:  dc,
		cs:  cs,
		rd:  rd,
		rst: rst,
		driver: &parallelDriver{
			d0: d0,
			wr: wr,
		},
	}
}

func (pd *parallelDriver) configure(config *Config) {
	output := machine.PinConfig{machine.PinOutput}
	for pin := pd.d0; pin < pd.d0+8; pin++ {
		pin.Configure(output)
		pin.Low()
	}
	pd.wr.Configure(output)
	pd.wr.High()

	pd.setPort, _ = pd.d0.PortMaskSet()
	pd.setMask = uint32(pd.d0) & 0x1f

	pd.clrPort, _ = (pd.d0).PortMaskClear()
	pd.clrMask = 0xFF << uint32(pd.d0)

	pd.wrPortSet, pd.wrMaskSet = pd.wr.PortMaskSet()
	pd.wrPortClr, pd.wrMaskClr = pd.wr.PortMaskClear()
}

func (pd *parallelDriver) write8(b byte) {
	volatile.StoreUint32(pd.clrPort, pd.clrMask)
	volatile.StoreUint32(pd.setPort, uint32(b)<<pd.setMask)
	volatile.StoreUint32(pd.wrPortClr, pd.wrMaskClr)
	volatile.StoreUint32(pd.wrPortSet, pd.wrMaskSet)
}

func (pd *parallelDriver) write16(data uint16) {
	// output the high byte
	volatile.StoreUint32(pd.clrPort, pd.clrMask)
	volatile.StoreUint32(pd.setPort, uint32(data>>8)<<pd.setMask)
	volatile.StoreUint32(pd.wrPortClr, pd.wrMaskClr)
	volatile.StoreUint32(pd.wrPortSet, pd.wrMaskSet)
	// output the low byte
	volatile.StoreUint32(pd.clrPort, pd.clrMask)
	volatile.StoreUint32(pd.setPort, uint32(byte(data))<<pd.setMask)
	volatile.StoreUint32(pd.wrPortClr, pd.wrMaskClr)
	volatile.StoreUint32(pd.wrPortSet, pd.wrMaskSet)
}

func (pd *parallelDriver) write16n(data uint16, n int) {
	setMaskHi := uint32(data>>8) << pd.setMask
	setMaskLo := uint32(byte(data)) << pd.setMask
	for i := 0; i < n; i++ {
		// output the high byte
		volatile.StoreUint32(pd.clrPort, pd.clrMask)
		volatile.StoreUint32(pd.setPort, setMaskHi)
		volatile.StoreUint32(pd.wrPortClr, pd.wrMaskClr)
		volatile.StoreUint32(pd.wrPortSet, pd.wrMaskSet)
		// output the low byte
		volatile.StoreUint32(pd.clrPort, pd.clrMask)
		volatile.StoreUint32(pd.setPort, setMaskLo)
		volatile.StoreUint32(pd.wrPortClr, pd.wrMaskClr)
		volatile.StoreUint32(pd.wrPortSet, pd.wrMaskSet)
	}
}

func (pd *parallelDriver) write16sl(data []uint16) {
	for _, d := range data {
		// output the high byte
		volatile.StoreUint32(pd.clrPort, pd.clrMask)
		volatile.StoreUint32(pd.setPort, uint32(d>>8)<<pd.setMask)
		volatile.StoreUint32(pd.wrPortClr, pd.wrMaskClr)
		volatile.StoreUint32(pd.wrPortSet, pd.wrMaskSet)
		// output the low byte
		volatile.StoreUint32(pd.clrPort, pd.clrMask)
		volatile.StoreUint32(pd.setPort, uint32(byte(d))<<pd.setMask)
		volatile.StoreUint32(pd.wrPortClr, pd.wrMaskClr)
		volatile.StoreUint32(pd.wrPortSet, pd.wrMaskSet)
	}
}
