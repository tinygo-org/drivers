package sdcard

import (
	"fmt"
)

type CSD struct {
	CSD_STRUCTURE      byte   //  2 R  [127:126]   0x01 : CSD Structure
	TAAC               byte   //  8 R  [119:112]   0x0E : Data Read Access-Time-1
	NSAC               byte   //  8 R  [111:104]   0x00 : Data Read Access-Time-2 in CLK Cycles (NSAC*100)
	TRAN_SPEED         byte   //  8 R  [103:96]    0x5A : Max. Data Transfer Rate
	CCC                uint16 // 12 R  [95:84]    0x5B5 : Card Command Classes
	READ_BL_LEN        byte   //  4 R  [83:80]     0x09 : Max. Read Data Block Length
	READ_BL_PARTIAL    byte   //  1 R  [79:79]     0x00 : Partial Blocks for Read Allowed
	WRITE_BLK_MISALIGN byte   //  1 R  [78:78]     0x00 : Write Block Misalignment
	READ_BLK_MISALIGN  byte   //  1 R  [77:77]     0x00 : Read Block Misalignment
	DSR_IMP            byte   //  1 R  [76:76]     0x00 : DSR Implemented
	C_SIZE             uint32 // 22 R  [69:48] 0xXXXXXX : Device Size
	ERASE_BLK_EN       byte   //  1 R  [46:46]     0x01 : Erase Single Block Enable
	SECTOR_SIZE        byte   //  7 R  [45:39]     0x7F : Erase Sector Size
	WP_GRP_SIZE        byte   //  7 R  [38:32]     0x00 : Write Protect Group Size
	WP_GRP_ENABLE      byte   //  1 R  [31:31]     0x00 : Write Protect Group Enable
	R2W_FACTOR         byte   //  3 R  [28:26]     0x02 : Write Speed Factor
	WRITE_BL_LEN       byte   //  4 R  [25:22]     0x09 : Max. Write data Block Length
	WRITE_BL_PARTIAL   byte   //  1 R  [21:21]     0x00 : Partial Blocks for Write Allowed
	FILE_FORMAT_GRP    byte   //  1 R  [15:15]     0x00 : File Format Group
	COPY               byte   //  1 RW [14:14]     0x00 :Copy Flag
	PERM_WRITE_PROTECT byte   //  1 RW [13:13]     0x00 : Permanent Write Protection
	TMP_WRITE_PROTECT  byte   //  1 RW [12:12]     0x00 : Temporary Write Protection
	FILE_FORMAT        byte   //  2 R  [11:10]     0x00 : File Format
	CRC                byte   //  7 RW [7:1]       0xXX : CRC
}

func NewCSD(buf []byte) *CSD {
	return &CSD{
		CSD_STRUCTURE:      (buf[0] & 0xC0) >> 6,
		TAAC:               buf[1],
		NSAC:               buf[2],
		TRAN_SPEED:         buf[3],
		CCC:                uint16(buf[4])<<4 | uint16(buf[5])>>4,
		READ_BL_LEN:        buf[5] & 0x0F,
		READ_BL_PARTIAL:    (buf[6] & 0x80) >> 7,
		WRITE_BLK_MISALIGN: (buf[6] & 0x40) >> 6,
		READ_BLK_MISALIGN:  (buf[6] & 0x20) >> 5,
		DSR_IMP:            (buf[6] & 0x10) >> 4,
		C_SIZE:             uint32(buf[7]&0x3F)<<16 | uint32(buf[8])<<8 | uint32(buf[9]),
		ERASE_BLK_EN:       (buf[10] & 0x40) >> 6,
		SECTOR_SIZE:        (buf[10]&0x3F)<<1 | (buf[11]&0x80)>>7,
		WP_GRP_SIZE:        buf[11] & 0x7F,
		WP_GRP_ENABLE:      (buf[12] & 0x80) >> 7,
		R2W_FACTOR:         (buf[12] & 0x1C) >> 2,
		WRITE_BL_LEN:       (buf[12]&0x03)<<2 | (buf[13]&0xC0)>>6,
		WRITE_BL_PARTIAL:   (buf[13] & 0x20) >> 5,
		FILE_FORMAT_GRP:    (buf[14] & 0x80) >> 7,
		COPY:               (buf[14] & 0x40) >> 6,
		PERM_WRITE_PROTECT: (buf[14] & 0x20) >> 5,
		TMP_WRITE_PROTECT:  (buf[14] & 0x10) >> 4,
		FILE_FORMAT:        (buf[14] & 0x0C) >> 2,
		CRC:                (buf[15] & 0xFE) >> 1,
	}
}

func (c *CSD) Dump() {
	fmt.Printf("CSD_STRUCTURE:      %X\r\n", c.CSD_STRUCTURE)
	fmt.Printf("TAAC:               %X\r\n", c.TAAC)
	fmt.Printf("NSAC:               %X\r\n", c.NSAC)
	fmt.Printf("TRAN_SPEED:         %X\r\n", c.TRAN_SPEED)
	fmt.Printf("CCC:                %X\r\n", c.CCC)
	fmt.Printf("READ_BL_LEN:        %X\r\n", c.READ_BL_LEN)
	fmt.Printf("READ_BL_PARTIAL:    %X\r\n", c.READ_BL_PARTIAL)
	fmt.Printf("WRITE_BLK_MISALIGN: %X\r\n", c.WRITE_BLK_MISALIGN)
	fmt.Printf("READ_BLK_MISALIGN:  %X\r\n", c.READ_BLK_MISALIGN)
	fmt.Printf("DSR_IMP:            %X\r\n", c.DSR_IMP)
	fmt.Printf("C_SIZE:             %X\r\n", c.C_SIZE)
	fmt.Printf("ERASE_BLK_EN:       %X\r\n", c.ERASE_BLK_EN)
	fmt.Printf("SECTOR_SIZE:        %X\r\n", c.SECTOR_SIZE)
	fmt.Printf("WP_GRP_SIZE:        %X\r\n", c.WP_GRP_SIZE)
	fmt.Printf("WP_GRP_ENABLE:      %X\r\n", c.WP_GRP_ENABLE)
	fmt.Printf("R2W_FACTOR:         %X\r\n", c.R2W_FACTOR)
	fmt.Printf("WRITE_BL_LEN:       %X\r\n", c.WRITE_BL_LEN)
	fmt.Printf("WRITE_BL_PARTIAL:   %X\r\n", c.WRITE_BL_PARTIAL)
	fmt.Printf("FILE_FORMAT_GRP:    %X\r\n", c.FILE_FORMAT_GRP)
	fmt.Printf("COPY:               %X\r\n", c.COPY)
	fmt.Printf("PERM_WRITE_PROTECT: %X\r\n", c.PERM_WRITE_PROTECT)
	fmt.Printf("TMP_WRITE_PROTECT:  %X\r\n", c.TMP_WRITE_PROTECT)
	fmt.Printf("FILE_FORMAT:        %X\r\n", c.FILE_FORMAT)
	fmt.Printf("CRC:                %X\r\n", c.CRC)
}

func (c *CSD) Sectors() (int64, error) {
	sectors := int64(0)
	if c.CSD_STRUCTURE == 0x01 {
		// CSD version 2.0
		sectors = (int64(c.C_SIZE) + 1) * 1024
	} else if c.CSD_STRUCTURE == 0x00 {
		// CSD version 1.0 (old, <=2GB)
		return 0, fmt.Errorf("CSD format version 1.0 is not supported")
	} else {
		return 0, fmt.Errorf("unknown CSD format")
	}
	return sectors, nil
}

func (c *CSD) Size() uint64 {
	return uint64(c.C_SIZE) * 512 * 1024
}
