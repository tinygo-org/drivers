package tcpip

import (
	"net"
	"time"

	"tinygo.org/x/drivers/netdev"
)

func (s *Stack) GetHostByName(name string) (net.IP, error) {
	return net.IP{}, netdev.ErrNotSupported
}

func (s *Stack) GetIPAddr() (net.IP, error) {
	return net.IP{}, netdev.ErrNotSupported
}

func (s *Stack) Socket(domain int, stype int, protocol int) (int, error) {
	return -1, netdev.ErrNotSupported
}

func (s *Stack) Bind(sockfd int, ip net.IP, port int) error {
	return netdev.ErrNotSupported
}

func (s *Stack) Connect(sockfd int, host string, ip net.IP, port int) error {
	return netdev.ErrNotSupported
}

func (s *Stack) Listen(sockfd int, backlog int) error {
	return netdev.ErrNotSupported
}

func (s *Stack) Accept(sockfd int, ip net.IP, port int) (int, error) {
	return -1, netdev.ErrNotSupported
}

func (s *Stack) Send(sockfd int, buf []byte, flags int, deadline time.Time) (int, error) {
	return 0, netdev.ErrNotSupported
}

func (s *Stack) Recv(sockfd int, buf []byte, flags int, deadline time.Time) (int, error) {
	return 0, netdev.ErrNotSupported
}

func (s *Stack) Close(sockfd int) error {
	return netdev.ErrNotSupported
}

func (s *Stack) SetSockOpt(sockfd int, level int, opt int, value interface{}) error {
	return netdev.ErrNotSupported
}
