// Package unor4wifi implements a WiFi network driver for the Arduino UNO R4
// WiFi. The on-board ESP32-S3 runs the Arduino "uno-r4-wifi-usb-bridge"
// firmware. This driver talks to that firmware over UART with the same AT
// style protocol that the Arduino WiFiS3 library uses.
//
// Supported: STA and AP mode, TCP client, TCP server, UDP, DNS lookup.
// Not supported: TLS offload (use Go TLS), socket options.
//
// References:
// https://github.com/arduino/uno-r4-wifi-usb-bridge/blob/main/UNOR4USBBridge/commands.h
// https://github.com/arduino/ArduinoCore-renesas/tree/main/libraries/WiFiS3/src
package unor4wifi // import "tinygo.org/x/drivers/unor4wifi"

import (
	"fmt"
	"io"
	"machine"
	"net"
	"net/netip"
	"strconv"
	"strings"
	"sync"
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/netdev"
	"tinygo.org/x/drivers/netlink"
)

var _debug debug = debugBasic

//var _debug debug = debugBasic | debugNetdev
//var _debug debug = debugBasic | debugNetdev | debugCmd

var (
	driverName = "TinyGo UNO R4 WiFi (ESP32-S3 bridge) network device driver (unor4wifi)"
)

const (
	defaultBaudRate = 115200
	defaultTimeout  = 10 * time.Second
	maxSockets      = 8
	// The bridge RX path is a UART, so keep transfers small.
	maxChunk = 1024
	// Poll interval while a Recv or Accept waits for data.
	pollInterval = 5 * time.Millisecond
	// TCP connect timeout passed to the firmware.
	connectTimeout = 10 * time.Second

	// WiFi status values, see WiFiS3/src/WiFiTypes.h.
	wlIdleStatus     = 0
	wlConnected      = 3
	wlConnectFailed  = 4
	wlConnectionLost = 5
	wlDisconnected   = 6
	wlAPListening    = 7

	// ip_type values, see commands.h.
	ipAddr      = 0
	gatewayAddr = 1
	netmaskAddr = 2
	dns1Addr    = 3

	modeSTA = 1
	modeAP  = 2
)

type Config struct {
	Uart     *machine.UART
	Tx       machine.Pin
	Rx       machine.Pin
	BaudRate uint32
	// Timeout for one modem transaction. Default 10 s.
	Timeout time.Duration
}

type socket struct {
	protocol int
	// Socket number inside the ESP32 firmware, -1 when not yet allocated.
	modemSock int
	server    bool
	laddr     netip.AddrPort
	raddr     netip.AddrPort
}

type device struct {
	cfg      *Config
	modem    modem
	mu       sync.Mutex
	notifyCb func(netlink.Event)

	params       *netlink.ConnectParams
	netConnected bool
	mac          net.HardwareAddr
	ip           netip.Addr
	gateway      netip.Addr
	netmask      netip.Addr
	dns          netip.Addr

	sockets [maxSockets]*socket
	rxbuf   [maxChunk]byte
}

// New returns a new driver instance. The UART is configured in NetConnect.
func New(cfg *Config) *device {
	d := &device{cfg: cfg}
	d.modem.timeout = cfg.Timeout
	if d.modem.timeout == 0 {
		d.modem.timeout = defaultTimeout
	}
	return d
}

// escapeArg escapes the characters that the bridge argument parser treats
// as special. See uno-r4-wifi-usb-bridge/UNOR4USBBridge/parser.cpp,
// ATParser::parse, state Argument.
func escapeArg(s string) string {
	if !strings.ContainsAny(s, `\",`) {
		return s
	}
	var sb strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '\\' || c == '"' || c == ',' {
			sb.WriteByte('\\')
		}
		sb.WriteByte(c)
	}
	return sb.String()
}

func (d *device) setupUART() {
	baud := d.cfg.BaudRate
	if baud == 0 {
		baud = defaultBaudRate
	}
	d.modem.uart = d.cfg.Uart
	d.modem.uart.Configure(machine.UARTConfig{
		BaudRate: baud,
		TX:       d.cfg.Tx,
		RX:       d.cfg.Rx,
	})
}

// reset clears all firmware sockets and WiFi state. Retries until the
// firmware answers.
func (d *device) reset() error {
	var err error
	for i := 0; i < 10; i++ {
		_, err = d.modem.execTimeout("AT+SOFTRESETWIFI", 500*time.Millisecond)
		if err == nil {
			return nil
		}
	}
	return err
}

func (d *device) firmwareVersion() string {
	res, err := d.modem.exec("AT+FWVERSION?")
	if err != nil {
		return "unknown"
	}
	return res
}

func (d *device) readMAC(cmd string) {
	res, err := d.modem.exec(cmd)
	if err != nil {
		return
	}
	if mac, err := net.ParseMAC(res); err == nil {
		d.mac = mac
	}
}

func (d *device) readAddr(cmd string) netip.Addr {
	res, err := d.modem.exec(cmd)
	if err != nil {
		return netip.Addr{}
	}
	addr, _ := netip.ParseAddr(strings.TrimSpace(res))
	return addr
}

func (d *device) status() int {
	n, err := d.modem.execInt("AT+GETSTATUS?")
	if err != nil {
		return wlIdleStatus
	}
	return n
}

func (d *device) connectSTA() error {
	p := d.params
	timeout := p.ConnectTimeout
	if timeout == 0 {
		timeout = netlink.DefaultConnectTimeout
	}

	if _, err := d.modem.exec("AT+WIFIMODE=" + strconv.Itoa(modeSTA)); err != nil {
		return err
	}
	if p.Hostname != "" {
		d.modem.exec("AT+HOSTNAME=" + escapeArg(p.Hostname))
	}

	cmd := "AT+BEGINSTA=" + escapeArg(p.Ssid)
	if p.Passphrase != "" {
		cmd += "," + escapeArg(p.Passphrase)
	}
	if _, err := d.modem.exec(cmd); err != nil {
		return netlink.ErrConnectFailed
	}

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		switch d.status() {
		case wlConnected:
			// The bridge reports WL_CONNECTED before DHCP completes.
			d.ip = d.readAddr("AT+IPSTA=" + strconv.Itoa(ipAddr))
			if !d.ip.IsValid() || d.ip.IsUnspecified() {
				break
			}
			d.gateway = d.readAddr("AT+IPSTA=" + strconv.Itoa(gatewayAddr))
			d.netmask = d.readAddr("AT+IPSTA=" + strconv.Itoa(netmaskAddr))
			d.dns = d.readAddr("AT+IPSTA=" + strconv.Itoa(dns1Addr))
			d.readMAC("AT+MACSTA?")
			return nil
		case wlConnectFailed:
			return netlink.ErrConnectFailed
		}
		time.Sleep(100 * time.Millisecond)
	}
	return netlink.ErrConnectTimeout
}

func (d *device) connectAP() error {
	p := d.params
	if _, err := d.modem.exec("AT+WIFIMODE=" + strconv.Itoa(modeAP)); err != nil {
		return err
	}
	// Channel 1 is the WiFiS3 default.
	cmd := "AT+BEGINSOFTAP=" + escapeArg(p.Ssid) + "," + escapeArg(p.Passphrase) + ",1"
	n, err := d.modem.execInt(cmd)
	if err != nil || n != wlAPListening {
		return netlink.ErrConnectFailed
	}
	d.ip = d.readAddr("AT+IPSOFTAP")
	d.gateway = d.ip
	d.readMAC("AT+MACSOFTAP?")
	return nil
}

func (d *device) NetConnect(params *netlink.ConnectParams) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.netConnected {
		return netlink.ErrConnected
	}
	if len(params.Ssid) == 0 {
		return netlink.ErrMissingSSID
	}
	if params.ConnectMode == netlink.ConnectModeSTA &&
		params.AuthType != netlink.AuthTypeOpen && len(params.Passphrase) < 8 {
		return netlink.ErrShortPassphrase
	}
	d.params = params

	d.setupUART()

	if debugging(debugBasic) {
		fmt.Printf("\r\n")
		fmt.Printf("%s\r\n\r\n", driverName)
		fmt.Printf("Driver version           : %s\r\n", drivers.Version)
	}

	if err := d.reset(); err != nil {
		return err
	}
	if debugging(debugBasic) {
		fmt.Printf("Bridge firmware version  : %s\r\n", d.firmwareVersion())
	}

	var err error
	switch params.ConnectMode {
	case netlink.ConnectModeSTA:
		if debugging(debugBasic) {
			fmt.Printf("Connecting to Wifi SSID '%s'...", params.Ssid)
		}
		err = d.connectSTA()
	case netlink.ConnectModeAP:
		if debugging(debugBasic) {
			fmt.Printf("Starting Wifi AP as SSID '%s'...", params.Ssid)
		}
		err = d.connectAP()
	default:
		return netlink.ErrConnectModeNoGood
	}
	if err != nil {
		if debugging(debugBasic) {
			fmt.Printf("FAILED (%s)\r\n", err.Error())
		}
		return err
	}
	if debugging(debugBasic) {
		fmt.Printf("CONNECTED\r\n")
		fmt.Printf("\r\n")
		fmt.Printf("MAC address              : %s\r\n", d.mac.String())
		fmt.Printf("IP address               : %s\r\n", d.ip)
		fmt.Printf("Gateway                  : %s\r\n", d.gateway)
		if d.dns.IsValid() {
			fmt.Printf("DNS server               : %s\r\n", d.dns)
		}
		fmt.Printf("\r\n")
	}

	d.netConnected = true
	if d.notifyCb != nil {
		d.notifyCb(netlink.EventNetUp)
	}
	return nil
}

func (d *device) NetDisconnect() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.netConnected {
		return
	}
	if d.params.ConnectMode == netlink.ConnectModeAP {
		d.modem.exec("AT+DISCONNECTSOFTAP")
	} else {
		d.modem.exec("AT+DISCONNECT")
	}
	d.netConnected = false

	if debugging(debugBasic) {
		fmt.Printf("\r\nDisconnected from Wifi SSID '%s'\r\n\r\n", d.params.Ssid)
	}

	if d.notifyCb != nil {
		d.notifyCb(netlink.EventNetDown)
	}
}

func (d *device) NetNotify(cb func(netlink.Event)) {
	d.notifyCb = cb
}

func (d *device) GetHardwareAddr() (net.HardwareAddr, error) {
	if debugging(debugNetdev) {
		fmt.Printf("[GetHardwareAddr]\r\n")
	}

	return d.mac, nil
}

func (d *device) GetHostByName(name string) (netip.Addr, error) {
	if debugging(debugNetdev) {
		fmt.Printf("[GetHostByName] name: %s\r\n", name)
	}

	if ip, err := netip.ParseAddr(name); err == nil {
		return ip, nil
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	res, err := d.modem.exec("AT+HOSTBYNAME=" + name)
	if err != nil {
		return netip.Addr{}, netdev.ErrHostUnknown
	}
	ip, err := netip.ParseAddr(strings.TrimSpace(res))
	if err != nil {
		return netip.Addr{}, netdev.ErrHostUnknown
	}
	return ip, nil
}

func (d *device) Addr() (netip.Addr, error) {
	if debugging(debugNetdev) {
		fmt.Printf("[GetIPAddr]\r\n")
	}

	return d.ip, nil
}

func (d *device) getSocket(sockfd int) (*socket, error) {
	if sockfd < 0 || sockfd >= maxSockets || d.sockets[sockfd] == nil {
		return nil, netdev.ErrInvalidSocketFd
	}
	return d.sockets[sockfd], nil
}

func (d *device) newSocket(protocol int, modemSock int) (int, *socket, error) {
	for fd := range d.sockets {
		if d.sockets[fd] == nil {
			s := &socket{protocol: protocol, modemSock: modemSock}
			d.sockets[fd] = s
			return fd, s, nil
		}
	}
	return -1, nil, netdev.ErrNoMoreSockets
}

func (d *device) Socket(domain int, stype int, protocol int) (int, error) {
	if debugging(debugNetdev) {
		fmt.Printf("[Socket] domain: %d, type: %d, protocol: %d\r\n",
			domain, stype, protocol)
	}

	if domain != netdev.AF_INET {
		return -1, netdev.ErrFamilyNotSupported
	}
	switch {
	case protocol == netdev.IPPROTO_TCP && stype == netdev.SOCK_STREAM:
	case protocol == netdev.IPPROTO_UDP && stype == netdev.SOCK_DGRAM:
	default:
		return -1, netdev.ErrProtocolNotSupported
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	fd, _, err := d.newSocket(protocol, -1)
	return fd, err
}

// udpBegin opens the firmware UDP socket on the local port, once.
func (d *device) udpBegin(s *socket) error {
	if s.modemSock >= 0 {
		return nil
	}
	n, err := d.modem.execInt("AT+UDPBEGIN=" + strconv.Itoa(int(s.laddr.Port())))
	if err != nil {
		return err
	}
	s.modemSock = n
	return nil
}

func (d *device) Bind(sockfd int, ip netip.AddrPort) error {
	if debugging(debugNetdev) {
		fmt.Printf("[Bind] sockfd: %d, addr: %s:%d\r\n", sockfd, ip.Addr(), ip.Port())
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	s, err := d.getSocket(sockfd)
	if err != nil {
		return err
	}
	s.laddr = ip
	if s.protocol == netdev.IPPROTO_UDP {
		return d.udpBegin(s)
	}
	return nil
}

func (d *device) Connect(sockfd int, host string, ip netip.AddrPort) error {
	if debugging(debugNetdev) {
		fmt.Printf("[Connect] sockfd: %d, host: %s, addr: %s\r\n", sockfd, host, ip)
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	s, err := d.getSocket(sockfd)
	if err != nil {
		return err
	}
	s.raddr = ip

	if s.protocol == netdev.IPPROTO_UDP {
		return d.udpBegin(s)
	}

	if s.modemSock < 0 {
		n, err := d.modem.execInt("AT+BEGINCLIENT")
		if err != nil {
			return err
		}
		s.modemSock = n
	}

	// The firmware resolves names and dotted addresses through the same
	// path, so always use +CLIENTCONNECT like the Arduino WiFiS3 library.
	if host == "" {
		host = ip.Addr().String()
	}
	cmd := "AT+CLIENTCONNECT=" + strconv.Itoa(s.modemSock) + "," + host + "," +
		strconv.Itoa(int(ip.Port())) + "," + strconv.Itoa(int(connectTimeout/time.Millisecond))
	if _, err := d.modem.execTimeout(cmd, connectTimeout+d.modem.timeout); err != nil {
		return netlink.ErrConnectFailed
	}
	return nil
}

func (d *device) Listen(sockfd int, backlog int) error {
	if debugging(debugNetdev) {
		fmt.Printf("[Listen] sockfd: %d\r\n", sockfd)
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	s, err := d.getSocket(sockfd)
	if err != nil {
		return err
	}
	if s.protocol != netdev.IPPROTO_TCP {
		return netdev.ErrProtocolNotSupported
	}
	n, err := d.modem.execInt("AT+BEGINSERVER=" + strconv.Itoa(int(s.laddr.Port())))
	if err != nil {
		return err
	}
	s.modemSock = n
	s.server = true
	return nil
}

func (d *device) remoteAddr(sock string) netip.AddrPort {
	ipStr, err := d.modem.exec("AT+REMOTEIP=" + sock)
	if err != nil {
		return netip.AddrPort{}
	}
	port, err := d.modem.execInt("AT+REMOTEPORT=" + sock)
	if err != nil {
		return netip.AddrPort{}
	}
	ip, _ := netip.ParseAddr(strings.TrimSpace(ipStr))
	return netip.AddrPortFrom(ip, uint16(port))
}

func (d *device) Accept(sockfd int) (int, netip.AddrPort, error) {
	if debugging(debugNetdev) {
		fmt.Printf("[Accept] sockfd: %d\r\n", sockfd)
	}

	for {
		d.mu.Lock()
		s, err := d.getSocket(sockfd)
		if err != nil {
			d.mu.Unlock()
			return -1, netip.AddrPort{}, err
		}
		if !s.server {
			d.mu.Unlock()
			return -1, netip.AddrPort{}, netdev.ErrNotSupported
		}
		n, err := d.modem.execInt("AT+SERVERACCEPT=" + strconv.Itoa(s.modemSock))
		if err == nil && n >= 0 {
			fd, cs, err := d.newSocket(netdev.IPPROTO_TCP, n)
			if err != nil {
				d.mu.Unlock()
				return -1, netip.AddrPort{}, err
			}
			cs.laddr = s.laddr
			cs.raddr = d.remoteAddr(strconv.Itoa(n))
			d.mu.Unlock()
			return fd, cs.raddr, nil
		}
		d.mu.Unlock()
		time.Sleep(pollInterval)
	}
}

func (d *device) sendTCP(s *socket, buf []byte) (int, error) {
	sock := strconv.Itoa(s.modemSock)
	sent := 0
	for sent < len(buf) {
		end := sent + maxChunk
		if end > len(buf) {
			end = len(buf)
		}
		chunk := buf[sent:end]
		cmd := "AT+CLIENTSEND=" + sock + "," + strconv.Itoa(len(chunk))
		if err := d.modem.execSend(cmd, chunk); err != nil {
			return sent, err
		}
		sent += len(chunk)
	}
	return sent, nil
}

func (d *device) sendUDP(s *socket, buf []byte) (int, error) {
	sock := strconv.Itoa(s.modemSock)
	cmd := "AT+UDPBEGINPACKETIP=" + sock + "," +
		strconv.Itoa(int(s.raddr.Port())) + "," + s.raddr.Addr().String()
	if _, err := d.modem.exec(cmd); err != nil {
		return 0, err
	}
	cmd = "AT+UDPWRITE=" + sock + "," + strconv.Itoa(len(buf))
	if err := d.modem.execSend(cmd, buf); err != nil {
		return 0, err
	}
	if _, err := d.modem.exec("AT+UDPENDPACKET=" + sock); err != nil {
		return 0, err
	}
	return len(buf), nil
}

func (d *device) Send(sockfd int, buf []byte, flags int, deadline time.Time) (int, error) {
	if debugging(debugNetdev) {
		fmt.Printf("[Send] sockfd: %d, len(buf): %d, flags: %d\r\n",
			sockfd, len(buf), flags)
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	s, err := d.getSocket(sockfd)
	if err != nil {
		return 0, err
	}
	if s.modemSock < 0 {
		return 0, netdev.ErrInvalidSocketFd
	}
	if s.protocol == netdev.IPPROTO_UDP {
		return d.sendUDP(s, buf)
	}
	return d.sendTCP(s, buf)
}

// recvOnce performs one receive attempt. Returns 0, nil when no data is
// available yet, and io.EOF when the connection is closed.
func (d *device) recvOnce(s *socket, buf []byte) (int, error) {
	sock := strconv.Itoa(s.modemSock)
	want := len(buf)
	if want > maxChunk {
		want = maxChunk
	}

	if s.protocol == netdev.IPPROTO_UDP {
		n, err := d.modem.execInt("AT+UDPPARSE=" + sock)
		if err != nil {
			return 0, err
		}
		if n == 0 {
			return 0, nil
		}
		if n < want {
			want = n
		}
		return d.modem.execSized("AT+UDPREAD="+sock+","+strconv.Itoa(want), buf)
	}

	n, err := d.modem.execSized("AT+CLIENTRECEIVE="+sock+","+strconv.Itoa(want), buf)
	if err != nil {
		return 0, err
	}
	if n > 0 {
		return n, nil
	}
	connected, err := d.modem.execInt("AT+CLIENTCONNECTED=" + sock)
	if err != nil || connected == 0 {
		return 0, io.EOF
	}
	return 0, nil
}

func (d *device) Recv(sockfd int, buf []byte, flags int, deadline time.Time) (int, error) {
	if debugging(debugNetdev) {
		fmt.Printf("[Recv] sockfd: %d, len(buf): %d, flags: %d\r\n",
			sockfd, len(buf), flags)
	}

	for {
		d.mu.Lock()
		s, err := d.getSocket(sockfd)
		if err != nil {
			d.mu.Unlock()
			return 0, err
		}
		if s.modemSock < 0 {
			d.mu.Unlock()
			return 0, netdev.ErrInvalidSocketFd
		}
		n, err := d.recvOnce(s, buf)
		d.mu.Unlock()
		if err != nil || n > 0 {
			return n, err
		}
		if !deadline.IsZero() && time.Now().After(deadline) {
			return 0, netdev.ErrTimeout
		}
		time.Sleep(pollInterval)
	}
}

func (d *device) Close(sockfd int) error {
	if debugging(debugNetdev) {
		fmt.Printf("[Close] sockfd: %d\r\n", sockfd)
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	s, err := d.getSocket(sockfd)
	if err != nil {
		return err
	}
	if s.modemSock >= 0 {
		sock := strconv.Itoa(s.modemSock)
		switch {
		case s.protocol == netdev.IPPROTO_UDP:
			_, err = d.modem.exec("AT+UDPSTOP=" + sock)
		case s.server:
			_, err = d.modem.exec("AT+SERVEREND=" + sock)
		default:
			_, err = d.modem.exec("AT+CLIENTCLOSE=" + sock)
		}
	}
	d.sockets[sockfd] = nil
	if err != nil {
		return netdev.ErrClosingSocket
	}
	return nil
}

func (d *device) SetSockOpt(sockfd int, level int, opt int, value interface{}) error {
	if debugging(debugNetdev) {
		fmt.Printf("[SetSockOpt] sockfd: %d\r\n", sockfd)
	}

	return netdev.ErrNotSupported
}
