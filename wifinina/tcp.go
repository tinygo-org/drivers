package wifinina

import (
	"fmt"
	"strconv"
	"time"

	"tinygo.org/x/drivers/net"
)

const (
	ReadBufferSize = 128
)

func (d *Device) NewDriver() net.DeviceDriver {
	return &Driver{dev: d, sock: NoSocketAvail}
}

type Driver struct {
	dev     *Device
	sock    uint8
	readBuf readBuffer
}

type readBuffer struct {
	data [ReadBufferSize]byte
	head int
	size int
}

func (drv *Driver) GetDNS(domain string) (string, error) {
	ipAddr, err := drv.dev.GetHostByName(domain)
	return ipAddr.String(), err
}

func (drv *Driver) ConnectTCPSocket(addr, portStr string) error {
	return drv.connectSocket(addr, portStr, ProtoModeTCP)
}

func (drv *Driver) ConnectSSLSocket(addr, portStr string) error {
	return drv.connectSocket(addr, portStr, ProtoModeTLS)
}

func (drv *Driver) connectSocket(addr, portStr string, mode uint8) error {

	// convert port to uint16
	p64, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return fmt.Errorf("could not convert port to uint16: %s", err.Error())
	}
	port := uint16(p64)

	// look up the hostname if necessary; if an IP address was specified, the
	// same will be returned.  Otherwise, an IPv4 for the hostname is returned.
	ipAddr, err := drv.dev.GetHostByName(addr)
	if err != nil {
		return err
	}
	ip := ipAddr.AsUint32()

	// check to see if socket is already set; if so, stop it
	if drv.sock != NoSocketAvail {
		if err := drv.stop(); err != nil {
			return err
		}
	}

	// get a socket from the device
	if drv.sock, err = drv.dev.GetSocket(); err != nil {
		return err
	}

	// attempt to start the client
	if err := drv.dev.StartClient(ip, port, drv.sock, mode); err != nil {
		return err
	}

	// FIXME: this 4 second timeout is simply mimicking the Arduino driver
	for now := time.Now(); time.Since(now) < 4*time.Second; {
		connected, err := drv.IsConnected()
		if err != nil {
			return err
		}
		if connected {
			return nil
		}
		wait(1 * time.Millisecond)
	}

	return ErrConnectionTimeout
}

func (drv *Driver) ConnectUDPSocket(addr, sport, lport string) error {
	return ErrNotImplemented
}

func (drv *Driver) DisconnectSocket() error {
	return drv.stop()
}

func (drv *Driver) StartSocketSend(size int) error {
	// not needed for WiFiNINA???
	return nil
}

func (drv *Driver) Response(timeout int) ([]byte, error) {
	return nil, nil
}

func (drv *Driver) Write(b []byte) (n int, err error) {
	if drv.sock == NoSocketAvail {
		return 0, ErrNoSocketAvail
	}
	if len(b) == 0 {
		return 0, ErrNoData
	}
	written, err := drv.dev.SendData(b, drv.sock)
	if err != nil {
		return 0, err
	}
	if written == 0 {
		return 0, ErrDataNotWritten
	}
	if sent, _ := drv.dev.CheckDataSent(drv.sock); !sent {
		return 0, ErrCheckDataError
	}
	return len(b), nil
}

func (drv *Driver) ReadSocket(b []byte) (n int, err error) {
	avail, err := drv.available()
	if err != nil {
		println("ReadSocket error: " + err.Error())
		return 0, err
	}
	if avail == 0 {
		return 0, nil
	}
	length := len(b)
	if avail < length {
		length = avail
	}
	copy(b, drv.readBuf.data[drv.readBuf.head:drv.readBuf.head+length])
	drv.readBuf.head += length
	drv.readBuf.size -= length
	return length, nil
}

// IsSocketDataAvailable returns of there is socket data available
func (drv *Driver) IsSocketDataAvailable() bool {
	n, err := drv.available()
	return err == nil && n > 0
}

func (drv *Driver) available() (int, error) {
	if drv.readBuf.size == 0 {
		n, err := drv.dev.GetDataBuf(drv.sock, drv.readBuf.data[:])
		if n > 0 {
			drv.readBuf.head = 0
			drv.readBuf.size = n
		}
		if err != nil {
			return int(n), err
		}
	}
	return drv.readBuf.size, nil
}

func (drv *Driver) IsConnected() (bool, error) {
	if drv.sock == NoSocketAvail {
		return false, nil
	}
	s, err := drv.status()
	if err != nil {
		return false, err
	}
	isConnected := !(s == TCPStateListen || s == TCPStateClosed ||
		s == TCPStateFinWait1 || s == TCPStateFinWait2 || s == TCPStateTimeWait ||
		s == TCPStateSynSent || s == TCPStateSynRcvd || s == TCPStateCloseWait)
	// TODO: investigate if the below is necessary (as per Arduino driver)
	//if !isConnected {
	//	//close socket buffer?
	//	WiFiSocketBuffer.close(_sock);
	//	_sock = 255;
	//}
	return isConnected, nil
}

func (drv *Driver) status() (uint8, error) {
	if drv.sock == NoSocketAvail {
		return TCPStateClosed, nil
	}
	return drv.dev.GetClientState(drv.sock)
}

func (drv *Driver) stop() error {
	if drv.sock == NoSocketAvail {
		return nil
	}
	drv.dev.StopClient(drv.sock)
	for now := time.Now(); time.Since(now) < 5*time.Second; {
		st, _ := drv.status()
		if st == TCPStateClosed {
			break
		}
		// FIXME: without the time.Sleep below this blocks until TCPStateClosed,
		// however with it got goroutine stack overflows; not sure if this is still
		// an issue so should investigate further
		//time.Sleep(1 * time.Millisecond)
	}
	drv.sock = NoSocketAvail
	return nil
}
