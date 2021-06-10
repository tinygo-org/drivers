// +build !atsamd51,!atsamd21

package ili9341

import (
	"device/sam"
	"machine"
	"unsafe"

	dma "github.com/sago35/tinygo-dma"
)

var (
	dbg5 = machine.D5
	dbg6 = machine.D6
)

func init() {
	dbg5.Configure(machine.PinConfig{Mode: machine.PinOutput})
	dbg6.Configure(machine.PinConfig{Mode: machine.PinOutput})
}

var buf [64]byte

type spiDriver struct {
	bus machine.SPI
}

var (
	dmatx *dma.DMA
	desc  *dma.DMADescriptor
)

func NewSPI(bus machine.SPI, dc, cs, rst machine.Pin) *Device {
	from := make([]byte, 256)
	for i := range from {
		from[i] = byte(i)
	}

	dmatx = dma.NewDMA(func(d *dma.DMA) {
		d.Wait()
		return
	})
	dmatx.SetTrigger(dma.DMAC_CHANNEL_CHCTRLA_TRIGSRC_SERCOM1_TX)
	dmatx.SetTriggerAction(sam.DMAC_CHANNEL_CHCTRLA_TRIGACT_BURST)
	desc = dmatx.GetDescriptor()

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
	if len(b) > 64 {
		desc.UpdateDescriptor(dma.DescriptorConfig{
			SRC:      unsafe.Pointer(&b[0]),
			DST:      unsafe.Pointer(&pd.bus.Bus.DATA.Reg),
			SRCINC:   dma.DMAC_SRAM_BTCTRL_SRCINC_ENABLE,
			DSTINC:   dma.DMAC_SRAM_BTCTRL_DSTINC_DISABLE,
			SIZE:     uint32(len(b)), // Total size of DMA transfer
			BLOCKACT: 1,
		})
		dmatx.Start()
		return
	}

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
	if len(data) > 64 {
		desc.UpdateDescriptor(dma.DescriptorConfig{
			SRC:      unsafe.Pointer(&data[0]),
			DST:      unsafe.Pointer(&pd.bus.Bus.DATA.Reg),
			SRCINC:   dma.DMAC_SRAM_BTCTRL_SRCINC_ENABLE,
			DSTINC:   dma.DMAC_SRAM_BTCTRL_DSTINC_DISABLE,
			SIZE:     uint32(len(data) * 2), // Total size of DMA transfer
			BLOCKACT: 1,
		})
		dmatx.Start()
		return
	}
	for i, c := 0, len(data); i < c; i++ {
		buf[0] = uint8(data[i] >> 8)
		buf[1] = uint8(data[i])
		pd.bus.Tx(buf[:2], nil)
	}
}
