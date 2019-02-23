package espat

import (
	"strconv"
	"time"
)

// DialUDP makes a UDP network connection. raadr is the port that the messages will
// be sent to, and laddr is the port that will be listened to in order to
// receive incoming messages.
func (d Device) DialUDP(network string, laddr, raddr *UDPAddr) (*SerialConn, error) {
	addr := raddr.IP.String()
	sendport := strconv.Itoa(raddr.Port)
	listenport := strconv.Itoa(laddr.Port)

	// disconnect any old socket
	d.DisconnectSocket()

	// connect new socket
	d.ConnectUDPSocket(addr, sendport, listenport)

	return &SerialConn{Adaptor: &d, laddr: laddr, raddr: raddr}, nil
}

// ListenUDP listens for UDP connections on the port listed in laddr.
func (d Device) ListenUDP(network string, laddr *UDPAddr) (*SerialConn, error) {
	addr := "0"
	sendport := "0"
	listenport := strconv.Itoa(laddr.Port)

	// disconnect any old socket
	d.DisconnectSocket()

	// connect new socket
	d.ConnectUDPSocket(addr, sendport, listenport)

	return &SerialConn{Adaptor: &d, laddr: laddr}, nil
}

// SerialConn is a loosely net.Conn compatible intended to support
// TCP/UDP over serial.
type SerialConn struct {
	Adaptor *Device
	laddr   *UDPAddr
	raddr   *UDPAddr
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
func (c *SerialConn) Write(b []byte) (n int, err error) {
	// specify that is a data transfer to the
	// currently open socket, not commands to the ESP8266/ESP32.
	c.Adaptor.StartSocketSend(len(b))
	return c.Adaptor.Write(b)
}

// Close closes the connection.
// Currently only supports a single Read or Write operations without blocking.
func (c *SerialConn) Close() error {
	c.Adaptor.DisconnectSocket()
	return nil
}

// LocalAddr returns the local network address.
func (c *SerialConn) LocalAddr() UDPAddr {
	return *c.laddr
}

// RemoteAddr returns the remote network address.
func (c *SerialConn) RemoteAddr() UDPAddr {
	return *c.laddr
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

// ParseIP parses s as an IP address, returning the result.
func ParseIP(s string) IP {
	return IP([]byte(s))
}

// String returns the string form of the IP address ip.
func (ip IP) String() string {
	return string(ip)
}
