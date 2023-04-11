// Package wifinina implements TCP wireless communication over SPI with an
// attached separate ESP32 SoC using the Arduino WiFiNINA protocol.
//
// In order to use this driver, the ESP32 must be flashed with specific
// firmware from Arduino.  For more information:
// https://github.com/arduino/nina-fw
//
// 12/2022    sfeldma@gmail.com    Heavily modified to use netdev interface

package wifinina // import "tinygo.org/x/drivers/wifinina"

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"machine"
	"math/bits"
	"net"
	"sync"
	"syscall"
	"time"

	"tinygo.org/x/drivers"
)

var _debug debug = debugBasic

//var _debug debug = debugBasic | debugNetdev
//var _debug debug = debugBasic | debugNetdev | debugCmd
//var _debug debug = debugBasic | debugNetdev | debugCmd | debugDetail

var (
	version    = "0.0.1"
	driverName = "Tinygo ESP32 Wifi network device driver (WiFiNINA)"
)

const (
	maxNetworks = 10

	statusNoShield       connectionStatus = 255
	statusIdle           connectionStatus = 0
	statusNoSSIDAvail    connectionStatus = 1
	statusScanCompleted  connectionStatus = 2
	statusConnected      connectionStatus = 3
	statusConnectFailed  connectionStatus = 4
	statusConnectionLost connectionStatus = 5
	statusDisconnected   connectionStatus = 6

	encTypeTKIP encryptionType = 2
	encTypeCCMP encryptionType = 4
	encTypeWEP  encryptionType = 5
	encTypeNone encryptionType = 7
	encTypeAuto encryptionType = 8

	tcpStateClosed      = 0
	tcpStateListen      = 1
	tcpStateSynSent     = 2
	tcpStateSynRcvd     = 3
	tcpStateEstablished = 4
	tcpStateFinWait1    = 5
	tcpStateFinWait2    = 6
	tcpStateCloseWait   = 7
	tcpStateClosing     = 8
	tcpStateLastACK     = 9
	tcpStateTimeWait    = 10

	flagCmd   = 0
	flagReply = 1 << 7
	flagData  = 0x40

	cmdStart = 0xE0
	cmdEnd   = 0xEE
	cmdErr   = 0xEF

	dummyData = 0xFF

	cmdSetNet            = 0x10
	cmdSetPassphrase     = 0x11
	cmdSetKey            = 0x12
	cmdSetIPConfig       = 0x14
	cmdSetDNSConfig      = 0x15
	cmdSetHostname       = 0x16
	cmdSetPowerMode      = 0x17
	cmdSetAPNet          = 0x18
	cmdSetAPPassphrase   = 0x19
	cmdSetDebug          = 0x1A
	cmdGetTemperature    = 0x1B
	cmdGetReasonCode     = 0x1F
	cmdGetConnStatus     = 0x20
	cmdGetIPAddr         = 0x21
	cmdGetMACAddr        = 0x22
	cmdGetCurrSSID       = 0x23
	cmdGetCurrBSSID      = 0x24
	cmdGetCurrRSSI       = 0x25
	cmdGetCurrEncrType   = 0x26
	cmdScanNetworks      = 0x27
	cmdStartServerTCP    = 0x28
	cmdGetStateTCP       = 0x29
	cmdDataSentTCP       = 0x2A
	cmdAvailDataTCP      = 0x2B
	cmdGetDataTCP        = 0x2C
	cmdStartClientTCP    = 0x2D
	cmdStopClientTCP     = 0x2E
	cmdGetClientStateTCP = 0x2F
	cmdDisconnect        = 0x30
	cmdGetIdxRSSI        = 0x32
	cmdGetIdxEncrType    = 0x33
	cmdReqHostByName     = 0x34
	cmdGetHostByName     = 0x35
	cmdStartScanNetworks = 0x36
	cmdGetFwVersion      = 0x37
	cmdSendDataUDP       = 0x39
	cmdGetRemoteData     = 0x3A
	cmdGetTime           = 0x3B
	cmdGetIdxBSSID       = 0x3C
	cmdGetIdxChannel     = 0x3D
	cmdPing              = 0x3E
	cmdGetSocket         = 0x3F

	// All commands with DATA_FLAG 0x4x send a 16bit Len
	cmdSendDataTCP   = 0x44
	cmdGetDatabufTCP = 0x45
	cmdInsertDataBuf = 0x46

	// Regular format commands
	cmdSetPinMode      = 0x50
	cmdSetDigitalWrite = 0x51
	cmdSetAnalogWrite  = 0x52

	errTimeoutChipReady  hwerr = 0x01
	errTimeoutChipSelect hwerr = 0x02
	errCheckStartCmd     hwerr = 0x03
	errWaitRsp           hwerr = 0x04
	errUnexpectedLength  hwerr = 0xE0
	errNoParamsReturned  hwerr = 0xE1
	errIncorrectSentinel hwerr = 0xE2
	errCmdErrorReceived  hwerr = 0xEF
	errNotImplemented    hwerr = 0xF0
	errUnknownHost       hwerr = 0xF1
	errSocketAlreadySet  hwerr = 0xF2
	errConnectionTimeout hwerr = 0xF3
	errNoData            hwerr = 0xF4
	errDataNotWritten    hwerr = 0xF5
	errCheckDataError    hwerr = 0xF6
	errBufferTooSmall    hwerr = 0xF7
	errNoSocketAvail     hwerr = 0xFF

	noSocketAvail sock = 0xFF
)

const (
	protoModeTCP = iota
	protoModeUDP
	protoModeTLS
	protoModeMul
)

type connectionStatus uint8
type encryptionType uint8
type sock uint8
type hwerr uint8

type socket struct {
	protocol int
	ip       net.IP
	port     int
	inuse    bool
}

type Config struct {
	// AP creditials
	Ssid       string
	Passphrase string

	// SPI config
	Spi  drivers.SPI
	Freq uint32
	Sdo  machine.Pin
	Sdi  machine.Pin
	Sck  machine.Pin

	// Device config
	Cs     machine.Pin
	Ack    machine.Pin
	Gpio0  machine.Pin
	Resetn machine.Pin

	// Retries is how many attempts to connect before returning with a
	// "Connect failed" error.  Zero means infinite retries.
	Retries int

	// Timeout duration for each connection attempt.  Default is 10sec.
	ConnectTimeo time.Duration

	// Watchdog ticker duration.  On tick, the watchdog will check for
	// downed connection or hardware fault and try to recover the
	// connection.  Default is 0secs, which means no watchdog.  Set to
	// non-zero to enable watchodog.
	WatchdogTimeo time.Duration
}

type wifinina struct {
	cfg      *Config
	notifyCb func(drivers.NetlinkEvent)
	mu       sync.Mutex

	spi    drivers.SPI
	cs     machine.Pin
	ack    machine.Pin
	gpio0  machine.Pin
	resetn machine.Pin

	buf   [64]byte
	ssids [maxNetworks]string

	netConnected bool
	driverShown  bool
	deviceShown  bool
	spiSetup     bool

	killWatchdog chan bool
	fault        error

	sockets map[sock]*socket // keyed by sock as returned by getSocket()
}

func newSocket(protocol int) *socket {
	return &socket{protocol: protocol, inuse: true}
}

func New(cfg *Config) *wifinina {
	w := wifinina{
		cfg:          cfg,
		sockets:      make(map[sock]*socket),
		killWatchdog: make(chan bool),
		cs:           cfg.Cs,
		ack:          cfg.Ack,
		gpio0:        cfg.Gpio0,
		resetn:       cfg.Resetn,
	}

	if w.cfg.ConnectTimeo == 0 {
		w.cfg.ConnectTimeo = 10 * time.Second
	}

	drivers.UseNetdev(&w)

	// assert that wifinina implements Netlinker
	var _ drivers.Netlinker = (*wifinina)(nil)

	return &w
}

func (err hwerr) Error() string {
	return "[wifinina] error: 0x" + hex.EncodeToString([]byte{uint8(err)})
}

func (w *wifinina) connectToAP(timeout time.Duration) error {

	if len(w.cfg.Ssid) == 0 {
		return drivers.ErrMissingSSID
	}

	if debugging(debugBasic) {
		fmt.Printf("Connecting to Wifi SSID '%s'...", w.cfg.Ssid)
	}

	start := time.Now()

	// Start the connection process
	w.setPassphrase(w.cfg.Ssid, w.cfg.Passphrase)

	// Check if we connected
	for time.Since(start) < timeout {
		status := w.getConnectionStatus()
		if status == statusConnected {
			if debugging(debugBasic) {
				fmt.Printf("CONNECTED\r\n")
			}
			if w.notifyCb != nil {
				w.notifyCb(drivers.NetlinkEventNetUp)
			}
			return nil
		}
		time.Sleep(1 * time.Second)
	}

	if debugging(debugBasic) {
		fmt.Printf("FAILED\r\n")
	}

	return drivers.ErrConnectTimeout
}

func (w *wifinina) netDisconnect() {
	w.disconnect()
}

func (w *wifinina) showDriver() {
	if w.driverShown {
		return
	}
	if debugging(debugBasic) {
		fmt.Printf("\r\n")
		fmt.Printf("%s\r\n\r\n", driverName)
		fmt.Printf("Driver version           : %s\r\n", version)
	}
	w.driverShown = true
}

func (w *wifinina) setupSPI() {
	if w.spiSetup {
		return
	}
	spi := machine.NINA_SPI
	spi.Configure(machine.SPIConfig{
		Frequency: w.cfg.Freq,
		SDO:       w.cfg.Sdo,
		SDI:       w.cfg.Sdi,
		SCK:       w.cfg.Sck,
	})
	w.spi = spi
	w.spiSetup = true
}

func (w *wifinina) start() {

	pinUseDevice(w)

	w.cs.Configure(machine.PinConfig{Mode: machine.PinOutput})
	w.ack.Configure(machine.PinConfig{Mode: machine.PinInput})
	w.resetn.Configure(machine.PinConfig{Mode: machine.PinOutput})
	w.gpio0.Configure(machine.PinConfig{Mode: machine.PinOutput})

	w.gpio0.High()
	w.cs.High()
	w.resetn.Low()
	time.Sleep(10 * time.Millisecond)
	w.resetn.High()
	time.Sleep(750 * time.Millisecond)

	w.gpio0.Low()
	w.gpio0.Configure(machine.PinConfig{Mode: machine.PinInput})
}

func (w *wifinina) stop() {
	w.resetn.Low()
	w.cs.Configure(machine.PinConfig{Mode: machine.PinInput})
}

func (w *wifinina) showDevice() {
	if w.deviceShown {
		return
	}
	if debugging(debugBasic) {
		fmt.Printf("ESP32 firmware version   : %s\r\n", w.getFwVersion())
		mac := w.getMACAddr()
		fmt.Printf("MAC address              : %s\r\n", mac.String())
		fmt.Printf("\r\n")
	}
	w.deviceShown = true
}

func (w *wifinina) showIP() {
	if debugging(debugBasic) {
		ip, subnet, gateway := w.getIP()
		fmt.Printf("\r\n")
		fmt.Printf("DHCP-assigned IP         : %s\r\n", ip.String())
		fmt.Printf("DHCP-assigned subnet     : %s\r\n", subnet.String())
		fmt.Printf("DHCP-assigned gateway    : %s\r\n", gateway.String())
		fmt.Printf("\r\n")
	}
}

func (w *wifinina) networkDown() bool {
	return w.getConnectionStatus() != statusConnected
}

func (w *wifinina) watchdog() {
	ticker := time.NewTicker(w.cfg.WatchdogTimeo)
	for {
		select {
		case <-w.killWatchdog:
			return
		case <-ticker.C:
			w.mu.Lock()
			if w.fault != nil {
				if debugging(debugBasic) {
					fmt.Printf("Watchdog: FAULT: %s\r\n", w.fault)
				}
				w.netDisconnect()
				w.netConnect(true)
				w.fault = nil
			} else if w.networkDown() {
				if debugging(debugBasic) {
					fmt.Printf("Watchdog: Wifi NOT CONNECTED, trying again...\r\n")
				}
				if w.notifyCb != nil {
					w.notifyCb(drivers.NetlinkEventNetDown)
				}
				w.netConnect(false)
			}
			w.mu.Unlock()
		}
	}
}

func (w *wifinina) netConnect(reset bool) error {
	if reset {
		w.start()
	}
	w.showDevice()

	for i := 0; w.cfg.Retries == 0 || i < w.cfg.Retries; i++ {
		if err := w.connectToAP(w.cfg.ConnectTimeo); err != nil {
			if err == drivers.ErrConnectTimeout {
				continue
			}
			return err
		}
		break
	}

	if w.networkDown() {
		return drivers.ErrConnectFailed
	}

	w.showIP()
	return nil
}

func (w *wifinina) NetConnect() error {

	w.mu.Lock()
	defer w.mu.Unlock()

	if w.netConnected {
		return drivers.ErrConnected
	}

	w.showDriver()
	w.setupSPI()

	if err := w.netConnect(true); err != nil {
		return err
	}

	w.netConnected = true

	if w.cfg.WatchdogTimeo != 0 {
		go w.watchdog()
	}

	return nil
}

func (w *wifinina) NetDisconnect() {

	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.netConnected {
		return
	}

	if w.cfg.WatchdogTimeo != 0 {
		w.killWatchdog <- true
	}

	w.netDisconnect()
	w.stop()

	w.netConnected = false

	if debugging(debugBasic) {
		fmt.Printf("\r\nDisconnected from Wifi SSID '%s'\r\n\r\n", w.cfg.Ssid)
	}

	if w.notifyCb != nil {
		w.notifyCb(drivers.NetlinkEventNetDown)
	}
}

func (w *wifinina) NetNotify(cb func(drivers.NetlinkEvent)) {
	w.notifyCb = cb
}

func (w *wifinina) GetHostByName(name string) (net.IP, error) {

	if debugging(debugNetdev) {
		fmt.Printf("[GetHostByName] name: %s\r\n", name)
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	ip := w.getHostByName(name)
	if ip == "" {
		return net.IP{}, drivers.ErrHostUnknown
	}

	return net.IP([]byte(ip)), nil
}

func (w *wifinina) GetHardwareAddr() (net.HardwareAddr, error) {

	if debugging(debugNetdev) {
		fmt.Printf("[GetHardwareAddr]\r\n")
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	return w.getMACAddr(), nil
}

func (w *wifinina) GetIPAddr() (net.IP, error) {

	if debugging(debugNetdev) {
		fmt.Printf("[GetIPAddr]\r\n")
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	ip, _, _ := w.getIP()

	return net.IP(ip), nil
}

// See man socket(2) for standard Berkely sockets for Socket, Bind, etc.
// The driver strives to meet the function and semantics of socket(2).

func (w *wifinina) Socket(domain int, stype int, protocol int) (int, error) {

	if debugging(debugNetdev) {
		fmt.Printf("[Socket] domain: %d, type: %d, protocol: %d\r\n",
			domain, stype, protocol)
	}

	switch domain {
	case syscall.AF_INET:
	default:
		return -1, drivers.ErrFamilyNotSupported
	}

	switch {
	case protocol == syscall.IPPROTO_TCP && stype == syscall.SOCK_STREAM:
	case protocol == drivers.IPPROTO_TLS && stype == syscall.SOCK_STREAM:
	case protocol == syscall.IPPROTO_UDP && stype == syscall.SOCK_DGRAM:
	default:
		return -1, drivers.ErrProtocolNotSupported
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	sock := w.getSocket()
	if sock == noSocketAvail {
		return -1, drivers.ErrNoMoreSockets
	}

	socket := newSocket(protocol)
	w.sockets[sock] = socket

	return int(sock), nil
}

func (w *wifinina) Bind(sockfd int, ip net.IP, port int) error {

	if debugging(debugNetdev) {
		fmt.Printf("[Bind] sockfd: %d, addr: %s:%d\r\n", sockfd, ip, port)
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	var sock = sock(sockfd)
	var socket = w.sockets[sock]

	switch socket.protocol {
	case syscall.IPPROTO_TCP:
	case drivers.IPPROTO_TLS:
	case syscall.IPPROTO_UDP:
		w.startServer(sock, uint16(port), protoModeUDP)
	}

	socket.ip, socket.port = ip, port

	return nil
}

func toUint32(ip net.IP) uint32 {
	ip = ip.To4()
	return uint32(ip[0])<<24 |
		uint32(ip[1])<<16 |
		uint32(ip[2])<<8 |
		uint32(ip[3])
}

func (w *wifinina) Connect(sockfd int, host string, ip net.IP, port int) error {

	if debugging(debugNetdev) {
		if host == "" {
			fmt.Printf("[Connect] sockfd: %d, addr: %s:%d\r\n", sockfd, ip, port)
		} else {
			fmt.Printf("[Connect] sockfd: %d, host: %s:%d\r\n", sockfd, host, port)
		}
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	var sock = sock(sockfd)
	var socket = w.sockets[sock]

	// Start the connection
	switch socket.protocol {
	case syscall.IPPROTO_TCP:
		w.startClient(sock, "", toUint32(ip), uint16(port), protoModeTCP)
	case drivers.IPPROTO_TLS:
		w.startClient(sock, host, 0, uint16(port), protoModeTLS)
	case syscall.IPPROTO_UDP:
		w.startClient(sock, "", toUint32(ip), uint16(port), protoModeUDP)
		return nil
	}

	// Wait for up to 10s to connect...
	expire := time.Now().Add(10 * time.Second)

	for time.Now().Before(expire) {
		if w.isConnected(sock) {
			return nil
		}

		// Check if we've faulted
		if w.fault != nil {
			w.stopClient(sock)
			return w.fault
		}

		w.mu.Unlock()
		time.Sleep(100 * time.Millisecond)
		w.mu.Lock()
	}

	w.stopClient(sock)

	if host == "" {
		return fmt.Errorf("Connect to %s:%d timed out", ip, port)
	} else {
		return fmt.Errorf("Connect to %s:%d timed out", host, port)
	}
}

func (w *wifinina) Listen(sockfd int, backlog int) error {

	if debugging(debugNetdev) {
		fmt.Printf("[Listen] sockfd: %d\r\n", sockfd)
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	var sock = sock(sockfd)
	var socket = w.sockets[sock]

	switch socket.protocol {
	case syscall.IPPROTO_TCP:
		w.startServer(sock, uint16(socket.port), protoModeTCP)
	case syscall.IPPROTO_UDP:
	default:
		return drivers.ErrProtocolNotSupported
	}

	return nil
}

func (w *wifinina) Accept(sockfd int, ip net.IP, port int) (int, error) {

	if debugging(debugNetdev) {
		fmt.Printf("[Accept] sockfd: %d, peer: %s:%d\r\n", sockfd, ip, port)
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	var client sock
	var sock = sock(sockfd)
	var socket = w.sockets[sock]

	switch socket.protocol {
	case syscall.IPPROTO_TCP:
	default:
		return -1, drivers.ErrProtocolNotSupported
	}

	for {
		// Accept() will be sleeping most of the time, checking for a
		// new clients every 1/10 sec.
		w.mu.Unlock()
		time.Sleep(100 * time.Millisecond)
		w.mu.Lock()

		// Check if we've faulted
		if w.fault != nil {
			return -1, w.fault
		}

		// TODO: BUG: Currently, a sock that is 100% busy will always be
		// TODO: returned by w.accept(sock), starving other socks
		// TODO: from begin serviced.  Need to figure out how to
		// TODO: service socks fairly (round-robin?) so no one sock
		// TODO: can dominate.

		// Check if a client has data
		client = w.accept(sock)
		if client == noSocketAvail {
			// None ready
			continue
		}

		// If we've already seen this socket, we can reuse
		// the socket and return it.  But, only if the socket
		// is closed.  If it's not closed, we'll just come back
		// later to reuse it.

		clientSocket, ok := w.sockets[client]
		if ok {
			// Wait for client to Close
			if clientSocket.inuse {
				continue
			}
			// Reuse client socket
			return int(client), nil
		}

		// Create new socket for client and return fd
		w.sockets[client] = newSocket(socket.protocol)
		return int(client), nil
	}
}

func (w *wifinina) sockDown(sock sock) bool {
	state := w.getClientState(sock)
	if state == tcpStateEstablished {
		return false
	}
	return true
}

func (w *wifinina) sendTCP(sock sock, buf []byte, timeout time.Duration) (int, error) {

	var timeoutDataSent = 25
	var expire = time.Now().Add(timeout)

	// Send it
	n := int(w.sendData(sock, buf))
	if n == 0 {
		return -1, io.EOF
	}

	// Check if data was sent
	for i := 0; i < timeoutDataSent; i++ {
		sent := w.checkDataSent(sock)
		if sent {
			return n, nil
		}

		// Check if we've timed out
		if timeout > 0 {
			if time.Now().Before(expire) {
				return -1, fmt.Errorf("Send timeout expired")
			}
		}

		// Check if socket went down
		if w.sockDown(sock) {
			return -1, io.EOF
		}

		// Check if we've faulted
		if w.fault != nil {
			return -1, w.fault
		}

		// Unlock while we sleep, so others can make progress
		w.mu.Unlock()
		time.Sleep(100 * time.Millisecond)
		w.mu.Lock()
	}

	return -1, fmt.Errorf("Send timed out")
}

func (w *wifinina) sendUDP(sock sock, buf []byte, timeout time.Duration) (int, error) {

	// Queue it
	ok := w.insertDataBuf(sock, buf)
	if !ok {
		return -1, fmt.Errorf("Insert UDP data failed, len(buf)=%d", len(buf))
	}

	// Send it
	ok = w.sendUDPData(sock)
	if !ok {
		return -1, fmt.Errorf("Send UDP data failed, len(buf)=%d", len(buf))
	}

	return len(buf), nil
}

func (w *wifinina) sendChunk(sockfd int, buf []byte, timeout time.Duration) (int, error) {
	var sock = sock(sockfd)
	var socket = w.sockets[sock]

	switch socket.protocol {
	case syscall.IPPROTO_TCP, drivers.IPPROTO_TLS:
		return w.sendTCP(sock, buf, timeout)
	case syscall.IPPROTO_UDP:
		return w.sendUDP(sock, buf, timeout)
	}

	return -1, drivers.ErrProtocolNotSupported
}

func (w *wifinina) Send(sockfd int, buf []byte, flags int,
	timeout time.Duration) (int, error) {

	if debugging(debugNetdev) {
		fmt.Printf("[Send] sockfd: %d, len(buf): %d, flags: %d\r\n",
			sockfd, len(buf), flags)
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	// Break large bufs into chunks so we don't overrun the hw queue

	chunkSize := 1436
	for i := 0; i < len(buf); i += chunkSize {
		end := i + chunkSize
		if end > len(buf) {
			end = len(buf)
		}
		_, err := w.sendChunk(sockfd, buf[i:end], timeout)
		if err != nil {
			return -1, err
		}
	}

	return len(buf), nil
}

func (w *wifinina) Recv(sockfd int, buf []byte, flags int,
	timeout time.Duration) (int, error) {

	if debugging(debugNetdev) {
		fmt.Printf("[Recv] sockfd: %d, len(buf): %d, flags: %d\r\n",
			sockfd, len(buf), flags)
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	var sock = sock(sockfd)
	var socket = w.sockets[sock]
	var max = len(buf)

	// Limit max read size to chunk large read requests
	if max > 1436 {
		max = 1436
	}

	expire := time.Now().Add(timeout)

	for {
		// Check if we've timed out
		if timeout > 0 {
			if time.Now().Before(expire) {
				return -1, drivers.ErrRecvTimeout
			}
		}

		// Receive into buf, if any data available.  It's ok if no data
		// is available, we'll just sleep a bit and recheck.  Recv()
		// doesn't return unless there is data, even a single byte, or
		// on error such as timeout or EOF.

		n := int(w.getDataBuf(sock, buf[:max]))
		if n > 0 {
			if debugging(debugNetdev) {
				fmt.Printf("[<--Recv] sockfd: %d, n: %d\r\n",
					sock, n)
			}
			return n, nil
		}

		// Check if socket went down
		if socket.protocol != syscall.IPPROTO_UDP && w.sockDown(sock) {
			// Get any last bytes
			n = int(w.getDataBuf(sock, buf[:max]))
			if debugging(debugNetdev) {
				fmt.Printf("[<--Recv] sockfd: %d, n: %d, EOF\r\n",
					sock, n)
			}
			if n > 0 {
				return n, io.EOF
			}
			return -1, io.EOF
		}

		// Check if we've faulted
		if w.fault != nil {
			return -1, w.fault
		}

		// Unlock while we sleep, so others can make progress
		w.mu.Unlock()
		time.Sleep(100 * time.Millisecond)
		w.mu.Lock()
	}
}

func (w *wifinina) Close(sockfd int) error {

	if debugging(debugNetdev) {
		fmt.Printf("[Close] sockfd: %d\r\n", sockfd)
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	var sock = sock(sockfd)
	var socket = w.sockets[sock]

	if !socket.inuse {
		return nil
	}

	w.stopClient(sock)

	if socket.protocol == syscall.IPPROTO_UDP {
		socket.inuse = false
		return nil
	}

	start := time.Now()
	for time.Since(start) < 5*time.Second {

		state := w.getClientState(sock)
		if state == tcpStateClosed {
			socket.inuse = false
			return nil
		}

		w.mu.Unlock()
		time.Sleep(100 * time.Millisecond)
		w.mu.Lock()
	}

	return drivers.ErrClosingSocket
}

func (w *wifinina) SetSockOpt(sockfd int, level int, opt int, value interface{}) error {

	if debugging(debugNetdev) {
		fmt.Printf("[SetSockOpt] sockfd: %d\r\n", sockfd)
	}

	return drivers.ErrNotSupported
}

// Is TCP/TLS socket connected?
func (w *wifinina) isConnected(sock sock) bool {
	s := w.getClientState(sock)

	connected := !(s == tcpStateListen || s == tcpStateClosed ||
		s == tcpStateFinWait1 || s == tcpStateFinWait2 || s == tcpStateTimeWait ||
		s == tcpStateSynSent || s == tcpStateSynRcvd || s == tcpStateCloseWait)

	return connected
}

func (w *wifinina) startClient(sock sock, hostname string, addr uint32, port uint16, mode uint8) {
	if debugging(debugCmd) {
		fmt.Printf("    [cmdStartClientTCP] sock: %d, hostname: \"%s\", addr: % 02X, port: %d, mode: %d\r\n",
			sock, hostname, addr, port, mode)
	}

	w.waitForChipReady()
	w.spiChipSelect()

	if len(hostname) > 0 {
		w.sendCmd(cmdStartClientTCP, 5)
		w.sendParamStr(hostname, false)
	} else {
		w.sendCmd(cmdStartClientTCP, 4)
	}

	w.sendParam32(addr, false)
	w.sendParam16(port, false)
	w.sendParam8(uint8(sock), false)
	w.sendParam8(mode, true)

	if len(hostname) > 0 {
		w.padTo4(17 + len(hostname))
	}

	w.spiChipDeselect()
	w.waitRspCmd1(cmdStartClientTCP)
}

func (w *wifinina) getSocket() sock {
	if debugging(debugCmd) {
		fmt.Printf("    [cmdGetSocket]\r\n")
	}
	return sock(w.getUint8(w.req0(cmdGetSocket)))
}

func (w *wifinina) getClientState(sock sock) uint8 {
	if debugging(debugCmd) {
		fmt.Printf("    [cmdGetClientStateTCP] sock: %d\r\n", sock)
	}
	return w.getUint8(w.reqUint8(cmdGetClientStateTCP, uint8(sock)))
}

func (w *wifinina) sendData(sock sock, buf []byte) uint16 {
	if debugging(debugCmd) {
		fmt.Printf("    [cmdSendDataTCP] sock: %d, len(buf): %d\r\n",
			sock, len(buf))
	}

	w.waitForChipReady()
	w.spiChipSelect()
	l := w.sendCmd(cmdSendDataTCP, 2)
	l += w.sendParamBuf([]byte{uint8(sock)}, false)
	l += w.sendParamBuf(buf, true)
	w.addPadding(l)
	w.spiChipDeselect()

	sent := w.getUint16(w.waitRspCmd1(cmdSendDataTCP))
	return bits.RotateLeft16(sent, 8)
}

func (w *wifinina) checkDataSent(sock sock) bool {
	if debugging(debugCmd) {
		fmt.Printf("    [cmdDataSentTCP] sock: %d\r\n", sock)
	}
	sent := w.getUint8(w.reqUint8(cmdDataSentTCP, uint8(sock)))
	return sent > 0
}

func (w *wifinina) getDataBuf(sock sock, buf []byte) uint16 {
	if debugging(debugCmd) {
		fmt.Printf("    [cmdGetDatabufTCP] sock: %d, len(buf): %d\r\n",
			sock, len(buf))
	}

	w.waitForChipReady()
	w.spiChipSelect()
	p := uint16(len(buf))
	l := w.sendCmd(cmdGetDatabufTCP, 2)
	l += w.sendParamBuf([]byte{uint8(sock)}, false)
	l += w.sendParamBuf([]byte{uint8(p & 0x00FF), uint8((p) >> 8)}, true)
	w.addPadding(l)
	w.spiChipDeselect()

	w.waitForChipReady()
	w.spiChipSelect()
	n := w.waitRspBuf16(cmdGetDatabufTCP, buf)
	w.spiChipDeselect()

	if n > 0 {
		if debugging(debugCmd) {
			fmt.Printf("    [<--cmdGetDatabufTCP] sock: %d, got n: %d\r\n",
				sock, n)
		}
	}

	return n
}

func (w *wifinina) stopClient(sock sock) {
	if debugging(debugCmd) {
		fmt.Printf("    [cmdStopClientTCP] sock: %d\r\n", sock)
	}
	w.getUint8(w.reqUint8(cmdStopClientTCP, uint8(sock)))
}

func (w *wifinina) startServer(sock sock, port uint16, mode uint8) {
	if debugging(debugCmd) {
		fmt.Printf("    [cmdStartServerTCP] sock: %d, port: %d, mode: %d\r\n",
			sock, port, mode)
	}

	w.waitForChipReady()
	w.spiChipSelect()
	l := w.sendCmd(cmdStartServerTCP, 3)
	l += w.sendParam16(port, false)
	l += w.sendParam8(uint8(sock), false)
	l += w.sendParam8(mode, true)
	w.addPadding(l)
	w.spiChipDeselect()

	w.waitRspCmd1(cmdStartServerTCP)
}

func (w *wifinina) accept(s sock) sock {

	if debugging(debugCmd) {
		fmt.Printf("    [cmdAvailDataTCP] sock: %d\r\n", s)
	}

	w.waitForChipReady()
	w.spiChipSelect()
	l := w.sendCmd(cmdAvailDataTCP, 1)
	l += w.sendParam8(uint8(s), true)
	w.addPadding(l)
	w.spiChipDeselect()

	newsock16 := w.getUint16(w.waitRspCmd1(cmdAvailDataTCP))
	newsock := sock(uint8(bits.RotateLeft16(newsock16, 8)))

	if newsock != noSocketAvail {
		if debugging(debugCmd) {
			fmt.Printf("    [cmdAvailDataTCP-->] sock: %d, got sock: %d\r\n",
				s, newsock)
		}
	}

	return newsock
}

// insertDataBuf adds data to the buffer used for sending UDP data
func (w *wifinina) insertDataBuf(sock sock, buf []byte) bool {

	if debugging(debugCmd) {
		fmt.Printf("    [cmdInsertDataBuf] sock: %d, len(buf): %d\r\n",
			sock, len(buf))
	}

	w.waitForChipReady()
	w.spiChipSelect()
	l := w.sendCmd(cmdInsertDataBuf, 2)
	l += w.sendParamBuf([]byte{uint8(sock)}, false)
	l += w.sendParamBuf(buf, true)
	w.addPadding(l)
	w.spiChipDeselect()

	n := w.getUint8(w.waitRspCmd1(cmdInsertDataBuf))
	return n == 1
}

// sendUDPData sends the data previously added to the UDP buffer
func (w *wifinina) sendUDPData(sock sock) bool {

	if debugging(debugCmd) {
		fmt.Printf("    [cmdSendDataUDP] sock: %d\r\n", sock)
	}

	w.waitForChipReady()
	w.spiChipSelect()
	l := w.sendCmd(cmdSendDataUDP, 1)
	l += w.sendParam8(uint8(sock), true)
	w.addPadding(l)
	w.spiChipDeselect()

	n := w.getUint8(w.waitRspCmd1(cmdSendDataUDP))
	return n == 1
}

func (w *wifinina) disconnect() {
	if debugging(debugCmd) {
		fmt.Printf("    [cmdDisconnect]\r\n")
	}
	w.req1(cmdDisconnect)
}

func (w *wifinina) getFwVersion() string {
	if debugging(debugCmd) {
		fmt.Printf("    [cmdGetFwVersion]\r\n")
	}
	return w.getString(w.req0(cmdGetFwVersion))
}

func (w *wifinina) getConnectionStatus() connectionStatus {
	if debugging(debugCmd) {
		fmt.Printf("    [cmdGetConnStatus]\r\n")
	}
	status := w.getUint8(w.req0(cmdGetConnStatus))
	return connectionStatus(status)
}

func (w *wifinina) getCurrentencryptionType() encryptionType {
	enctype := w.getUint8(w.req1(cmdGetCurrEncrType))
	return encryptionType(enctype)
}

func (w *wifinina) getCurrentBSSID() net.HardwareAddr {
	return w.getMACAddress(w.req1(cmdGetCurrBSSID))
}

func (w *wifinina) getCurrentRSSI() int32 {
	return w.getInt32(w.req1(cmdGetCurrRSSI))
}

func (w *wifinina) getCurrentSSID() string {
	return w.getString(w.req1(cmdGetCurrSSID))
}

func (w *wifinina) getMACAddr() net.HardwareAddr {
	if debugging(debugCmd) {
		fmt.Printf("    [cmdGetMACAddr]\r\n")
	}
	return w.getMACAddress(w.req1(cmdGetMACAddr))
}

func (w *wifinina) faultf(f string, args ...any) {
	// Only record the first fault
	if w.fault == nil {
		w.fault = fmt.Errorf(f, args...)
	}
}

func (w *wifinina) getIP() (ip, subnet, gateway net.IP) {
	if debugging(debugCmd) {
		fmt.Printf("    [cmdGetIPAddr]\r\n")
	}
	sl := make([]string, 3)
	l := w.reqRspStr1(cmdGetIPAddr, dummyData, sl)
	if l != 3 {
		w.faultf("getIP wanted l=3, got l=%d", l)
		return
	}
	ip, subnet, gateway = make([]byte, 4), make([]byte, 4), make([]byte, 4)
	copy(ip[:], sl[0])
	copy(subnet[:], sl[1])
	copy(gateway[:], sl[2])
	return
}

func (w *wifinina) getHostByName(name string) string {
	if debugging(debugCmd) {
		fmt.Printf("    [cmdGetHostByName]\r\n")
	}
	ok := w.getUint8(w.reqStr(cmdReqHostByName, name))
	if ok != 1 {
		return ""
	}
	return w.getString(w.req0(cmdGetHostByName))
}

func (w *wifinina) getNetworkBSSID(idx int) net.HardwareAddr {
	if idx < 0 || idx >= maxNetworks {
		return net.HardwareAddr{}
	}
	return w.getMACAddress(w.reqUint8(cmdGetIdxBSSID, uint8(idx)))
}

func (w *wifinina) getNetworkChannel(idx int) uint8 {
	if idx < 0 || idx >= maxNetworks {
		return 0
	}
	return w.getUint8(w.reqUint8(cmdGetIdxChannel, uint8(idx)))
}

func (w *wifinina) getNetworkEncrType(idx int) encryptionType {
	if idx < 0 || idx >= maxNetworks {
		return 0
	}
	enctype := w.getUint8(w.reqUint8(cmdGetIdxEncrType, uint8(idx)))
	return encryptionType(enctype)
}

func (w *wifinina) getNetworkRSSI(idx int) int32 {
	if idx < 0 || idx >= maxNetworks {
		return 0
	}
	return w.getInt32(w.reqUint8(cmdGetIdxRSSI, uint8(idx)))
}

func (w *wifinina) getNetworkSSID(idx int) string {
	if idx < 0 || idx >= maxNetworks {
		return ""
	}
	return w.ssids[idx]
}

func (w *wifinina) getReasonCode() uint8 {
	return w.getUint8(w.req0(cmdGetReasonCode))
}

// getTime is the time as a Unix timestamp
func (w *wifinina) getTime() uint32 {
	return w.getUint32(w.req0(cmdGetTime))
}

func (w *wifinina) getTemperature() float32 {
	return w.getFloat32(w.req0(cmdGetTemperature))
}

func (w *wifinina) setDebug(on bool) {
	var v uint8
	if on {
		v = 1
	}
	w.reqUint8(cmdSetDebug, v)
}

func (w *wifinina) setNetwork(ssid string) {
	w.reqStr(cmdSetNet, ssid)
}

func (w *wifinina) setPassphrase(ssid string, passphrase string) {

	if debugging(debugCmd) {
		fmt.Printf("    [cmdSetPassphrase] ssid: %s, passphrase: ******\r\n",
			ssid)
	}

	// Dont' show passphrase in debug output
	saveDebug := _debug
	_debug = _debug & ^debugDetail
	w.reqStr2(cmdSetPassphrase, ssid, passphrase)
	_debug = saveDebug
}

func (w *wifinina) setKey(ssid string, index uint8, key string) {

	w.waitForChipReady()
	w.spiChipSelect()
	w.sendCmd(cmdSetKey, 3)
	w.sendParamStr(ssid, false)
	w.sendParam8(index, false)
	w.sendParamStr(key, true)
	w.padTo4(8 + len(ssid) + len(key))
	w.spiChipDeselect()

	w.waitRspCmd1(cmdSetKey)
}

func (w *wifinina) setNetworkForAP(ssid string) {
	w.reqStr(cmdSetAPNet, ssid)
}

func (w *wifinina) setPassphraseForAP(ssid string, passphrase string) {
	w.reqStr2(cmdSetAPPassphrase, ssid, passphrase)
}

func (w *wifinina) setDNS(which uint8, dns1 uint32, dns2 uint32) {
	w.waitForChipReady()
	w.spiChipSelect()
	w.sendCmd(cmdSetDNSConfig, 3)
	w.sendParam8(which, false)
	w.sendParam32(dns1, false)
	w.sendParam32(dns2, true)
	//pad??
	w.spiChipDeselect()

	w.waitRspCmd1(cmdSetDNSConfig)
}

func (w *wifinina) setHostname(hostname string) {
	w.waitForChipReady()
	w.spiChipSelect()
	w.sendCmd(cmdSetHostname, 3)
	w.sendParamStr(hostname, true)
	w.padTo4(5 + len(hostname))
	w.spiChipDeselect()

	w.waitRspCmd1(cmdSetHostname)
}

func (w *wifinina) setPowerMode(mode uint8) {
	w.reqUint8(cmdSetPowerMode, mode)
}

func (w *wifinina) scanNetworks() uint8 {
	return w.reqRspStr0(cmdScanNetworks, w.ssids[:])
}

func (w *wifinina) startScanNetworks() uint8 {
	return w.getUint8(w.req0(cmdStartScanNetworks))
}

func (w *wifinina) PinMode(pin uint8, mode uint8) {
	if debugging(debugCmd) {
		fmt.Printf("    [cmdSetPinMode] pin: %d, mode: %d\r\n", pin, mode)
	}
	w.req2Uint8(cmdSetPinMode, pin, mode)
}

func (w *wifinina) DigitalWrite(pin uint8, value uint8) {
	if debugging(debugCmd) {
		fmt.Printf("    [cmdSetDigitialWrite] pin: %d, value: %d\r\n", pin, value)
	}
	w.req2Uint8(cmdSetDigitalWrite, pin, value)
}

func (w *wifinina) AnalogWrite(pin uint8, value uint8) {
	w.req2Uint8(cmdSetAnalogWrite, pin, value)
}

func (w *wifinina) getString(l uint8) string {
	return string(w.buf[0:l])
}

func (w *wifinina) getUint8(l uint8) uint8 {
	if l == 1 {
		return w.buf[0]
	}
	w.faultf("expected length 1, was actually %d", l)
	return 0
}

func (w *wifinina) getUint16(l uint8) uint16 {
	if l == 2 {
		return binary.BigEndian.Uint16(w.buf[0:2])
	}
	w.faultf("expected length 2, was actually %d", l)
	return 0
}

func (w *wifinina) getUint32(l uint8) uint32 {
	if l == 4 {
		return binary.BigEndian.Uint32(w.buf[0:4])
	}
	w.faultf("expected length 4, was actually %d", l)
	return 0
}

func (w *wifinina) getInt32(l uint8) int32 {
	return int32(w.getUint32(l))
}

func (w *wifinina) getFloat32(l uint8) float32 {
	return float32(w.getUint32(l))
}

func (w *wifinina) getMACAddress(l uint8) net.HardwareAddr {
	if l == 6 {
		mac := w.buf[0:6]
		// Reverse the bytes
		for i, j := 0, len(mac)-1; i < j; i, j = i+1, j-1 {
			mac[i], mac[j] = mac[j], mac[i]
		}
		return mac
	}
	w.faultf("expected length 6, was actually %d", l)
	return net.HardwareAddr{}
}

func (w *wifinina) transfer(b byte) byte {
	v, err := w.spi.Transfer(b)
	if err != nil {
		w.faultf("SPI.Transfer")
		return 0
	}
	return v
}

// Cmd Struct Message */
// ._______________________________________________________________________.
// | START CMD | C/R  | CMD  | N.PARAM | PARAM LEN | PARAM  | .. | END CMD |
// |___________|______|______|_________|___________|________|____|_________|
// |   8 bit   | 1bit | 7bit |  8bit   |   8bit    | nbytes | .. |   8bit  |
// |___________|______|______|_________|___________|________|____|_________|

// req0 sends a command to the device with no request parameters
func (w *wifinina) req0(cmd uint8) uint8 {
	w.sendCmd0(cmd)
	return w.waitRspCmd1(cmd)
}

// req1 sends a command to the device with a single dummy parameters of 0xFF
func (w *wifinina) req1(cmd uint8) uint8 {
	return w.reqUint8(cmd, dummyData)
}

// reqUint8 sends a command to the device with a single uint8 parameter
func (w *wifinina) reqUint8(cmd uint8, data uint8) uint8 {
	w.sendCmdPadded1(cmd, data)
	return w.waitRspCmd1(cmd)
}

// req2Uint8 sends a command to the device with two uint8 parameters
func (w *wifinina) req2Uint8(cmd, p1, p2 uint8) uint8 {
	w.sendCmdPadded2(cmd, p1, p2)
	return w.waitRspCmd1(cmd)
}

// reqStr sends a command to the device with a single string parameter
func (w *wifinina) reqStr(cmd uint8, p1 string) uint8 {
	w.sendCmdStr(cmd, p1)
	return w.waitRspCmd1(cmd)
}

// reqStr2 sends a command to the device with 2 string parameters
func (w *wifinina) reqStr2(cmd uint8, p1 string, p2 string) {
	w.sendCmdStr2(cmd, p1, p2)
	w.waitRspCmd1(cmd)
}

// reqStrRsp0 sends a command passing a string slice for the response
func (w *wifinina) reqRspStr0(cmd uint8, sl []string) (l uint8) {
	w.sendCmd0(cmd)
	w.waitForChipReady()
	w.spiChipSelect()
	l = w.waitRspStr(cmd, sl)
	w.spiChipDeselect()
	return
}

// reqStrRsp1 sends a command with a uint8 param and a string slice for the response
func (w *wifinina) reqRspStr1(cmd uint8, data uint8, sl []string) uint8 {
	w.sendCmdPadded1(cmd, data)
	w.waitForChipReady()
	w.spiChipSelect()
	l := w.waitRspStr(cmd, sl)
	w.spiChipDeselect()
	return l
}

func (w *wifinina) sendCmd0(cmd uint8) {
	w.waitForChipReady()
	w.spiChipSelect()
	w.sendCmd(cmd, 0)
	w.spiChipDeselect()
}

func (w *wifinina) sendCmdPadded1(cmd uint8, data uint8) {
	w.waitForChipReady()
	w.spiChipSelect()
	w.sendCmd(cmd, 1)
	w.sendParam8(data, true)
	w.transfer(dummyData)
	w.transfer(dummyData)
	w.spiChipDeselect()
	return
}

func (w *wifinina) sendCmdPadded2(cmd, data1, data2 uint8) {
	w.waitForChipReady()
	w.spiChipSelect()
	w.sendCmd(cmd, 1)
	w.sendParam8(data1, false)
	w.sendParam8(data2, true)
	w.transfer(dummyData)
	w.spiChipDeselect()
}

func (w *wifinina) sendCmdStr(cmd uint8, p1 string) {
	w.waitForChipReady()
	w.spiChipSelect()
	w.sendCmd(cmd, 1)
	w.sendParamStr(p1, true)
	w.padTo4(5 + len(p1))
	w.spiChipDeselect()
}

func (w *wifinina) sendCmdStr2(cmd uint8, p1 string, p2 string) {
	w.waitForChipReady()
	w.spiChipSelect()
	w.sendCmd(cmd, 2)
	w.sendParamStr(p1, false)
	w.sendParamStr(p2, true)
	w.padTo4(6 + len(p1) + len(p2))
	w.spiChipDeselect()
}

func (w *wifinina) waitRspCmd1(cmd uint8) uint8 {
	w.waitForChipReady()
	w.spiChipSelect()
	l := w.waitRspCmd(cmd, 1)
	w.spiChipDeselect()
	return l
}

func (w *wifinina) sendCmd(cmd uint8, numParam uint8) (l int) {
	if debugging(debugDetail) {
		fmt.Printf("        sendCmd: %02X %02X %02X",
			cmdStart, cmd & ^(uint8(flagReply)), numParam)
	}

	l = 3
	w.transfer(cmdStart)
	w.transfer(cmd & ^(uint8(flagReply)))
	w.transfer(numParam)
	if numParam == 0 {
		w.transfer(cmdEnd)
		l += 1
		if debugging(debugDetail) {
			fmt.Printf(" %02X", cmdEnd)
		}
	}

	if debugging(debugDetail) {
		fmt.Printf(" (%d)\r\n", l)
	}
	return
}

func (w *wifinina) sendParamLen16(p uint16) (l int) {
	w.transfer(uint8(p >> 8))
	w.transfer(uint8(p & 0xFF))
	if debugging(debugDetail) {
		fmt.Printf("        %02X %02X", uint8(p>>8), uint8(p&0xFF))
	}
	return 2
}

func (w *wifinina) sendParamBuf(p []byte, isLastParam bool) (l int) {
	if debugging(debugDetail) {
		fmt.Printf("        sendParamBuf:")
	}
	l += w.sendParamLen16(uint16(len(p)))
	for _, b := range p {
		if debugging(debugDetail) {
			fmt.Printf(" %02X", b)
		}
		w.transfer(b)
		l += 1
	}
	if isLastParam {
		if debugging(debugDetail) {
			fmt.Printf(" %02X", cmdEnd)
		}
		w.transfer(cmdEnd)
		l += 1
	}
	if debugging(debugDetail) {
		fmt.Printf(" (%d) \r\n", l)
	}
	return
}

func (w *wifinina) sendParamStr(p string, isLastParam bool) (l int) {
	if debugging(debugDetail) {
		fmt.Printf("        sendParamStr: p: %s, lastParam: %t\r\n", p, isLastParam)
	}
	l = len(p)
	w.transfer(uint8(l))
	if l > 0 {
		w.spi.Tx([]byte(p), nil)
	}
	if isLastParam {
		w.transfer(cmdEnd)
		l += 1
	}
	return
}

func (w *wifinina) sendParam8(p uint8, isLastParam bool) (l int) {
	if debugging(debugDetail) {
		fmt.Printf("        sendParam8: p: %d, lastParam: %t\r\n", p, isLastParam)
	}
	l = 2
	w.transfer(1)
	w.transfer(p)
	if isLastParam {
		w.transfer(cmdEnd)
		l += 1
	}
	return
}

func (w *wifinina) sendParam16(p uint16, isLastParam bool) (l int) {
	if debugging(debugDetail) {
		fmt.Printf("        sendParam16: p: %d, lastParam: %t\r\n", p, isLastParam)
	}
	l = 3
	w.transfer(2)
	w.transfer(uint8(p >> 8))
	w.transfer(uint8(p & 0xFF))
	if isLastParam {
		w.transfer(cmdEnd)
		l += 1
	}
	return
}

func (w *wifinina) sendParam32(p uint32, isLastParam bool) (l int) {
	if debugging(debugDetail) {
		fmt.Printf("        sendParam32: p: %d, lastParam: %t\r\n", p, isLastParam)
	}
	l = 5
	w.transfer(4)
	w.transfer(uint8(p >> 24))
	w.transfer(uint8(p >> 16))
	w.transfer(uint8(p >> 8))
	w.transfer(uint8(p & 0xFF))
	if isLastParam {
		w.transfer(cmdEnd)
		l += 1
	}
	return
}

func (w *wifinina) waitForChipReady() {
	if debugging(debugDetail) {
		fmt.Printf("        waitForChipReady\r\n")
	}

	for i := 0; w.ack.Get(); i++ {
		time.Sleep(1 * time.Millisecond)
		if i == 10000 {
			w.faultf("hung in waitForChipReady")
			return
		}
	}
}

func (w *wifinina) spiChipSelect() {
	if debugging(debugDetail) {
		fmt.Printf("        spiChipSelect\r\n")
	}
	w.cs.Low()
	start := time.Now()
	for time.Since(start) < 5*time.Millisecond {
		if w.ack.Get() {
			return
		}
		time.Sleep(100 * time.Microsecond)
	}
	w.faultf("hung in spiChipSelect")
}

func (w *wifinina) spiChipDeselect() {
	if debugging(debugDetail) {
		fmt.Printf("        spiChipDeselect\r\n")
	}
	w.cs.High()
}

func (w *wifinina) waitSpiChar(wait byte) {

	if debugging(debugDetail) {
		fmt.Printf("        waitSpiChar: wait: %02X\r\n", wait)
	}

	var read byte

	for timeout := 1000; read != wait && timeout > 0; timeout-- {
		w.readParam(&read)
		if read == cmdErr {
			w.faultf("cmdErr received, waiting for %d", wait)
			return
		}
	}

	if read != wait {
		w.faultf("timeout waiting for SPI char %02X\r\n", wait)
	}
}

func (w *wifinina) waitRspCmd(cmd uint8, np uint8) (l uint8) {

	if debugging(debugDetail) {
		fmt.Printf("        waitRspCmd: cmd: %02X, np: %d\r\n", cmd, np)
	}

	var data byte

	w.waitSpiChar(cmdStart)

	if !w.readAndCheckByte(cmd|flagReply, &data) {
		w.faultf("expected cmd %02X, read %02X", cmd, data)
		return
	}

	if w.readAndCheckByte(np, &data) {

		w.readParam(&l)
		for i := uint8(0); i < l; i++ {
			w.readParam(&w.buf[i])
		}

		if !w.readAndCheckByte(cmdEnd, &data) {
			w.faultf("expected cmdEnd, read %02X", data)
		}
	}

	return
}

func (w *wifinina) waitRspBuf16(cmd uint8, buf []byte) (l uint16) {

	if debugging(debugDetail) {
		fmt.Printf("        waitRspBuf16: cmd: %02X, len(buf): %d\r\n", cmd, len(buf))
	}

	var data byte

	w.waitSpiChar(cmdStart)

	if !w.readAndCheckByte(cmd|flagReply, &data) {
		w.faultf("expected cmd %02X, read %02X", cmd, data)
		return
	}

	if w.readAndCheckByte(1, &data) {
		l = w.readParamLen16()
		for i := uint16(0); i < l; i++ {
			w.readParam(&buf[i])
		}
		if !w.readAndCheckByte(cmdEnd, &data) {
			w.faultf("expected cmdEnd, read %02X", data)
		}
	}

	return
}

func (w *wifinina) waitRspStr(cmd uint8, sl []string) (numRead uint8) {

	if debugging(debugDetail) {
		fmt.Printf("        waitRspStr: cmd: %02X, len(sl): %d\r\n", cmd, len(sl))
	}

	var data byte

	w.waitSpiChar(cmdStart)

	if !w.readAndCheckByte(cmd|flagReply, &data) {
		w.faultf("expected cmd %02X, read %02X", cmd, data)
		return
	}

	numRead = w.transfer(dummyData)
	if numRead == 0 {
		w.faultf("waitRspStr numRead == 0")
		return
	}

	maxNumRead := uint8(len(sl))
	for j, l := uint8(0), uint8(0); j < numRead; j++ {
		w.readParam(&l)
		for i := uint8(0); i < l; i++ {
			w.readParam(&w.buf[i])
		}
		if j < maxNumRead {
			sl[j] = string(w.buf[0:l])
			if debugging(debugDetail) {
				fmt.Printf("            str: %d (%d) - %08X\r\n", j, l, []byte(sl[j]))
			}
		}
	}

	for j := numRead; j < maxNumRead; j++ {
		if debugging(debugDetail) {
			fmt.Printf("            str: ", j, "\"\"\r")
		}
		sl[j] = ""
	}

	if !w.readAndCheckByte(cmdEnd, &data) {
		w.faultf("expected cmdEnd, read %02X", data)
		return
	}

	if numRead > maxNumRead {
		numRead = maxNumRead
	}
	return
}

func (w *wifinina) readAndCheckByte(check byte, read *byte) bool {
	w.readParam(read)
	return *read == check
}

// readParamLen16 reads 2 bytes from the SPI bus (MSB first), returning uint16
func (w *wifinina) readParamLen16() (v uint16) {
	b := w.transfer(0xFF)
	v = uint16(b) << 8
	b = w.transfer(0xFF)
	v |= uint16(b)
	return
}

func (w *wifinina) readParam(b *byte) {
	*b = w.transfer(0xFF)
}

func (w *wifinina) addPadding(l int) {
	if debugging(debugDetail) {
		fmt.Printf("        addPadding: l: %d\r\n", l)
	}
	for i := (4 - (l % 4)) & 3; i > 0; i-- {
		if debugging(debugDetail) {
			fmt.Printf("            padding\r\n")
		}
		w.transfer(dummyData)
	}
}

func (w *wifinina) padTo4(l int) {
	if debugging(debugDetail) {
		fmt.Printf("        padTo4: l: %d\r\n", l)
	}

	for l%4 != 0 {
		if debugging(debugDetail) {
			fmt.Printf("            padding\r\n")
		}
		w.transfer(dummyData)
		l++
	}
}
