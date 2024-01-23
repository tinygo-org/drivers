package sd

import (
	"bytes"
	"encoding/binary"
	"io"
	"strconv"
	"time"
)

// For reference of CID/CSD structs see:
// See https://github.com/arduino-libraries/SD/blob/1c56f58252553c7537f7baf62798cacc625aa543/src/utility/SdInfo.h#L110

type CardKind uint8

func isTimeout(err error) bool {
	return err == errReadTimeout || err == errWriteTimeout || err == errBusyTimeout
}

const (
	// card types
	TypeSD1  CardKind = 1 // Standard capacity V1 SD card
	TypeSD2  CardKind = 2 // Standard capacity V2 SD card
	TypeSDHC CardKind = 3 // High Capacity SD card
)

type CID struct {
	data [16]byte
}

func DecodeCID(b []byte) (cid CID, _ error) {
	if len(b) < 16 {
		return CID{}, io.ErrShortBuffer
	}
	copy(cid.data[:], b)
	if !cid.IsValid() {
		return cid, errBadCSDCID
	}
	return cid, nil
}

// RawCopy returns a copy of the raw CID data.
func (c *CID) RawCopy() [16]byte { return c.data }

// ManufacturerID is an 8-bit binary number that identifies the card manufacturer. The MID number is controlled, defined, and allocated to a SD Memory Card manufacturer by the SD-3C, LLC.
func (c *CID) ManufacturerID() uint8 { return c.data[0] }

// OEMApplicationID A 2-character ASCII string that identifies the card OEM and/or the card contents (when used as a
// distribution media either on ROM or FLASH cards). The OID number is controlled, defined, and allocated
// to a SD Memory Card manufacturer by the SD-3C, LLC
func (c *CID) OEMApplicationID() uint16 {
	return binary.BigEndian.Uint16(c.data[1:3])
}

// The product name is a string, 5-character ASCII string.
func (c *CID) ProductName() string {
	return string(upToNull(c.data[3:8]))
}

// ProductRevision is composed of two Binary Coded Decimal (BCD) digits, four bits each, representing
// an "n.m" revision number. The "n" is the most significant nibble and "m" is the least significant nibble.
// As an example, the PRV binary value field for product revision "6.2" will be: 0110 0010b
func (c *CID) ProductRevision() (n, m uint8) {
	rev := c.data[8]
	return rev >> 4, rev & 0x0F
}

// The Serial Number is 32 bits of binary number.
func (c *CID) ProductSerialNumber() uint32 {
	return binary.BigEndian.Uint32(c.data[9:13])
}

// ManufacturingDate returns the manufacturing date of the card.
func (c *CID) ManufacturingDate() (year uint16, month uint8) {
	date := binary.BigEndian.Uint16(c.data[13:15])
	return (date >> 4) + 2000, uint8(date & 0x0F)
}

// CRC7 returns the CRC7 checksum for this CID. May be invalid. Use [IsValid] to check validity of CRC7+Always1 fields.
func (c *CID) CRC7() uint8 { return c.data[15] >> 1 }

// Always1 checks the presence of the Always 1 bit. Should return true for valid CIDs.
func (c *CID) Always1() bool { return c.data[15]&1 != 0 }

// IsValid checks if the CRC and always1 fields are expected values.
func (c *CID) IsValid() bool { return c.Always1() && CRC7(c.data[:15]) == c.CRC7() }

// CSD is the Card Specific Data register, a 128-bit (16-byte) register that defines how
// the SD card standard communicates with the memory field or register. This type is
// shared among V1 and V2 type devices.
type CSD struct {
	data [16]byte
}

// CSDv1 is the Card Specific Data register for V1 devices. See [CSD] for more info.
type CSDv1 struct {
	CSD
}

// CSDv2 is the Card Specific Data register for V2 devices. See [CSD] for more info.
type CSDv2 struct {
	CSD
}

// DecodeCSD decodes the CSD from a 16-byte slice.
func DecodeCSD(b []byte) (CSD, error) {
	if len(b) < 16 {
		return CSD{}, io.ErrShortBuffer
	}
	csd := CSD{}
	copy(csd.data[:], b)
	if !csd.IsValid() {
		return csd, errBadCSDCID
	}
	return csd, nil
}

// csdStructure returns the version of the CSD structure.
func (c *CSD) csdStructure() uint8 { return c.data[0] >> 6 }

// Version returns the version of the CSD structure. Effectively returns 1+CSDStructure.
func (c *CSD) Version() uint8 { return 1 + c.csdStructure() }

// MustV1 returns the CSD as a CSDv1. Panics if the CSD is not version 1.0.
func (c CSD) MustV1() CSDv1 {
	if c.csdStructure() != 0 {
		panic("CSD is not version 1.0")
	}
	return CSDv1{CSD: c}
}

func (c CSD) MustV2() CSDv2 {
	if c.csdStructure() != 1 {
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
func (c *CSD) ReadBlockLen() int        { return 1 << c.ReadBlockLenShift() }
func (c *CSD) ReadBlockLenShift() uint8 { return c.data[5] & 0x0F }

// AllowsReadBlockPartial should always return true. Indicates that
func (c *CSD) AllowsReadBlockPartial() bool { return c.data[6]&(1<<7) != 0 }

// AllowsWriteBlockMisalignment defines if the data block to be written by one command
// can be spread over more than one physical block of the memory device.
func (c *CSD) AllowsWriteBlockMisalignment() bool { return c.data[6]&(1<<6) != 0 }

// AllowsReadBlockMisalignment defines if the data block to be read by one command
// can be spread over more than one physical block of the memory device.
func (c *CSD) AllowsReadBlockMisalignment() bool { return c.data[6]&(1<<5) != 0 }

// CRC7 returns the CRC read for this CSD. May be invalid. Use [IsValid] to check validity of CRC7+Always1 fields.
func (c *CSD) CRC7() uint8 { return c.data[15] >> 1 }

// Always1 checks the Always 1 bit. Should always evaluate to true for valid CSDs.
func (c *CSD) Always1() bool { return c.data[15]&1 != 0 }

// IsValid checks if the CRC and always1 fields are expected values.
func (c *CSD) IsValid() bool {
	// Compare last byte with CRC and also the always1 bit.
	return c.Always1() && CRC7(c.data[:15]) == c.CRC7()
}

// ImplementsDSR defines if the configurable driver stage is integrated on the card.
func (c *CSD) ImplementsDSR() bool { return c.data[6]&(1<<4) != 0 }

// EraseSectorSizeInBlocks represents how much memory is erased in an erase
// command in multiple of block size.
func (c *CSDv1) EraseSectorSizeInBytes() int64 {
	blklen := c.WriteBlockLen()
	numblocks := c.SectorSize()
	return int64(numblocks) * blklen
}

// SectorSize varies in meaning depending on the version.
func (c *CSD) SectorSize() uint8 {
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

// WriteBlockLen represents maximum write data block length in bytes.
func (c *CSD) WriteBlockLen() int64 {
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

func (c *CSD) DeviceCapacity() (size int64) {
	switch c.csdStructure() {
	case 0:
		v1 := c.MustV1()
		size = int64(v1.DeviceCapacity())
	case 1:
		v2 := c.MustV2()
		size = v2.DeviceCapacity()
	}
	return size
}

// NumberOfBlocks returns amount of readable blocks in the device given by Capacity/ReadBlockLength.
func (c *CSD) NumberOfBlocks() (numBlocks int64) {
	rblocks := c.ReadBlockLen()
	if rblocks == 0 {
		return 0
	}
	return c.DeviceCapacity() / int64(rblocks)
}

// After byte 5 CSDv1 and CSDv2  differ in structure at some fields.

// DeviceCapacity returns the device capacity in bytes.
func (c *CSDv2) DeviceCapacity() int64 {
	csize := c.csize()
	return int64(csize) * 512_000
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
	version := c.csdStructure() + 1
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
	b = appendnum(b, "Version", uint64(c.Version()), delim)
	b = appendnum(b, "Capacity(bytes)", uint64(c.DeviceCapacity()), delim)
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
	b = appendnum(b, "WriteBlockLen", uint64(c.WriteBlockLen()), delim)
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

type (
	command    byte
	appcommand byte
)

// SD commands and application commands.
const (
	cmdGoIdleState         command = 0
	cmdSendOpCnd           command = 1
	cmdAllSendCID          command = 2
	cmdSendRelativeAddr    command = 3
	cmdSetDSR              command = 4
	cmdSwitchFunc          command = 6
	cmdSelectDeselectCard  command = 7
	cmdSendIfCond          command = 8
	cmdSendCSD             command = 9
	cmdSendCID             command = 10
	cmdStopTransmission    command = 12
	cmdSendStatus          command = 13
	cmdGoInactiveState     command = 15
	cmdSetBlocklen         command = 16
	cmdReadSingleBlock     command = 17
	cmdReadMultipleBlock   command = 18
	cmdWriteBlock          command = 24
	cmdWriteMultipleBlock  command = 25
	cmdProgramCSD          command = 27
	cmdSetWriteProt        command = 28
	cmdClrWriteProt        command = 29
	cmdSendWriteProt       command = 30
	cmdEraseWrBlkStartAddr command = 32
	cmdEraseWrBlkEndAddr   command = 33
	cmdErase               command = 38
	cmdLockUnlock          command = 42
	cmdAppCmd              command = 55
	cmdGenCmd              command = 56
	cmdReadOCR             command = 58
	cmdCRCOnOff            command = 59

	acmdSET_BUS_WIDTH            appcommand = 6
	acmdSD_STATUS                appcommand = 13
	acmdSEND_NUM_WR_BLOCKS       appcommand = 22
	acmdSET_WR_BLK_ERASE_COUNT   appcommand = 23
	acmdSD_APP_OP_COND           appcommand = 41
	acmdSET_CLR_CARD_DETECT      appcommand = 42
	acmdSEND_SCR                 appcommand = 51
	acmdSECURE_READ_MULTI_BLOCK  appcommand = 18
	acmdSECURE_WRITE_MULTI_BLOCK appcommand = 25
	acmdSECURE_WRITE_MKB         appcommand = 26
	acmdSECURE_ERASE             appcommand = 38
	acmdGET_MKB                  appcommand = 43
	acmdGET_MID                  appcommand = 44
	acmdSET_CER_RN1              appcommand = 45
	acmdSET_CER_RN2              appcommand = 46
	acmdSET_CER_RES2             appcommand = 47
	acmdSET_CER_RES1             appcommand = 48
	acmdCHANGE_SECURE_AREA       appcommand = 49
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

// RateMegabits returns the transfer rate in kilobits per second.
func (t TransferSpeed) RateKilobits() int64 {
	return 100 * log10table[t&0b111]
}

func (t TAAC) AccessTime() (d time.Duration) {
	return time.Duration(log10table[t&0b111]) * time.Nanosecond
}

func b2u8(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}

// CRC16 computes the CRC16 checksum for a given payload using the CRC-16-CCITT polynomial.
func CRC16(buf []byte) (crc uint16) {
	const poly uint16 = 0x1021 // Generator polynomial G(x) = x^16 + x^12 + x^5 + 1
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
func CRC7(data []byte) (crc uint8) {
	return crc7noshift(data) >> 1
}

func crc7noshift(data []byte) (crc uint8) {
	for _, b := range data {
		crc = crc7_table[crc^b]
	}
	return crc
}

var crc7_table = [256]byte{
	0x00, 0x12, 0x24, 0x36, 0x48, 0x5a, 0x6c, 0x7e,
	0x90, 0x82, 0xb4, 0xa6, 0xd8, 0xca, 0xfc, 0xee,
	0x32, 0x20, 0x16, 0x04, 0x7a, 0x68, 0x5e, 0x4c,
	0xa2, 0xb0, 0x86, 0x94, 0xea, 0xf8, 0xce, 0xdc,
	0x64, 0x76, 0x40, 0x52, 0x2c, 0x3e, 0x08, 0x1a,
	0xf4, 0xe6, 0xd0, 0xc2, 0xbc, 0xae, 0x98, 0x8a,
	0x56, 0x44, 0x72, 0x60, 0x1e, 0x0c, 0x3a, 0x28,
	0xc6, 0xd4, 0xe2, 0xf0, 0x8e, 0x9c, 0xaa, 0xb8,
	0xc8, 0xda, 0xec, 0xfe, 0x80, 0x92, 0xa4, 0xb6,
	0x58, 0x4a, 0x7c, 0x6e, 0x10, 0x02, 0x34, 0x26,
	0xfa, 0xe8, 0xde, 0xcc, 0xb2, 0xa0, 0x96, 0x84,
	0x6a, 0x78, 0x4e, 0x5c, 0x22, 0x30, 0x06, 0x14,
	0xac, 0xbe, 0x88, 0x9a, 0xe4, 0xf6, 0xc0, 0xd2,
	0x3c, 0x2e, 0x18, 0x0a, 0x74, 0x66, 0x50, 0x42,
	0x9e, 0x8c, 0xba, 0xa8, 0xd6, 0xc4, 0xf2, 0xe0,
	0x0e, 0x1c, 0x2a, 0x38, 0x46, 0x54, 0x62, 0x70,
	0x82, 0x90, 0xa6, 0xb4, 0xca, 0xd8, 0xee, 0xfc,
	0x12, 0x00, 0x36, 0x24, 0x5a, 0x48, 0x7e, 0x6c,
	0xb0, 0xa2, 0x94, 0x86, 0xf8, 0xea, 0xdc, 0xce,
	0x20, 0x32, 0x04, 0x16, 0x68, 0x7a, 0x4c, 0x5e,
	0xe6, 0xf4, 0xc2, 0xd0, 0xae, 0xbc, 0x8a, 0x98,
	0x76, 0x64, 0x52, 0x40, 0x3e, 0x2c, 0x1a, 0x08,
	0xd4, 0xc6, 0xf0, 0xe2, 0x9c, 0x8e, 0xb8, 0xaa,
	0x44, 0x56, 0x60, 0x72, 0x0c, 0x1e, 0x28, 0x3a,
	0x4a, 0x58, 0x6e, 0x7c, 0x02, 0x10, 0x26, 0x34,
	0xda, 0xc8, 0xfe, 0xec, 0x92, 0x80, 0xb6, 0xa4,
	0x78, 0x6a, 0x5c, 0x4e, 0x30, 0x22, 0x14, 0x06,
	0xe8, 0xfa, 0xcc, 0xde, 0xa0, 0xb2, 0x84, 0x96,
	0x2e, 0x3c, 0x0a, 0x18, 0x66, 0x74, 0x42, 0x50,
	0xbe, 0xac, 0x9a, 0x88, 0xf6, 0xe4, 0xd2, 0xc0,
	0x1c, 0x0e, 0x38, 0x2a, 0x54, 0x46, 0x70, 0x62,
	0x8c, 0x9e, 0xa8, 0xba, 0xc4, 0xd6, 0xe0, 0xf2,
}
