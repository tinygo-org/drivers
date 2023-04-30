package wifinina

type CommandType uint8

//go:generate stringer -type=CommandType -trimprefix=Cmd
const (
	CmdStart CommandType = 0xE0
	CmdEnd   CommandType = 0xEE
	CmdErr   CommandType = 0xEF

	CmdSetNet          CommandType = 0x10
	CmdSetPassphrase   CommandType = 0x11
	CmdSetKey          CommandType = 0x12
	CmdSetIPConfig     CommandType = 0x14
	CmdSetDNSConfig    CommandType = 0x15
	CmdSetHostname     CommandType = 0x16
	CmdSetPowerMode    CommandType = 0x17
	CmdSetAPNet        CommandType = 0x18
	CmdSetAPPassphrase CommandType = 0x19
	CmdSetDebug        CommandType = 0x1A
	CmdGetTemperature  CommandType = 0x1B
	CmdGetReasonCode   CommandType = 0x1F
	//	TEST_CMD	        = 0x13

	CmdGetConnStatus     CommandType = 0x20
	CmdGetIPAddr         CommandType = 0x21
	CmdGetMACAddr        CommandType = 0x22
	CmdGetCurrSSID       CommandType = 0x23
	CmdGetCurrBSSID      CommandType = 0x24
	CmdGetCurrRSSI       CommandType = 0x25
	CmdGetCurrEncrType   CommandType = 0x26
	CmdScanNetworks      CommandType = 0x27
	CmdStartServerTCP    CommandType = 0x28
	CmdGetStateTCP       CommandType = 0x29
	CmdDataSentTCP       CommandType = 0x2A
	CmdAvailDataTCP      CommandType = 0x2B
	CmdGetDataTCP        CommandType = 0x2C
	CmdStartClientTCP    CommandType = 0x2D
	CmdStopClientTCP     CommandType = 0x2E
	CmdGetClientStateTCP CommandType = 0x2F
	CmdDisconnect        CommandType = 0x30
	CmdGetIdxRSSI        CommandType = 0x32
	CmdGetIdxEncrType    CommandType = 0x33
	CmdReqHostByName     CommandType = 0x34
	CmdGetHostByName     CommandType = 0x35
	CmdStartScanNetworks CommandType = 0x36
	CmdGetFwVersion      CommandType = 0x37
	CmdSendDataUDP       CommandType = 0x39
	CmdGetRemoteData     CommandType = 0x3A
	CmdGetTime           CommandType = 0x3B
	CmdGetIdxBSSID       CommandType = 0x3C
	CmdGetIdxChannel     CommandType = 0x3D
	CmdPing              CommandType = 0x3E
	CmdGetSocket         CommandType = 0x3F
	//	GET_IDX_SSID_CMD	= 0x31,
	//	GET_TEST_CMD		= 0x38

	// All command with DATA_FLAG 0x40 send a 16bit Len
	CmdSendDataTCP   CommandType = 0x44
	CmdGetDatabufTCP CommandType = 0x45
	CmdInsertDataBuf CommandType = 0x46

	// regular format commands
	CmdSetPinMode      CommandType = 0x50
	CmdSetDigitalWrite CommandType = 0x51
	CmdSetAnalogWrite  CommandType = 0x52
)
