package rtl8720dn

import (
	"fmt"
	"strconv"
)

// Here is the implementation of tinygo-org/x/drivers/net.DeviceDriver.

func (d *Driver) GetDNS(domain string) (string, error) {
	if d.debug {
		fmt.Printf("GetDNS(%q)\r\n", domain)
	}

	ipaddr := make([]byte, 4)
	_, err := d.Rpc_netconn_gethostbyname(domain, &ipaddr)
	if err != nil {
		return "", err
	}

	ret, err := fmt.Sprintf("%d.%d.%d.%d", ipaddr[0], ipaddr[1], ipaddr[2], ipaddr[3]), nil
	if d.debug {
		fmt.Printf("-> %s\r\n", ret)
		fmt.Printf("-> %02X.%02X.%02X.%02X\r\n", ipaddr[0], ipaddr[1], ipaddr[2], ipaddr[3])
	}
	return ret, err
}

func (d *Driver) ConnectTCPSocket(addr, port string) error {
	if d.debug {
		fmt.Printf("ConnectTCPSocket(%q, %q)\r\n", addr, port)
	}

	ipaddr := make([]byte, 4)
	if len(addr) == 4 {
		copy(ipaddr, addr)
	} else {
		_, err := d.Rpc_netconn_gethostbyname(addr, &ipaddr)
		if err != nil {
			return err
		}
	}

	portNum, err := strconv.ParseUint(port, 0, 0)
	if err != nil {
		return err
	}

	socket, err := d.Rpc_lwip_socket(0x02, 0x01, 0x00)
	if err != nil {
		return err
	}
	d.socket = socket
	d.connectionType = ConnectionTypeTCP

	_, err = d.Rpc_lwip_fcntl(socket, 0x00000003, 0x00000000)
	if err != nil {
		return err
	}

	_, err = d.Rpc_lwip_fcntl(socket, 0x00000004, 0x00000001)
	if err != nil {
		return err
	}

	name := []byte{0x00, 0x02, 0x00, 0x50, 0xC0, 0xA8, 0x01, 0x76, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	name[2] = byte(portNum >> 8)
	name[3] = byte(portNum)
	name[4] = byte(ipaddr[0])
	name[5] = byte(ipaddr[1])
	name[6] = byte(ipaddr[2])
	name[7] = byte(ipaddr[3])

	_, err = d.Rpc_lwip_connect(socket, name, uint32(len(name)))
	if err != nil {
		return err
	}

	readset := []byte{}
	writeset := []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	exceptset := []byte{}
	timeout := []byte{}
	_, err = d.Rpc_lwip_select(0x01, readset, writeset, exceptset, timeout)
	if err != nil {
		return err
	}

	optval := make([]byte, 4)
	optlen := uint32(len(optval))
	_, err = d.Rpc_lwip_getsockopt(socket, 0x00000FFF, 0x00001007, []byte{0xA5, 0xA5, 0xA5, 0xA5}, &optval, &optlen)
	if err != nil {
		return err
	}

	_, err = d.Rpc_lwip_fcntl(socket, 0x00000003, 0x00000000)
	if err != nil {
		return err
	}

	_, err = d.Rpc_lwip_fcntl(socket, 0x00000004, 0x00000000)
	if err != nil {
		return err
	}

	readset = []byte{}
	writeset = []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	exceptset = []byte{}
	timeout = []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x40, 0x42, 0x0F, 0x00, 0xFF, 0xFF, 0xFF, 0xFF}
	_, err = d.Rpc_lwip_select(0x01, readset, writeset, exceptset, timeout)
	if err != nil {
		return err
	}

	return nil
}

func (d *Driver) ConnectSSLSocket(addr, port string) error {
	if d.debug {
		fmt.Printf("ConnectSSLSocket(%q, %q)\r\n", addr, port)
	}
	if d.root_ca == nil {
		return fmt.Errorf("root_ca is not set")
	}

	client, err := d.Rpc_wifi_ssl_client_create()
	if err != nil {
		return err
	}
	d.client = client
	d.connectionType = ConnectionTypeTLS

	err = d.Rpc_wifi_ssl_init(client)
	if err != nil {
		return err
	}

	err = d.Rpc_wifi_ssl_set_timeout(client, 0x0001D4C0)
	if err != nil {
		return err
	}

	_, err = d.Rpc_wifi_ssl_set_rootCA(client, *d.root_ca)
	if err != nil {
		return err
	}

	// TODO: use port
	_, err = d.Rpc_wifi_start_ssl_client(client, addr, 443, 0x0001D4C0)
	if err != nil {
		return err
	}

	_, err = d.Rpc_wifi_ssl_get_socket(client)
	if err != nil {
		return err
	}
	return nil
}

func (d *Driver) ConnectUDPSocket(addr, sendport, listenport string) error {
	if d.debug {
		fmt.Printf("ConnectUDPSocket(\"%d.%d.%d.%d\", %q, %q)\r\n", byte(addr[0]), byte(addr[1]), byte(addr[2]), byte(addr[3]), sendport, listenport)
	}

	socket, err := d.Rpc_lwip_socket(0x02, 0x02, 0x00)
	if err != nil {
		return err
	}
	d.socket = socket
	d.connectionType = ConnectionTypeUDP

	optval := []byte{0x01, 0x00, 0x00, 0x00}
	_, err = d.Rpc_lwip_setsockopt(socket, 0x00000FFF, 0x00000004, optval, uint32(len(optval)))
	if err != nil {
		return err
	}

	port, err := strconv.ParseUint(sendport, 10, 0)
	if err != nil {
		return err
	}

	ip := []byte(addr)

	// remote info
	d.udpInfo[0] = byte(port >> 8)
	d.udpInfo[1] = byte(port)
	d.udpInfo[2] = ip[0]
	d.udpInfo[3] = ip[1]
	d.udpInfo[4] = ip[2]
	d.udpInfo[5] = ip[3]

	port, err = strconv.ParseUint(listenport, 10, 0)
	if err != nil {
		return err
	}

	ip_info := make([]byte, 12)
	_, err = d.Rpc_tcpip_adapter_get_ip_info(0, &ip_info)
	if err != nil {
		return err
	}

	name := []byte{0x00, 0x02, 0x0D, 0x05, 0xC0, 0xA8, 0x01, 0x78, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	name[2] = byte(port >> 8)
	name[3] = byte(port)
	name[4] = ip_info[0]
	name[5] = ip_info[1]
	name[6] = ip_info[2]
	name[7] = ip_info[3]

	_, err = d.Rpc_lwip_bind(socket, name, uint32(len(name)))
	if err != nil {
		return err
	}

	_, err = d.Rpc_lwip_fcntl(socket, 0x00000004, 0x00000000)
	if err != nil {
		return err
	}

	return nil
}

func (d *Driver) DisconnectSocket() error {
	if d.debug {
		fmt.Printf("DisconnectSocket()\r\n")
	}
	switch d.connectionType {
	case ConnectionTypeTCP, ConnectionTypeUDP:
		_, err := d.Rpc_lwip_close(d.socket)
		if err != nil {
			return err
		}
	case ConnectionTypeTLS:
		err := d.Rpc_wifi_stop_ssl_socket(d.client)
		if err != nil {
			return err
		}

		err = d.Rpc_wifi_ssl_client_destroy(d.client)
		if err != nil {
			return err
		}
	default:
	}
	d.connectionType = ConnectionTypeNone
	return nil
}

func (d *Driver) StartSocketSend(size int) error {
	if d.debug {
		fmt.Printf("StartSocketSend(%d)\r\n", size)
	}
	// No implementation required
	return nil
}

func (d *Driver) Write(b []byte) (n int, err error) {
	if d.debug {
		fmt.Printf("Write(%#v)\r\n", b)
	}

	switch d.connectionType {
	case ConnectionTypeTCP:
		sn, err := d.Rpc_lwip_send(d.socket, b, 0x00000008)
		if err != nil {
			return 0, err
		}
		n = int(sn)
	case ConnectionTypeUDP:
		to := []byte{0x00, 0x02, 0x0D, 0x05, 0xC0, 0xA8, 0x01, 0x76, 0xEB, 0x43, 0x00, 0x00, 0xD5, 0x27, 0x01, 0x00}
		copy(to[2:], d.udpInfo[:])
		sn, err := d.Rpc_lwip_sendto(d.socket, b, 0x00000000, to, uint32(len(to)))
		if err != nil {
			return 0, err
		}
		n = int(sn)
	case ConnectionTypeTLS:
		sn, err := d.Rpc_wifi_send_ssl_data(d.client, b, uint16(len(b)))
		if err != nil {
			return 0, err
		}
		n = int(sn)
	default:
		return 0, nil
	}
	return n, nil
}

func (d *Driver) ReadSocket(b []byte) (n int, err error) {
	if d.debug {
		//fmt.Printf("ReadSocket(b)\r\n")
	}
	if d.connectionType == ConnectionTypeNone {
		return 0, nil
	}

	switch d.connectionType {
	case ConnectionTypeTCP:
		length := len(b)
		if length > maxUartRecvSize-16 {
			length = maxUartRecvSize - 16
		}
		buf := b[:length]
		nn, err := d.Rpc_lwip_recv(d.socket, &buf, uint32(length), 0x00000008, 0x00002800)
		if err != nil {
			return 0, err
		}

		if nn == -1 {
			return 0, nil
		} else if nn == 0 {
			return 0, d.DisconnectSocket()
		}
		n = int(nn)
	case ConnectionTypeUDP:
		length := len(b)
		if length > maxUartRecvSize-32 {
			length = maxUartRecvSize - 32
		}
		buf := b[:length]
		from := make([]byte, 16)
		fromLen := uint32(len(from))
		nn, err := d.Rpc_lwip_recvfrom(d.socket, &buf, uint32(length), 0x00000008, &from, &fromLen, 10000)
		if err != nil {
			return 0, err
		}

		if nn == -1 {
			return 0, nil
		}
		n = int(nn)
	case ConnectionTypeTLS:
		length := len(b)
		if length > maxUartRecvSize-16 {
			length = maxUartRecvSize - 16
		}
		buf := b[:length]
		nn, err := d.Rpc_wifi_get_ssl_receive(d.client, &buf, int32(length))
		if err != nil {
			return 0, err
		}
		if nn < 0 {
			return 0, fmt.Errorf("error %d", n)
		} else if nn == 0 || nn == -30848 {
			return 0, d.DisconnectSocket()
		}
		n = int(nn)
	default:
	}

	return n, nil
}

func (d *Driver) IsSocketDataAvailable() bool {
	if d.debug {
		fmt.Printf("IsSocketDataAvailable()\r\n")
	}
	ret, err := d.Rpc_lwip_available(d.socket)
	if err != nil {
		fmt.Printf("error: %s\r\n", err.Error())
		return false
	}
	if ret == 1 {
		return true
	}
	return false
}

func (d *Driver) Response(timeout int) ([]byte, error) {
	if d.debug {
		fmt.Printf("Response(%d))\r\n", timeout)
	}
	// No implementation required
	return nil, nil
}
