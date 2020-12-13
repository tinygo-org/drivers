package flash

import (
	"time"
)

const (
	// BlockSize is the number of bytes in a block for most/all NOR flash memory
	BlockSize = 64 * 1024

	// SectorSize is the number of bytes in a sector for most/all NOR flash memory
	SectorSize = 4 * 1024

	// PageSize is the number of bytes in a page for most/all NOR flash memory
	PageSize = 256
)

// Device represents a NOR flash memory device accessible using SPI
type Device struct {
	trans transport
	attrs Attrs
}

// DeviceConfig contains the parameters that can be set when configuring a
// flash memory device.
type DeviceConfig struct {
	Identifier DeviceIdentifier
}

// JedecID encapsules the ID values that unique identify a flash memory device.
type JedecID struct {
	ManufID  uint8
	MemType  uint8
	Capacity uint8
}

// Uint32 returns the JEDEC ID packed into a uint32
func (id JedecID) Uint32() uint32 {
	return uint32(id.ManufID)<<16 | uint32(id.MemType)<<8 | uint32(id.Capacity)
}

// SerialNumber represents a serial number read from a flash memory device
type SerialNumber uint64

// Attrs represent the differences in hardware characteristics and capabilities
// of various SPI flash memory devices.
type Attrs struct {

	// TotalSize is the number of bytes that the flash device can store
	TotalSize uint32

	// StartUp is the duration of time between when the device is reset and when
	// it is ready to operation
	StartUp time.Duration

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
	WriteStatusSplit bool

	// True when the status register is a single byte. This implies the Quad
	// Enable bit is in the first byte and the Read Status Register 2 command
	// (0x35) is unsupported.
	SingleStatusByte bool
}

// Configure sets up the device and the underlying transport mechanism.  The
// DeviceConfig argument allows the caller to specify an instance of the
// DeviceIdentifier interface that, if provided, will be used to retrieve the
// attributes of the device based on the JEDEC ID.
func (dev *Device) Configure(config *DeviceConfig) (err error) {

	dev.trans.configure(config)

	var id JedecID
	if id, err = dev.ReadJEDEC(); err != nil {
		return err
	}

	// try to ascertain the vendor-specific attributes of the chip using the
	// provided Identifier
	if config.Identifier != nil {
		dev.attrs = config.Identifier.Identify(id)
	} else {
		dev.attrs = Attrs{JedecID: id}
	}

	// We don't know what state the flash is in so wait for any remaining
	// writes and then reset.

	// The write in progress bit should be low.
	for s, err := dev.ReadStatus(); (s & 0x01) > 0; s, err = dev.ReadStatus() {
		if err != nil {
			return err
		}
	}
	// The suspended write/erase bit should be low.
	for s, err := dev.ReadStatus2(); (s & 0x80) > 0; s, err = dev.ReadStatus2() {
		if err != nil {
			return err
		}
	}
	// perform device reset
	if err := dev.trans.runCommand(cmdEnableReset); err != nil {
		return err
	}
	if err := dev.trans.runCommand(cmdReset); err != nil {
		return err
	}

	// Wait for the reset - 30us by default
	time.Sleep(30 * time.Microsecond)

	// Speed up to max device frequency
	// I propose a check here for max frequency, but not put that functionality directly into the driver.
	// Either that or we have to change the signature of the SPI interface in the machine package itself.
	if dev.attrs.MaxClockSpeedMHz > 0 {
		err := dev.trans.setClockSpeed(uint32(dev.attrs.MaxClockSpeedMHz) * 1e6)
		if err != nil {
			return err
		}
	}

	// Enable Quad Mode if available
	if dev.trans.supportQuadMode() && dev.attrs.QuadEnableBitMask > 0 {
		// Verify that QSPI mode is enabled.
		var status byte
		if dev.attrs.SingleStatusByte {
			status, err = dev.ReadStatus()
		} else {
			status, err = dev.ReadStatus2()
		}
		if err != nil {
			return err
		}
		// Check and set the quad enable bit.
		if status&dev.attrs.QuadEnableBitMask == 0 {
			if err := dev.WriteEnable(); err != nil {
				return err
			}
			fullStatus := []byte{0x00, dev.attrs.QuadEnableBitMask}
			if dev.attrs.WriteStatusSplit {
				err = dev.trans.writeCommand(cmdWriteStatus2, fullStatus[1:])
			} else if dev.attrs.SingleStatusByte {
				err = dev.trans.writeCommand(cmdWriteStatus, fullStatus[1:])
			} else {
				err = dev.trans.writeCommand(cmdWriteStatus, fullStatus)
			}
			if err != nil {
				return err
			}
		}
	}

	// disable sector protection if the chip has it
	if dev.attrs.HasSectorProtection {
		if err := dev.WriteEnable(); err != nil {
			return err
		}
		if err := dev.trans.writeCommand(cmdWriteStatus, []byte{0x00}); err != nil {
			return err
		}
	}

	// write disable
	if err := dev.trans.runCommand(cmdWriteDisable); err != nil {
		return err
	}
	return dev.WaitUntilReady()
}

// Attrs returns the attributes of the device determined from the most recent
// call to Configure(). If no call to Configure() has been made, this will be
// the zero value of the Attrs struct.
func (dev *Device) Attrs() Attrs {
	return dev.attrs
}

// ReadJEDEC reads the JEDEC ID from the device; this ID can then be used to
// ascertain the attributes of the chip from a list of known devices.
func (dev *Device) ReadJEDEC() (JedecID, error) {
	jedecID := make([]byte, 3)
	if err := dev.trans.readCommand(cmdReadJedecID, jedecID); err != nil {
		return JedecID{}, err
	}
	return JedecID{jedecID[0], jedecID[1], jedecID[2]}, nil
}

// ReadSerialNumber reads the serial numbers from the connected device.
// TODO: maybe check if byte order / endianess is correct, probably is not
func (dev *Device) ReadSerialNumber() (SerialNumber, error) {
	sn := make([]byte, 12)
	if err := dev.trans.readCommand(0x4B, sn); err != nil {
		return 0, err
	}
	return SerialNumber(uint64(sn[11]) | uint64(sn[10])<<0x8 |
		uint64(sn[9])<<0x10 | uint64(sn[8])<<0x18 | uint64(sn[7])<<0x20 |
		uint64(sn[6])<<0x28 | uint64(sn[5])<<0x30 | uint64(sn[4])<<0x38), nil
}

// Size returns the size of this memory, in bytes.
func (dev *Device) Size() int64 {
	if dev.attrs.TotalSize < 1 {
		// in case a DeviceIdentifier function wasn't used, use the capacity
		// specified in the JEDEC ID instead
		return int64(dev.attrs.Capacity)
	}
	return int64(dev.attrs.TotalSize)
}

// ReadAt satisfies the io.ReaderAt interface, and fills the provided buffer
// with memory read from the device starting at the provided address.
func (dev *Device) ReadAt(buf []byte, addr int64) (int, error) {
	if err := dev.WaitUntilReady(); err != nil {
		return 0, err
	}
	if err := dev.trans.readMemory(uint32(addr), buf); err != nil {
		return 0, err
	}
	return len(buf), nil
}

// WriteAt satisfies the io.WriterAt interface and writes data to the device,
// one page at a time, starting at the provided address. This method assumes
// that the destination is already erased.
func (dev *Device) WriteAt(buf []byte, addr int64) (n int, err error) {
	remain := uint32(len(buf))
	idx := uint32(0)
	loc := uint32(addr)
	for remain > 0 {
		if err = dev.WaitUntilReady(); err != nil {
			return
		}
		if err = dev.WriteEnable(); err != nil {
			return
		}
		leftOnPage := PageSize - (loc & (PageSize - 1))
		toWrite := remain
		if leftOnPage < remain {
			toWrite = leftOnPage
		}
		if err = dev.trans.writeMemory(loc, buf[idx:idx+toWrite]); err != nil {
			return
		}
		idx += toWrite
		loc += toWrite
		remain -= toWrite
	}
	return len(buf) - int(remain), nil
}

// WriteBlockSize returns the block size in which data can be written to
// memory. It can be used by a client to optimize writes, non-aligned writes
// should always work correctly.
// For SPI NOR flash this is the page size, usually/always 256.
func (dev *Device) WriteBlockSize() int64 {
	return PageSize
}

// EraseBlockSize returns the smallest erasable area on this particular chip
// in bytes. This is used for the block size in EraseBlocks.
// For SPI NOR flash this is the sector size, usually/always 4096.
func (dev *Device) EraseBlockSize() int64 {
	return SectorSize
}

// EraseBlocks erases the given number of blocks. An implementation may
// transparently coalesce ranges of blocks into larger bundles if the chip
// supports this. The start and len parameters are in block numbers, use
// EraseBlockSize to map addresses to blocks.
func (dev *Device) EraseBlocks(start, len int64) error {
	// TODO: maybe combine sector erase operations into block erase operations
	for i := start; i < start+len; i++ {
		if err := dev.EraseSector(uint32(i)); err != nil {
			return err
		}
	}
	return nil
}

func (dev *Device) WriteEnable() error {
	return dev.trans.runCommand(cmdWriteEnable)
}

// EraseBlock erases a block of memory at the specified index
func (dev *Device) EraseBlock(blockNumber uint32) error {
	if err := dev.WaitUntilReady(); err != nil {
		return err
	}
	if err := dev.WriteEnable(); err != nil {
		return err
	}
	return dev.trans.eraseCommand(cmdEraseBlock, blockNumber*BlockSize)
}

// EraseSector erases a sector of memory at the given index
func (dev *Device) EraseSector(sectorNumber uint32) error {
	if err := dev.WaitUntilReady(); err != nil {
		return err
	}
	if err := dev.WriteEnable(); err != nil {
		return err
	}
	return dev.trans.eraseCommand(cmdEraseSector, sectorNumber*SectorSize)
}

// EraseChip erases the entire flash memory chip
func (dev *Device) EraseAll() error {
	if err := dev.WaitUntilReady(); err != nil {
		return err
	}
	if err := dev.WriteEnable(); err != nil {
		return err
	}
	return dev.trans.runCommand(cmdEraseChip)
}

// ReadStatus reads the value from status register 1 of the device
func (dev *Device) ReadStatus() (status byte, err error) {
	buf := make([]byte, 1)
	err = dev.trans.readCommand(cmdReadStatus, buf)
	return buf[0], err
}

// ReadStatus2 reads the value from status register 2 of the device
func (dev *Device) ReadStatus2() (status byte, err error) {
	buf := make([]byte, 1)
	err = dev.trans.readCommand(cmdReadStatus2, buf)
	return buf[0], err
}

// WaitUntilReady queries the status register until the device is ready for the
// next operation.
func (dev *Device) WaitUntilReady() error {
	expire := time.Now().UnixNano() + int64(1*time.Second)
	for s, err := dev.ReadStatus(); (s & 0x03) > 0; s, err = dev.ReadStatus() {
		if err != nil {
			return err
		}
		if time.Now().UnixNano() > expire {
			return ErrWaitExpired
		}
	}
	return nil
}

const (
	cmdRead            = 0x03 // read memory using single-bit transfer
	cmdQuadRead        = 0x6B // read with 1 line address, 4 line data
	cmdReadJedecID     = 0x9F // read the JEDEC ID from the device
	cmdPageProgram     = 0x02 // write a page of memory using single-bit transfer
	cmdQuadPageProgram = 0x32 // write with 1 line address, 4 line data
	cmdReadStatus      = 0x05 // read status register 1
	cmdReadStatus2     = 0x35 // read status register 2
	cmdWriteStatus     = 0x01 // write status register 1
	cmdWriteStatus2    = 0x31 // write status register 2
	cmdEnableReset     = 0x66 // enable reset
	cmdReset           = 0x99 // perform reset
	cmdWriteEnable     = 0x06 // write-enable memory
	cmdWriteDisable    = 0x04 // write-protect memory
	cmdEraseSector     = 0x20 // erase a sector of memory
	cmdEraseBlock      = 0xD8 // erase a block of memory
	cmdEraseChip       = 0xC7 // erase the entire chip
)

type Error uint8

const (
	_                          = iota
	ErrInvalidClockSpeed Error = iota
	ErrInvalidAddrRange
	ErrWaitExpired
)

func (err Error) Error() string {
	switch err {
	case ErrInvalidClockSpeed:
		return "flash: invalid clock speed"
	case ErrInvalidAddrRange:
		return "flash: invalid address range"
	case ErrWaitExpired:
		return "flash: wait until ready expired"
	default:
		return "flash: unspecified error"
	}
}
