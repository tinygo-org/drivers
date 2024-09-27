package sd

import (
	"encoding/binary"
	"strconv"
)

const (
	_CMD_TIMEOUT = 100

	_R1_IDLE_STATE           = 1 << 0
	_R1_ERASE_RESET          = 1 << 1
	_R1_ILLEGAL_COMMAND      = 1 << 2
	_R1_COM_CRC_ERROR        = 1 << 3
	_R1_ERASE_SEQUENCE_ERROR = 1 << 4
	_R1_ADDRESS_ERROR        = 1 << 5
	_R1_PARAMETER_ERROR      = 1 << 6

	_DATA_RES_MASK     = 0x1F
	_DATA_RES_ACCEPTED = 0x05
)

type response1 uint8

func (r response1) IsIdle() bool          { return r&_R1_IDLE_STATE != 0 }
func (r response1) IllegalCmdError() bool { return r&_R1_ILLEGAL_COMMAND != 0 }
func (r response1) CRCError() bool        { return r&_R1_COM_CRC_ERROR != 0 }
func (r response1) EraseReset() bool      { return r&_R1_ERASE_RESET != 0 }
func (r response1) EraseSeqError() bool   { return r&_R1_ERASE_SEQUENCE_ERROR != 0 }
func (r response1) AddressError() bool    { return r&_R1_ADDRESS_ERROR != 0 }
func (r response1) ParamError() bool      { return r&_R1_PARAMETER_ERROR != 0 }

type response1Err struct {
	context string
	status  response1
}

func (e response1Err) Error() string {
	return e.status.Response()
	if e.context != "" {
		return "sd:" + e.context + " " + strconv.Itoa(int(e.status))
	}
	return "sd:status " + strconv.Itoa(int(e.status))
}

func (e response1) Response() string {
	b := make([]byte, 0, 8)
	return string(e.appendf(b))
}

func (r response1) appendf(b []byte) []byte {
	b = append(b, '[')
	if r.IsIdle() {
		b = append(b, "idle,"...)
	}
	if r.EraseReset() {
		b = append(b, "erase-rst,"...)
	}
	if r.EraseSeqError() {
		b = append(b, "erase-seq,"...)
	}
	if r.CRCError() {
		b = append(b, "crc-err,"...)
	}
	if r.AddressError() {
		b = append(b, "addr-err,"...)
	}
	if r.ParamError() {
		b = append(b, "param-err,"...)
	}
	if r.IllegalCmdError() {
		b = append(b, "illegal-cmd,"...)
	}
	if len(b) > 1 {
		b = b[:len(b)-1]
	}
	b = append(b, ']')
	return b
}

func makeResponseError(status response1) error {
	return response1Err{
		status: status,
	}
}

// Commands used to help generate this file:
//   - stringer -type=state -trimprefix=state -output=state_string.go
//   - stringer -type=status -trimprefix=status -output=status_string.go

// Tokens that are sent by card during polling.
// https://github.com/arduino-libraries/SD/blob/master/src/utility/SdInfo.h
const (
	tokSTART_BLOCK = 0xfe
	tokSTOP_TRAN   = 0xfd
	tokWRITE_MULT  = 0xfc
)

type state uint8

const (
	stateIdle state = iota
	stateReady
	stateIdent
	stateStby
	stateTran
	stateData
	stateRcv
	statePrg
	stateDis
)

// status represents the Card Status Register (R1), as per section 4.10.1.
type status uint32

func (s status) state() state {
	return state(s >> 9 & 0xf)
}

// First status bits.
const (
	statusRsvd0 status = iota
	statusRsvd1
	statusRsvd2
	statusAuthSeqError
	statusRsvdSDIO
	statusAppCmd
	statusFXEvent
	statusRsvd7
	statusReadyForData
)

// Upper bound status bits.
const (
	statusEraseReset status = iota + 13
	statusECCDisabled
	statusWPEraseSkip
	statusCSDOverwrite
	_
	_
	statusGenericError
	statusControllerError // internal card controller error
	statusECCFailed
	statusIllegalCommand
	statusComCRCError // CRC check of previous command failed
	statusLockUnlockFailed
	statusCardIsLocked    // Signals that the card is locked by the host.
	statusWPViolation     // Write protected violation
	statusEraseParamError // invalid write block selection for erase
	statusEraseSeqError   // error in erase sequence
	statusBlockLenError   // tx block length not allowed
	statusAddrError       // misaligned address
	statusAddrOutOfRange  // address out of range
)

// r1 is the normal response to a command.
type r1 struct {
	data [48 / 8]byte // 48 bits of response.
}

func (r *r1) RawCopy() [6]byte { return r.data }
func (r *r1) startbit() bool {
	return r.data[0]&(1<<7) != 0
}
func (r *r1) txbit() bool {
	return r.data[0]&(1<<6) != 0
}
func (r *r1) cmdidx() uint8 {
	return r.data[0] & 0b11_1111
}
func (r *r1) cardstatus() status {
	return status(binary.BigEndian.Uint32(r.data[1:5]))
}
func (r *r1) CRC7() uint8  { return r.data[5] >> 1 }
func (r *r1) endbit() bool { return r.data[5]&1 != 0 }

func (r *r1) IsValid() bool {
	return r.endbit() && CRC7(r.data[:5]) == r.CRC7()
}

type r6 struct {
	data [48 / 8]byte
}

func (r *r6) RawCopy() [6]byte { return r.data }
func (r *r6) startbit() bool {
	return r.data[0]&(1<<7) != 0
}
func (r *r6) txbit() bool {
	return r.data[0]&(1<<6) != 0
}
func (r *r6) cmdidx() uint8 {
	return r.data[0] & 0b11_1111
}
func (r *r6) rca() uint16 {
	return binary.BigEndian.Uint16(r.data[1:3])
}
func (r *r6) CardStatus() status {
	moveBit := func(b status, from, to uint) status {
		return (b & (1 << from)) >> from << to
	}
	// See 4.9.5 R6 (Published RCA response) of the SD Simplified Specification.
	s := status(binary.BigEndian.Uint16(r.data[1:5]))
	s = moveBit(s, 13, 19)
	s = moveBit(s, 14, 22)
	s = moveBit(s, 15, 23)
	return s
}
func (r *r6) CRC7() uint8  { return r.data[5] >> 1 }
func (r *r6) endbit() bool { return r.data[5]&1 != 0 }

func (r *r6) IsValid() bool {
	return r.endbit() && CRC7(r.data[:5]) == r.CRC7()
}

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[statusRsvd0-0]
	_ = x[statusRsvd1-1]
	_ = x[statusRsvd2-2]
	_ = x[statusAuthSeqError-3]
	_ = x[statusRsvdSDIO-4]
	_ = x[statusAppCmd-5]
	_ = x[statusFXEvent-6]
	_ = x[statusRsvd7-7]
	_ = x[statusReadyForData-8]
	_ = x[statusEraseReset-13]
	_ = x[statusECCDisabled-14]
	_ = x[statusWPEraseSkip-15]
	_ = x[statusCSDOverwrite-16]
	_ = x[statusGenericError-19]
	_ = x[statusControllerError-20]
	_ = x[statusECCFailed-21]
	_ = x[statusIllegalCommand-22]
	_ = x[statusComCRCError-23]
	_ = x[statusLockUnlockFailed-24]
	_ = x[statusCardIsLocked-25]
	_ = x[statusWPViolation-26]
	_ = x[statusEraseParamError-27]
	_ = x[statusEraseSeqError-28]
	_ = x[statusBlockLenError-29]
	_ = x[statusAddrError-30]
	_ = x[statusAddrOutOfRange-31]
}

const (
	_status_name_0 = "Rsvd0Rsvd1Rsvd2AuthSeqErrorRsvdSDIOAppCmdFXEventRsvd7ReadyForData"
	_status_name_1 = "EraseResetECCDisabledWPEraseSkipCSDOverwrite"
	_status_name_2 = "GenericErrorControllerErrorECCFailedIllegalCommandComCRCErrorLockUnlockFailedCardIsLockedWPViolationEraseParamErrorEraseSeqErrorBlockLenErrorAddrErrorAddrOutOfRange"
)

var (
	_status_index_0 = [...]uint8{0, 5, 10, 15, 27, 35, 41, 48, 53, 65}
	_status_index_1 = [...]uint8{0, 10, 21, 32, 44}
	_status_index_2 = [...]uint8{0, 12, 27, 36, 50, 61, 77, 89, 100, 115, 128, 141, 150, 164}
)

func (i status) string() string {
	switch {
	case i <= 8:
		return _status_name_0[_status_index_0[i]:_status_index_0[i+1]]
	case 13 <= i && i <= 16:
		i -= 13
		return _status_name_1[_status_index_1[i]:_status_index_1[i+1]]
	case 19 <= i && i <= 31:
		i -= 19
		return _status_name_2[_status_index_2[i]:_status_index_2[i+1]]
	default:
		return ""
	}
}

func (s status) String() string {
	return string(s.appendf(nil, ','))
}

func (s status) appendf(b []byte, delim byte) []byte {
	b = append(b, s.state().String()...)
	b = append(b, '[')
	if s == 0 {
		return append(b, ']')
	}
	for bit := 0; bit < 32; bit++ {
		if s&(1<<bit) != 0 {
			b = append(b, status(bit).string()...)
			b = append(b, delim)
		}
	}
	b = append(b[:len(b)-1], ']')
	return b
}

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[stateIdle-0]
	_ = x[stateReady-1]
	_ = x[stateIdent-2]
	_ = x[stateStby-3]
	_ = x[stateTran-4]
	_ = x[stateData-5]
	_ = x[stateRcv-6]
	_ = x[statePrg-7]
	_ = x[stateDis-8]
}

const _state_name = "IdleReadyIdentStbyTranDataRcvPrgDis"

var _state_index = [...]uint8{0, 4, 9, 14, 18, 22, 26, 29, 32, 35}

func (i state) String() string {
	if i >= state(len(_state_index)-1) {
		return "<reserved state>"
	}
	return _state_name[_state_index[i]:_state_index[i+1]]
}
