package stusb4500

const (
	// Control Message Types
	ctrlMsgReserved1       = 0x00
	ctrlMsgGoodCRC         = 0x01
	ctrlMsgGotoMin         = 0x02
	ctrlMsgAccept          = 0x03
	ctrlMsgReject          = 0x04
	ctrlMsgPing            = 0x05
	ctrlMsgPSRDY           = 0x06
	ctrlMsgGetSourceCap    = 0x07
	ctrlMsgGetSinkCap      = 0x08
	ctrlMsgDRSwap          = 0x09
	ctrlMsgPRSwap          = 0x0A
	ctrlMsgVCONNSwap       = 0x0B
	ctrlMsgWait            = 0x0C
	ctrlMsgSoftReset       = 0x0D
	ctrlMsgReserved2       = 0x0E
	ctrlMsgReserved3       = 0x0F
	ctrlMsgNotSupported    = 0x10
	ctrlMsgGetSourceCapExt = 0x11
	ctrlMsgGetStatus       = 0x12
	ctrlMsgFRSwap          = 0x13
	ctrlMsgGetPPSStatus    = 0x14
	ctrlMsgGetCountryCodes = 0x15
	ctrlMsgReserved4       = 0x16
	ctrlMsgReserved5       = 0x1F
	// Data Message Types
	dataMsgReserved1      = 0x00
	dataMsgSourceCap      = 0x01
	dataMsgRequest        = 0x02
	dataMsgBIST           = 0x03
	dataMsgSinkCap        = 0x04
	dataMsgBatteryStatus  = 0x05
	dataMsgAlert          = 0x06
	dataMsgGetCountryInfo = 0x07
	dataMsgReserved2      = 0x08
	dataMsgReserved3      = 0x0E
	dataMsgVendorDefined  = 0x0F
	dataMsgReserved4      = 0x10
	dataMsgReserved5      = 0x1F
)
