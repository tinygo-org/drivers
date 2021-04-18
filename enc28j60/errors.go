package enc28j60

type ErrorCode uint8

const (
	errOutOfBound ErrorCode = iota
	errBufferSize
	errBadRev // errors.New("got rev=0. is dev connected?")
	errBadMac //errors.New("mac addr len not 6")
	errBadIP
	errBadARP
	errUnableToResolveARP //= errors.New("unable to resolve ARP")
	errARPViolation       //= errors.New("ARP protocol violation")
	errIPNotImplemented
	errIO
)

func (err ErrorCode) Error() string {
	switch err {
	case errOutOfBound:
		return "out of buff bounds"
	case errBufferSize:
		return "buff size not in 64..1500"
	case errBadRev:
		return "got rev=0. is dev connected?"
	case errBadIP:
		return "invalid IP address"
	case errBadMac:
		return "mac addr len not 6"
	case errUnableToResolveARP:
		return "unable to resolve ARP"
	case errARPViolation:
		return "ARP protocol violation"
	case errIO:
		return "I/O"
	case errIPNotImplemented:
		return "internet protocol procedure not implemented"
	}
	return "undefined"
}
