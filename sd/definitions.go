package sd

import (
	"bytes"
	"encoding/binary"
	"io"
	"strconv"
	"time"
)

type CardKind uint8

const (
	// card types
	TypeSD1  CardKind = 1 // Standard capacity V1 SD card
	TypeSD2  CardKind = 2 // Standard capacity V2 SD card
	TypeSDHC CardKind = 3 // High Capacity SD card
)

type CID struct {
	ManufacturerID   uint8   // 0:1
	OEMApplicationID uint16  // 1:3
	prodName         [5]byte // 3:8
	// productRevision n.m
	productRev          byte   // 8:9
	ProductSerialNumber uint32 // 9:13
	// Manufacturing date bitfield:
	//  - yearhi=0:4
	//  - reserved=4:8
	//  - month=8:12
	//  - yearlo=12:16
	date [2]byte // 13:15
}

func DecodeCID(b []byte) (CID, error) {
	if len(b) < 16 {
		return CID{}, io.ErrShortBuffer
	}
	cid := CID{
		ManufacturerID:      b[0],
		OEMApplicationID:    binary.BigEndian.Uint16(b[1:3]),
		prodName:            [5]byte{b[3], b[4], b[5], b[6], b[7]},
		productRev:          b[8],
		ProductSerialNumber: binary.BigEndian.Uint32(b[9:13]),
		date:                [2]byte{b[13], b[14]},
	}

	return cid, nil
}

func (c *CID) ProductName() string {
	return string(upToNull(c.prodName[:]))
}

func (c *CID) ProductRevision() (n, m uint8) {
	return c.productRev >> 4, c.productRev & 0x0F
}

// CSD is the Card Specific Data register, a 128-bit (16-byte) register that defines how
// the SD card standard communicates with the memory field or register. This type is
// shared among V1 and V2 type devices.
type CSD struct {
	data [16]byte
}

type CSDv1 struct {
	CSD
}

type CSDv2 struct {
	CSD
}

func DecodeCSD(b []byte) (CSD, error) {
	if len(b) < 16 {
		return CSD{}, io.ErrShortBuffer
	}
	csd := CSD{}
	copy(csd.data[:], b)
	return csd, nil
}

// CSDStructure returns the version of the CSD structure.
func (c *CSD) CSDStructure() uint8 { return c.data[0] >> 6 }

func (c CSD) MustV1() CSDv1 {
	if c.CSDStructure() != 0 {
		panic("CSD is not version 1.0")
	}
	return CSDv1{CSD: c}
}

func (c CSD) MustV2() CSDv2 {
	if c.CSDStructure() != 1 {
		panic("CSD is not version 2.0")
	}
	return CSDv2{CSD: c}
}

func (c *CSD) RawCopy() [16]byte { return c.data }

// TAAC returns the Time Access Attribute Class (data read access-time-1).
func (c *CSD) TAAC() TAAC { return TAAC(c.data[1]) }

// NSAC returns the Data Read Access-time 2 in CLK cycles (NSAC*100).
func (c *CSD) NSAC() NSAC { return NSAC(c.data[2]) }

// TransferSpeed returns the Max Data Transfer Rate. Either 0x32 or 0x5A.
func (c *CSD) TransferSpeed() TransferSpeed { return TransferSpeed(c.data[3]) }

// CommandClasses returns the supported Card Command Classes.
// This is a bitfield, each bit position indicates whether the
func (c *CSD) CommandClasses() CommandClasses {
	return CommandClasses(uint16(c.data[4])<<4 | uint16(c.data[5]&0xf0)>>4)
}

// ReadBlockLen returns the Max Read Data Block Length in bytes.
func (c *CSD) ReadBlockLen() uint16 { return 1 << (c.data[5] & 0x0F) }

// AllowsReadBlockPartial should always return true. Indicates that
func (c *CSD) AllowsReadBlockPartial() bool { return c.data[6]&(1<<7) != 0 }

// AllowsWriteBlockMisalignment defines if the data block to be written by one command
// can be spread over more than one physical block of the memory device.
func (c *CSD) AllowsWriteBlockMisalignment() bool { return c.data[6]&(1<<6) != 0 }

// AllowsReadBlockMisalignment defines if the data block to be read by one command
// can be spread over more than one physical block of the memory device.
func (c *CSD) AllowsReadBlockMisalignment() bool { return c.data[6]&(1<<5) != 0 }

// CRC7 returns the CRC read for this CSD. May be invalid. Use [IsValid] to check validity of CRC7+Always1 fields.
func (c *CSD) CRC7() uint8 { return c.data[15] & 0b111_1111 }

// IsValid checks if the CRC and always1 fields are expected values.
func (c *CSD) IsValid() bool {
	// Compare last byte with CRC and also the always1 bit.
	got := CRC7(c.data[:15])
	return got|(1<<7) == c.data[15]
}

// ImplementsDSR defines if the configurable driver stage is integrated on the card.
func (c *CSD) ImplementsDSR() bool { return c.data[6]&(1<<4) != 0 }

// EraseSectorSizeInBlocks represents how much memory is erased in an erase
// command in multiple of block size.
func (c *CSD) EraseSectorSizeInBlocks() uint8 {
	return 1 + ((c.data[10]&0b11_1111)<<1 | (c.data[11] >> 7))
}

// EraseBlockEnabled defines granularity of unit size of data to be erased.
// If enabled the erase operation can erase either one or multiple units of 512 bytes.
func (c *CSD) EraseBlockEnabled() bool { return (c.data[10]>>6)&1 != 0 }

func (c *CSD) ReadToWriteFactor() uint8 { return (c.data[12] >> 2) & 0b111 }

// WriteProtectGroupSizeInSectors indicates the size of a write protected
// group in multiple of erasable sectors.
func (c *CSD) WriteProtectGroupSizeInSectors() uint8 {
	return 1 + (c.data[11] & 0b111_1111)
}

// WriteBlockLength represents maximum write data block length in bytes.
func (c *CSD) WriteBlockLength() uint16 {
	return 1 << ((c.data[12]&0b11)<<2 | (c.data[13] >> 6))
}

// WriteGroupEnabled indicates if write group protection is available.
func (c *CSD) WriteGroupEnabled() bool { return c.data[12]&(1<<7) != 0 }

// AllowsWritePartial Defines whether partial block sizes can be used in write block sizes.
func (c *CSD) AllowsWritePartial() bool { return c.data[13]&(1<<5) != 0 }

// FileFormat returns the file format on the card. This field is read-only for ROM.
func (c *CSD) FileFormat() FileFormat { return FileFormat(c.data[14]>>2) & 0b11 }

// TmpWriteProtected indicates temporary protection over the entire card content from being overwritten or erased.
func (c *CSD) TmpWriteProtected() bool { return c.data[14]&(1<<4) != 0 }

// PermWriteProtected indicates permanent protecttion of entire card content against overwriting or erasing (write+erase permanently disabled).
func (c *CSD) PermWriteProtected() bool { return c.data[14]&(1<<5) != 0 }

// IsCopy whether contents are original or have been copied.
func (c *CSD) IsCopy() bool { return c.data[14]&(1<<6) != 0 }

func (c *CSD) FileFormatGroup() bool { return c.data[14]&(1<<7) != 0 }

func (c *CSD) DeviceCapacity() (size uint64) {
	switch c.CSDStructure() {
	case 0:
		v1 := c.MustV1()
		size = uint64(v1.DeviceCapacity())
	case 1:
		v2 := c.MustV2()
		size = v2.DeviceCapacity()
	}
	return size
}

// After byte 5 CSDv1 and CSDv2  differ in structure at some fields.

// DeviceCapacity returns the device capacity in bytes.
func (c *CSDv2) DeviceCapacity() uint64 {
	csize := c.csize()
	return uint64(csize) * 512_000
}

func (c *CSDv2) csize() uint32 {
	return uint32(c.data[7]>>2)<<16 | uint32(c.data[8])<<8 | uint32(c.data[9])
}

// DeviceCapacity returns the total memory capacity of the SDCard in bytes. Max is 2GB for V1.
func (c *CSDv1) DeviceCapacity() uint32 {
	mult := c.mult()
	csize := c.csize()
	blklen := c.ReadBlockLen()
	blockNR := uint32(csize+1) * uint32(mult)
	return blockNR * uint32(blklen)
}

func (c *CSDv1) csize() uint16 {
	// Jesus, why did SD make this so complicated?
	return uint16(c.data[8]>>6) | uint16(c.data[7])<<2 | uint16(c.data[6]&0b11)<<10
}

// mult is a factor for computing total device size with csize and csizemult.
func (c *CSDv1) mult() uint16 { return 1 << (2 + c.csizemult()) }

func (c *CSDv1) csizemult() uint8 {
	return (c.data[9]&0b11)<<1 | (c.data[10] >> 7)
}

// VddReadCurrent indicates min and max values for read power supply currents.
//   - values min: 0=0.5mA; 1=1mA; 2=5mA; 3=10mA; 4=25mA; 5=35mA; 6=60mA; 7=100mA
//   - values max: 0=1mA; 1=5mA; 2=10mA; 3=25mA; 4=35mA; 5=45mA; 6=80mA; 7=200mA
func (c *CSDv1) VddReadCurrent() (min, max uint8) {
	return (c.data[8] >> 3) & 0b111, c.data[8] & 0b111
}

// VddWriteCurrent indicates min and max values for write power supply currents.
//   - values min: 0=0.5mA; 1=1mA; 2=5mA; 3=10mA; 4=25mA; 5=35mA; 6=60mA; 7=100mA
//   - values max: 0=1mA; 1=5mA; 2=10mA; 3=25mA; 4=35mA; 5=45mA; 6=80mA; 7=200mA
func (c *CSDv1) VddWriteCurrent() (min, max uint8) {
	return c.data[9] >> 5, (c.data[9] >> 3) & 0b111
}

func (c *CSD) String() string {
	version := c.CSDStructure() + 1
	if version > 2 {
		return "<unsupported CSD version>"
	}
	const delim = '\n'
	buf := make([]byte, 0, 64)
	buf = c.appendf(buf, delim)
	return string(buf)
}

func (c *CSDv1) String() string { return c.CSD.String() }

func (c *CSDv2) String() string { return c.CSD.String() }

func (c *CSD) appendf(b []byte, delim byte) []byte {
	b = appendnum(b, "Version", uint64(c.CSDStructure()+1), delim)
	b = appendnum(b, "Capacity(bytes)", c.DeviceCapacity(), delim)
	b = appendnum(b, "TimeAccess_ns", uint64(c.TAAC().AccessTime()), delim)
	b = appendnum(b, "NSAC", uint64(c.NSAC()), delim)
	b = appendnum(b, "Tx_kb/s", uint64(c.TransferSpeed().RateKilobits()), delim)
	b = appendnum(b, "CCC", uint64(c.CommandClasses()), delim)
	b = appendnum(b, "ReadBlockLen", uint64(c.ReadBlockLen()), delim)
	b = appendbit(b, "ReadBlockPartial", c.AllowsReadBlockPartial(), delim)
	b = appendbit(b, "AllowWriteBlockMisalignment", c.AllowsWriteBlockMisalignment(), delim)
	b = appendbit(b, "AllowReadBlockMisalignment", c.AllowsReadBlockMisalignment(), delim)
	b = appendbit(b, "ImplementsDSR", c.ImplementsDSR(), delim)
	b = appendnum(b, "WProtectNumSectors", uint64(c.WriteProtectGroupSizeInSectors()), delim)
	b = appendnum(b, "WriteBlockLen", uint64(c.WriteBlockLength()), delim)
	b = appendbit(b, "WGrpEnable", c.WriteGroupEnabled(), delim)
	b = appendbit(b, "WPartialAllow", c.AllowsWritePartial(), delim)
	b = append(b, "FileFmt:"...)
	b = append(b, c.FileFormat().String()...)
	b = append(b, delim)
	b = appendbit(b, "TmpWriteProtect", c.TmpWriteProtected(), delim)
	b = appendbit(b, "PermWriteProtect", c.PermWriteProtected(), delim)
	b = appendbit(b, "IsCopy", c.IsCopy(), delim)
	b = appendbit(b, "FileFormatGrp", c.FileFormatGroup(), delim)
	return b
}

func appendnum(b []byte, label string, n uint64, delim byte) []byte {
	b = append(b, label...)
	b = append(b, ':')
	b = strconv.AppendUint(b, n, 10)
	b = append(b, delim)
	return b
}

func appendbit(b []byte, label string, n bool, delim byte) []byte {
	b = append(b, label...)
	b = append(b, ':')
	b = append(b, '0'+b2u8(n))
	b = append(b, delim)
	return b
}

func upToNull(buf []byte) []byte {
	nullIdx := bytes.IndexByte(buf, 0)
	if nullIdx < 0 {
		return buf
	}
	return buf[:nullIdx]
}

const (
	CMD0_GO_IDLE_STATE              = 0
	CMD1_SEND_OP_CND                = 1
	CMD2_ALL_SEND_CID               = 2
	CMD3_SEND_RELATIVE_ADDR         = 3
	CMD4_SET_DSR                    = 4
	CMD6_SWITCH_FUNC                = 6
	CMD7_SELECT_DESELECT_CARD       = 7
	CMD8_SEND_IF_COND               = 8
	CMD9_SEND_CSD                   = 9
	CMD10_SEND_CID                  = 10
	CMD12_STOP_TRANSMISSION         = 12
	CMD13_SEND_STATUS               = 13
	CMD15_GO_INACTIVE_STATE         = 15
	CMD16_SET_BLOCKLEN              = 16
	CMD17_READ_SINGLE_BLOCK         = 17
	CMD18_READ_MULTIPLE_BLOCK       = 18
	CMD24_WRITE_BLOCK               = 24
	CMD25_WRITE_MULTIPLE_BLOCK      = 25
	CMD27_PROGRAM_CSD               = 27
	CMD28_SET_WRITE_PROT            = 28
	CMD29_CLR_WRITE_PROT            = 29
	CMD30_SEND_WRITE_PROT           = 30
	CMD32_ERASE_WR_BLK_START_ADDR   = 32
	CMD33_ERASE_WR_BLK_END_ADDR     = 33
	CMD38_ERASE                     = 38
	CMD42_LOCK_UNLOCK               = 42
	CMD55_APP_CMD                   = 55
	CMD56_GEN_CMD                   = 56
	CMD58_READ_OCR                  = 58
	CMD59_CRC_ON_OFF                = 59
	ACMD6_SET_BUS_WIDTH             = 6
	ACMD13_SD_STATUS                = 13
	ACMD22_SEND_NUM_WR_BLOCKS       = 22
	ACMD23_SET_WR_BLK_ERASE_COUNT   = 23
	ACMD41_SD_APP_OP_COND           = 41
	ACMD42_SET_CLR_CARD_DETECT      = 42
	ACMD51_SEND_SCR                 = 51
	ACMD18_SECURE_READ_MULTI_BLOCK  = 18
	ACMD25_SECURE_WRITE_MULTI_BLOCK = 25
	ACMD26_SECURE_WRITE_MKB         = 26
	ACMD38_SECURE_ERASE             = 38
	ACMD43_GET_MKB                  = 43
	ACMD44_GET_MID                  = 44
	ACMD45_SET_CER_RN1              = 45
	ACMD46_SET_CER_RN2              = 46
	ACMD47_SET_CER_RES2             = 47
	ACMD48_SET_CER_RES1             = 48
	ACMD49_CHANGE_SECURE_AREA       = 49
)

// CSD enum types.
type (
	TransferSpeed  uint8
	TAAC           uint8
	FileFormat     uint8
	CommandClasses uint16
	NSAC           uint8
)

const (
	FileFmtPartition FileFormat = iota // Hard disk like file system with partition table.
	FileFmtDOSFAT                      // DOS FAT (floppy like)
	FileFmtUFF                         // Universal File Format
	FileFmtUnknown
)

func (ff FileFormat) String() (s string) {
	switch ff {
	case FileFmtPartition:
		s = "partition"
	case FileFmtDOSFAT:
		s = "DOS/FAT"
	case FileFmtUFF:
		s = "UFF"
	case FileFmtUnknown:
		s = "unknown"
	default:
		s = "<invalid format>"
	}
	return s
}

var log10table = [...]int64{
	1,
	10,
	100,
	1000,
	10000,
	100000,
	1000000,
}

// RateMegabits returns the transfer rate in megabits per second.
func (t TransferSpeed) RateKilobits() int64 {
	return 100 * log10table[t&0b111]
}

func (t TAAC) AccessTime() (d time.Duration) {
	return time.Duration(log10table[t&0b111]) * time.Nanosecond
}

const (
	_CMD_TIMEOUT = 100

	_R1_IDLE_STATE           = 1 << 0
	_R1_ERASE_RESET          = 1 << 1
	_R1_ILLEGAL_COMMAND      = 1 << 2
	_R1_COM_CRC_ERROR        = 1 << 3
	_R1_ERASE_SEQUENCE_ERROR = 1 << 4
	_R1_ADDRESS_ERROR        = 1 << 5
	_R1_PARAMETER_ERROR      = 1 << 6
)

type response1 uint8

func (r response1) IsIdle() bool          { return r&_R1_IDLE_STATE != 0 }
func (r response1) IllegalCmdError() bool { return r&_R1_ILLEGAL_COMMAND != 0 }
func (r response1) CRCError() bool        { return r&_R1_COM_CRC_ERROR != 0 }
func (r response1) EraseReset() bool      { return r&_R1_ERASE_RESET != 0 }
func (r response1) EraseSeqError() bool   { return r&_R1_ERASE_SEQUENCE_ERROR != 0 }
func (r response1) AddressError() bool    { return r&_R1_ADDRESS_ERROR != 0 }
func (r response1) ParamError() bool      { return r&_R1_PARAMETER_ERROR != 0 }

func b2u8(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}

// CRC16 computes the CRC16 checksum for a given payload using the CRC-16-CCITT polynomial.
func CRC16(buf []byte) (sum uint16) {
	const poly uint16 = 0x1021 // Generator polynomial G(x) = x^16 + x^12 + x^5 + 1
	var crc uint16 = 0x0000    // Initial value
	for _, b := range buf {
		crc ^= (uint16(b) << 8)  // Shift byte into MSB of crc
		for i := 0; i < 8; i++ { // Process each bit
			if crc&0x8000 != 0 {
				crc = (crc << 1) ^ poly
			} else {
				crc <<= 1
			}
		}
	}
	return crc
}

// CRC7 computes the CRC7 checksum for a given payload using the polynomial x^7 + x^3 + 1.
func CRC7(data []byte) uint8 {
	const poly uint8 = 0x09 // Generator polynomial G(x) = x^7 + x^3 + 1
	var crc uint8 = 0x00    // Initial value

	for _, b := range data {
		crc ^= b                 // Initial XOR
		for i := 0; i < 8; i++ { // Process each bit
			if crc&0x80 != 0 {
				crc = (crc << 1) ^ poly
			} else {
				crc <<= 1
			}
		}
	}
	return crc >> 1
}
