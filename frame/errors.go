package frame

type ErrorCode uint8

const (
	errUndefined ErrorCode = iota
	// out of buff bounds
	ErrOutOfBound
	// buff size not in 64..1500
	ErrBufferSize
	// buffer too small for unmarshal/marshal
	ErrBufferTooSmall
	ErrBadRev // errors.New("got rev=0. is dev connected?")
	ErrBadMac //errors.New("mac addr len not 6")
	ErrBadIP
	ErrBadARP
	ErrUnableToResolveARP //= errors.New("unable to resolve ARP")
	ErrARPViolation       //= errors.New("ARP protocol violation")
	ErrIPNotImplemented
	ErrIO
	ErrNoTCPPseudoHeader
	ErrCodeMax
)

func (err ErrorCode) Error() string {
	switch err {
	case ErrOutOfBound:
		return "out of buff bounds"
	case ErrBufferSize:
		return "buff size not in 64..1500"
	case ErrBufferTooSmall:
		return "buffer too small for unmarshal/marshal"
	case ErrBadRev:
		return "got rev=0. is dev connected?"
	case ErrBadIP:
		return "invalid IP address"
	case ErrBadMac:
		return "mac addr len not 6"
	case ErrUnableToResolveARP:
		return "unable to resolve ARP"
	case ErrARPViolation:
		return "ARP protocol violation"
	case ErrIO:
		return "I/O"
	case ErrIPNotImplemented:
		return "internet protocol procedure not implemented"
	case ErrNoTCPPseudoHeader:
		return "could not form pseudo header for TCP"
	}
	return "undefined"
}
