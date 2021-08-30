package wifinina

import (
	"errors"
	"strconv"
	"time"

	"tinygo.org/x/drivers/net"
)

const (
	ReadBufferSize = 128
)

// Configure resets wifinina device, creates network driver and tells network stack to use it.
// Wifi connection can either be handled esplicitly in user code or transparently by using WithAccessPoint.
func (d *Device) Configure() *Driver {
	d.Reset()
	drv := &Driver{dev: d, sock: NoSocketAvail}
	net.UseDriver(drv)
	return drv
}

// WithAccessPoint instructs the driver to establish and maintain access point connection transparently for the user.
// Multiple access points can be configured.
// In such case and when multiple access points available, active access point is selected in random.
// Each access point has an individual timeout after which connection attempt deemed failed and reconnect shall be tried.
func (drv *Driver) WithAccessPoint(ssid, password string, timeout time.Duration) *Driver {
	if drv.accessPoints == nil {
		drv.accessPoints = make([]accessPoint, 1)
	}
	drv.accessPoints = append(drv.accessPoints, accessPoint{ssid, password, timeout})
	return drv
}

// WithTimeout configures global timeout for transparent access point connection handling.
// Once this timeout elapses, no more connection and scan retries attempted and network operation returns with timeout error.
// When timeout not set, driver tries reconnect to access points forever.
func (drv *Driver) WithTimeout(timeout time.Duration) *Driver {
	drv.timeout = timeout
	return drv
}

type Driver struct {
	dev     *Device
	sock    uint8
	readBuf readBuffer

	proto uint8
	ip    uint32
	port  uint16

	accessPoints []accessPoint
	timeout      time.Duration
}

type readBuffer struct {
	data [ReadBufferSize]byte
	head int
	size int
}

type accessPoint struct {
	ssid     string
	password string
	timeout  time.Duration
}

// HandleWifi ensures device is connected to an access point (if any) and is called before any network operation
func (drv *Driver) HandleWifi() error {

	// return fast when no access points configured
	if drv.accessPoints == nil {
		return nil
	}

	var activeAccessPoint *accessPoint
	var reconnectTime time.Time
	var stopTime time.Time

	if drv.timeout != 0 {
		stopTime = time.Now().Add(drv.timeout)
	}

	for {

		// global timeout elapsed, throw
		if !stopTime.IsZero() && time.Now().After(stopTime) {
			return errors.New("global wifi timeout elapsed")
		}

		st, err := drv.dev.GetConnectionStatus()
		// problems communicating with the device, reset device
		if err != nil {
			drv.dev.Reset()
			time.Sleep(time.Second)
			continue
		}

		// already connected, return
		if st == StatusConnected {
			return nil
		}

		// not connected and connection timeout not expired yet, just wait
		if st == StatusNoSSIDAvail && activeAccessPoint != nil && time.Now().Before(reconnectTime) {
			time.Sleep(time.Second)
			continue
		}

		// search for any known access point and if not found, just loop around
		activeAccessPoint = drv.searchAccessPoint()
		if activeAccessPoint == nil {
			continue
		}

		// try connect to the access point
		reconnectTime = time.Now().Add(activeAccessPoint.timeout)
		err = drv.dev.SetPassphrase(activeAccessPoint.ssid, activeAccessPoint.password)
		// problems communicating with the device, reset device
		if err != nil {
			drv.dev.Reset()
			time.Sleep(time.Second)
		}

	}

}

// TODO: check RSSI and select best AP?
func (drv *Driver) searchAccessPoint() *accessPoint {
	count, err := drv.dev.ScanNetworks()
	// problems communicating with the device, reset device
	if err != nil {
		drv.dev.Reset()
		time.Sleep(time.Second)
		return nil
	}
	for i := 0; i < int(count); i++ {
		ssid := drv.dev.GetNetworkSSID(i)
		for _, ap := range drv.accessPoints {
			if ap.ssid == ssid {
				return &ap
			}
		}
	}
	return nil
}

func (drv *Driver) GetDNS(domain string) (string, error) {
	if err := drv.HandleWifi(); err != nil {
		return "", errors.New("could not connect to AP: " + err.Error())
	}
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

	if err := drv.HandleWifi(); err != nil {
		return errors.New("could not connect to WiFi: " + err.Error())
	}

	drv.proto, drv.ip, drv.port = mode, 0, 0

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
		ipAddr, err := drv.dev.GetHostByName(addr)
		if err != nil {
			return err
		}
		hostname = ""
		ip = ipAddr.AsUint32()
	}

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
	if err := drv.dev.StartClient(hostname, ip, port, drv.sock, mode); err != nil {
		return err
	}

	// FIXME: this 4 second timeout is simply mimicking the Arduino driver
	for t := newTimer(4 * time.Second); !t.Expired(); {
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

func convertPort(portStr string) (uint16, error) {
	p64, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return 0, errors.New("could not convert port to uint16: " + err.Error())
	}
	return uint16(p64), nil
}

func (drv *Driver) ConnectUDPSocket(addr, portStr, lportStr string) (err error) {

	if err := drv.HandleWifi(); err != nil {
		return errors.New("could not connect to WiFi: " + err.Error())
	}

	drv.proto, drv.ip, drv.port = ProtoModeUDP, 0, 0

	// convert remote port to uint16
	if drv.port, err = convertPort(portStr); err != nil {
		return err
	}

	// convert local port to uint16
	var lport uint16
	if lport, err = convertPort(lportStr); err != nil {
		return err
	}

	// look up the hostname if necessary; if an IP address was specified, the
	// same will be returned.  Otherwise, an IPv4 for the hostname is returned.
	ipAddr, err := drv.dev.GetHostByName(addr)
	if err != nil {
		return err
	}
	drv.ip = ipAddr.AsUint32()

	// check to see if socket is already set; if so, stop it
	// TODO: we can probably have more than one socket at once right?
	if drv.sock != NoSocketAvail {
		if err := drv.stop(); err != nil {
			return err
		}
	}

	// get a socket from the device
	if drv.sock, err = drv.dev.GetSocket(); err != nil {
		return err
	}

	// start listening for UDP packets on the local port
	if err := drv.dev.StartServer(lport, drv.sock, drv.proto); err != nil {
		return err
	}

	return nil
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
	if err := drv.HandleWifi(); err != nil {
		return 0, errors.New("could not connect to WiFi: " + err.Error())
	}
	if drv.sock == NoSocketAvail {
		return 0, ErrNoSocketAvail
	}
	if len(b) == 0 {
		return 0, ErrNoData
	}
	if drv.proto == ProtoModeUDP {
		if err := drv.dev.StartClient("", drv.ip, drv.port, drv.sock, drv.proto); err != nil {
			return 0, errors.New("error in startClient: " + err.Error())
		}
		if _, err := drv.dev.InsertDataBuf(b, drv.sock); err != nil {
			return 0, errors.New("error in insertDataBuf: " + err.Error())
		}
		if _, err := drv.dev.SendUDPData(drv.sock); err != nil {
			return 0, errors.New("error in sendUDPData: " + err.Error())
		}
		return len(b), nil
	} else {
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
		return int(written), nil
	}
}

func (drv *Driver) ReadSocket(b []byte) (n int, err error) {
	if err := drv.HandleWifi(); err != nil {
		return 0, errors.New("could not connect to WiFi: " + err.Error())
	}
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
	if err := drv.HandleWifi(); err != nil {
		return false
	}
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
	for t := newTimer(5 * time.Second); !t.Expired(); {
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
