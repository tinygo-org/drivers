package net

import "strings"

// SplitHostPort splits a network address of the form "host:port",
// "host%zone:port", "[host]:port" or "[host%zone]:port" into host or
// host%zone and port.
//
// A literal IPv6 address in hostport must be enclosed in square
// brackets, as in "[::1]:80", "[::1%lo0]:80".
//
// See func Dial for a description of the hostport parameter, and host
// and port results.
func SplitHostPort(hostport string) (host, port string, err error) {

	if strings.Contains(hostport, ":") {
		spl := strings.Split(hostport, ":")
		host = spl[0]
		port = spl[1]
	} else {
		host = hostport
		port = "80"
	}

	return host, port, nil
}
