// +build atsamd51

package flash

import (
	. "device/sam"
	"machine"
	"runtime/volatile"
	"unsafe"
)

// NewQSPI returns a pointer to a flash device that uses the QSPI peripheral to
// communicate with a serial memory chip.
func NewQSPI(cs, sck, d0, d1, d2, d3 machine.Pin) *Device {
	return &Device{
		transport: &qspi{
			cs:  cs,
			sck: sck,
			d0:  d0,
			d1:  d1,
			d2:  d2,
			d3:  d3,
		},
	}
}

const (
	// QSPI address space on SAMD51 is 0x04000000 to 0x05000000
	qspi_AHB_LO = 0x04000000
	qspi_AHB_HI = 0x05000000
)

type qspi struct {
	cs  machine.Pin
	sck machine.Pin
	d0  machine.Pin
	d1  machine.Pin
	d2  machine.Pin
	d3  machine.Pin
}

func (q qspi) begin() {

	// enable main clocks
	MCLK.APBCMASK.SetBits(MCLK_APBCMASK_QSPI_)
	MCLK.AHBMASK.SetBits(MCLK_AHBMASK_QSPI_)
	MCLK.AHBMASK.ClearBits(MCLK_AHBMASK_QSPI_2X_)

	QSPI.CTRLA.SetBits(QSPI_CTRLA_SWRST)

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
	QSPI.CTRLB.Reg = QSPI_CTRLB_MODE_MEMORY |
		(QSPI_CTRLB_DATALEN_Msk & (QSPI_CTRLB_DATALEN_8BITS << QSPI_CTRLB_DATALEN_Pos)) |
		(QSPI_CTRLB_CSMODE_Msk & (QSPI_CTRLB_CSMODE_LASTXFER << QSPI_CTRLB_CSMODE_Pos))

	// enable the peripheral
	QSPI.CTRLA.SetBits(QSPI_CTRLA_ENABLE)
}

func (q qspi) supportQuadMode() bool {
	return true
}

func (q qspi) setClockSpeed(hz uint32) error {
	if divider := machine.CPUFrequency() / hz; divider < 256 {
		QSPI.BAUD.Reg = QSPI_BAUD_BAUD_Msk & (divider << QSPI_BAUD_BAUD_Pos)
	}
	return ErrInvalidClockSpeed
}

func (q qspi) runCommand(cmd Command) (err error) {
	QSPI.INSTRCTRL.Set(uint32(cmd))
	QSPI.INSTRFRAME.Set(QSPI_INSTRFRAME_WIDTH_SINGLE_BIT_SPI |
		QSPI_INSTRFRAME_ADDRLEN_24BITS | QSPI_INSTRFRAME_INSTREN |
		(QSPI_INSTRFRAME_TFRTYPE_READ << QSPI_INSTRFRAME_TFRTYPE_Pos))
	QSPI.INSTRFRAME.Get() // dummy read for synchronization, as per datasheet
	q.endTransfer()
	return
}

func (q qspi) readCommand(cmd Command, buf []byte) (err error) {
	q.disableAndClearCache()
	QSPI.INSTRCTRL.Set(uint32(cmd))
	const iframe = QSPI_INSTRFRAME_WIDTH_SINGLE_BIT_SPI | QSPI_INSTRFRAME_DATAEN |
		QSPI_INSTRFRAME_ADDRLEN_24BITS | QSPI_INSTRFRAME_INSTREN |
		(QSPI_INSTRFRAME_TFRTYPE_READ << QSPI_INSTRFRAME_TFRTYPE_Pos)
	QSPI.INSTRFRAME.Set(iframe)
	QSPI.INSTRFRAME.Get() // dummy read for synchronization, as per datasheet
	var ptr uintptr = qspi_AHB_LO
	for i := range buf {
		buf[i] = volatile.LoadUint8((*uint8)(unsafe.Pointer(ptr)))
		ptr++
	}
	q.endTransfer()
	q.enableCache()
	return
}

func (q qspi) readMemory(addr uint32, buf []byte) (err error) {
	if (addr + uint32(len(buf))) > (qspi_AHB_HI - qspi_AHB_LO) {
		return ErrInvalidAddrRange
	}
	q.disableAndClearCache()
	QSPI.INSTRCTRL.Set(uint32(CmdQuadRead))
	const iframe = QSPI_INSTRFRAME_WIDTH_QUAD_OUTPUT | QSPI_INSTRFRAME_DATAEN |
		QSPI_INSTRFRAME_ADDRLEN_24BITS | QSPI_INSTRFRAME_INSTREN |
		QSPI_INSTRFRAME_ADDREN | (8 << QSPI_INSTRFRAME_DUMMYLEN_Pos) |
		(QSPI_INSTRFRAME_TFRTYPE_READMEMORY << QSPI_INSTRFRAME_TFRTYPE_Pos)
	QSPI.INSTRFRAME.Set(iframe)
	QSPI.INSTRFRAME.Get() // dummy read for synchronization, as per datasheet
	ln := len(buf)
	sl := (*[1 << 28]byte)(unsafe.Pointer(uintptr(qspi_AHB_LO + addr)))[:ln:ln]
	copy(buf, sl)
	q.endTransfer()
	q.enableCache()
	return
}

func (q qspi) writeCommand(cmd Command, data []byte) (err error) {
	iframe := uint32(QSPI_INSTRFRAME_WIDTH_SINGLE_BIT_SPI |
		QSPI_INSTRFRAME_ADDRLEN_24BITS | QSPI_INSTRFRAME_INSTREN |
		(QSPI_INSTRFRAME_TFRTYPE_WRITE << QSPI_INSTRFRAME_TFRTYPE_Pos))
	if len(data) > 0 {
		iframe |= QSPI_INSTRFRAME_DATAEN
	}
	q.disableAndClearCache()
	QSPI.INSTRCTRL.Set(uint32(cmd))
	QSPI.INSTRFRAME.Set(iframe)
	QSPI.INSTRFRAME.Get() // dummy read for synchronization, as per datasheet
	var ptr uintptr = qspi_AHB_LO
	for i := range data {
		volatile.StoreUint8((*uint8)(unsafe.Pointer(ptr)), data[i])
		ptr++
	}
	q.endTransfer()
	q.enableCache()
	return
}

func (q qspi) eraseCommand(cmd Command, address uint32) (err error) {
	panic("implement me")
}

func (q qspi) writeMemory(addr uint32, data []byte) (err error) {
	panic("implement me")
}

//go:inline
func (q qspi) enableCache() {
	CMCC.CTRL.SetBits(CMCC_CTRL_CEN)
}

//go:inline
func (q qspi) disableAndClearCache() {
	CMCC.CTRL.ClearBits(CMCC_CTRL_CEN)
	for CMCC.SR.HasBits(CMCC_SR_CSTS) {
	}
	CMCC.MAINT0.SetBits(CMCC_MAINT0_INVALL)
}

//go:inline
func (q qspi) endTransfer() {
	QSPI.CTRLA.Set(QSPI_CTRLA_ENABLE | QSPI_CTRLA_LASTXFER)
	for !QSPI.INTFLAG.HasBits(QSPI_INTFLAG_INSTREND) {
	}
	QSPI.INTFLAG.Set(QSPI_INTFLAG_INSTREND)
}
