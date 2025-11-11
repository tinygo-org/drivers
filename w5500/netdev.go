package w5500

import (
	"errors"
	"net"
	"net/netip"
	"os"
	"runtime"
	"time"

	"tinygo.org/x/drivers/netdev"
)

type socket struct {
	sockn    uint8
	protocol uint8
	port     uint16
	inUse    bool
	closed   bool
}

func (s *socket) setProtocol(proto byte) *socket {
	s.protocol = proto
	return s
}

func (s *socket) setPort(port uint16) *socket {
	s.port = port
	return s
}

func (s *socket) setInUse(inUse bool) *socket {
	s.inUse = inUse
	return s
}

func (s *socket) setClosed(closed bool) *socket {
	s.closed = closed
	return s
}

func (s *socket) reset() {
	s.protocol = 0
	s.port = 0
	s.inUse = false
	s.closed = false
}

// GetHostByName resolves the given host name to an IP address.
func (d *Device) GetHostByName(name string) (netip.Addr, error) {
	d.mu.Lock()
	dns := d.dns
	d.mu.Unlock()

	if dns == nil {
		return netip.Addr{}, netdev.ErrNotSupported
	}
	return dns(name)
}

func (d *Device) Socket(domain int, stype int, protocol int) (int, error) {
	if domain != netdev.AF_INET {
		return -1, netdev.ErrFamilyNotSupported
	}
	switch {
	case stype == netdev.SOCK_STREAM && protocol == netdev.IPPROTO_TCP:
	case stype == netdev.SOCK_DGRAM && protocol == netdev.IPPROTO_UDP:
	default:
		return -1, errors.New("unsupported combination of socket type and protocol")
	}

	var proto byte
	switch protocol {
	case netdev.IPPROTO_TCP:
		proto = 1 // TCP
	case netdev.IPPROTO_UDP:
		proto = 2 // UDP
	default:
		return -1, netdev.ErrNotSupported
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	sockfd, sock, err := d.nextSocket()
	if err != nil {
		return -1, err
	}

	d.openSocket(sock.sockn, proto)

	sock.setProtocol(proto).setInUse(true)
	return sockfd, nil
}

func (d *Device) openSocket(sockn uint8, proto byte) {
	d.writeByte(sockMode, sockAddr(sockn), proto&0x0F)
}

func (d *Device) Bind(sockfd int, ip netip.AddrPort) error {
	// The IP address is irrelevant. The configured ip will always be used.
	port := ip.Port()

	d.mu.Lock()
	defer d.mu.Unlock()

	sock, err := d.socket(sockfd)
	if err != nil {
		return err
	}

	if err = d.bindSocket(sock.sockn, port); err != nil {
		return errors.New("could not set socket port: " + err.Error())
	}

	sock.setPort(port)
	return nil
}

func (d *Device) bindSocket(sockn uint8, port uint16) error {
	d.writeUint16(sockSrcPort, sockAddr(sockn), port)
	d.socketSendCmd(sockn, sockCmdOpen)
	if d.sockStatus(sockn) == sockStatusClosed {
		return errors.New("socket is closed after binding")
	}
	return nil
}

// SetSockOpt sets the socket option for the given socket file descriptor.
// It is not supported by the W5500, so it always returns an error.
func (d *Device) SetSockOpt(int, int, int, any) error {
	return netdev.ErrNotSupported
}

// Connect establishes a connection to the specified host and port or ip and port.
//
// If the host is an empty string, it will use the provided ip address and port,
// otherwise it will resolve the host name to an IP address.
func (d *Device) Connect(sockfd int, host string, ip netip.AddrPort) error {
	destIP := ip.Addr()
	if host != "" {
		var err error
		destIP, err = d.GetHostByName(host)
		if err != nil {
			return errors.New("could not resolve host " + host + ":" + err.Error())
		}
	}
	if !destIP.IsValid() || !destIP.Is4() {
		return errors.New("invalid destination IP address: " + destIP.String())
	}
	port := ip.Port()

	d.mu.Lock()
	defer d.mu.Unlock()

	sock, err := d.socket(sockfd)
	if err != nil {
		return err
	}

	d.write(sockDestIP, sockAddr(sock.sockn), destIP.AsSlice())
	d.writeUint16(sockDestPort, sockAddr(sock.sockn), port)
	d.socketSendCmd(sock.sockn, sockCmdOpen)
	return nil
}

// Listen sets the socket to listen for incoming connections on the specified socket file descriptor.
//
// The backlog parameter is ignored, as the W5500 does not support it.
func (d *Device) Listen(sockfd int, _ int) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	sock, err := d.socket(sockfd)
	if err != nil {
		return err
	}

	if sock.protocol != 1 { // Only TCP sockets can listen
		return errors.New("not a TCP socket")
	}

	if err = d.listen(sock.sockn); err != nil {
		return errors.New("could not send listen command: " + err.Error())
	}
	return nil
}

func (d *Device) listen(sockn uint8) error {
	state := d.sockStatus(sockn)
	if state != sockStatusInit {
		return errors.New("socket is not in the initial state")
	}
	d.socketSendCmd(sockn, sockCmdListen)
	return nil
}

// Accept waits for an incoming connection on the specified socket file descriptor.
func (d *Device) Accept(sockfd int) (int, netip.AddrPort, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	lsock, err := d.socket(sockfd)
	if err != nil {
		return -1, netip.AddrPort{}, errors.New("could not get socket: " + err.Error())
	}

	if err = d.waitForEstablished(lsock.sockn); err != nil {
		return -1, netip.AddrPort{}, err
	}

	// Acquire a new socket for the listening connection.
	csockfd, csock, err := d.nextSocket()
	if err != nil {
		return -1, netip.AddrPort{}, err
	}
	// Swap the socket numbers of the client and listening sockets.
	lsock.sockn, csock.sockn = csock.sockn, lsock.sockn

	// Rebind the listening socket to the local address and port and start listening.
	d.openSocket(lsock.sockn, lsock.protocol)
	if err = d.bindSocket(lsock.sockn, lsock.port); err != nil {
		return -1, netip.AddrPort{}, errors.New("could not bind listening socket: " + err.Error())
	}
	if err = d.listen(lsock.sockn); err != nil {
		return -1, netip.AddrPort{}, errors.New("could not set listening socket: " + err.Error())
	}

	csock.setInUse(true)

	remoteIP := d.remoteIP(csock.sockn)
	return csockfd, remoteIP, nil
}

func (d *Device) waitForEstablished(sockn uint8) error {
	for {
		status := d.sockStatus(sockn)
		switch status {
		case sockStatusEstablished:
			return nil
		case sockStatusClosed:
			return net.ErrClosed
		case sockStatusCloseWait:
			// The server closed the connection, so we need to reset the socket
			// and set it to listen again.
			if err := d.listen(sockn); err != nil {
				return errors.New("could not set socket to listen: " + err.Error())
			}
		}

		d.irqPoll(sockn, sockIntConnect|sockIntDisconnect, time.Time{})
	}
}

func (d *Device) remoteIP(sockn uint8) netip.AddrPort {
	var rip [4]byte
	d.read(sockDestIP, sockAddr(sockn), rip[:])

	var rport [2]byte
	d.read(sockDestPort, sockAddr(sockn), rport[:])

	return netip.AddrPortFrom(netip.AddrFrom4(rip), uint16(rport[0])<<8|uint16(rport[1]))
}

// Send sends data to the socket with the given file descriptor.
// It blocks until all data is sent or the deadline is reached.
func (d *Device) Send(sockfd int, buf []byte, _ int, deadline time.Time) (int, error) {
	bufLen := len(buf)
	if bufLen <= d.maxSockSize {
		// Fast path for small buffers.
		return d.sendChunk(sockfd, buf, deadline)
	}

	var n int
	for i := 0; i < bufLen; i += d.maxSockSize {
		end := i + d.maxSockSize
		if end > bufLen {
			end = bufLen
		}

		sent, err := d.sendChunk(sockfd, buf[i:end], deadline)
		if err != nil {
			return n, errors.New("could not send chunk: " + err.Error())
		}
		n += sent
	}
	return n, nil
}

func (d *Device) sendChunk(sockfd int, buf []byte, deadline time.Time) (int, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	sock, err := d.socket(sockfd)
	if err != nil {
		return 0, errors.New("could not get socket: " + err.Error())
	}
	if sock.closed {
		return 0, os.ErrClosed
	}

	bufLen := uint16(len(buf))
	if err = d.waitForFreeBuffer(sock.sockn, bufLen, deadline); err != nil {
		return 0, err
	}

	sendPtr := d.readUint16(sockTXWritePtr, sockAddr(sock.sockn))

	d.write(sendPtr, sock.sockn<<2|0b10, buf)
	d.writeUint16(sockTXWritePtr, sockAddr(sock.sockn), sendPtr+bufLen)
	d.writeByte(sockCmd, sockAddr(sock.sockn), sockCmdSend)

	irq := d.irqPoll(sock.sockn, sockIntSendOK|sockIntDisconnect|sockIntTimeout, deadline)
	switch {
	case irq == sockIntUnknown:
		return 0, os.ErrDeadlineExceeded
	case irq&sockIntDisconnect != 0:
		sock.setClosed(true)
		return 0, net.ErrClosed
	case irq&sockIntTimeout != 0:
		return 0, netdev.ErrTimeout
	default:
		return int(bufLen), nil
	}
}

func (d *Device) waitForFreeBuffer(sockn uint8, len uint16, deadline time.Time) error {
	for {
		freeSize := d.readUint16(sockTXFreeSize, sockAddr(sockn))
		if freeSize >= len {
			return nil
		}

		if !deadline.IsZero() && time.Now().After(deadline) {
			return netdev.ErrTimeout
		}

		status := d.sockStatus(sockn)
		switch status {
		case sockStatusEstablished, sockStatusCloseWait:
		default:
			return errors.New("socket is not in a valid state for sending data")
		}

		d.mu.Unlock()

		time.Sleep(time.Millisecond)

		d.mu.Lock()
	}
}

// Recv reads data from the socket with the given file descriptor into the provided buffer.
// It blocks until data is available or the deadline is reached.
func (d *Device) Recv(sockfd int, buf []byte, _ int, deadline time.Time) (int, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	sock, err := d.socket(sockfd)
	if err != nil {
		return 0, errors.New("could not get socket: " + err.Error())
	}
	if sock.closed {
		return 0, os.ErrClosed
	}

	size, err := d.waitForData(sock, deadline)
	if err != nil {
		return 0, err
	}

	recvPtr := d.readUint16(sockRXReadPtr, sockAddr(sock.sockn))

	buf = buf[:min(size, len(buf))]
	d.read(recvPtr, sock.sockn<<2|0b00011, buf)
	d.writeUint16(sockRXReadPtr, sockAddr(sock.sockn), recvPtr+uint16(len(buf)))
	d.socketSendCmd(sock.sockn, sockCmdRecv)

	return len(buf), nil
}

func (d *Device) waitForData(sock *socket, deadline time.Time) (int, error) {
	for {
		recvdSize := d.readUint16(sockRXReceivedSize, sockAddr(sock.sockn))
		if recvdSize > 0 {
			return int(recvdSize), nil
		}

		irq := d.irqPoll(sock.sockn, sockIntReceive|sockIntDisconnect, deadline)
		switch {
		case irq == sockIntUnknown:
			return 0, os.ErrDeadlineExceeded
		case irq&sockIntDisconnect != 0:
			sock.setClosed(true)
			return 0, net.ErrClosed
		}
	}
}

// Close closes the socket with the given file descriptor.
func (d *Device) Close(sockfd int) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	sock, err := d.socket(sockfd)
	if err != nil {
		return err
	}

	d.socketSendCmd(sock.sockn, sockCmdClose)
	sock.reset()
	return nil
}

func (d *Device) nextSocket() (int, *socket, error) {
	for i, sock := range d.sockets {
		if sock.inUse {
			continue
		}
		return i, sock, nil
	}
	return -1, nil, netdev.ErrNoMoreSockets
}

func (d *Device) socket(sockfd int) (*socket, error) {
	if sockfd < 0 || sockfd >= len(d.sockets) {
		return nil, netdev.ErrInvalidSocketFd
	}
	return d.sockets[sockfd], nil
}

func (d *Device) socketSendCmd(sockn uint8, cmd byte) {
	d.writeByte(sockCmd, sockAddr(sockn), cmd)
	for d.readByte(sockCmd, sockAddr(sockn)) != 0 {
		runtime.Gosched()
	}
}

func (d *Device) sockStatus(sockn uint8) int {
	return int(d.readByte(sockStatus, sockAddr(sockn)))
}

func sockAddr(sockn uint8) uint8 {
	return sockn<<2 | 0b0001
}
