package wifinina

import (
	"errors"
	"strconv"
	"time"
)

const (
	ReadBufferSize = 128
)

type readBuffer struct {
	data [ReadBufferSize]byte
	head int
	size int
}

func (d *Device) GetDNS(domain string) (string, error) {
	ipAddr, err := d.GetHostByName(domain)
	return ipAddr.String(), err
}

func (d *Device) ConnectTCPSocket(addr, portStr string) error {
	return d.connectSocket(addr, portStr, ProtoModeTCP)
}

func (d *Device) ConnectSSLSocket(addr, portStr string) error {
	return d.connectSocket(addr, portStr, ProtoModeTLS)
}

func (d *Device) connectSocket(addr, portStr string, mode uint8) error {

	d.proto, d.ip, d.port = mode, 0, 0

	// convert port to uint16
	port, err := convertPort(portStr)
	if err != nil {
		return err
	}

	hostname := addr
	ip := uint32(0)

	if mode != ProtoModeTLS {
		// look up the hostname if necessary; if an IP address was specified, the
		// same will be returned.  Otherwise, an IPv4 for the hostname is returned.
		ipAddr, err := d.GetHostByName(addr)
		if err != nil {
			return err
		}
		hostname = ""
		ip = ipAddr.AsUint32()
	}

	// check to see if socket is already set; if so, stop it
	if d.sock != NoSocketAvail {
		if err := d.stop(); err != nil {
			return err
		}
	}

	// get a socket from the device
	if d.sock, err = d.GetSocket(); err != nil {
		return err
	}

	// attempt to start the client
	if err := d.StartClient(hostname, ip, port, d.sock, mode); err != nil {
		return err
	}

	// FIXME: this 4 second timeout is simply mimicking the Arduino driver
	start := time.Now()
	for time.Since(start) < 4*time.Second {
		connected, err := d.IsConnected()
		if err != nil {
			return err
		}
		if connected {
			return nil
		}
		time.Sleep(1 * time.Millisecond)
	}

	return ErrConnectionTimeout
}

func convertPort(portStr string) (uint16, error) {
	p64, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return 0, errors.New("could not convert port to uint16: " + err.Error())
	}
	return uint16(p64), nil
}

func (d *Device) ConnectUDPSocket(addr, portStr, lportStr string) (err error) {

	d.proto, d.ip, d.port = ProtoModeUDP, 0, 0

	// convert remote port to uint16
	if d.port, err = convertPort(portStr); err != nil {
		return err
	}

	// convert local port to uint16
	var lport uint16
	if lport, err = convertPort(lportStr); err != nil {
		return err
	}

	// look up the hostname if necessary; if an IP address was specified, the
	// same will be returned.  Otherwise, an IPv4 for the hostname is returned.
	ipAddr, err := d.GetHostByName(addr)
	if err != nil {
		return err
	}
	d.ip = ipAddr.AsUint32()

	// check to see if socket is already set; if so, stop it
	// TODO: we can probably have more than one socket at once right?
	if d.sock != NoSocketAvail {
		if err := d.stop(); err != nil {
			return err
		}
	}

	// get a socket from the device
	if d.sock, err = d.GetSocket(); err != nil {
		return err
	}

	// start listening for UDP packets on the local port
	if err := d.StartServer(lport, d.sock, d.proto); err != nil {
		return err
	}

	return nil
}

func (d *Device) DisconnectSocket() error {
	return d.stop()
}

func (d *Device) StartSocketSend(size int) error {
	// not needed for WiFiNINA???
	return nil
}

func (d *Device) Response(timeout int) ([]byte, error) {
	return nil, nil
}

func (d *Device) Write(b []byte) (n int, err error) {
	if d.sock == NoSocketAvail {
		return 0, ErrNoSocketAvail
	}
	if len(b) == 0 {
		return 0, ErrNoData
	}
	if d.proto == ProtoModeUDP {
		if err := d.StartClient("", d.ip, d.port, d.sock, d.proto); err != nil {
			return 0, errors.New("error in startClient: " + err.Error())
		}
		if _, err := d.InsertDataBuf(b, d.sock); err != nil {
			return 0, errors.New("error in insertDataBuf: " + err.Error())
		}
		if _, err := d.SendUDPData(d.sock); err != nil {
			return 0, errors.New("error in sendUDPData: " + err.Error())
		}
		return len(b), nil
	} else {
		written, err := d.SendData(b, d.sock)
		if err != nil {
			return 0, err
		}
		if written == 0 {
			return 0, ErrDataNotWritten
		}
		if sent, _ := d.CheckDataSent(d.sock); !sent {
			return 0, ErrCheckDataError
		}
		return len(b), nil
	}

	return len(b), nil
}

func (d *Device) ReadSocket(b []byte) (n int, err error) {
	avail, err := d.available()
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
	copy(b, d.readBuf.data[d.readBuf.head:d.readBuf.head+length])
	d.readBuf.head += length
	d.readBuf.size -= length
	return length, nil
}

// IsSocketDataAvailable returns of there is socket data available
func (d *Device) IsSocketDataAvailable() bool {
	n, err := d.available()
	return err == nil && n > 0
}

func (d *Device) available() (int, error) {
	if d.readBuf.size == 0 {
		n, err := d.GetDataBuf(d.sock, d.readBuf.data[:])
		if n > 0 {
			d.readBuf.head = 0
			d.readBuf.size = n
		}
		if err != nil {
			return int(n), err
		}
	}
	return d.readBuf.size, nil
}

func (d *Device) IsConnected() (bool, error) {
	if d.sock == NoSocketAvail {
		return false, nil
	}
	s, err := d.status()
	if err != nil {
		return false, err
	}
	isConnected := !(s == uint8(TCPStateListen) || s == uint8(TCPStateClosed) ||
		s == uint8(TCPStateFinWait1) || s == uint8(TCPStateFinWait2) || s == uint8(TCPStateTimeWait) ||
		s == uint8(TCPStateSynSent) || s == uint8(TCPStateSynRcvd) || s == uint8(TCPStateCloseWait))
	// TODO: investigate if the below is necessary (as per Arduino driver)
	//if !isConnected {
	//	//close socket buffer?
	//	WiFiSocketBuffer.close(_sock);
	//	_sock = 255;
	//}
	return isConnected, nil
}

func (d *Device) status() (uint8, error) {
	if d.sock == NoSocketAvail {
		return uint8(TCPStateClosed), nil
	}
	return d.GetClientState(d.sock)
}

func (d *Device) stop() error {
	if d.sock == NoSocketAvail {
		return nil
	}
	d.StopClient(d.sock)
	start := time.Now()
	for time.Since(start) < 5*time.Second {
		st, _ := d.status()
		if st == uint8(TCPStateClosed) {
			break
		}
		time.Sleep(1 * time.Millisecond)
	}
	d.sock = NoSocketAvail
	return nil
}
