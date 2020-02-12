package flash

import (
	"fmt"
	"time"
)

const (
	cmdRead            = 0x03 // Single Read
	cmdQuadRead        = 0x6B // 1 line address, 4 line data
	cmdReadJedecID     = 0x9f
	cmdPageProgram     = 0x02
	cmdQuadPageProgram = 0x32 // 1 line address, 4 line data
	cmdReadStatus      = 0x05
	cmdReadStatus2     = 0x35
	cmdWriteStatus     = 0x01
	cmdWriteStatus2    = 0x31
	cmdEnableReset     = 0x66
	cmdReset           = 0x99
	cmdWriteEnable     = 0x06
	cmdWriteDisable    = 0x04
	cmdEraseSector     = 0x20
	cmdEraseBlock      = 0xD8
	cmdEraseChip       = 0xC7
)

const (
	BlockSize  = 64 * 1024
	SectorSize = 4 * 1024
	PageSize   = 256
)

type Error uint8

const (
	_                          = iota
	ErrInvalidClockSpeed Error = iota
	ErrInvalidAddrRange
)

func (err Error) Error() string {
	switch err {
	case ErrInvalidClockSpeed:
		return "invalid clock speed"
	case ErrInvalidAddrRange:
		return "invalid address range"
	default:
		return "unspecified error"
	}
}

type JedecID struct {
	ManufID  uint8
	MemType  uint8
	Capacity uint8
}

func (id *JedecID) Uint32() uint32 {
	return uint32(id.ManufID)<<16 | uint32(id.MemType)<<8 | uint32(id.Capacity)
}

func (id *JedecID) String() string {
	return fmt.Sprintf("%06X", id.Uint32())
}

type SerialNumber uint64

func (sn SerialNumber) String() string {
	return fmt.Sprintf("%8X", uint64(sn))
}

type Device struct {
	transport transport
	attrs     Attrs
}

type transport interface {
	begin()
	supportQuadMode() bool
	setClockSpeed(hz uint32) (err error)
	runCommand(cmd byte) (err error)
	readCommand(cmd byte, rsp []byte) (err error)
	writeCommand(cmd byte, data []byte) (err error)
	eraseCommand(cmd byte, address uint32) (err error)
	readMemory(addr uint32, rsp []byte) (err error)
	writeMemory(addr uint32, data []byte) (err error)
}

type Attrs struct {
	TotalSize uint32
	//start_up_time_us uint16

	// Three response bytes to 0x9f JEDEC ID command.
	JedecID

	// Max clock speed for all operations and the fastest read mode.
	MaxClockSpeedMHz uint8

	// Bitmask for Quad Enable bit if present. 0x00 otherwise. This is for the
	// highest byte in the status register.
	QuadEnableBitMask uint8

	HasSectorProtection bool

	// Supports the 0x0b fast read command with 8 dummy cycles.
	SupportsFastRead bool

	// Supports the fast read, quad output command 0x6b with 8 dummy cycles.
	SupportsQSPI bool

	// Supports the quad input page program command 0x32. This is known as 1-1-4
	// because it only uses all four lines for data.
	SupportsQSPIWrites bool

	// Requires a separate command 0x31 to write to the second byte of the status
	// register. Otherwise two byte are written via 0x01.
	WriteStatusRegisterSplit bool

	// True when the status register is a single byte. This implies the Quad
	// Enable bit is in the first byte and the Read Status Register 2 command
	// (0x35) is unsupported.
	SingleStatusByte bool
}

func (dev *Device) Begin() (err error) {

	dev.transport.begin()

	// TODO: should check JEDEC ID against list of known devices
	/*
		if dev.ID, err = dev.ReadJEDEC(); err != nil {
			return err
		}
		println("JEDEC:", dev.ID.String())
	*/

	// We don't know what state the flash is in so wait for any remaining writes and then reset.

	var s byte // status

	// The write in progress bit should be low.
	for s, err = dev.ReadStatus(); (s & 0x01) > 0; s, err = dev.ReadStatus() {
		if err != nil {
			return err
		}
	}

	// The suspended write/erase bit should be low.
	for s, err = dev.ReadStatus2(); (s & 0x80) > 0; s, err = dev.ReadStatus2() {
		if err != nil {
			return err
		}
	}

	if err = dev.transport.runCommand(cmdEnableReset); err != nil {
		return err
	}
	if err = dev.transport.runCommand(cmdReset); err != nil {
		return err
	}

	// Wait 30us for the reset
	stop := time.Now().UnixNano() + int64(30*time.Microsecond)
	for stop > time.Now().UnixNano() {
	}

	// Speed up to max device frequency
	//_trans->setClockSpeed(_flash_dev->max_clock_speed_mhz*1000000UL);

	if err = dev.transport.runCommand(cmdWriteDisable); err != nil {
		return err
	}

	err = dev.WaitUntilReady()
	return err
}

func (dev *Device) ReadJEDEC() (JedecID, error) {
	jedecID := make([]byte, 3)
	if err := dev.transport.readCommand(cmdReadJedecID, jedecID); err != nil {
		return JedecID{}, err
	}
	return JedecID{jedecID[0], jedecID[1], jedecID[2]}, nil
}

func (dev *Device) ReadSerialNumber() (SerialNumber, error) {
	sn := make([]byte, 12)
	if err := dev.transport.readCommand(0x4B, sn); err != nil {
		return 0, err
	}
	return SerialNumber(uint64(sn[11]) | uint64(sn[10])<<0x8 |
		uint64(sn[9])<<0x10 | uint64(sn[8])<<0x18 | uint64(sn[7])<<0x20 |
		uint64(sn[6])<<0x28 | uint64(sn[5])<<0x30 | uint64(sn[4])<<0x38), nil
}

func (dev *Device) ReadBuffer(addr uint32, buf []byte) error {
	// TODO: check if Begin() was successful
	if err := dev.WaitUntilReady(); err != nil {
		return err
	}
	return dev.transport.readMemory(addr, buf)
}

func (dev *Device) WriteBuffer(addr uint32, buf []byte) (n int, err error) {
	remain := uint32(len(buf))
	idx := uint32(0)
	for remain > 0 {
		if err = dev.WaitUntilReady(); err != nil {
			return
		}
		if err = dev.WriteEnable(); err != nil {
			return
		}
		leftOnPage := PageSize - (addr & (PageSize - 1))
		toWrite := remain
		if leftOnPage < remain {
			toWrite = leftOnPage
		}
		if err = dev.transport.writeMemory(addr, buf[idx:idx+toWrite]); err != nil {
			return
		}
		idx += toWrite
		addr += toWrite
		remain -= toWrite
	}
	return len(buf) - int(remain), nil
}

func (dev *Device) WriteEnable() error {
	return dev.transport.runCommand(cmdWriteEnable)
}

func (dev *Device) EraseBlock(addr uint32) error {
	if err := dev.WaitUntilReady(); err != nil {
		return err
	}
	if err := dev.WriteEnable(); err != nil {
		return err
	}
	return dev.transport.eraseCommand(cmdEraseBlock, addr*BlockSize)
}

func (dev *Device) EraseSector(addr uint32) error {
	if err := dev.WaitUntilReady(); err != nil {
		return err
	}
	if err := dev.WriteEnable(); err != nil {
		return err
	}
	return dev.transport.eraseCommand(cmdEraseSector, addr*SectorSize)
}

func (dev *Device) EraseChip() error {
	if err := dev.WaitUntilReady(); err != nil {
		return err
	}
	if err := dev.WriteEnable(); err != nil {
		return err
	}
	return dev.transport.runCommand(cmdEraseChip)
}

// ReadStatus reads the value from status register 1 of the device
func (dev *Device) ReadStatus() (status byte, err error) {
	buf := make([]byte, 1)
	err = dev.transport.readCommand(cmdReadStatus, buf)
	return buf[0], err
}

// ReadStatus2 reads the value from status register 2 of the device
func (dev *Device) ReadStatus2() (status byte, err error) {
	buf := make([]byte, 1)
	err = dev.transport.readCommand(cmdReadStatus2, buf)
	return buf[0], err
}

func (dev *Device) WaitUntilReady() error {
	expire := time.Now().UnixNano() + int64(1*time.Second)
	for s, err := dev.ReadStatus(); (s & 0x03) > 0; s, err = dev.ReadStatus() {
		println("wait until ready status", s)
		if err != nil {
			return err
		}
		if time.Now().UnixNano() > expire {
			return fmt.Errorf("WaitUntilReady expired")
		}
	}
	return nil
}
