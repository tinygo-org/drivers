package enc28j60

type ErrorCode uint8

const (
	errUndefined ErrorCode = iota
	// written packet exceeds size 64..1500
	ErrBufferSize
	// got rev=0. is dev connected?
	ErrBadRev
	// mac addr len not 6
	ErrBadMac
	// read deadline exceeded
	ErrRXDeadlineExceeded
	// CRC checksum fail
	ErrCRC
	// IO error
	ErrIO
)

// Implements error interface.
func (err ErrorCode) Error() string {
	switch err {
	case ErrBufferSize:
		return "written packet exceeds size 64..1500"
	case ErrBadRev:
		return "got rev=0. is dev connected?"
	case ErrBadMac:
		return "mac addr len not 6"
	case ErrRXDeadlineExceeded:
		return "rx deadline exceeded"
	case ErrCRC:
		return "CRC error"
	case ErrIO:
		return "IO error"
	}
	return "undefined"
}

// Gets the numeric representation of an ENC28J60
// error. Avoids usage of strings if memory consumption is an object.
func GetErrorCode(c error) uint8 {
	if u, ok := c.(ErrorCode); ok {
		return uint8(u)
	}
	return 0
}
