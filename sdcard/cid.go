package sdcard

import "fmt"

type CID struct {
	//// byte 0
	ManufacturerID byte
	//uint8_t mid;  // Manufacturer ID
	//// byte 1-2
	OEMApplicationID uint16
	//char oid[2];  // OEM/Application ID
	//// byte 3-7
	ProductName string
	//char pnm[5];  // Product name
	//// byte 8
	ProductVersion string
	//unsigned prv_m : 4;  // Product revision n.m
	//unsigned prv_n : 4;
	//// byte 9-12
	ProductSerialNumber uint32
	//uint32_t psn;  // Product serial number
	//// byte 13
	ManufacturingYear  byte
	ManufacturingMonth byte
	//unsigned mdt_year_high : 4;  // Manufacturing date
	//unsigned reserved : 4;
	//// byte 14
	//unsigned mdt_month : 4;
	//unsigned mdt_year_low : 4;
	//// byte 15
	Always1 byte
	//unsigned always1 : 1;
	CRC byte
	//unsigned crc : 7;
}

func NewCID(buf []byte) *CID {
	return &CID{
		ManufacturerID:      buf[0],
		OEMApplicationID:    (uint16(buf[0]) << 8) | uint16(buf[1]),
		ProductName:         string(buf[3:8]),
		ProductVersion:      fmt.Sprintf("%d.%d", (buf[8]&0xF0)>>4, buf[8]&0x0F),
		ProductSerialNumber: (uint32(buf[9]) << 24) | (uint32(buf[10]) << 16) | (uint32(buf[11]) << 8) | uint32(buf[12]),
		ManufacturingYear:   (buf[13] & 0xF0) | (buf[14] & 0x0F),
		ManufacturingMonth:  (buf[14] & 0xF0) >> 4,
		Always1:             (buf[15] & 0x80) >> 7,
		CRC:                 buf[15] & 0x7F,
	}
}
