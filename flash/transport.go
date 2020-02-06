package flash

type Command byte

const (
	CmdRead            Command = 0x03 // Single Read
	CmdQuadRead                = 0x6B // 1 line address, 4 line data
	CmdReadJedecID             = 0x9f
	CmdPageProgram             = 0x02
	CmdQuadPageProgram         = 0x32 // 1 line address, 4 line data
	CmdReadStatus              = 0x05
	CmdReadStatus2             = 0x35
	CmdWriteStatus             = 0x01
	CmdWriteStatus2            = 0x31
	CmdEnableReset             = 0x66
	CmdReset                   = 0x99
	CmdWriteEnable             = 0x06
	CmdWriteDisable            = 0x04
	CmdEraseSector             = 0x20
	CmdEraseBlock              = 0xD8
	CmdEraseChip               = 0xC7
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

type transport interface {
	begin()
	supportQuadMode() bool
	setClockSpeed(hz uint32) (err error)
	runCommand(cmd Command) (err error)
	readCommand(cmd Command, rsp []byte) (err error)
	writeCommand(cmd Command, data []byte) (err error)
	eraseCommand(cmd Command, address uint32) (err error)
	readMemory(addr uint32, rsp []byte) (err error)
	writeMemory(addr uint32, data []byte) (err error)
}
