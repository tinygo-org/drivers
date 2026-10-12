package rtl8720dn

import (
	"fmt"
	"net/netip"

	"tinygo.org/x/drivers/netdev"
)

type sock int32

type socket struct {
	protocol      int
	inuse         bool
	connectFailed bool
}

func newSocket(protocol int) *socket {
	return &socket{protocol: protocol, inuse: true, connectFailed: protocol == netdev.IPPROTO_TLS}
}

func ipToName(ip netip.AddrPort) []byte {
	name := make([]byte, 16)
	name[0] = 0x00
	name[1] = netdev.AF_INET
	name[2] = byte(ip.Port() >> 8)
	name[3] = byte(ip.Port())
	if ip.Addr().Is4() {
		addr := ip.Addr().As4()
		name[4] = byte(addr[0])
		name[5] = byte(addr[1])
		name[6] = byte(addr[2])
		name[7] = byte(addr[3])
	}
	return name
}

func (r *rtl8720dn) Connect(sockfd int, host string, ip netip.AddrPort) error {

	port := ip.Port()

	if debugging(debugNetdev) {
		if host == "" {
			fmt.Printf("[Connect] sockfd: %d, addr: %s\r\n", sockfd, ip)
		} else {
			fmt.Printf("[Connect] sockfd: %d, host: %s:%d\r\n", sockfd, host, port)
		}
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	var sock = sock(sockfd)
	var socket = r.sockets[sock]
	var name = ipToName(ip)
	if socket == nil || !socket.inuse {
		return netdev.ErrInvalidSocketFd
	}

	switch socket.protocol {
	case netdev.IPPROTO_TCP, netdev.IPPROTO_UDP:
		result := r.rpc_lwip_connect(int32(sock), name, uint32(len(name)))
		if result == -1 {
			return fmt.Errorf("Connect to %s failed", ip)
		}
	case netdev.IPPROTO_TLS:
		socket.connectFailed = true
		result, err := r.rpc_wifi_start_ssl_client(uint32(sock), host, uint32(port), 0)
		if err != nil {
			return fmt.Errorf("TLS connection to %s:%d failed: %w", host, port, err)
		}
		if result < 0 {
			return fmt.Errorf("TLS connection to %s:%d failed: code %d", host, port, result)
		}
		internalSocket, err := r.rpc_wifi_ssl_get_socket(uint32(sock))
		if err != nil {
			return fmt.Errorf("TLS socket check for %s:%d failed: %w", host, port, err)
		}
		if internalSocket < 0 {
			return fmt.Errorf("TLS connection to %s:%d failed: socket %d", host, port, internalSocket)
		}
		socket.connectFailed = false
	}

	return nil
}

func (r *rtl8720dn) Close(sockfd int) error {

	if debugging(debugNetdev) {
		fmt.Printf("[Close] sockfd: %d\r\n", sockfd)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	var sock = sock(sockfd)
	var socket = r.sockets[sock]
	var result int32

	if socket == nil || !socket.inuse {
		return nil
	}

	switch socket.protocol {
	case netdev.IPPROTO_TCP, netdev.IPPROTO_UDP:
		result = r.rpc_lwip_close(int32(sock))
	case netdev.IPPROTO_TLS:
		r.rpc_wifi_stop_ssl_socket(uint32(sock))
		r.rpc_wifi_ssl_client_destroy(uint32(sock))
	}

	if result == -1 {
		return netdev.ErrClosingSocket
	}

	socket.inuse = false
	if socket.protocol == netdev.IPPROTO_TLS {
		delete(r.sockets, sock)
	}

	return nil
}
