package w5500

import (
	"fmt"
	"net"
	"net/netip"
	"sync"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/netdev"
)

var _ netdev.Netdever = &Device{}

var _debug debug = debugOff

// Resolver is a function that resolves a hostname to an IP address.
type Resolver func(host string) (netip.Addr, error)

// PinOutput is a function that sets a pin high or low.
type PinOutput func(level bool)

// Device is a driver for the W5500 Ethernet controller.
type Device struct {
	maxSockets  int
	maxSockSize int

	mu  sync.Mutex
	bus drivers.SPI
	cs  PinOutput
	dns Resolver

	sockets []*socket
	laddr   netip.Addr

	cmdBuf [3]byte
}

// New returns a new w5500 driver.
func New(bus drivers.SPI, cs PinOutput) *Device {
	return &Device{
		bus: bus,
		cs:  cs,
	}
}

// Config is the configuration for the device.
//
// The SPI bus must be fully configured.
type Config struct {
	DNS Resolver

	MAC        net.HardwareAddr
	IP         netip.Addr
	SubnetMask netip.Addr
	Gateway    netip.Addr

	// Optional, default is 8.
	MaxSockets int
}

// Configure sets up the device.
//
// MAC address must be provided. The other fields are optional.
func (d *Device) Configure(cfg Config) error {
	d.cs(true)

	d.mu.Lock()
	defer d.mu.Unlock()

	d.dns = cfg.DNS

	d.reset()

	if err := d.setupSockets(cfg.MaxSockets); err != nil {
		return fmt.Errorf("could not setup sockets: %w", err)
	}

	// Set the MAC address and IP configuration.
	d.write(regMAC, 0, cfg.MAC)
	d.write(regIPAddr, 0, cfg.IP.AsSlice())
	d.write(regSubnetMask, 0, cfg.SubnetMask.AsSlice())
	d.write(regGatewayAddr, 0, cfg.Gateway.AsSlice())
	d.laddr = cfg.IP

	return nil
}

func (d *Device) setupSockets(maxSockets int) error {
	if maxSockets == 0 {
		maxSockets = 8 // Default to 8 sockets if not specified.
	}
	switch maxSockets {
	case 1, 2, 4, 8:
		// Valid socket counts.
	default:
		return fmt.Errorf("invalid number of sockets: %d, must be one of 1, 2, 4, or 8", maxSockets)
	}

	socks := make([]*socket, maxSockets)
	for i := range socks {
		socks[i] = &socket{
			sockn: uint8(i),
		}
	}

	d.maxSockets = maxSockets
	d.maxSockSize = 16 * 1024 / maxSockets
	d.sockets = socks

	// Set the RX and TX buffer sizes for each socket.
	for i := 0; i < 8; i++ {
		size := byte(d.maxSockSize >> 10)
		if i >= maxSockets {
			size = 0
		}

		d.writeByte(sockRXBUFSize, sockAddr(uint8(i)), size)
		d.writeByte(sockTXBUFSize, sockAddr(uint8(i)), size)
	}

	mask := byte(0b11111111)
	switch maxSockets {
	case 1:
		mask = 0b00000001
	case 2:
		mask = 0b00000011
	case 4:
		mask = 0b00001111
	}
	d.writeByte(regSockIntMask, 0, mask)
	d.writeByte(regIntMask, 0, 0)
	return nil
}

func (d *Device) findSocket(sockn uint8) *socket {
	for _, sock := range d.sockets {
		if sock.sockn == sockn {
			return sock
		}
	}
	// This should not be possible
	panic(fmt.Sprintf("could not find socket with sockn %d", sockn))
}

// Reset performs a soft reset.
func (d *Device) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.reset()
}

func (d *Device) reset() {
	// RST is bit 7 of regMode.
	d.writeByte(regMode, 0, 0x80)
}

// GetHardwareAddr returns the hardware address of the device.
func (d *Device) GetHardwareAddr() (net.HardwareAddr, error) {
	if debugging(debugNetdev) {
		fmt.Printf("[GetHardwareAddr]\r\n")
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	mac := make([]byte, 6)
	d.read(regMAC, 0, mac)
	return mac, nil
}

// Addr returns the IP address of the device.
func (d *Device) Addr() (netip.Addr, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	var ip [4]byte
	d.read(regIPAddr, 0, ip[:])
	return netip.AddrFrom4(ip), nil
}

// SetAddr sets the IP address of the device.
//
// The IP address must be a valid IPv4 address.
func (d *Device) SetAddr(ip netip.Addr) error {
	if err := d.setAddress(regIPAddr, ip); err != nil {
		return fmt.Errorf("could not set IP address: %w", err)
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	d.laddr = ip
	return nil
}

// SetSubnetMask sets the subnet mask of the device.
//
// The subnet mask must be a valid IPv4 address.
// It is not checked if the subnet mask is valid for the device's IP address.
func (d *Device) SetSubnetMask(mask netip.Addr) error {
	return d.setAddress(regSubnetMask, mask)
}

// SetGateway sets the gateway address of the device.
//
// The gateway must be a valid IPv4 address.
// It is not checked if the gateway is in the same subnet as the device.
func (d *Device) SetGateway(gateway netip.Addr) error {
	return d.setAddress(regGatewayAddr, gateway)
}

func (d *Device) setAddress(addr uint16, ip netip.Addr) error {
	if !ip.IsValid() || !ip.Is4() {
		return fmt.Errorf("invalid IP address: %s", ip)
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	d.write(addr, 0, ip.AsSlice())
	return nil
}

// LinkStatus is the link status of the device.
type LinkStatus = uint8

// LinkStatus values.
const (
	LinkStatusDown LinkStatus = iota
	LinkStatusUp
)

// LinkStatus returns the current link status of the device.
func (d *Device) LinkStatus() LinkStatus {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.readByte(regPHYCfg, 0) & 0b00000001
}

// LinkInfo returns the current link information of the device.
func (d *Device) LinkInfo() string {
	d.mu.Lock()
	defer d.mu.Unlock()

	linkInfo := d.readByte(regPHYCfg, 0) & 0b00000110

	speed := "10Mbps"
	if linkInfo&0b00000010 != 0 {
		speed = "100Mbps"
	}
	duplex := "Half Duplex"
	if linkInfo&0b00000100 != 0 {
		duplex = "Full Duplex"
	}
	return speed + " " + duplex
}
