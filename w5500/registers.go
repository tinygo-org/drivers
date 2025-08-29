package w5500

// Common Registers.
const (
	regMode        = 0x0000
	regGatewayAddr = 0x0001
	regSubnetMask  = 0x0005
	regMAC         = 0x0009
	regIPAddr      = 0x000F
	regIntLevel    = 0x0013
	regInt         = 0x0015
	regIntMask     = 0x0016
	regSockInt     = 0x0017
	regSockIntMask = 0x0018
	regRetryTime   = 0x0019
	regRetryN      = 0x001B
	// ... PPP registers, not needed
	regPHYCfg  = 0x002E
	regChipVer = 0x0039
)

// Socket Registers.
const (
	sockMode           = 0x0000
	sockCmd            = 0x0001
	sockInt            = 0x0002
	sockStatus         = 0x0003
	sockSrcPort        = 0x0004
	sockDestMAC        = 0x0006
	sockDestIP         = 0x000C
	sockDestPort       = 0x0010
	sockMaxSegSize     = 0x0012
	sockIPTOS          = 0x0015
	sockIPTTL          = 0x0016
	sockRXBUFSize      = 0x001E
	sockTXBUFSize      = 0x001F
	sockTXFreeSize     = 0x0020
	sockTXReadPtr      = 0x0022
	sockTXWritePtr     = 0x0024
	sockRXReceivedSize = 0x0026
	sockRXReadPtr      = 0x0028
	sockRXWritePtr     = 0x002A
	sockIntMask        = 0x002C
	sockKeepInt        = 0x002F
)

// Socket Commands.
const (
	sockCmdOpen       = 0x01
	sockCmdClose      = 0x10
	sockCmdListen     = 0x02
	sockCmdConnect    = 0x04
	sockCmdDisconnect = 0x08
	sockCmdSend       = 0x20
	sockCmdSendMacRaw = 0x21
	sockCmdSendKeep   = 0x22
	sockCmdRecv       = 0x40
)

// Socket Statuses.
const (
	sockStatusClosed      = 0x00
	sockStatusInit        = 0x13
	sockStatusListen      = 0x14
	sockStatusEstablished = 0x17
	sockStatusCloseWait   = 0x1C
	sockStatusUdp         = 0x22
	sockStatusMacRaw      = 0x42
	// Temporary TCP states
	sockStatusSynSent  = 0x15
	sockStatusSynRecv  = 0x16
	sockStatusFinWait  = 0x18
	sockStatusClosing  = 0x1A
	sockStatusTimeWait = 0x1B
	sockStatusLastAck  = 0x1D
	sockStatusUnknown  = 0xFF
)

// Socket Interrupts.
const (
	sockIntConnect uint8 = 1 << iota
	sockIntDisconnect
	sockIntReceive
	sockIntTimeout
	sockIntSendOK

	sockIntUnknown uint8 = 0
)
