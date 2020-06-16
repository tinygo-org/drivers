// package net is intended to provide compatible interfaces with the
// Go standard library's net package.
package net

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

// DialUDP makes a UDP network connection. raadr is the port that the messages will
// be sent to, and laddr is the port that will be listened to in order to
// receive incoming messages.
func DialUDP(network string, laddr, raddr *UDPAddr) (*UDPSerialConn, error) {
	addr := raddr.IP.String()
	sendport := strconv.Itoa(raddr.Port)
	listenport := strconv.Itoa(laddr.Port)

	// disconnect any old socket
	//ActiveDevice.DisconnectSocket()

	// connect new socket
	err := ActiveDevice.ConnectUDPSocket(addr, sendport, listenport)
	if err != nil {
		return nil, err
	}

	return &UDPSerialConn{SerialConn: SerialConn{Adaptor: ActiveDevice}, laddr: laddr, raddr: raddr}, nil
}

// ListenUDP listens for UDP connections on the port listed in laddr.
func ListenUDP(network string, laddr *UDPAddr) (*UDPSerialConn, error) {
	addr := "0"
	sendport := "0"
	listenport := strconv.Itoa(laddr.Port)

	// disconnect any old socket
	ActiveDevice.DisconnectSocket()

	// connect new socket
	err := ActiveDevice.ConnectUDPSocket(addr, sendport, listenport)
	if err != nil {
		return nil, err
	}

	return &UDPSerialConn{SerialConn: SerialConn{Adaptor: ActiveDevice}, laddr: laddr}, nil
}

// DialTCP makes a TCP network connection. raadr is the port that the messages will
// be sent to, and laddr is the port that will be listened to in order to
// receive incoming messages.
func DialTCP(network string, laddr, raddr *TCPAddr) (*TCPSerialConn, error) {
	addr := raddr.IP.String()
	sendport := strconv.Itoa(raddr.Port)

	// disconnect any old socket?
	//ActiveDevice.DisconnectSocket()

	// connect new socket
	err := ActiveDevice.ConnectTCPSocket(addr, sendport)
	if err != nil {
		return nil, err
	}

	return &TCPSerialConn{SerialConn: SerialConn{Adaptor: ActiveDevice}, laddr: laddr, raddr: raddr}, nil
}

// Dial connects to the address on the named network.
// It tries to provide a mostly compatible interface
// to net.Dial().
func Dial(network, address string) (Conn, error) {
	switch network {
	case "tcp":
		raddr, err := ResolveTCPAddr(network, address)
		if err != nil {
			return nil, err
		}

		c, e := DialTCP(network, &TCPAddr{}, raddr)
		return c.opConn(), e
	case "udp":
		raddr, err := ResolveUDPAddr(network, address)
		if err != nil {
			return nil, err
		}

		c, e := DialUDP(network, &UDPAddr{}, raddr)
		return c.opConn(), e
	default:
		return nil, errors.New("invalid network for dial")
	}
}

// SerialConn is a loosely net.Conn compatible implementation
type SerialConn struct {
	Adaptor DeviceDriver
}

// UDPSerialConn is a loosely net.Conn compatible intended to support
// UDP over serial.
type UDPSerialConn struct {
	SerialConn
	laddr *UDPAddr
	raddr *UDPAddr
}

// NewUDPSerialConn returns a new UDPSerialConn/
func NewUDPSerialConn(c SerialConn, laddr, raddr *UDPAddr) *UDPSerialConn {
	return &UDPSerialConn{SerialConn: c, raddr: raddr}
}

// TCPSerialConn is a loosely net.Conn compatible intended to support
// TCP over serial.
type TCPSerialConn struct {
	SerialConn
	laddr *TCPAddr
	raddr *TCPAddr
}

// NewTCPSerialConn returns a new TCPSerialConn/
func NewTCPSerialConn(c SerialConn, laddr, raddr *TCPAddr) *TCPSerialConn {
	return &TCPSerialConn{SerialConn: c, raddr: raddr}
}

// Read reads data from the connection.
// TODO: implement the full method functionality:
// Read can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetReadDeadline.
func (c *SerialConn) Read(b []byte) (n int, err error) {
	// read only the data that has been received via "+IPD" socket
	return c.Adaptor.ReadSocket(b)
}

// Write writes data to the connection.
// TODO: implement the full method functionality for timeouts.
// Write can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetWriteDeadline.
func (c *UDPSerialConn) Write(b []byte) (n int, err error) {
	n, err = c.Adaptor.Write(b)
	return n, err
}

// Write writes data to the connection.
// TODO: implement the full method functionality for timeouts.
// Write can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetWriteDeadline.
func (c *SerialConn) Write(b []byte) (n int, err error) {
	// specify that is a data transfer to the
	// currently open socket, not commands to the ESP8266/ESP32.
	err = c.Adaptor.StartSocketSend(len(b))
	if err != nil {
		return
	}
	n, err = c.Adaptor.Write(b)
	if err != nil {
		return n, err
	}
	/* TODO(bcg): this is kind of specific to espat, should maybe refactor */
	_, err = c.Adaptor.Response(1000)
	if err != nil {
		return n, err
	}
	return n, err
}

// Close closes the connection.
// Currently only supports a single Read or Write operations without blocking.
func (c *SerialConn) Close() error {
	c.Adaptor.DisconnectSocket()
	return nil
}

// LocalAddr returns the local network address.
func (c *UDPSerialConn) LocalAddr() Addr {
	return c.laddr.opAddr()
}

// RemoteAddr returns the remote network address.
func (c *UDPSerialConn) RemoteAddr() Addr {
	return c.laddr.opAddr()
}

func (c *UDPSerialConn) opConn() Conn {
	if c == nil {
		return nil
	}
	return c
}

// LocalAddr returns the local network address.
func (c *TCPSerialConn) LocalAddr() Addr {
	return c.laddr.opAddr()
}

// RemoteAddr returns the remote network address.
func (c *TCPSerialConn) RemoteAddr() Addr {
	return c.laddr.opAddr()
}

func (c *TCPSerialConn) opConn() Conn {
	if c == nil {
		return nil
	}
	return c
}

// SetDeadline sets the read and write deadlines associated
// with the connection. It is equivalent to calling both
// SetReadDeadline and SetWriteDeadline.
//
// A deadline is an absolute time after which I/O operations
// fail with a timeout (see type Error) instead of
// blocking. The deadline applies to all future and pending
// I/O, not just the immediately following call to Read or
// Write. After a deadline has been exceeded, the connection
// can be refreshed by setting a deadline in the future.
//
// An idle timeout can be implemented by repeatedly extending
// the deadline after successful Read or Write calls.
//
// A zero value for t means I/O operations will not time out.
func (c *SerialConn) SetDeadline(t time.Time) error {
	return nil
}

// SetReadDeadline sets the deadline for future Read calls
// and any currently-blocked Read call.
// A zero value for t means Read will not time out.
func (c *SerialConn) SetReadDeadline(t time.Time) error {
	return nil
}

// SetWriteDeadline sets the deadline for future Write calls
// and any currently-blocked Write call.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means Write will not time out.
func (c *SerialConn) SetWriteDeadline(t time.Time) error {
	return nil
}

// ResolveTCPAddr returns an address of TCP end point.
//
// The network must be a TCP network name.
//
func ResolveTCPAddr(network, address string) (*TCPAddr, error) {
	// TODO: make sure network is 'tcp'
	// separate domain from port, if any
	r := strings.Split(address, ":")
	addr, err := ActiveDevice.GetDNS(r[0])
	if err != nil {
		return nil, err
	}
	ip := IP(addr)
	if len(r) > 1 {
		port, e := strconv.Atoi(r[1])
		if e != nil {
			return nil, e
		}
		return &TCPAddr{IP: ip, Port: port}, nil
	}
	return &TCPAddr{IP: ip}, nil
}

// ResolveUDPAddr returns an address of UDP end point.
//
// The network must be a UDP network name.
//
func ResolveUDPAddr(network, address string) (*UDPAddr, error) {
	// TODO: make sure network is 'udp'
	// separate domain from port, if any
	r := strings.Split(address, ":")
	addr, err := ActiveDevice.GetDNS(r[0])
	if err != nil {
		return nil, err
	}
	ip := IP(addr)
	if len(r) > 1 {
		port, e := strconv.Atoi(r[1])
		if e != nil {
			return nil, e
		}
		return &UDPAddr{IP: ip, Port: port}, nil
	}

	return &UDPAddr{IP: ip}, nil
}

// The following definitions are here to support a Golang standard package
// net-compatible interface for IP until TinyGo can compile the net package.

// IP is an IP address. Unlike the standard implementation, it is only
// a buffer of bytes that contains the string form of the IP address, not the
// full byte format used by the Go standard .
type IP []byte

// UDPAddr here to serve as compatible type. until TinyGo can compile the net package.
type UDPAddr struct {
	IP   IP
	Port int
	Zone string // IPv6 scoped addressing zone; added in Go 1.1
}

// Network returns the address's network name, "udp".
func (a *UDPAddr) Network() string { return "udp" }

func (a *UDPAddr) String() string {
	if a == nil {
		return "<nil>"
	}
	if a.Port != 0 {
		return a.IP.String() + ":" + strconv.Itoa(a.Port)
	}
	return a.IP.String()
}

func (a *UDPAddr) opAddr() Addr {
	if a == nil {
		return nil
	}
	return a
}

// TCPAddr here to serve as compatible type. until TinyGo can compile the net package.
type TCPAddr struct {
	IP   IP
	Port int
	Zone string // IPv6 scoped addressing zone
}

// Network returns the address's network name, "tcp".
func (a *TCPAddr) Network() string { return "tcp" }

func (a *TCPAddr) String() string {
	if a == nil {
		return "<nil>"
	}
	if a.Port != 0 {
		return a.IP.String() + ":" + strconv.Itoa(a.Port)
	}
	return a.IP.String()
}

func (a *TCPAddr) opAddr() Addr {
	if a == nil {
		return nil
	}
	return a
}

// ParseIP parses s as an IP address, returning the result.
func ParseIP(s string) IP {
	return IP([]byte(s))
}

// String returns the string form of the IP address ip.
func (ip IP) String() string {
	return string(ip)
}

// Conn is a generic stream-oriented network connection.
// This interface is from the Go standard library.
type Conn interface {
	// Read reads data from the connection.
	// Read can be made to time out and return an Error with Timeout() == true
	// after a fixed time limit; see SetDeadline and SetReadDeadline.
	Read(b []byte) (n int, err error)

	// Write writes data to the connection.
	// Write can be made to time out and return an Error with Timeout() == true
	// after a fixed time limit; see SetDeadline and SetWriteDeadline.
	Write(b []byte) (n int, err error)

	// Close closes the connection.
	// Any blocked Read or Write operations will be unblocked and return errors.
	Close() error

	// LocalAddr returns the local network address.
	LocalAddr() Addr

	// RemoteAddr returns the remote network address.
	RemoteAddr() Addr

	// SetDeadline sets the read and write deadlines associated
	// with the connection. It is equivalent to calling both
	// SetReadDeadline and SetWriteDeadline.
	//
	// A deadline is an absolute time after which I/O operations
	// fail with a timeout (see type Error) instead of
	// blocking. The deadline applies to all future and pending
	// I/O, not just the immediately following call to Read or
	// Write. After a deadline has been exceeded, the connection
	// can be refreshed by setting a deadline in the future.
	//
	// An idle timeout can be implemented by repeatedly extending
	// the deadline after successful Read or Write calls.
	//
	// A zero value for t means I/O operations will not time out.
	SetDeadline(t time.Time) error

	// SetReadDeadline sets the deadline for future Read calls
	// and any currently-blocked Read call.
	// A zero value for t means Read will not time out.
	SetReadDeadline(t time.Time) error

	// SetWriteDeadline sets the deadline for future Write calls
	// and any currently-blocked Write call.
	// Even if write times out, it may return n > 0, indicating that
	// some of the data was successfully written.
	// A zero value for t means Write will not time out.
	SetWriteDeadline(t time.Time) error
}

// Addr represents a network end point address.
type Addr interface {
	Network() string // name of the network (for example, "tcp", "udp")
	String() string  // string form of address (for example, "192.0.2.1:25", "[2001:db8::1]:80")
}
