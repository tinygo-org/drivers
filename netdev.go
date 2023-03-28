package drivers

import (
	"errors"
	"net"
	"time"
	_ "unsafe" // to use go:linkname
)

const (
	AF_INET       = 0x2
	SOCK_STREAM   = 0x1
	SOCK_DGRAM    = 0x2
	SOL_SOCKET    = 0x1
	SO_KEEPALIVE  = 0x9
	SOL_TCP       = 0x6
	TCP_KEEPINTVL = 0x5
	IPPROTO_TCP   = 0x6
	IPPROTO_UDP   = 0x11
	// Made up, not a real IP protocol number.  This is used to create a
	// TLS socket on the device, assuming the device supports mbed TLS.
	IPPROTO_TLS = 0xFE
	F_SETFL     = 0x4
)

// GethostByName() errors
var (
	ErrHostUnknown = errors.New("Host unknown")
)

// Socket errors
var (
	ErrFamilyNotSupported   = errors.New("Address family not supported")
	ErrProtocolNotSupported = errors.New("Socket protocol/type not supported")
	ErrNoMoreSockets        = errors.New("No more sockets")
	ErrClosingSocket        = errors.New("Error closing socket")
	ErrNotSupported         = errors.New("Not supported")
)

// Duplicate of non-exported net.errTimeout
var ErrTimeout error = &timeoutError{}

type timeoutError struct{}

func (e *timeoutError) Error() string   { return "i/o timeout" }
func (e *timeoutError) Timeout() bool   { return true }
func (e *timeoutError) Temporary() bool { return true }

//go:linkname UseNetdev net.useNetdev
func UseNetdev(dev netdever)

// Netdev is TinyGo's network device driver model.  Network drivers implement
// the netdever interface, providing a common network I/O interface to TinyGo's
// "net" package.  The interface is modeled after the BSD socket interface.
// net.Conn implementations (TCPConn, UDPConn, and TLSConn) use the netdev
// interface for device I/O access.
//
// A netdever is passed to the "net" package using UseNetdev().
//
// Just like a net.Conn, multiple goroutines may invoke methods on a netdever
// simultaneously.
//
// NOTE: The netdever interface is mirrored in tinygo/src/net/netdev.go.
// NOTE: If making changes to this interface, mirror the changes in
// NOTE: tinygo/src/net/netdev.go, and vice-versa.

type netdever interface {

	// GetHostByName returns the IP address of either a hostname or IPv4
	// address in standard dot notation
	GetHostByName(name string) (net.IP, error)

	// Berkely Sockets-like interface, Go-ified.  See man page for socket(2), etc.
	Socket(domain int, stype int, protocol int) (int, error)
	Bind(sockfd int, ip net.IP, port int) error
	Connect(sockfd int, host string, ip net.IP, port int) error
	Listen(sockfd int, backlog int) error
	Accept(sockfd int, ip net.IP, port int) (int, error)
	Send(sockfd int, buf []byte, flags int, deadline time.Time) (int, error)
	Recv(sockfd int, buf []byte, flags int, deadline time.Time) (int, error)
	Close(sockfd int) error
	SetSockOpt(sockfd int, level int, opt int, value interface{}) error
}
