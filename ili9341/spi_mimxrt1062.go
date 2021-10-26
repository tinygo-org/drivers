// +build mimxrt1062

package ili9341

import (
	"device/nxp"
	"machine"
)

const (
	statusTxDataRequest    = nxp.LPSPI_SR_TDF // Transmit data flag
	statusRxDataReady      = nxp.LPSPI_SR_RDF // Receive data flag
	statusWordComplete     = nxp.LPSPI_SR_WCF // Word Complete flag
	statusFrameComplete    = nxp.LPSPI_SR_FCF // Frame Complete flag
	statusTransferComplete = nxp.LPSPI_SR_TCF // Transfer Complete flag
	statusTransmitError    = nxp.LPSPI_SR_TEF // Transmit Error flag (FIFO underrun)
	statusReceiveError     = nxp.LPSPI_SR_REF // Receive Error flag (FIFO overrun)
	statusDataMatch        = nxp.LPSPI_SR_DMF // Data Match flag
	statusModuleBusy       = nxp.LPSPI_SR_MBF // Module Busy flag
	statusAll              = nxp.LPSPI_SR_TDF | nxp.LPSPI_SR_RDF | nxp.LPSPI_SR_WCF |
		nxp.LPSPI_SR_FCF | nxp.LPSPI_SR_TCF | nxp.LPSPI_SR_TEF | nxp.LPSPI_SR_REF |
		nxp.LPSPI_SR_DMF | nxp.LPSPI_SR_MBF
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
	if !pd.setWord(b, true, true) {
		panic("setWord failed!")
	}
}

func (pd *spiDriver) write8n(b byte, n int) {
	for i := 0; i < n; i++ {
		if !pd.setWord(b, i == 0, i+1 == n) {
			panic("setWord failed!")
		}
	}
}

func (pd *spiDriver) write8sl(b []byte) {
	for i, w := range b {
		if !pd.setWord(w, i == 0, i+1 == len(b)) {
			panic("setWord failed!")
		}
	}
}

func (pd *spiDriver) write16(data uint16) {
	if !pd.setWord(uint8(data>>8), true, false) {
		panic("setWord failed!")
	}
	if !pd.setWord(uint8(data), false, true) {
		panic("setWord failed!")
	}
}

func (pd *spiDriver) write16n(data uint16, n int) {
	for i := 0; i < n; i++ {
		if !pd.setWord(uint8(data>>8), i == 0, false) {
			panic("setWord failed!")
		}
		if !pd.setWord(uint8(data), false, i+1 == n) {
			panic("setWord failed!")
		}
	}
}

func (pd *spiDriver) write16sl(data []uint16) {
	for i, w := range data {
		if !pd.setWord(uint8(w>>8), i == 0, false) {
			panic("setWord failed!")
		}
		if !pd.setWord(uint8(w), false, i+1 == len(data)) {
			panic("setWord failed!")
		}
	}
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
func (pd *spiDriver) setWord(word uint8, first bool, last bool) (ok bool) {

	const tryMax = 10000

	canWrite := false
	if first {
		for pd.bus.Bus.SR.HasBits(statusModuleBusy) {
		} // wait for SPI busy bit to clear
		pd.bus.Bus.CR.SetBits(nxp.LPSPI_CR_RRF | nxp.LPSPI_CR_RTF) // flush FIFOs
		pd.bus.Bus.SR.Set(statusAll)                               // clear all status flags (W1C)
		pd.bus.Bus.TCR.SetBits(nxp.LPSPI_TCR_RXMSK)                // mask receive data
		canWrite = true
	} else {
		// wait for TX FIFO to not be full
		txFIFOSize := uint32(1) << ((pd.bus.Bus.PARAM.Get() & nxp.LPSPI_PARAM_TXFIFO_Msk) >> nxp.LPSPI_PARAM_TXFIFO_Pos)
		for i := 0; !canWrite && (i < tryMax); i++ {
			canWrite = ((pd.bus.Bus.FSR.Get() & nxp.LPSPI_FSR_TXCOUNT_Msk) >> nxp.LPSPI_FSR_TXCOUNT_Pos) < txFIFOSize
		}
	}
	if !canWrite {
		ok = false
		return ok
	}

	pd.bus.Bus.TDR.Set(uint32(word))

	if last {
		canReturn := false
		for i := 0; (!canReturn) && (i < tryMax); i++ {
			canReturn = pd.bus.Bus.SR.HasBits(nxp.LPSPI_SR_TCF)
		}
		if !canReturn {
			ok = false
			return ok // timeout
		}
	}

	ok = true
	return ok
}
