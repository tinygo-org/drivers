// Package rtl8720dn implements TCP wireless communication over UART
// talking to a RealTek rtl8720dn module.
//
// 01/2023    sfeldma@gmail.com    Heavily modified to use netdev interface

package rtl8720dn // import "tinygo.org/x/drivers/rtl8720dn"

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"machine"
	"strings"
	"sync"
	"time"

	"tinygo.org/x/drivers/netdev"
)

var _debug debug = debugBasic

//var _debug debug = debugBasic | debugNetdev
//var _debug debug = debugBasic | debugNetdev | debugRpc

var (
	version    = "0.0.1"
	driverName = "Realtek rtl8720dn Wifi network device driver (rtl8720dn)"
)

const (
	F_SETFL      = 4
	O_NONBLOCK   = 1
	RTW_MODE_STA = 0x00000001
)

type macAddress netdev.HardwareAddr
type sock int32

type socket struct {
	protocol netdev.Protocol
	inuse    bool
}

type Config struct {
	// AP creditials
	Ssid       string
	Passphrase string

	// Enable
	En machine.Pin

	// UART config
	Uart     *machine.UART
	Tx       machine.Pin
	Rx       machine.Pin
	Baudrate uint32

	// Retries is how many attempts to connect before returning with a
	// "Connect failed" error.  Zero means infinite retries.
	Retries int

	// Watchdog ticker duration.  On tick, the watchdog will check for
	// downed connection and try to recover the connection.  Default is
	// 0secs, which means no watchdog.  Set to non-zero to enable
	// watchodog.
	WatchdogTimeo time.Duration
}

type rtl8720dn struct {
	cfg      *Config
	notifyCb func(netdev.Event)
	mu       sync.Mutex

	uart *machine.UART
	seq  uint64

	debug bool

	netConnected bool
	driverShown  bool
	deviceShown  bool

	killWatchdog chan bool

	// keyed by sock as returned by rpc_lwip_socket()
	sockets map[sock]*socket
}

func newSocket(protocol netdev.Protocol) *socket {
	return &socket{protocol: protocol, inuse: true}
}

func New(cfg *Config) *rtl8720dn {
	return &rtl8720dn{
		debug:        (_debug & debugRpc) != 0,
		cfg:          cfg,
		sockets:      make(map[sock]*socket),
		killWatchdog: make(chan bool),
	}
}

func (r *rtl8720dn) startDhcpc() error {
	if result := r.rpc_tcpip_adapter_dhcpc_start(0); result == -1 {
		return netdev.ErrStartingDHCPClient
	}
	return nil
}

func (r *rtl8720dn) connectToAP() error {

	if len(r.cfg.Ssid) == 0 {
		return netdev.ErrMissingSSID
	}

	if debugging(debugBasic) {
		fmt.Printf("Connecting to Wifi SSID '%s'...", r.cfg.Ssid)
	}

	// Start the connection process
	securityType := uint32(0x00400004)
	result := r.rpc_wifi_connect(r.cfg.Ssid, r.cfg.Passphrase, securityType, -1, 0)
	if result == -1 {
		if debugging(debugBasic) {
			fmt.Printf("FAILED\r\n")
		}
		return netdev.ErrConnectFailed
	}

	if debugging(debugBasic) {
		fmt.Printf("CONNECTED\r\n")
	}

	if r.notifyCb != nil {
		r.notifyCb(netdev.EventNetUp)
	}

	return r.startDhcpc()
}

func (r *rtl8720dn) showDriver() {
	if r.driverShown {
		return
	}
	if debugging(debugBasic) {
		fmt.Printf("\r\n")
		fmt.Printf("%s\r\n\r\n", driverName)
		fmt.Printf("Driver version           : %s\r\n", version)
	}
	r.driverShown = true
}

func (r *rtl8720dn) initWifi() error {
	if result := r.rpc_tcpip_adapter_init(); result == -1 {
		return fmt.Errorf("TCP/IP adapter init failed")
	}
	if result := r.rpc_wifi_off(); result == -1 {
		return errors.New("Error turning off WiFi")
	}
	if result := r.rpc_wifi_on(RTW_MODE_STA); result == -1 {
		return errors.New("Error turning on WiFi")
	}
	if result := r.rpc_wifi_disconnect(); result == -1 {
		return errors.New("Error disconnecting WiFi")
	}
	return nil
}

func (r *rtl8720dn) setupUART() {
	r.uart = r.cfg.Uart
	r.uart.Configure(machine.UARTConfig{TX: r.cfg.Tx,
		RX: r.cfg.Rx, BaudRate: r.cfg.Baudrate})
}

func (r *rtl8720dn) start() error {
	en := r.cfg.En
	if en == 0 {
		return fmt.Errorf("Must set Config.En")
	}
	en.Configure(machine.PinConfig{Mode: machine.PinOutput})
	en.Low()
	time.Sleep(100 * time.Millisecond)
	en.High()
	time.Sleep(1000 * time.Millisecond)
	r.setupUART()
	return r.initWifi()
}

func (r *rtl8720dn) stop() {
	r.rpc_tcpip_adapter_stop(0)
	r.cfg.En.Low()
}

func (r *rtl8720dn) showDevice() {
	if r.deviceShown {
		return
	}
	if debugging(debugBasic) {
		fmt.Printf("RTL8720 firmware version : %s\r\n", r.getFwVersion())
		fmt.Printf("MAC address              : %s\r\n", r.getMACAddr())
		fmt.Printf("\r\n")
	}
	r.deviceShown = true
}

func (r *rtl8720dn) showIP() {
	if debugging(debugBasic) {
		ip, subnet, gateway, _ := r.getIP()
		fmt.Printf("\r\n")
		fmt.Printf("DHCP-assigned IP         : %s\r\n", netdev.IP(ip).String())
		fmt.Printf("DHCP-assigned subnet     : %s\r\n", netdev.IP(subnet).String())
		fmt.Printf("DHCP-assigned gateway    : %s\r\n", netdev.IP(gateway).String())
		fmt.Printf("\r\n")
	}
}

func (r *rtl8720dn) networkDown() bool {
	result := r.rpc_wifi_is_connected_to_ap()
	return result != 0
}

func (r *rtl8720dn) watchdog() {
	ticker := time.NewTicker(r.cfg.WatchdogTimeo)
	for {
		select {
		case <-r.killWatchdog:
			return
		case <-ticker.C:
			r.mu.Lock()
			if r.networkDown() {
				if debugging(debugBasic) {
					fmt.Printf("Watchdog: Wifi NOT CONNECTED, trying again...\r\n")
				}
				if r.notifyCb != nil {
					r.notifyCb(netdev.EventNetDown)
				}
				r.netConnect(false)
			}
			r.mu.Unlock()
		}
	}
}

func (r *rtl8720dn) netConnect(reset bool) error {
	if reset {
		if err := r.start(); err != nil {
			return err
		}
	}
	r.showDevice()

	for i := 0; r.cfg.Retries == 0 || i < r.cfg.Retries; i++ {
		if err := r.connectToAP(); err != nil {
			if err == netdev.ErrConnectFailed {
				continue
			}
			return err
		}
		break
	}

	if r.networkDown() {
		return netdev.ErrConnectFailed
	}

	r.showIP()
	return nil
}

func (r *rtl8720dn) NetConnect() error {

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.netConnected {
		return netdev.ErrConnected
	}

	r.showDriver()

	if err := r.netConnect(true); err != nil {
		return err
	}

	r.netConnected = true

	if r.cfg.WatchdogTimeo != 0 {
		go r.watchdog()
	}

	return nil
}

func (r *rtl8720dn) netDisconnect() {
	r.disconnect()
}

func (r *rtl8720dn) NetDisconnect() {

	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.netConnected {
		return
	}

	if r.cfg.WatchdogTimeo != 0 {
		r.killWatchdog <- true
	}
	r.netDisconnect()
	r.stop()

	r.netConnected = false

	if debugging(debugBasic) {
		fmt.Printf("\r\nDisconnected from Wifi SSID '%s'\r\n\r\n", r.cfg.Ssid)
	}

	if r.notifyCb != nil {
		r.notifyCb(netdev.EventNetDown)
	}
}

func (r *rtl8720dn) NetNotify(cb func(netdev.Event)) {
	r.notifyCb = cb
}

func (r *rtl8720dn) GetHostByName(name string) (netdev.IP, error) {

	if debugging(debugNetdev) {
		fmt.Printf("[GetHostByName] name: %s\r\n", name)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	var addr [4]byte
	result := r.rpc_netconn_gethostbyname(name, addr[:])
	if result == -1 {
		return netdev.IP{}, fmt.Errorf("Get IP of host '%s' failed", name)
	}

	var ip netdev.IP
	copy(ip[:], addr[:])

	return ip, nil
}

func (r *rtl8720dn) GetHardwareAddr() (netdev.HardwareAddr, error) {

	if debugging(debugNetdev) {
		fmt.Printf("[GetHardwareAddr]\r\n")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	mac := strings.ReplaceAll(r.getMACAddr(), ":", "")
	addr, err := hex.DecodeString(mac)

	return netdev.HardwareAddr(addr), err
}

func (r *rtl8720dn) GetIPAddr() (netdev.IP, error) {

	if debugging(debugNetdev) {
		fmt.Printf("[GetIPAddr]\r\n")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	ip, _, _, err := r.getIP()

	return netdev.IP(ip), err
}

func (r *rtl8720dn) clientTLS() uint32 {
	client := r.rpc_wifi_ssl_client_create()
	r.rpc_wifi_ssl_init(client)
	r.rpc_wifi_ssl_set_timeout(client, 120*1000 /* usec? */)
	return client
}

// See man socket(2) for standard Berkely sockets for Socket, Bind, etc.
// The driver strives to meet the function and semantics of socket(2).

func (r *rtl8720dn) Socket(family netdev.AddressFamily, sockType netdev.SockType,
	protocol netdev.Protocol) (netdev.Sockfd, error) {

	if debugging(debugNetdev) {
		fmt.Printf("[Socket] family: %s, sockType: %s, protocol: %s\r\n",
			family, sockType, protocol)
	}

	switch family {
	case netdev.AF_INET:
	default:
		return -1, netdev.ErrFamilyNotSupported
	}

	var newSock int32

	r.mu.Lock()
	defer r.mu.Unlock()

	switch {
	case protocol == netdev.IPPROTO_TCP && sockType == netdev.SOCK_STREAM:
		newSock = r.rpc_lwip_socket(netdev.AF_INET, netdev.SOCK_STREAM,
			netdev.IPPROTO_TCP)
	case protocol == netdev.IPPROTO_TLS && sockType == netdev.SOCK_STREAM:
		// TODO Investigate: using client number as socket number;
		// TODO this may cause a problem if mixing TLS and non-TLS sockets?
		newSock = int32(r.clientTLS())
	case protocol == netdev.IPPROTO_UDP && sockType == netdev.SOCK_DGRAM:
		newSock = r.rpc_lwip_socket(netdev.AF_INET, netdev.SOCK_DGRAM,
			netdev.IPPROTO_UDP)
	default:
		return -1, netdev.ErrProtocolNotSupported
	}

	if newSock == -1 {
		return -1, netdev.ErrNoMoreSockets
	}

	socket := newSocket(protocol)
	r.sockets[sock(newSock)] = socket

	return netdev.Sockfd(newSock), nil
}

func addrToName(addr netdev.SockAddr) []byte {
	port := addr.Port()
	ip := addr.IpBytes()

	name := make([]byte, 16)
	name[0] = 0x00
	name[1] = netdev.AF_INET
	name[2] = byte(port >> 8)
	name[3] = byte(port)
	name[4] = byte(ip[0])
	name[5] = byte(ip[1])
	name[6] = byte(ip[2])
	name[7] = byte(ip[3])

	return name
}

func (r *rtl8720dn) Bind(sockfd netdev.Sockfd, addr netdev.SockAddr) error {

	if debugging(debugNetdev) {
		fmt.Printf("[Bind] sockfd: %d, addr: %s\r\n", sockfd, addr)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	var sock = sock(sockfd)
	var socket = r.sockets[sock]
	var name = addrToName(addr)

	switch socket.protocol {
	case netdev.IPPROTO_TCP, netdev.IPPROTO_UDP:
		result := r.rpc_lwip_bind(int32(sock), name, uint32(len(name)))
		if result == -1 {
			return fmt.Errorf("Bind to %s failed", addr.String())
		}
	default:
		return netdev.ErrProtocolNotSupported
	}

	return nil
}

func (r *rtl8720dn) Connect(sockfd netdev.Sockfd, servaddr netdev.SockAddr) error {

	if debugging(debugNetdev) {
		fmt.Printf("[Connect] sockfd: %d, servaddr: %s\r\n", sockfd, servaddr)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	var sock = sock(sockfd)
	var socket = r.sockets[sock]
	var name = addrToName(servaddr)

	// Start the connection
	switch socket.protocol {
	case netdev.IPPROTO_TCP, netdev.IPPROTO_UDP:
		result := r.rpc_lwip_connect(int32(sock), name, uint32(len(name)))
		if result == -1 {
			return fmt.Errorf("Connect to %s failed", servaddr.String())
		}
	case netdev.IPPROTO_TLS:
		result := r.rpc_wifi_start_ssl_client(uint32(sock),
			servaddr.Host(), uint32(servaddr.Port()), 0)
		if result == -1 {
			return fmt.Errorf("Connect to %s failed", servaddr.String())
		}
	}

	return nil
}

func (r *rtl8720dn) Listen(sockfd netdev.Sockfd, backlog int) error {

	if debugging(debugNetdev) {
		fmt.Printf("[Listen] sockfd: %d\r\n", sockfd)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	var sock = sock(sockfd)
	var socket = r.sockets[sock]

	switch socket.protocol {
	case netdev.IPPROTO_TCP:
		result := r.rpc_lwip_listen(int32(sock), int32(backlog))
		if result == -1 {
			return fmt.Errorf("Listen failed")
		}
		result = r.rpc_lwip_fcntl(int32(sock), F_SETFL, O_NONBLOCK)
		if result == -1 {
			return fmt.Errorf("Fcntl failed")
		}
	case netdev.IPPROTO_UDP:
		result := r.rpc_lwip_listen(int32(sock), int32(backlog))
		if result == -1 {
			return fmt.Errorf("Listen failed")
		}
	default:
		return netdev.ErrProtocolNotSupported
	}

	return nil
}

func (r *rtl8720dn) Accept(sockfd netdev.Sockfd, peer netdev.SockAddr) (netdev.Sockfd, error) {

	if debugging(debugNetdev) {
		fmt.Printf("[Accept] sockfd: %d, peer: %s\r\n", sockfd, peer)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	var newSock int32
	var lsock = sock(sockfd)
	var socket = r.sockets[lsock]
	var addr = addrToName(peer)

	switch socket.protocol {
	case netdev.IPPROTO_TCP:
	default:
		return -1, netdev.ErrProtocolNotSupported
	}

	for {
		// Accept() will be sleeping most of the time, checking for a
		// new clients every 1/10 sec.
		r.mu.Unlock()
		time.Sleep(100 * time.Millisecond)
		r.mu.Lock()

		// Check if a client connected.  O_NONBLOCK is set on lsock.
		addrlen := uint32(len(addr))
		newSock = r.rpc_lwip_accept(int32(lsock), addr, &addrlen)
		if newSock == -1 {
			// No new client
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// If we've already seen this socket, we can re-use
		// the socket and return it.  But, only if the socket
		// is closed.  If it's not closed, we'll just come back
		// later to reuse it.

		clientSocket, ok := r.sockets[sock(newSock)]
		if ok {
			// Wait for client to Close
			if clientSocket.inuse {
				continue
			}
			// Reuse client socket
			return netdev.Sockfd(newSock), nil
		}

		// Create new socket for client and return fd
		r.sockets[sock(newSock)] = newSocket(socket.protocol)
		return netdev.Sockfd(newSock), nil
	}
}

func (r *rtl8720dn) sendChunk(sockfd netdev.Sockfd, buf []byte) (int, error) {
	var sock = sock(sockfd)
	var socket = r.sockets[sock]

	switch socket.protocol {
	case netdev.IPPROTO_TCP, netdev.IPPROTO_UDP:
		result := r.rpc_lwip_send(int32(sock), buf, 0x00000008)
		if result == -1 {
			return -1, fmt.Errorf("Send error")
		}
		return int(result), nil
	case netdev.IPPROTO_TLS:
		result := r.rpc_wifi_send_ssl_data(uint32(sock), buf, uint16(len(buf)))
		if result == -1 {
			return -1, fmt.Errorf("TLS Send error")
		}
		return int(result), nil
	}

	return -1, netdev.ErrProtocolNotSupported
}

func (r *rtl8720dn) Send(sockfd netdev.Sockfd, buf []byte, flags netdev.SockFlags,
	timeout time.Duration) (int, error) {

	if debugging(debugNetdev) {
		fmt.Printf("[Send] sockfd: %d, len(buf): %d, flags: %d\r\n",
			sockfd, len(buf), flags)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Break large bufs into chunks

	// TODO handle timeout

	chunkSize := 1436
	for i := 0; i < len(buf); i += chunkSize {
		end := i + chunkSize
		if end > len(buf) {
			end = len(buf)
		}
		_, err := r.sendChunk(sockfd, buf[i:end])
		if err != nil {
			return -1, err
		}
	}

	return len(buf), nil
}

func (r *rtl8720dn) Recv(sockfd netdev.Sockfd, buf []byte, flags netdev.SockFlags,
	timeout time.Duration) (int, error) {

	if debugging(debugNetdev) {
		fmt.Printf("[Recv] sockfd: %d, len(buf): %d, flags: %d\r\n",
			sockfd, len(buf), flags)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	var sock = sock(sockfd)
	var socket = r.sockets[sock]
	var length = len(buf)
	var expire = time.Now().Add(timeout)
	var n int32

	// Limit length read size to chunk large read requests
	if length > 1436 {
		length = 1436
	}

	for {
		// Check if we've timed out
		if timeout > 0 {
			if time.Now().Before(expire) {
				return -1, netdev.ErrRecvTimeout
			}
		}

		switch socket.protocol {
		case netdev.IPPROTO_TCP, netdev.IPPROTO_UDP:
			n = r.rpc_lwip_recv(int32(sock), buf[:length],
				uint32(length), 0x00000008, 0)
		case netdev.IPPROTO_TLS:
			n = r.rpc_wifi_get_ssl_receive(uint32(sock),
				buf[:length], int32(length))
		}

		if n < 0 {
			r.mu.Unlock()
			time.Sleep(100 * time.Millisecond)
			r.mu.Lock()
			continue
		} else if n == 0 {
			if debugging(debugNetdev) {
				fmt.Printf("[<--Recv] sockfd: %d, n: %d EOF\r\n",
					sock, n)
			}
			return -1, io.EOF
		}

		if debugging(debugNetdev) {
			fmt.Printf("[<--Recv] sockfd: %d, n: %d\r\n",
				sock, n)
		}

		return int(n), nil
	}
}

func (r *rtl8720dn) Close(sockfd netdev.Sockfd) error {

	if debugging(debugNetdev) {
		fmt.Printf("[Close] sockfd: %d\r\n", sockfd)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	var sock = sock(sockfd)
	var socket = r.sockets[sock]
	var result int32

	if !socket.inuse {
		return nil
	}

	switch socket.protocol {
	case netdev.IPPROTO_TCP, netdev.IPPROTO_UDP:
		result = r.rpc_lwip_close(int32(sock))
	case netdev.IPPROTO_TLS:
		r.rpc_wifi_stop_ssl_socket(uint32(sock))
		r.rpc_wifi_ssl_client_destroy(uint32(sock))
	}

	if result == -1 {
		return netdev.ErrClosingSocket
	}

	socket.inuse = false

	return nil
}

func (r *rtl8720dn) SetSockOpt(sockfd netdev.Sockfd, level netdev.SockOptLevel,
	opt netdev.SockOpt, value any) error {

	if debugging(debugNetdev) {
		fmt.Printf("[SetSockOpt] sockfd: %d\r\n", sockfd)
	}

	return netdev.ErrNotSupported
}

func (r *rtl8720dn) disconnect() error {
	result := r.rpc_wifi_disconnect()
	if result == -1 {
		return fmt.Errorf("Error disconnecting Wifi")
	}
	return nil
}

func (r *rtl8720dn) getFwVersion() string {
	return r.rpc_system_version()
}

func (r *rtl8720dn) getMACAddr() string {
	var mac [18]uint8
	r.rpc_wifi_get_mac_address(mac[:])
	return string(mac[:])
}

func (r *rtl8720dn) getIP() (ip, subnet, gateway netdev.IP, err error) {
	var ip_info [12]byte
	result := r.rpc_tcpip_adapter_get_ip_info(0, ip_info[:])
	if result == -1 {
		err = fmt.Errorf("Get IP info failed")
		return
	}
	copy(ip[:], ip_info[0:4])
	copy(subnet[:], ip_info[4:8])
	copy(gateway[:], ip_info[8:12])
	return
}
