package esp8266

import (
	"strconv"
	"time"
)

// IP here to serve as compatible type. until TinyGo can compile the net package.
type IP []byte

// UDPAddr here to serve as compatible type. until TinyGo can compile the net package.
type UDPAddr struct {
	IP   IP
	Port int
	Zone string // IPv6 scoped addressing zone; added in Go 1.1
}

// DialUDP makes a UDP network connection. raadr is the port that the messages will
// be sent to, and laddr is the port that will be listened to in order to
// receive incoming messages.
func (d Device) DialUDP(network string, laddr, raddr *UDPAddr) (*SerialConn, error) {
	// TODO: get addr out of the raddr.IP
	addr := "0"
	sendport := strconv.Itoa(raddr.Port)
	listenport := strconv.Itoa(laddr.Port)

	d.ConnectUDPSocket(addr, sendport, listenport)

	return &SerialConn{Adaptor: d}, nil
}

// ListenUDP listens for UDP connections on the port listed in laddr.
func (d Device) ListenUDP(network string, laddr *UDPAddr) (*SerialConn, error) {
	addr := "0"
	sendport := "0"
	listenport := strconv.Itoa(laddr.Port)

	d.ConnectUDPSocket(addr, sendport, listenport)

	return &SerialConn{Adaptor: d}, nil
}

// SerialConn is a loosely net.Conn compatible intended to support
// TCP/UDP over serial.
type SerialConn struct {
	Adaptor Device
}

// Read reads data from the connection.
// Read can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetReadDeadline.
func (c *SerialConn) Read(b []byte) (n int, err error) {
	// this should read only the data that has been received via "+IPD"
	return c.Adaptor.Read(b)
}

// Write writes data to the connection.
// Write can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetWriteDeadline.
func (c *SerialConn) Write(b []byte) (n int, err error) {
	// specify that is a data transfer to the
	// currently open socket, not commands to the ESP8266.
	c.Adaptor.StartSocketSend(len(b))
	return c.Adaptor.Write(b)
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (c *SerialConn) Close() error {
	c.Adaptor.DisconnectSocket()
	return nil
}

// LocalAddr returns the local network address.
// func (c SLIPConn) LocalAddr() net.Addr {
// 	return nil
// }

// RemoteAddr returns the remote network address.
// func (c SLIPConn) RemoteAddr() net.Addr {
// 	return nil
// }

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
