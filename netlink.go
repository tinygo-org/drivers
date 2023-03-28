package drivers

import (
	"errors"
	"net"
)

// NetConnect() errors
var (
	ErrConnected          = errors.New("Already connected")
	ErrConnectFailed      = errors.New("Connect failed")
	ErrConnectTimeout     = errors.New("Connect timed out")
	ErrMissingSSID        = errors.New("Missing WiFi SSID")
	ErrStartingDHCPClient = errors.New("Error starting DHPC client")
)

type NetlinkEvent int

// Netlink network events
const (
	// The device's network connection is now UP
	NetlinkEventNetUp NetlinkEvent = iota
	// The device's network connection is now DOWN
	NetlinkEventNetDown
)

// Network drivers (optionally) implement the Netlinker interface.  This
// interface is not used by TinyGo's "net" package, but rather provides the
// TinyGo application direct access to the network device for common settings
// and control that fall outside of netdev's socket interface.

type Netlinker interface {

	// NetConnect device to IP network
	NetConnect() error

	// NetDisconnect device from IP network
	NetDisconnect()

	// NetNotify to register callback for network events
	NetNotify(func(NetlinkEvent))

	// GetHardwareAddr returns device MAC address
	GetHardwareAddr() (net.HardwareAddr, error)

	// GetIPAddr returns IP address assigned to device, either by DHCP or
	// statically
	GetIPAddr() (net.IP, error)
}
