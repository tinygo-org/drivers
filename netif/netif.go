package netif

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"time"
	_ "unsafe"
)

//go:linkname UseStack net.useNetdev
func UseStack(stack Stack)

// Socket errors.
var (
	ErrProtocolNotSupported = errors.New("socket protocol/type not supported")
)

// Wifi errors.
var (
	ErrAuthFailure     = errors.New("wifi authentication failure")
	ErrAuthUnsupported = errors.New("wifi authorization type not supported")
)

const (
	_AF_INET       = 0x2
	_SOCK_STREAM   = 0x1
	_SOCK_DGRAM    = 0x2
	_SOL_SOCKET    = 0x1
	_SO_KEEPALIVE  = 0x9
	_SOL_TCP       = 0x6
	_TCP_KEEPINTVL = 0x5
	_IPPROTO_TCP   = 0x6
	_IPPROTO_UDP   = 0x11
	// Made up, not a real IP protocol number.  This is used to create a
	// TLS socket on the device, assuming the device supports mbed TLS.
	_IPPROTO_TLS = 0xFE
	_F_SETFL     = 0x4
)

// Interface is the minimum interface that need be implemented by any network
// device driver and is based on [net.Interface].
type Interface interface {
	// HardwareAddr6 returns the device's 6-byte [MAC address].
	//
	// [MAC address]: https://en.wikipedia.org/wiki/MAC_address
	HardwareAddr6() ([6]byte, error)
	// NetFlags returns the net.Flag values for the interface. It includes state of connection.
	NetFlags() net.Flags
	// MTU returns the maximum transmission unit size.
	MTU() int
	// Notify to register callback for network events. May not be supported for certain devices.
	NetNotify(cb func(Event)) error
}

// InterfaceEthPoller is implemented by devices that send/receive ethernet packets.
type InterfaceEthPoller interface {
	Interface
	// SendEth sends an Ethernet packet
	SendEth(pkt []byte) error
	// RecvEthHandle sets recieve Ethernet packet callback function
	RecvEthHandle(func(pkt []byte) error)
	// PollOne tries to receive one Ethernet packet and returns true if one was
	PollOne() (bool, error)
}

// InterfaceWifi is implemented by a interface device that has the capacity
// to connect to wifi networks.
type InterfaceWifi interface {
	Interface
	// Connect device to network
	NetConnect(params WifiParams) error
	// Disconnect device from network
	NetDisconnect()
}

// InterfaceEthPollWifi is implemented by devices that connect to wifi networks
// and send/receive ethernet packets (OSI level 2).
type InterfaceEthPollWifi interface {
	InterfaceWifi
	InterfaceEthPoller
}

type Stack interface {
	// GetHostByName returns the IP address of either a hostname or IPv4
	// address in standard dot notation
	// GetHostByName(name string) (netip.Addr, error)

	// Addr returns IP address assigned to the interface, either by
	// DHCP or statically
	Addr() (netip.Addr, error)

	// Berkely Sockets-like interface, Go-ified.  See man page for socket(2), etc.
	Socket(domain int, stype int, protocol int) (int, error)
	Bind(sockfd int, ip netip.AddrPort) error
	Connect(sockfd int, host string, ip netip.AddrPort) error
	Listen(sockfd int, backlog int) error
	Accept(sockfd int) (int, netip.AddrPort, error)
	Send(sockfd int, buf []byte, flags int, deadline time.Time) (int, error)
	Recv(sockfd int, buf []byte, flags int, deadline time.Time) (int, error)
	Close(sockfd int) error
	SetSockOpt(sockfd int, level int, opt int, value interface{}) error
}

// Resolver is implemented by DNS resolvers, notably Go's [net.DefaultResolver].
type Resolver interface {
	// LookupNetIP looks up host using the local resolver.
	// It returns a slice of that host's IP addresses of the type specified by
	// network.
	// The network must be one of "ip", "ip4" or "ip6".
	LookupNetIP(ctx context.Context, network, host string) ([]netip.Addr, error)
}

// StackWifi is returned by `Probe` function for devices that communicate
// on the OSI level 4 (transport) layer.
type StackWifi interface {
	InterfaceWifi
	Stack
}

type WifiParams struct {
	// Connect mode
	ConnectMode ConnectMode

	// SSID of Wifi AP
	SSID string

	// Passphrase of Wifi AP
	Passphrase string

	// Wifi authorization type
	Auth AuthType

	// Wifi country code as two-char string.  E.g. "XX" for world-wide,
	// "US" for USA, etc.
	CountryCode string
}

type Event uint8

// Network events
const (
	// The device's network connection is now UP
	EventNetUp Event = iota
	// The device's network connection is now DOWN
	EventNetDown
)

type ConnectMode uint8

// Connect modes
const (
	ConnectModeSTA = iota // Connect as Wifi station (default)
	ConnectModeAP         // Connect as Wifi Access Point
)

type AuthType uint8

// Wifi authorization types.  Used when setting up an access point, or
// connecting to an access point
const (
	AuthTypeWPA2      = iota // WPA2 authorization (default)
	AuthTypeOpen             // No authorization required (open)
	AuthTypeWPA              // WPA authorization
	AuthTypeWPA2Mixed        // WPA2/WPA mixed authorization
)

type WifiAutoconnectParams struct {
	WifiParams

	// Retries is how many attempts to connect before returning with a
	// "Connect failed" error.  Zero means infinite retries.
	// Retries int // Probably should be implemented as a function

	// Timeout duration for each connection attempt.  The default zero
	// value means 10sec.
	ConnectTimeout time.Duration

	// Watchdog ticker duration.  On tick, the watchdog will check for
	// downed connection or hardware fault and try to recover the
	// connection.  Set to zero to disable watchodog.
	WatchdogTimeout time.Duration
}

func StartWifiAutoconnect(dev InterfaceWifi, cfg WifiAutoconnectParams) error {
	if dev == nil {
		return errors.New("nil device")
	}
	go func() {
		// Wifi autoconnect algorithm in one place,
		// no need to implement for every single netdever.
	RECONNECT:
		for i := 0; i < 4; i++ {
			err := dev.NetConnect(cfg.WifiParams)
			if err != nil {
				time.Sleep(cfg.ConnectTimeout)
				goto RECONNECT
			}
			// Once connected reset the retry counter.
			i = 0
			for cfg.WatchdogTimeout != 0 {
				time.Sleep(cfg.WatchdogTimeout)
				if dev.NetFlags()&net.FlagRunning == 0 {
					goto RECONNECT
				}
			}
		}
	}()
	return nil
}
