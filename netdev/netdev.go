// L3/L4 network/transport layer

package netdev

import (
	"errors"
	"net/netip"
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
	ErrMalAddr     = errors.New("Malformed address")
)

// Socket errors
var (
	ErrFamilyNotSupported   = errors.New("Address family not supported")
	ErrProtocolNotSupported = errors.New("Socket protocol/type not supported")
	ErrStartingDHCPClient   = errors.New("Error starting DHPC client")
	ErrNoMoreSockets        = errors.New("No more sockets")
	ErrClosingSocket        = errors.New("Error closing socket")
	ErrNotSupported         = errors.New("Not supported")
	ErrInvalidSocketFd      = errors.New("Invalid socket fd")
)

// Duplicate of non-exported net.errTimeout
var ErrTimeout error = &timeoutError{}

type timeoutError struct{}

func (e *timeoutError) Error() string   { return "i/o timeout" }
func (e *timeoutError) Timeout() bool   { return true }
func (e *timeoutError) Temporary() bool { return true }

//go:linkname UseNetdev net.useNetdev
func UseNetdev(dev Netdever)

// Netdever is TinyGo's OSI L3/L4 network/transport layer interface.  Network
// drivers implement the Netdever interface, providing a common network L3/L4
// interface to TinyGo's "net" package.  net.Conn implementations (TCPConn,
// UDPConn, and TLSConn) use the Netdever interface for device I/O access.
//
// A Netdever is passed to the "net" package using UseNetdev().
//
// Just like a net.Conn, multiple goroutines may invoke methods on a Netdever
// simultaneously.
//
// NOTE: The Netdever interface is mirrored in tinygo/src/net/netdev.go.
// NOTE: If making changes to this interface, mirror the changes in
// NOTE: tinygo/src/net/netdev.go, and vice-versa.

type Netdever interface {

	// GetHostByName returns the IP address of either a hostname or IPv4
	// address in standard dot notation
	GetHostByName(name string) (netip.Addr, error)

	// Addr returns IP address assigned to the interface, either by
	// DHCP or statically
	Addr() (netip.Addr, error)

	// Berkely Sockets-like interface, Go-ified.  See man page for socket(2), etc.
	Socket(domain int, stype int, protocol int) (int, error)
	Bind(sockfd int, ip netip.AddrPort) error
	Connect(sockfd int, host string, ip netip.AddrPort) error
	Listen(sockfd int, backlog int) error
	Accept(sockfd int) (int, netip.AddrPort, error)
	Send(sockfd int, buf []byte, flags int, deadline time.Time) (int, error)
	Recv(sockfd int, buf []byte, flags int, deadline time.Time) (int, error)
	Close(sockfd int) error
	SetSockOpt(sockfd int, level int, opt int, value interface{}) error
}
