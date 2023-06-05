package wifinina

type ConnectionStatus uint8

//go:generate stringer -type=ConnectionStatus -trimprefix=Status
const (
	StatusNoShield       ConnectionStatus = 255
	StatusIdle           ConnectionStatus = 0
	StatusNoSSIDAvail    ConnectionStatus = 1
	StatusScanCompleted  ConnectionStatus = 2
	StatusConnected      ConnectionStatus = 3
	StatusConnectFailed  ConnectionStatus = 4
	StatusConnectionLost ConnectionStatus = 5
	StatusDisconnected   ConnectionStatus = 6
)

// Default state value for Wifi state field
// #define NA_STATE -1

type EncryptionType uint8

//go:generate stringer -type=EncryptionType -trimprefix=EncType
const (
	EncTypeTKIP EncryptionType = 2
	EncTypeCCMP EncryptionType = 4
	EncTypeWEP  EncryptionType = 5
	EncTypeNone EncryptionType = 7
	EncTypeAuto EncryptionType = 8
)

type TCPState uint8

//go:generate stringer -type=TCPState -trimprefix=TCPState
const (
	TCPStateClosed      TCPState = 0
	TCPStateListen      TCPState = 1
	TCPStateSynSent     TCPState = 2
	TCPStateSynRcvd     TCPState = 3
	TCPStateEstablished TCPState = 4
	TCPStateFinWait1    TCPState = 5
	TCPStateFinWait2    TCPState = 6
	TCPStateCloseWait   TCPState = 7
	TCPStateClosing     TCPState = 8
	TCPStateLastACK     TCPState = 9
	TCPStateTimeWait    TCPState = 10
)

type Error uint8

func (err Error) Error() string {
	return "wifinina error: " + err.String()
}

//go:generate stringer -type=Error -trimprefix=Err
const (
	ErrTimeoutChipReady  Error = 0x01
	ErrTimeoutChipSelect Error = 0x02
	ErrCheckStartCmd     Error = 0x03
	ErrWaitRsp           Error = 0x04
	ErrUnexpectedLength  Error = 0xE0
	ErrNoParamsReturned  Error = 0xE1
	ErrIncorrectSentinel Error = 0xE2
	ErrCmdErrorReceived  Error = 0xEF
	ErrNotImplemented    Error = 0xF0
	ErrUnknownHost       Error = 0xF1
	ErrSocketAlreadySet  Error = 0xF2
	ErrConnectionTimeout Error = 0xF3
	ErrNoData            Error = 0xF4
	ErrDataNotWritten    Error = 0xF5
	ErrCheckDataError    Error = 0xF6
	ErrBufferTooSmall    Error = 0xF7
	ErrNoSocketAvail     Error = 0xFF

	NoSocketAvail uint8 = 0xFF
)
