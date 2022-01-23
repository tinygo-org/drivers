package net

import (
	"errors"
	"time"
)

var (
	ErrWiFiMissingSSID    = errors.New("missing SSID")
	ErrWiFiConnectTimeout = errors.New("WiFi connect timeout")
)

// Adapter interface is used to communicate with the network adapter.
type Adapter interface {
	// functions used to connect/disconnect to/from an access point
	ConnectToAccessPoint(ssid, pass string, timeout time.Duration) error
	Disconnect() error
	GetClientIP() (string, error)

	// these functions are used once the adapter is connected to the network
	GetDNS(domain string) (string, error)
	ConnectTCPSocket(addr, port string) error
	ConnectSSLSocket(addr, port string) error
	ConnectUDPSocket(addr, sendport, listenport string) error
	DisconnectSocket() error
	StartSocketSend(size int) error
	Write(b []byte) (n int, err error)
	ReadSocket(b []byte) (n int, err error)
	IsSocketDataAvailable() bool

	// FIXME: this is really specific to espat, and maybe shouldn't be part
	// of the driver interface
	Response(timeout int) ([]byte, error)
}

var ActiveDevice Adapter

func UseDriver(a Adapter) {
	// TODO: rethink and refactor this
	if ActiveDevice != nil {
		panic("net.ActiveDevice is already set")
	}
	ActiveDevice = a
}
