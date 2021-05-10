// Package tls is intended to provide a minimal set of compatible interfaces with the
// Go standard library's tls package.
package tls

import (
	"strconv"
	"strings"

	"tinygo.org/x/drivers/net"
)

// Dial makes a TLS network connection. It tries to provide a mostly compatible interface
// to tls.Dial().
// Dial connects to the given network address.
func Dial(network, address string, config *Config) (*net.TCPSerialConn, error) {
	raddr, err := net.ResolveTCPAddr(network, address)
	if err != nil {
		return nil, err
	}

	hostname := strings.Split(address, ":")[0]
	sendport := strconv.Itoa(raddr.Port)
	if sendport == "0" {
		sendport = "443"
	}

	// disconnect any old socket
	net.ActiveDevice.DisconnectSocket()

	// connect new socket
	err = net.ActiveDevice.ConnectSSLSocket(hostname, sendport)
	if err != nil {
		return nil, err
	}

	return net.NewTCPSerialConn(net.SerialConn{Adaptor: net.ActiveDevice}, nil, raddr), nil
}

// Config is a placeholder for future compatibility with
// tls.Config.
type Config struct {
}
