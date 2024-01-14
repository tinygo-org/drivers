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

func (c *CID) ProductName() []byte {
	return upToNull(c.prodName[:])
}

func (c *CID) ProductRevision() (n, m uint8) {
	return c.productRev >> 4, c.productRev & 0x0F
}

/*
CSD Register Fields:
Name:                   Field:                  Width: Value:          CSD-Slice:
CSD Structure           CSD_STRUCTURE           2      00b              R[127:128]
TAAC                    TAAC                    8      00h             R[119:113]
NSAC                    NSAC                    8      00h             R[111:105]
TRAN_SPEED              TRAN_SPEED              8      32h or 5Ah      R[103:97]
CCC                     CCC                     12     01x110110101b   R[95:85]
READ_BL_LEN             READ_BL_LEN             4      xh              R[83:81]
READ_BL_PARTIAL         READ_BL_PARTIAL         1      1b              R[79:80]
WRITE_BLK_MISALIGN      WRITE_BLK_MISALIGN      1      xb              R[78:79]
READ_BLK_MISALIGN       READ_BLK_MISALIGN       1      xb              R[77:78]
DSR_IMP                 DSR_IMP                 1      xb              R[76:77]
C_SIZE                  C_SIZE                  12     xxxh            R[73:63]
VDD_R_CURR_MIN          VDD_R_CURR_MIN          3      xxxb            R[61:60]
VDD_R_CURR_MAX          VDD_R_CURR_MAX          3      xxxb            R[58:57]
VDD_W_CURR_MIN          VDD_W_CURR_MIN          3      xxxb            R[55:54]
VDD_W_CURR_MAX          VDD_W_CURR_MAX          3      xxxb            R[52:51]
C_SIZE_MULT             C_SIZE_MULT             3      xxxb            R[49:48]
ERASE_BLK_EN            ERASE_BLK_EN            1      xb              R[46:47]
SECTOR_SIZE             SECTOR_SIZE             7      xxxxxxxb        R[45:40]
WP_GRP_SIZE             WP_GRP_SIZE             7      xxxxxxxb        R[38:33]
WP_GRP_ENABLE           WP_GRP_ENABLE           1      xb              R[31:32]
R2W_FACTOR              R2W_FACTOR              2      xxxb            R[28:27]
WRITE_BL_LEN            WRITE_BL_LEN            4      xxxxb           R[25:23]
WRITE_BL_PARTIAL        WRITE_BL_PARTIAL        1      xb              R[21:22]
FILE_FORMAT_GRP         FILE_FORMAT_GRP         1      xb              R[15:16]
COPY                    COPY                    1      xb              R[14:15]
PERM_WRITE_PROTECT      PERM_WRITE_PROTECT      1      xb              R[13:14]
TMP_WRITE_PROTECT       TMP_WRITE_PROTECT       1      xb              R[12:13]
FILE_FORMAT             FILE_FORMAT             2      xxb             R[11:11]
CRC                     CRC                     7      xxxxxxxb        R[7:2]
Not Used                -                       1      1b              R[0:1]

Note: 'R' indicates read-only fields, 'R/W' indicates read/write fields.
The values in the 'CSD-Slice' column indicate the bit positions in the CSD register.
*/

// CSD is the Card Specific Data register, a 128-bit (16-byte) register that defines how
// the SD card standard communicates with the memory field or register.
type CSD struct {
	data [16]byte
}

type CSDv1 struct {
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
	if c.CSDStructure() != 1 {
		panic("CSD is not version 1")
	}
	return CSDv1{CSD: c}
}

// TAAC returns the Time Access Attribute Class (data read access-time-1).
func (c *CSD) TAAC() TAAC { return TAAC(c.data[1]) }

// NSAC returns the Data Read Access-time 2 in CLK cycles (NSAC*100).
func (c *CSD) NSAC() uint8 { return c.data[2] }

// TransferSpeed returns the Max Data Transfer Rate. Either 0x32 or 0x5A.
func (c *CSD) TransferSpeed() TransferSpeed { return TransferSpeed(c.data[3]) }

// CCC returns the Card Command Classes.
func (c *CSD) CCC() uint16 {
	return uint16(c.data[4])<<4 | uint16(c.data[5]&0xf0)>>4
}

// ReadBlockLen returns the Max Read Data Block Length in bytes.
func (c *CSD) ReadBlockLen() uint16 { return 1 << (c.data[5] & 0x0F) }

func (c *CSD) ReadBlockPartial() bool       { return c.data[6]&(1<<7) != 0 }
func (c *CSD) WriteBlockMisalignment() bool { return c.data[6]&(1<<6) != 0 }
func (c *CSD) ReadBlockMisalignment() bool  { return c.data[6]&(1<<5) != 0 }

// ImplementsDSR returns whether the card implements the DSR register.
func (c *CSD) ImplementsDSR() bool { return c.data[6]&(1<<4) != 0 }

func (c *CSDv1) CSize() uint16 {
	// Jesus, why did SD make this so complicated?
	return uint16(c.data[8]>>6) | uint16(c.data[7])<<2 | uint16(c.data[6]&0b11)<<10
}

func (c *CSD) String() string {
	buf := make([]byte, 0, 64)
	return string(c.appendf(buf, '\n'))
}

func (c *CSD) appendf(b []byte, delim byte) []byte {
	b = appendnum(b, "CSDStructure", uint64(c.CSDStructure()), delim)
	b = appendnum(b, "TimeAccess_ns", uint64(c.TAAC().AccessTime()), delim)
	b = appendnum(b, "NSAC", uint64(c.NSAC()), delim)
	b = appendnum(b, "Tx_kb/s", uint64(c.TransferSpeed().RateKilobits()), delim)
	b = appendnum(b, "CCC", uint64(c.CCC()), delim)
	b = appendnum(b, "ReadBlockLen", uint64(c.ReadBlockLen()), delim)
	b = appendbit(b, "ReadBlockPartial", c.ReadBlockPartial(), delim)
	b = appendbit(b, "WriteBlockMisalignment", c.WriteBlockMisalignment(), delim)
	b = appendbit(b, "ReadBlockMisalignment", c.ReadBlockMisalignment(), delim)
	b = appendbit(b, "ImplementsDSR", c.ImplementsDSR(), delim)
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

type (
	TransferSpeed uint8
	TAAC          uint8
)

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
