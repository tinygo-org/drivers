// +build atsamd51

package flash

import (
	"device/sam"
	"machine"
	"runtime/volatile"
	"unsafe"
)

// NewQSPI returns a pointer to a flash device that uses the QSPI peripheral to
// communicate with a serial memory chip.
func NewQSPI(cs, sck, d0, d1, d2, d3 machine.Pin) *Device {
	return &Device{
		transport: &qspiTransport{
			cs:  cs,
			sck: sck,
			d0:  d0,
			d1:  d1,
			d2:  d2,
			d3:  d3,
		},
	}
}

// QSPI address space on SAMD51 is 0x04000000 to 0x05000000
const (
	// Low address of the QSPI address space on SAMD51
	qspi_AHB_LO = 0x04000000

	// High address of the QSPI address space on SAMD51
	qspi_AHB_HI = 0x05000000
)

type qspiTransport struct {
	cs  machine.Pin
	sck machine.Pin
	d0  machine.Pin
	d1  machine.Pin
	d2  machine.Pin
	d3  machine.Pin
}

func (q qspiTransport) begin() {

	// enable main clocks
	sam.MCLK.APBCMASK.SetBits(sam.MCLK_APBCMASK_QSPI_)
	sam.MCLK.AHBMASK.SetBits(sam.MCLK_AHBMASK_QSPI_)
	sam.MCLK.AHBMASK.ClearBits(sam.MCLK_AHBMASK_QSPI_2X_)

	sam.QSPI.CTRLA.SetBits(sam.QSPI_CTRLA_SWRST)

	// enable all pins to be PinCom
	q.d0.Configure(machine.PinConfig{Mode: machine.PinCom})
	q.d1.Configure(machine.PinConfig{Mode: machine.PinCom})
	q.d2.Configure(machine.PinConfig{Mode: machine.PinCom})
	q.d3.Configure(machine.PinConfig{Mode: machine.PinCom})
	q.cs.Configure(machine.PinConfig{Mode: machine.PinCom})
	q.sck.Configure(machine.PinConfig{Mode: machine.PinCom})

	// start out with 4Mhz
	// can ignore the error, 4Mhz is always a valid speed
	_ = q.setClockSpeed(4e6)

	// configure the CTRLB register
	sam.QSPI.CTRLB.Reg = sam.QSPI_CTRLB_MODE_MEMORY |
		(sam.QSPI_CTRLB_DATALEN_8BITS << sam.QSPI_CTRLB_DATALEN_Pos) |
		(sam.QSPI_CTRLB_CSMODE_LASTXFER << sam.QSPI_CTRLB_CSMODE_Pos)

	// enable the peripheral
	sam.QSPI.CTRLA.SetBits(sam.QSPI_CTRLA_ENABLE)
}

func (q qspiTransport) supportQuadMode() bool {
	return true
}

func (q qspiTransport) setClockSpeed(hz uint32) error {
	if divider := machine.CPUFrequency() / hz; divider < 256 {
		sam.QSPI.BAUD.Reg = sam.QSPI_BAUD_BAUD_Msk & (divider << sam.QSPI_BAUD_BAUD_Pos)
	}
	return ErrInvalidClockSpeed
}

func (q qspiTransport) runCommand(cmd byte) (err error) {
	sam.QSPI.INSTRCTRL.Set(uint32(cmd))
	sam.QSPI.INSTRFRAME.Set(sam.QSPI_INSTRFRAME_WIDTH_SINGLE_BIT_SPI |
		sam.QSPI_INSTRFRAME_ADDRLEN_24BITS | sam.QSPI_INSTRFRAME_INSTREN |
		(sam.QSPI_INSTRFRAME_TFRTYPE_READ << sam.QSPI_INSTRFRAME_TFRTYPE_Pos))
	sam.QSPI.INSTRFRAME.Get() // dummy read for synchronization, as per datasheet
	q.endTransfer()
	return
}

func (q qspiTransport) readCommand(cmd byte, buf []byte) (err error) {
	q.disableAndClearCache()
	sam.QSPI.INSTRCTRL.Set(uint32(cmd))
	const iframe = sam.QSPI_INSTRFRAME_WIDTH_SINGLE_BIT_SPI | sam.QSPI_INSTRFRAME_DATAEN |
		sam.QSPI_INSTRFRAME_ADDRLEN_24BITS | sam.QSPI_INSTRFRAME_INSTREN |
		(sam.QSPI_INSTRFRAME_TFRTYPE_READ << sam.QSPI_INSTRFRAME_TFRTYPE_Pos)
	sam.QSPI.INSTRFRAME.Set(iframe)
	sam.QSPI.INSTRFRAME.Get() // dummy read for synchronization, as per datasheet
	var ptr uintptr = qspi_AHB_LO
	for i := range buf {
		buf[i] = volatile.LoadUint8((*uint8)(unsafe.Pointer(ptr)))
		ptr++
	}
	q.endTransfer()
	q.enableCache()
	return
}

func (q qspiTransport) readMemory(addr uint32, buf []byte) (err error) {
	if (addr + uint32(len(buf))) > (qspi_AHB_HI - qspi_AHB_LO) {
		return ErrInvalidAddrRange
	}
	q.disableAndClearCache()
	sam.QSPI.INSTRCTRL.Set(uint32(cmdQuadRead))
	const iframe = sam.QSPI_INSTRFRAME_WIDTH_QUAD_OUTPUT | sam.QSPI_INSTRFRAME_DATAEN |
		sam.QSPI_INSTRFRAME_ADDRLEN_24BITS | sam.QSPI_INSTRFRAME_INSTREN |
		sam.QSPI_INSTRFRAME_ADDREN | (8 << sam.QSPI_INSTRFRAME_DUMMYLEN_Pos) |
		(sam.QSPI_INSTRFRAME_TFRTYPE_READMEMORY << sam.QSPI_INSTRFRAME_TFRTYPE_Pos)
	sam.QSPI.INSTRFRAME.Set(iframe)
	sam.QSPI.INSTRFRAME.Get() // dummy read for synchronization, as per datasheet
	ln := len(buf)
	sl := (*[1 << 28]byte)(unsafe.Pointer(uintptr(qspi_AHB_LO + addr)))[:ln:ln]
	copy(buf, sl)
	q.endTransfer()
	q.enableCache()
	return
}

func (q qspiTransport) writeCommand(cmd byte, data []byte) (err error) {
	iframe := uint32(sam.QSPI_INSTRFRAME_WIDTH_SINGLE_BIT_SPI |
		sam.QSPI_INSTRFRAME_ADDRLEN_24BITS | sam.QSPI_INSTRFRAME_INSTREN |
		(sam.QSPI_INSTRFRAME_TFRTYPE_WRITE << sam.QSPI_INSTRFRAME_TFRTYPE_Pos))
	if len(data) > 0 {
		iframe |= sam.QSPI_INSTRFRAME_DATAEN
	}
	q.disableAndClearCache()
	sam.QSPI.INSTRCTRL.Set(uint32(cmd))
	sam.QSPI.INSTRFRAME.Set(iframe)
	sam.QSPI.INSTRFRAME.Get() // dummy read for synchronization, as per datasheet
	var ptr uintptr = qspi_AHB_LO
	for i := range data {
		volatile.StoreUint8((*uint8)(unsafe.Pointer(ptr)), data[i])
		ptr++
	}
	q.endTransfer()
	q.enableCache()
	return
}

func (q qspiTransport) eraseCommand(cmd byte, address uint32) (err error) {
	panic("implement me")
}

func (q qspiTransport) writeMemory(addr uint32, data []byte) (err error) {
	panic("implement me")
}

func (q qspiTransport) enableCache() {
	sam.CMCC.CTRL.SetBits(sam.CMCC_CTRL_CEN)
}

func (q qspiTransport) disableAndClearCache() {
	sam.CMCC.CTRL.ClearBits(sam.CMCC_CTRL_CEN)
	for sam.CMCC.SR.HasBits(sam.CMCC_SR_CSTS) {
	}
	sam.CMCC.MAINT0.SetBits(sam.CMCC_MAINT0_INVALL)
}

func (q qspiTransport) endTransfer() {
	sam.QSPI.CTRLA.Set(sam.QSPI_CTRLA_ENABLE | sam.QSPI_CTRLA_LASTXFER)
	for !sam.QSPI.INTFLAG.HasBits(sam.QSPI_INTFLAG_INSTREND) {
	}
	sam.QSPI.INTFLAG.Set(sam.QSPI_INTFLAG_INSTREND)
}
