// Package espat implements TCP/UDP wireless communication over serial
// with a separate ESP8266 or ESP32 board using the Espressif AT command set
// across a UART interface.
//
// In order to use this driver, the ESP8266/ESP32 must be flashed with firmware
// supporting the AT command set. Many ESP8266/ESP32 chips already have this firmware
// installed by default. You will need to install this firmware if you have an
// ESP8266 that has been flashed with NodeMCU (Lua) or Arduino firmware.
//
// AT Command Core repository:
// https://github.com/espressif/esp32-at
//
// Datasheet:
// https://www.espressif.com/sites/default/files/documentation/0a-esp8266ex_datasheet_en.pdf
//
// AT command set:
// https://www.espressif.com/sites/default/files/documentation/4a-esp8266_at_instruction_set_en.pdf
//
// 02/2023    sfeldma@gmail.com    Heavily modified to use netdev interface

package espat // import "tinygo.org/x/drivers/espat"

import (
	"errors"
	"fmt"
	"machine"
	"net"
	"net/netip"
	"strconv"
	"strings"
	"sync"
	"time"

	"tinygo.org/x/drivers/netdev"
	"tinygo.org/x/drivers/netlink"
)

type Config struct {
	// UART config
	Uart *machine.UART
	Tx   machine.Pin
	Rx   machine.Pin
}

type socket struct {
	inUse    bool
	protocol int
	laddr    netip.AddrPort
}

type Device struct {
	cfg  *Config
	uart *machine.UART
	// command responses that come back from the ESP8266/ESP32
	response []byte
	// data received from a TCP/UDP connection forwarded by the ESP8266/ESP32
	data   []byte
	socket socket
	mu     sync.Mutex
}

func NewDevice(cfg *Config) *Device {
	return &Device{
		cfg:      cfg,
		response: make([]byte, 1500),
		data:     make([]byte, 0, 1500),
	}
}

func (d *Device) NetConnect(params *netlink.ConnectParams) error {

	if len(params.Ssid) == 0 {
		return netlink.ErrMissingSSID
	}

	d.uart = d.cfg.Uart
	d.uart.Configure(machine.UARTConfig{TX: d.cfg.Tx, RX: d.cfg.Rx})

	// Connect to ESP8266/ESP32
	fmt.Printf("Connecting to device...")

	for i := 0; i < 5; i++ {
		if d.Connected() {
			break
		}
		time.Sleep(1 * time.Second)
	}

	if !d.Connected() {
		fmt.Printf("FAILED\r\n")
		return netlink.ErrConnectFailed
	}

	fmt.Printf("CONNECTED\r\n")

	// Connect to Wifi AP
	fmt.Printf("Connecting to Wifi SSID '%s'...", params.Ssid)

	d.SetWifiMode(WifiModeClient)

	err := d.ConnectToAP(params.Ssid, params.Passphrase, 10 /* secs */)
	if err != nil {
		fmt.Printf("FAILED\r\n")
		return err
	}

	fmt.Printf("CONNECTED\r\n")

	ip, err := d.Addr()
	if err != nil {
		return err
	}
	fmt.Printf("DHCP-assigned IP: %s\r\n", ip)
	fmt.Printf("\r\n")

	return nil
}

func (d *Device) NetDisconnect() {
	d.DisconnectFromAP()
	fmt.Printf("\r\nDisconnected from Wifi\r\n\r\n")
}

func (d *Device) NetNotify(cb func(netlink.Event)) {
	// Not supported
}

func (d *Device) GetHostByName(name string) (netip.Addr, error) {
	ip, err := d.GetDNS(name)
	if err != nil {
		return netip.Addr{}, err
	}
	return netip.ParseAddr(ip)
}

func (d *Device) GetHardwareAddr() (net.HardwareAddr, error) {
	return net.HardwareAddr{}, netlink.ErrNotSupported
}

func (d *Device) Addr() (netip.Addr, error) {
	resp, err := d.GetClientIP()
	if err != nil {
		return netip.Addr{}, err
	}
	prefix := "+CIPSTA:ip:"
	for _, line := range strings.Split(resp, "\n") {
		if ok := strings.HasPrefix(line, prefix); ok {
			ip := line[len(prefix)+1 : len(line)-2]
			return netip.ParseAddr(ip)
		}
	}
	return netip.Addr{}, fmt.Errorf("Error getting IP address")
}

func (d *Device) Socket(domain int, stype int, protocol int) (int, error) {

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

	// Only supporting single connection mode, so only one socket at a time
	if d.socket.inUse {
		return -1, netdev.ErrNoMoreSockets
	}
	d.socket.inUse = true
	d.socket.protocol = protocol

	return 0, nil
}

func (d *Device) Bind(sockfd int, ip netip.AddrPort) error {
	d.socket.laddr = ip
	return nil
}

func (d *Device) Connect(sockfd int, host string, ip netip.AddrPort) error {
	var err error
	var addr = ip.Addr().String()
	var rport = strconv.Itoa(int(ip.Port()))
	var lport = strconv.Itoa(int(d.socket.laddr.Port()))

	switch d.socket.protocol {
	case netdev.IPPROTO_TCP:
		err = d.ConnectTCPSocket(addr, rport)
	case netdev.IPPROTO_UDP:
		err = d.ConnectUDPSocket(addr, rport, lport)
	case netdev.IPPROTO_TLS:
		err = d.ConnectSSLSocket(host, rport)
	}

	if err != nil {
		if host == "" {
			return fmt.Errorf("Connect to %s timed out", ip)
		} else {
			return fmt.Errorf("Connect to %s:%d timed out", host, ip.Port())
		}
	}

	return nil
}

func (d *Device) Listen(sockfd int, backlog int) error {
	switch d.socket.protocol {
	case netdev.IPPROTO_UDP:
	default:
		return netdev.ErrProtocolNotSupported
	}
	return nil
}

func (d *Device) Accept(sockfd int) (int, netip.AddrPort, error) {
	return -1, netip.AddrPort{}, netdev.ErrNotSupported
}

func (d *Device) sendChunk(sockfd int, buf []byte, deadline time.Time) (int, error) {
	// Check if we've timed out
	if !deadline.IsZero() {
		if time.Now().After(deadline) {
			return -1, netdev.ErrTimeout
		}
	}
	err := d.StartSocketSend(len(buf))
	if err != nil {
		return -1, err
	}
	n, err := d.Write(buf)
	if err != nil {
		return -1, err
	}
	_, err = d.Response(1000)
	if err != nil {
		return -1, err
	}
	return n, err
}

func (d *Device) Send(sockfd int, buf []byte, flags int, deadline time.Time) (int, error) {

	d.mu.Lock()
	defer d.mu.Unlock()

	// Break large bufs into chunks so we don't overrun the hw queue

	chunkSize := 1436
	for i := 0; i < len(buf); i += chunkSize {
		end := i + chunkSize
		if end > len(buf) {
			end = len(buf)
		}
		_, err := d.sendChunk(sockfd, buf[i:end], deadline)
		if err != nil {
			return -1, err
		}
	}

	return len(buf), nil
}

func (d *Device) Recv(sockfd int, buf []byte, flags int, deadline time.Time) (int, error) {

	d.mu.Lock()
	defer d.mu.Unlock()

	var length = len(buf)

	// Limit length read size to chunk large read requests
	if length > 1436 {
		length = 1436
	}

	for {
		// Check if we've timed out
		if !deadline.IsZero() {
			if time.Now().After(deadline) {
				return -1, netdev.ErrTimeout
			}
		}

		n, err := d.ReadSocket(buf[:length])
		if err != nil {
			return -1, err
		}
		if n == 0 {
			d.mu.Unlock()
			time.Sleep(100 * time.Millisecond)
			d.mu.Lock()
			continue
		}

		return n, nil
	}
}

func (d *Device) Close(sockfd int) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.socket.inUse = false
	return d.DisconnectSocket()
}

func (d *Device) SetSockOpt(sockfd int, level int, opt int, value interface{}) error {
	return netdev.ErrNotSupported
}

// Connected checks if there is communication with the ESP8266/ESP32.
func (d *Device) Connected() bool {
	d.Execute(Test)

	// handle response here, should include "OK"
	_, err := d.Response(1000)
	if err != nil {
		return false
	}
	return true
}

// Write raw bytes to the UART.
func (d *Device) Write(b []byte) (n int, err error) {
	return d.uart.Write(b)
}

// Read raw bytes from the UART.
func (d *Device) Read(b []byte) (n int, err error) {
	return d.uart.Read(b)
}

// how long in milliseconds to pause after sending AT commands
const pause = 300

// Execute sends an AT command to the ESP8266/ESP32.
func (d Device) Execute(cmd string) error {
	_, err := d.Write([]byte("AT" + cmd + "\r\n"))
	return err
}

// Query sends an AT command to the ESP8266/ESP32 that returns the
// current value for some configuration parameter.
func (d Device) Query(cmd string) (string, error) {
	_, err := d.Write([]byte("AT" + cmd + "?\r\n"))
	return "", err
}

// Set sends an AT command with params to the ESP8266/ESP32 for a
// configuration value to be set.
func (d Device) Set(cmd, params string) error {
	_, err := d.Write([]byte("AT" + cmd + "=" + params + "\r\n"))
	return err
}

// Version returns the ESP8266/ESP32 firmware version info.
func (d Device) Version() []byte {
	d.Execute(Version)
	r, err := d.Response(2000)
	if err != nil {
		//return []byte("unknown")
		return []byte(err.Error())
	}
	return r
}

// Echo sets the ESP8266/ESP32 echo setting.
func (d Device) Echo(set bool) {
	if set {
		d.Execute(EchoConfigOn)
	} else {
		d.Execute(EchoConfigOff)
	}
	// TODO: check for success
	d.Response(100)
}

// Reset restarts the ESP8266/ESP32 firmware. Due to how the baud rate changes,
// this messes up communication with the ESP8266/ESP32 module. So make sure you know
// what you are doing when you call this.
func (d Device) Reset() {
	d.Execute(Restart)
	d.Response(100)
}

// ReadSocket returns the data that has already been read in from the responses.
func (d *Device) ReadSocket(b []byte) (n int, err error) {
	// make sure no data in buffer
	d.Response(300)

	count := len(b)
	if len(b) >= len(d.data) {
		// copy it all, then clear socket data
		count = len(d.data)
		copy(b, d.data[:count])
		d.data = d.data[:0]
	} else {
		// copy all we can, then keep the remaining socket data around
		copy(b, d.data[:count])
		copy(d.data, d.data[count:])
		d.data = d.data[:len(d.data)-count]
	}

	return count, nil
}

// Response gets the next response bytes from the ESP8266/ESP32.
// The call will retry for up to timeout milliseconds before returning nothing.
func (d *Device) Response(timeout int) ([]byte, error) {
	// read data
	var size int
	var start, end int
	pause := 100 // pause to wait for 100 ms
	retries := timeout / pause

	for {
		size = d.uart.Buffered()

		if size > 0 {
			end += size
			d.uart.Read(d.response[start:end])

			// if "+IPD" then read socket data
			if strings.Contains(string(d.response[:end]), "+IPD") {
				// handle socket data
				return nil, d.parseIPD(end)
			}

			// if "OK" then the command worked
			if strings.Contains(string(d.response[:end]), "OK") {
				return d.response[start:end], nil
			}

			// if "Error" then the command failed
			if strings.Contains(string(d.response[:end]), "ERROR") {
				return d.response[start:end], errors.New("response error:" + string(d.response[start:end]))
			}

			// if anything else, then keep reading data in?
			start = end
		}

		// wait longer?
		retries--
		if retries == 0 {
			return nil, errors.New("response timeout error:" + string(d.response[start:end]))
		}

		time.Sleep(time.Duration(pause) * time.Millisecond)
	}
}

func (d *Device) parseIPD(end int) error {
	// find the "+IPD," to get length
	s := strings.Index(string(d.response[:end]), "+IPD,")

	// find the ":"
	e := strings.Index(string(d.response[:end]), ":")

	// find the data length
	val := string(d.response[s+5 : e])

	// TODO: verify count
	v, err := strconv.Atoi(val)
	if err != nil {
		// not expected data here. what to do?
		return err
	}

	// load up the socket data
	//d.data = append(d.data, d.response[e+1:end]...)
	d.data = append(d.data, d.response[e+1:e+1+v]...)
	return nil
}

// IsSocketDataAvailable returns of there is socket data available
func (d *Device) IsSocketDataAvailable() bool {
	return len(d.data) > 0 || d.uart.Buffered() > 0
}
