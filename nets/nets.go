package nets

import (
	"errors"
	"net/netip"
	"time"
)

//go:linkname UseStack net.useNetdev
func UseStack(stack Stack)

// GethostByName() errors
var (
	ErrHostUnknown = errors.New("host unknown")
	ErrMalAddr     = errors.New("malformed address")
)

// Socket errors
var (
	ErrFamilyNotSupported   = errors.New("address family not supported")
	ErrProtocolNotSupported = errors.New("socket protocol/type not supported")
	ErrStartingDHCPClient   = errors.New("error starting DHPC client")
	ErrNoMoreSockets        = errors.New("no more sockets")
	ErrClosingSocket        = errors.New("error closing socket")
)

var (
	ErrConnected         = errors.New("already connected")
	ErrConnectFailed     = errors.New("connect failed")
	ErrConnectTimeout    = errors.New("connect timed out")
	ErrMissingSSID       = errors.New("missing WiFi SSID")
	ErrAuthFailure       = errors.New("wifi authentication failure")
	ErrAuthTypeNoGood    = errors.New("wifi authorization type not supported")
	ErrConnectModeNoGood = errors.New("connect mode not supported")
	ErrNotSupported      = errors.New("not supported")
)

const (
	AF_INET       = 0x2
	SOCK_STREAM   = 0x1
	SOCK_DGRAM    = 0x2
	SOL_SOCKET    = 0x1
	SO_KEEPALIVE  = 0x9
	SOL_TCP       = 0x6
	TCP_KEEPINTVL = 0x5
	IPPROTO_TCP   = 0x6
	IPPROTO_UDP   = 0x11
	// Made up, not a real IP protocol number.  This is used to create a
	// TLS socket on the device, assuming the device supports mbed TLS.
	IPPROTO_TLS = 0xFE
	F_SETFL     = 0x4
)

// Link is the minimum interface that need be implemented by any network
// device driver.
type Link interface {
	// HardwareAddr6 returns the device's 6-byte [MAC address], a.k.a EUI-48.
	//
	// [MAC address]: https://en.wikipedia.org/wiki/MAC_address
	HardwareAddr6() ([6]byte, error)
	// LinkStatus returns the state of the connection.
	LinkStatus() LinkStatus
	// MTU returns the maximum transmission unit size.
	MTU() int
}

type EthLinkPoller interface {
	Link
	// SendEth sends an Ethernet packet
	SendEth(pkt []byte) error
	// RecvEthHandle sets recieve Ethernet packet callback function
	RecvEthHandle(func(pkt []byte) error)
	// PollOne tries to receive one Ethernet packet and returns true if one was
	PollOne() (bool, error)
}

type LinkWifi interface {
	Link
	// Connect device to network
	NetConnect(params WifiParams) error
	// Disconnect device from network
	NetDisconnect()
	// Notify to register callback for network events
	NetNotify(cb func(Event))
}

type Stack interface {
	// GetHostByName returns the IP address of either a hostname or IPv4
	// address in standard dot notation
	GetHostByName(name string) (netip.Addr, error)

	// Addr returns IP address assigned to the interface, either by
	// DHCP or statically
	Addr() (netip.Addr, error)

	// Berkely Sockets-like interface, Go-ified.  See man page for socket(2), etc.
	Socket(domain int, stype int, protocol int) (int, error)
	Bind(sockfd int, ip netip.AddrPort) error
	Connect(sockfd int, host string, ip netip.AddrPort) error
	Listen(sockfd int, backlog int) error
	Accept(sockfd int, ip netip.AddrPort) (int, error)
	Send(sockfd int, buf []byte, flags int, deadline time.Time) (int, error)
	Recv(sockfd int, buf []byte, flags int, deadline time.Time) (int, error)
	Close(sockfd int) error
	SetSockOpt(sockfd int, level int, opt int, value interface{}) error
}

// Netdev is returned by `Probe` function.
type Netdev interface {
	LinkWifi
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

type Event int

// Network events
const (
	// The device's network connection is now UP
	EventNetUp Event = iota
	// The device's network connection is now DOWN
	EventNetDown
)

type ConnectMode int

// Connect modes
const (
	ConnectModeSTA = iota // Connect as Wifi station (default)
	ConnectModeAP         // Connect as Wifi Access Point
)

type AuthType int

// Wifi authorization types.  Used when setting up an access point, or
// connecting to an access point
const (
	AuthTypeWPA2      = iota // WPA2 authorization (default)
	AuthTypeOpen             // No authorization required (open)
	AuthTypeWPA              // WPA authorization
	AuthTypeWPA2Mixed        // WPA2/WPA mixed authorization
)

type LinkStatus uint8

const (
	LinkDown LinkStatus = iota
	LinkConnecting
	LinkUp
)

type WifiAutoconnectParams struct {
	WifiParams
	// NOTE ON FIELDS BELOW: These fields seem like a leaky abstraction- they can
	// be implemented as an external "AutoConnect" function that is called in
	// a goroutine that performs auto connect and whatnot.

	// Retries is how many attempts to connect before returning with a
	// "Connect failed" error.  Zero means infinite retries.
	Retries int // Probably should be implemented as a function

	// Timeout duration for each connection attempt.  The default zero
	// value means 10sec.
	ConnectTimeout time.Duration

	// Watchdog ticker duration.  On tick, the watchdog will check for
	// downed connection or hardware fault and try to recover the
	// connection.  Set to zero to disable watchodog.
	WatchdogTimeout time.Duration
}

func StartWifiAutoconnect(dev Netdev, cfg WifiAutoconnectParams) error {
	if dev == nil {
		return ErrConnectModeNoGood
	}
	go func() {
		// Wifi autoconnect algorithm in one place,
		// no need to implement for every single netdever.
	RECONNECT:
		for i := 0; i < cfg.Retries; i++ {
			err := dev.NetConnect(cfg.WifiParams)
			if err != nil {
				time.Sleep(cfg.ConnectTimeout)
				goto RECONNECT
			}
			for cfg.WatchdogTimeout != 0 {
				time.Sleep(cfg.WatchdogTimeout)
				if dev.LinkStatus() == LinkDown {
					i = 0
					goto RECONNECT
				}
			}
		}
	}()
	return nil
}
