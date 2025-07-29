// Package comboat implements WiFi driver for the Aithinker-Combo-AT WiFi
// device found on the Elecrow W5 rp2040 and rp2350 devices.  Ths WiFi device
// is a RTL8720d variant.  The driver interface is via AT command set over UART
// (see reference docs below).
//
// NOTE: the driver doesn't support UDP/TCP server connections in STA mode,
// currently.  UDP/TCP/TLS client connections are supported in STA mode.
//
// https://aithinker-combo-guide.readthedocs.io/en/latest/docs/instruction/index.html
// https://aithinker-combo-guide.readthedocs.io/en/latest/docs/command-set/index.html
// https://aithinker-combo-guide.readthedocs.io/en/latest/docs/command-examples/index.html

package comboat // import "tinygo.org/x/drivers/comboat"

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"machine"
	"net"
	"net/netip"
	"strconv"
	"sync"
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/netdev"
	"tinygo.org/x/drivers/netlink"
)

type Config struct {
	BaudRate uint32
	Uart     *machine.UART
	Tx       machine.Pin
	Rx       machine.Pin
}

type socket struct {
	protocol  int
	id        string
	rx        chan []byte
	remainder []byte
	laddr     netip.AddrPort // Set in Bind()
}

type device struct {
	cfg     *Config
	uart    *machine.UART
	uartMu  sync.Mutex
	mac     net.HardwareAddr
	ip      netip.Addr
	gateway netip.Addr
	buf     [1500]byte
	pos     int
	last    []byte
	ok      chan bool
	txReady chan bool
	accept  chan string
	err     chan error
	sockets [8]*socket
	sync.Mutex
}

func NewDevice(cfg *Config) *device {
	return &device{
		cfg:     cfg,
		ok:      make(chan bool),
		txReady: make(chan bool),
		accept:  make(chan string),
		err:     make(chan error),
	}
}

func logDebug(msg string) {
	//println("[DEBUG] " + msg)
}

func logError(msg string) {
	println("[ERROR] " + msg)
}

func split(resp []byte, part int, del, on string) string {
	parts := bytes.Split(resp, []byte(del))
	if part >= len(parts) {
		return "Split parts error getting " + on
	}
	return string(parts[part])
}

func (d *device) getFWVersion() string {
	return split(d.last, 1, ":", "FW version")
}

func (d *device) saveMAC() {
	raw := split(d.last, 1, ":", "MAC")
	if len(raw) > 11 {
		macStr := fmt.Sprintf("%s:%s:%s:%s:%s:%s",
			raw[0:2], raw[2:4], raw[4:6],
			raw[6:8], raw[8:10], raw[10:12])
		d.mac, _ = net.ParseMAC(macStr)
	}
}

var countryCodes = map[int]string{
	1:  "JP Japan",
	2:  "American Samoa",
	3:  "CA Canada",
	4:  "US",
	5:  "CN China",
	6:  "Hong Kong, China",
	7:  "Taiwan, China",
	8:  "MO Macau, China",
	9:  "IL Israel",
	10: "Singapore",
	11: "KR South Korea",
	12: "TR Türkiye",
	13: "AU Australia",
	14: "ZA South Africa",
	15: "BR Brazil",
}

func (d *device) getCountry() (code string) {
	code = split(d.last, 1, ":", "county code")
	codeNum, err := strconv.Atoi(code)
	if err != nil {
		return
	}
	if val, ok := countryCodes[codeNum]; ok {
		code = val
	}
	return
}

func (d *device) saveIP() {
	ipStr := split(d.last, 7, ",", "IP address")
	gwStr := split(d.last, 8, ",", "gateway address")
	d.ip, _ = netip.ParseAddr(ipStr)
	d.gateway, _ = netip.ParseAddr(gwStr)
}

func (d *device) execute(cmd string, timeout int) (err error) {
	logDebug("EXECUTE " + cmd)

	d.uartMu.Lock()
	_, err = d.uart.Write([]byte(cmd + "\r\n"))
	d.uartMu.Unlock()

	if err != nil {
		return
	}

	t := time.NewTicker(time.Duration(timeout) * time.Millisecond)
	defer t.Stop()

	select {
	case <-t.C:
		return errors.New("Timed out")
	case <-d.ok:
		return
	case err = <-d.err:
		return
	}
}

func (d *device) send(cmd string, timeout int) (err error) {
	logDebug("EXECUTE " + cmd)

	d.uartMu.Lock()
	_, err = d.uart.Write([]byte(cmd + "\r\n"))
	d.uartMu.Unlock()

	if err != nil {
		return
	}

	t := time.NewTicker(time.Duration(timeout) * time.Millisecond)
	defer t.Stop()

	select {
	case <-t.C:
		return errors.New("Timed out")
	case <-d.txReady:
		return
	case err = <-d.err:
		return
	}
}

func (d *device) findSocket(id string) (*socket, error) {
	for _, s := range d.sockets {
		if s.id == id {
			return s, nil
		}
	}
	return nil, errors.New("Socket not found with id: " + id)
}

func (d *device) getSocket(sockfd int) (*socket, error) {
	if sockfd < 0 || sockfd+1 > len(d.sockets) {
		return nil, netdev.ErrInvalidSocketFd
	}
	if d.sockets[sockfd] == nil {
		return nil, netdev.ErrInvalidSocketFd
	}
	return d.sockets[sockfd], nil
}

func (d *device) handle(event []byte) {
	logDebug("GOT EVENT " + string(event))
	switch {

	// SocketDisconnect,<id>
	case bytes.HasPrefix(event, []byte("SocketDisconnect")):
		id := split(event, 1, ",", "SocketDisconnect")
		s, err := d.findSocket(id)
		if err == nil {
			close(s.rx) // Sends io.EOF
		}

	// SocketSeed,<id>,<server id>
	case bytes.HasPrefix(event, []byte("SocketSeed,2,1")):
		//d.uart.Write([]byte("AT+SOCKET?" + "\r\n"))
	}
}

func (d *device) processUART() {

	if d.pos == 1 && d.buf[0] == '>' {
		d.pos = 0
		logDebug("GOT >")
		d.txReady <- true
	}

	sofar := d.buf[:d.pos]

	if !bytes.HasSuffix(sofar, []byte("\r\n")) {
		return
	}

	// Strip CR/LF off end
	sofar = sofar[:len(sofar)-2]

	switch {

	case bytes.HasPrefix(sofar, []byte("+EVENT:SocketDown")):
		// +EVENT:SocketDown,<id>,<length>,<data>
		parts := bytes.SplitN(sofar, []byte(","), 4)
		if len(parts) != 4 {
			logError("Error parsing +EVENT:SocketDown: " + string(sofar))
			d.pos = 0
			return
		}
		id := string(parts[1])
		length, err := strconv.Atoi(string(parts[2]))
		if err != nil {
			logError("Error parsing length from: " + string(parts[2]))
			d.pos = 0
			return
		}
		if length != len(parts[3]) {
			// This can happen if <data> actually contains a CR/LF.
			// Return without resetting d.pos to continue reading
			// in the full <data>.
			return
		}
		s, err := d.findSocket(id)
		if err != nil {
			logError(err.Error())
			d.pos = 0
			return
		}
		logDebug("GOT +EVENT:SocketDown," + id + "," + string(parts[2]))
		d.pos = 0
		data := make([]byte, len(parts[3]))
		copy(data, parts[3])
		s.rx <- data

	case bytes.HasPrefix(sofar, []byte("OK")):
		d.pos = 0
		logDebug("GOT OK")
		d.ok <- true

	case bytes.HasPrefix(sofar, []byte("ERROR")):
		d.pos = 0
		logDebug("GOT ERROR")
		errStr := getErrStr(d.last)
		d.err <- errors.New(errStr)

	case bytes.HasPrefix(sofar, []byte("+EVENT:")):
		d.pos = 0
		event := sofar[len("+EVENT:"):]
		d.handle(event)

	default:
		// Catch everything else and store in d.last
		d.pos = 0
		size := len(sofar)
		if size > 0 {
			d.last = make([]byte, size)
			copy(d.last, sofar[:size])
			logDebug("GOT LINE " + string(d.last))
		}
	}
}

func (d *device) serviceUART() {
	for {
		d.uartMu.Lock()
		for d.uart.Buffered() > 0 {
			if d.pos >= len(d.buf) {
				println("Trying to write past buffer")
				d.pos = 0
				break
			}
			var err error
			d.buf[d.pos], err = d.uart.ReadByte()
			if err == nil {
				d.pos++
				d.processUART()
			}
		}
		d.uartMu.Unlock()
		time.Sleep(10 * time.Millisecond)
	}
}

func (d *device) NetConnect(params *netlink.ConnectParams) error {

	d.Lock()
	defer d.Unlock()

	d.uart = d.cfg.Uart
	d.uart.Configure(machine.UARTConfig{
		BaudRate: d.cfg.BaudRate,
		TX:       d.cfg.Tx,
		RX:       d.cfg.Rx,
	})

	go d.serviceUART()

	fmt.Printf("\r\n")
	fmt.Printf("TinyGo Combo-AT WiFi network device driver\r\n")

	fmt.Printf("\r\n")
	fmt.Printf("Driver version            : %s\r\n", drivers.Version)

	if len(params.Ssid) == 0 {
		return netlink.ErrMissingSSID
	}

	// AT Test to see if device is alive
	if err := d.execute("AT", 1000); err != nil {
		return err
	}

	// Disable echo
	if err := d.execute("ATE0", 1000); err != nil {
		return err
	}

	// Get FW version
	if err := d.execute("AT+GMR", 1000); err != nil {
		return err
	}
	fmt.Printf("Combo-AT firmware version : %s\r\n", d.getFWVersion())

	// Get/save MAC addresses
	if err := d.execute("AT+CIPSTAMAC_DEF?", 1000); err != nil {
		return err
	}
	d.saveMAC()
	fmt.Printf("MAC address               : %s\r\n", d.mac.String())

	// Set country code US
	if err := d.execute("AT+WCOUNTRY=4", 1000); err != nil {
		return err
	}

	// Get country code
	if err := d.execute("AT+WCOUNTRY?", 1000); err != nil {
		return err
	}
	fmt.Printf("WiFi country code         : %s\r\n", d.getCountry())

	// Set Wi-Fi working mode to STA and save to flash
	if err := d.execute("AT+WMODE=1,1", 1000); err != nil {
		return err
	}

	// Connect to Wifi AP (keep trying until connected)
	fmt.Printf("\r\n")
	cmd := "AT+WJAP=" + params.Ssid + "," + params.Passphrase

	for {
		fmt.Printf("Connecting to WiFi SSID '%s'...", params.Ssid)
		if err := d.execute(cmd, 20000); err != nil {
			fmt.Printf("FAILED (%s)\r\n", err.Error())
			continue
		}
		break
	}

	fmt.Printf("CONNECTED\r\n")

	// Automatically reconnect to Wi-Fi after power on
	if err := d.execute("AT+WAUTOCONN=1", 1000); err != nil {
		return err
	}

	// Get/save IP/gateway addresses
	if err := d.execute("AT+WJAP?", 1000); err != nil {
		return err
	}
	d.saveIP()

	fmt.Printf("\r\n")
	fmt.Printf("DHCP-assigned IP          : %s\r\n", d.ip)
	fmt.Printf("DHCP-assigned gateway     : %s\r\n", d.gateway)
	fmt.Printf("\r\n")

	// Set socket receiving mode to active
	if err := d.execute("AT+SOCKETRECVCFG=1", 1000); err != nil {
		return err
	}

	return nil
}

func (d *device) NetDisconnect() {
	d.Lock()
	defer d.Unlock()
	// Disconnect from WiFi AP
	d.execute("AT+WDISCONNECT", 1000)
}

func (d *device) NetNotify(cb func(netlink.Event)) {
	fmt.Printf("\r\n%s\r\n", netlink.ErrNotSupported)
}

func (d *device) GetHardwareAddr() (net.HardwareAddr, error) {
	return d.mac, nil
}

func (d *device) _getHostByName(name string) (ip netip.Addr, err error) {
	if err = d.execute("AT+WDOMAIN="+name, 1000); err != nil {
		return
	}
	ipStr := split(d.last, 1, ":", "host by name")
	return netip.ParseAddr(ipStr)
}

func (d *device) GetHostByName(name string) (ip netip.Addr, err error) {

	// If it's already a dotted-network address, and not a host name,
	// return it
	ip, err = netip.ParseAddr(name)
	if err == nil {
		return
	}

	d.Lock()
	defer d.Unlock()

	return d._getHostByName(name)
}

func (d *device) Addr() (netip.Addr, error) {
	return d.ip, nil
}

func (d *device) Socket(domain, stype, protocol int) (int, error) {

	switch domain {
	case netdev.AF_INET:
	default:
		return -1, netdev.ErrFamilyNotSupported
	}

	switch {
	case protocol == netdev.IPPROTO_TCP && stype == netdev.SOCK_STREAM:
	case protocol == netdev.IPPROTO_TLS && stype == netdev.SOCK_STREAM:
	case protocol == netdev.IPPROTO_UDP && stype == netdev.SOCK_DGRAM:
	default:
		return -1, netdev.ErrProtocolNotSupported
	}

	d.Lock()
	defer d.Unlock()

	// Search for empty slot in sockets array
	for fd, s := range d.sockets {
		if s == nil {
			// Found one
			d.sockets[fd] = &socket{
				protocol: protocol,
				rx:       make(chan []byte, 10),
			}
			return fd, nil
		}
	}

	return -1, netdev.ErrNoMoreSockets
}

func (d *device) Bind(sockfd int, ip netip.AddrPort) error {

	d.Lock()
	defer d.Unlock()

	s, err := d.getSocket(sockfd)
	if err != nil {
		return err
	}

	s.laddr = ip
	return nil
}

func (d *device) Connect(sockfd int, host string, ip netip.AddrPort) error {

	var addr string
	var cmd string

	d.Lock()
	defer d.Unlock()

	s, err := d.getSocket(sockfd)
	if err != nil {
		return err
	}

	if host == "" {
		addr = ip.Addr().String()
	} else {
		ip, err := d._getHostByName(host)
		if err != nil {
			return err
		}
		addr = ip.String()
	}
	port := strconv.Itoa(int(ip.Port()))

	switch s.protocol {
	case netdev.IPPROTO_UDP:
		cmd = "AT+SOCKET=2," + addr + "," + port
	case netdev.IPPROTO_TCP:
		cmd = "AT+SOCKET=4," + addr + "," + port
	case netdev.IPPROTO_TLS:
		cmd = "AT+SOCKET=7," + addr + "," + port
	}

	if cmd == "" {
		return netdev.ErrProtocolNotSupported
	}

	if err := d.execute(cmd, 20000); err != nil {
		return err
	}

	s.id = split(d.last, 1, "=", "connection ID")

	return nil
}

func (d *device) Listen(sockfd, backlog int) error {

	// TODO Creating a TCP server socket isn't working when in STA mode,
	// TODO returning error "Socket bind error".
	// TODO The reference example shows a TCP server example in AP mode.

	/*
		var cmd string

		d.Lock()
		defer d.Unlock()

		s, err := d.getSocket(sockfd)
		if err != nil {
			return err
		}

		port := strconv.Itoa(int(s.laddr.Port()))

		switch s.protocol {
		case netdev.IPPROTO_UDP:
			cmd = "AT+SOCKET=1," + port
		case netdev.IPPROTO_TCP:
			cmd = "AT+SOCKET=3," + port
		}

		if cmd == "" {
			return netdev.ErrProtocolNotSupported
		}

		if err := d.execute(cmd, 20000); err != nil {
			return err
		}

		s.id = split(d.last, 1, "=", "connection ID")
	*/

	return netdev.ErrNotSupported
}

func (d *device) Accept(sockfd int) (int, netip.AddrPort, error) {
	return 0, netip.AddrPort{}, netdev.ErrNotSupported
}

func (d *device) Send(sockfd int, buf []byte, flags int, deadline time.Time) (int, error) {

	d.Lock()
	defer d.Unlock()

	s, err := d.getSocket(sockfd)
	if err != nil {
		return 0, err
	}

	cmd := fmt.Sprintf("AT+SOCKETSEND=%s,%d", s.id, len(buf))

	if err := d.send(cmd, 1000); err != nil {
		return 0, err
	}

	// AT+SOCKETSEND will sub-packet send data into 1024-byte chunks,
	// automatically, so send the full buffer in one shot, even if it's
	// bigger than 1024 bytes.

	d.uartMu.Lock()
	n, err := d.uart.Write(buf)
	d.uartMu.Unlock()

	if err != nil {
		return 0, err
	}

	// Expecting "OK" after good send, or "ERROR"

	t := time.NewTicker(time.Duration(1000) * time.Millisecond)
	defer t.Stop()

	select {
	case <-t.C:
		return 0, errors.New("Timed out")
	case <-d.ok:
		return n, nil
	case err = <-d.err:
		return 0, err
	}
}

func (d *device) Recv(sockfd int, buf []byte, flags int, deadline time.Time) (int, error) {

	d.Lock()
	defer d.Unlock()

	s, err := d.getSocket(sockfd)
	if err != nil {
		return 0, err
	}

	// 1. Use leftover data first
	if len(s.remainder) > 0 {
		n := copy(buf, s.remainder)
		s.remainder = s.remainder[n:]
		return n, nil
	}

	// 2. Get new data from the channel
	data, ok := <-s.rx
	if !ok {
		// Socket closed, return EOF
		return 0, io.EOF
	}

	// 3. Copy data, handle leftovers
	n := copy(buf, data)
	if n < len(data) {
		s.remainder = data[n:]
	}

	return n, nil
}

func (d *device) Close(sockfd int) error {

	d.Lock()
	defer d.Unlock()

	s, err := d.getSocket(sockfd)
	if err != nil {
		return err
	}

	// Delete socket only if connection was successful (s.id is set)
	if s.id != "" {
		cmd := fmt.Sprintf("AT+SOCKETDEL=%s", s.id)
		if err = d.execute(cmd, 1000); err != nil {
			return err
		}
	}

	d.sockets[sockfd] = nil

	return nil
}

func (d *device) SetSockOpt(sockfd, level, opt int, value interface{}) error {
	return netdev.ErrNotSupported
}
