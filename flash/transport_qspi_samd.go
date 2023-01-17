//go:build atsamd51

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
		trans: &qspiTransport{
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

	// Instruction frame for running sending a command to the device
	iframeRunCommand = 0x0 |
		sam.QSPI_INSTRFRAME_WIDTH_SINGLE_BIT_SPI |
		sam.QSPI_INSTRFRAME_ADDRLEN_24BITS |
		sam.QSPI_INSTRFRAME_INSTREN |
		(sam.QSPI_INSTRFRAME_TFRTYPE_READ << sam.QSPI_INSTRFRAME_TFRTYPE_Pos)

	// Instruction frame for running a command that returns data
	iframeReadCommand = 0x0 |
		sam.QSPI_INSTRFRAME_WIDTH_SINGLE_BIT_SPI |
		sam.QSPI_INSTRFRAME_ADDRLEN_24BITS |
		sam.QSPI_INSTRFRAME_INSTREN |
		sam.QSPI_INSTRFRAME_DATAEN |
		(sam.QSPI_INSTRFRAME_TFRTYPE_READ << sam.QSPI_INSTRFRAME_TFRTYPE_Pos)

	// Instruction frame to set up the device to read from memory
	iframeReadMemory = 0x0 |
		sam.QSPI_INSTRFRAME_WIDTH_QUAD_OUTPUT |
		sam.QSPI_INSTRFRAME_ADDRLEN_24BITS |
		sam.QSPI_INSTRFRAME_INSTREN |
		sam.QSPI_INSTRFRAME_DATAEN |
		sam.QSPI_INSTRFRAME_ADDREN |
		(8 << sam.QSPI_INSTRFRAME_DUMMYLEN_Pos) |
		(sam.QSPI_INSTRFRAME_TFRTYPE_READMEMORY << sam.QSPI_INSTRFRAME_TFRTYPE_Pos)

	// Instruction frame for running a command that requires parameter data
	iframeWriteCommand = 0x0 |
		sam.QSPI_INSTRFRAME_WIDTH_SINGLE_BIT_SPI |
		sam.QSPI_INSTRFRAME_ADDRLEN_24BITS |
		sam.QSPI_INSTRFRAME_INSTREN |
		(sam.QSPI_INSTRFRAME_TFRTYPE_WRITE << sam.QSPI_INSTRFRAME_TFRTYPE_Pos)

	// Instruction frame to set up the device for writing to memory
	iframeWriteMemory = 0x0 |
		sam.QSPI_INSTRFRAME_WIDTH_QUAD_OUTPUT |
		sam.QSPI_INSTRFRAME_ADDRLEN_24BITS |
		sam.QSPI_INSTRFRAME_INSTREN |
		sam.QSPI_INSTRFRAME_ADDREN |
		sam.QSPI_INSTRFRAME_DATAEN |
		(sam.QSPI_INSTRFRAME_TFRTYPE_WRITEMEMORY << sam.QSPI_INSTRFRAME_TFRTYPE_Pos)

	// Instruction frame for running an erase command that requires and address
	iframeEraseCommand = 0x0 |
		sam.QSPI_INSTRFRAME_WIDTH_SINGLE_BIT_SPI |
		sam.QSPI_INSTRFRAME_ADDRLEN_24BITS |
		sam.QSPI_INSTRFRAME_INSTREN |
		sam.QSPI_INSTRFRAME_ADDREN |
		(sam.QSPI_INSTRFRAME_TFRTYPE_WRITE << sam.QSPI_INSTRFRAME_TFRTYPE_Pos)
)

type qspiTransport struct {
	cs  machine.Pin
	sck machine.Pin
	d0  machine.Pin
	d1  machine.Pin
	d2  machine.Pin
	d3  machine.Pin
}

func (q qspiTransport) configure(config *DeviceConfig) {

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
	sam.QSPI.CTRLB.Set(sam.QSPI_CTRLB_MODE_MEMORY |
		(sam.QSPI_CTRLB_DATALEN_8BITS << sam.QSPI_CTRLB_DATALEN_Pos) |
		(sam.QSPI_CTRLB_CSMODE_LASTXFER << sam.QSPI_CTRLB_CSMODE_Pos))

	// enable the peripheral
	sam.QSPI.CTRLA.SetBits(sam.QSPI_CTRLA_ENABLE)
}

func (q qspiTransport) supportQuadMode() bool {
	return true
}

func (q qspiTransport) setClockSpeed(hz uint32) error {
	// The clock speed for the QSPI peripheral is controlled by a divider, so
	// we can't set the requested speed exactly. Instead we will increment the
	// divider until the speed is less than or equal to the speed requested.
	for div, freq := uint32(1), machine.CPUFrequency(); div < 256; div++ {
		if freq/div <= hz {
			sam.QSPI.BAUD.Set(div << sam.QSPI_BAUD_BAUD_Pos)
			return nil
		}
	}
	return ErrInvalidClockSpeed
}

func (q qspiTransport) runCommand(cmd byte) (err error) {
	q.runInstruction(cmd, iframeRunCommand)
	q.endTransfer()
	return
}

func (q qspiTransport) readCommand(cmd byte, buf []byte) (err error) {
	q.disableAndClearCache()
	q.runInstruction(cmd, iframeReadCommand)
	q.readInto(buf, 0)
	q.endTransfer()
	q.enableCache()
	return
}

func (q qspiTransport) readMemory(addr uint32, buf []byte) (err error) {
	if (addr + uint32(len(buf))) > (qspi_AHB_HI - qspi_AHB_LO) {
		return ErrInvalidAddrRange
	}
	q.disableAndClearCache()
	q.runInstruction(cmdQuadRead, iframeReadMemory)
	q.readInto(buf, addr)
	q.endTransfer()
	q.enableCache()
	return
}

func (q qspiTransport) writeCommand(cmd byte, data []byte) (err error) {
	var dataen uint32
	if len(data) > 0 {
		dataen = sam.QSPI_INSTRFRAME_DATAEN
	}
	q.disableAndClearCache()
	q.runInstruction(cmd, iframeWriteCommand|dataen)
	q.writeFrom(data, 0)
	q.endTransfer()
	q.enableCache()
	return
}

func (q qspiTransport) writeMemory(addr uint32, data []byte) (err error) {
	if (addr + uint32(len(data))) > (qspi_AHB_HI - qspi_AHB_LO) {
		return ErrInvalidAddrRange
	}
	q.disableAndClearCache()
	q.runInstruction(cmdQuadPageProgram, iframeWriteMemory)
	q.writeFrom(data, addr)
	q.endTransfer()
	q.enableCache()
	return
}

func (q qspiTransport) eraseCommand(cmd byte, addr uint32) (err error) {
	q.disableAndClearCache()
	sam.QSPI.INSTRADDR.Set(addr)
	q.runInstruction(cmd, iframeEraseCommand)
	q.endTransfer()
	q.enableCache()
	return
}

func (q qspiTransport) runInstruction(cmd byte, iframe uint32) {
	sam.QSPI.INSTRCTRL.Set(uint32(cmd))
	sam.QSPI.INSTRFRAME.Set(iframe)
	sam.QSPI.INSTRFRAME.Get() // dummy read for synchronization, as per datasheet
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

func (q qspiTransport) readInto(buf []byte, addr uint32) {
	var ptr = qspi_AHB_LO + uintptr(addr)
	for i := range buf {
		buf[i] = volatile.LoadUint8((*uint8)(unsafe.Pointer(ptr)))
		ptr++
	}
	/* // NB(bcg): for some reason this reads data that results from commands in
	   // a different byte order than the loop above, but works fine for reading
	   // from memory. Oddly, the above loop seems to work fine in both cases.
		ln := len(buf)
		sl := (*[1 << 28]byte)(unsafe.Pointer(uintptr(qspi_AHB_LO + addr)))[:ln:ln]
		copy(buf, sl)
	*/
}

func (q qspiTransport) writeFrom(buf []byte, addr uint32) {
	var ptr = qspi_AHB_LO + uintptr(addr)
	for i := range buf {
		volatile.StoreUint8((*uint8)(unsafe.Pointer(ptr)), buf[i])
		ptr++
	}
}
