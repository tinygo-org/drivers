// +build stm32f4

package ili9341

import (
	"device/stm32"
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
	if !pd.bus.Bus.CR1.HasBits(stm32.SPI_CR1_SPE) {
		pd.bus.Bus.CR1.SetBits(stm32.SPI_CR1_SPE)
	}

	pd.setWord(b, true, true)

	pd.bus.Bus.CR1.ClearBits(stm32.SPI_CR1_SPE)
}

func (pd *spiDriver) write8n(b byte, n int) {
	if !pd.bus.Bus.CR1.HasBits(stm32.SPI_CR1_SPE) {
		pd.bus.Bus.CR1.SetBits(stm32.SPI_CR1_SPE)
	}

	for i := 0; i < n-1; i++ {
		pd.setWord(b, i == 0, i+1 == n)
	}

	pd.bus.Bus.CR1.ClearBits(stm32.SPI_CR1_SPE)
}

func (pd *spiDriver) write8sl(b []byte) {
	if !pd.bus.Bus.CR1.HasBits(stm32.SPI_CR1_SPE) {
		pd.bus.Bus.CR1.SetBits(stm32.SPI_CR1_SPE)
	}

	for i, w := range b {
		pd.setWord(w, i == 0, i+1 == len(b))
	}

	pd.bus.Bus.CR1.ClearBits(stm32.SPI_CR1_SPE)
}

func (pd *spiDriver) write16(data uint16) {
	if !pd.bus.Bus.CR1.HasBits(stm32.SPI_CR1_SPE) {
		pd.bus.Bus.CR1.SetBits(stm32.SPI_CR1_SPE)
	}

	pd.setWord(uint8(data>>8), true, false)
	pd.setWord(uint8(data), false, true)

	pd.bus.Bus.CR1.ClearBits(stm32.SPI_CR1_SPE)
}

func (pd *spiDriver) write16n(data uint16, n int) {
	if !pd.bus.Bus.CR1.HasBits(stm32.SPI_CR1_SPE) {
		pd.bus.Bus.CR1.SetBits(stm32.SPI_CR1_SPE)
	}

	for i := 0; i < n; i++ {
		pd.setWord(uint8(data>>8), i == 0, false)
		pd.setWord(uint8(data), false, i+1 == n)
	}

	pd.bus.Bus.CR1.ClearBits(stm32.SPI_CR1_SPE)
}

func (pd *spiDriver) write16sl(data []uint16) {
	if !pd.bus.Bus.CR1.HasBits(stm32.SPI_CR1_SPE) {
		pd.bus.Bus.CR1.SetBits(stm32.SPI_CR1_SPE)
	}

	for i, w := range data {
		pd.setWord(uint8(w>>8), i == 0, false)
		pd.setWord(uint8(w), false, i+1 == len(data))
	}

	pd.bus.Bus.CR1.ClearBits(stm32.SPI_CR1_SPE)
}

// puts a single 8-bit word in the SPI data register (DR).
// if first (first word being transmitted) is false, waits for the SPI transmit
// buffer empty bit (TXE) is set before putting the word in DR.
// if last (last word being transmitted) is true, waits for the SPI transmit
// buffer empty bit (TXE) is set and SPI bus busy bit (BSY) is clear before
// returning.
// for all wait operations, a fixed number of wait iterations (const tryMax) are
// performed before a timeout is assumed.
// if timeout occurs, returns false. otherwise, returns true.
func (pd *spiDriver) setWord(word uint8, first bool, last bool) bool {

	const tryMax = 10000

	canWrite := first
	for i := 0; (!canWrite) && (i < tryMax); i++ {
		canWrite = pd.bus.Bus.SR.HasBits(stm32.SPI_SR_TXE)
	}
	if !canWrite {
		return false // timeout
	}

	pd.bus.Bus.DR.Set(uint32(word))

	if last {
		canReturn := false
		for i := 0; (!canReturn) && (i < tryMax); i++ {
			canReturn = pd.bus.Bus.SR.HasBits(stm32.SPI_SR_TXE)
		}
		if !canReturn {
			return false // timeout
		}

		canReturn = false
		for i := 0; (!canReturn) && (i < tryMax); i++ {
			canReturn = !pd.bus.Bus.SR.HasBits(stm32.SPI_SR_BSY)
		}
		if !canReturn {
			return false // timeout
		}
	}

	return true
}
