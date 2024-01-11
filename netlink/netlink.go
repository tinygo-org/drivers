// L2 data link layer

package netlink

import (
	"errors"
	"net"
	"time"
)

var (
	ErrConnected         = errors.New("Already connected")
	ErrConnectFailed     = errors.New("Connect failed")
	ErrConnectTimeout    = errors.New("Connect timed out")
	ErrMissingSSID       = errors.New("Missing WiFi SSID")
	ErrShortPassphrase   = errors.New("Invalid Wifi Passphrase < 8 chars")
	ErrAuthFailure       = errors.New("Wifi authentication failure")
	ErrAuthTypeNoGood    = errors.New("Wifi authorization type not supported")
	ErrConnectModeNoGood = errors.New("Connect mode not supported")
	ErrNotSupported      = errors.New("Not supported")
)

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

const DefaultConnectTimeout = 10 * time.Second

type ConnectParams struct {

	// Connect mode
	ConnectMode

	// SSID of Wifi AP
	Ssid string

	// Passphrase of Wifi AP
	Passphrase string

	// Wifi authorization type
	AuthType

	// Wifi country code as two-char string.  E.g. "XX" for world-wide,
	// "US" for USA, etc.
	Country string

	// Retries is how many attempts to connect before returning with a
	// "Connect failed" error.  Zero means infinite retries.
	Retries int

	// Timeout duration for each connection attempt.  The default zero
	// value means 10sec.
	ConnectTimeout time.Duration

	// Watchdog ticker duration.  On tick, the watchdog will check for
	// downed connection or hardware fault and try to recover the
	// connection.  Set to zero to disable watchodog.
	WatchdogTimeout time.Duration
}

// Netlinker is TinyGo's OSI L2 data link layer interface.  Network device
// drivers implement Netlinker to expose the device's L2 functionality.

type Netlinker interface {

	// Connect device to network
	NetConnect(params *ConnectParams) error

	// Disconnect device from network
	NetDisconnect()

	// Notify to register callback for network events
	NetNotify(cb func(Event))

	// GetHardwareAddr returns device MAC address
	GetHardwareAddr() (net.HardwareAddr, error)
}
