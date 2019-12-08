package net

type DeviceDriver interface {
	GetDNS(domain string) (string, error)
	ConnectTCPSocket(addr, port string) error
	ConnectSSLSocket(addr, port string) error
	ConnectUDPSocket(addr, sendport, listenport string) error
	DisconnectSocket() error
	StartSocketSend(size int) error
	Write(b []byte) (n int, err error)
	ReadSocket(b []byte) (n int, err error)
	IsSocketDataAvailable() bool
}

var ActiveDevice DeviceDriver

func UseDriver(driver DeviceDriver) {
	// TODO: rethink and refactor this
	if ActiveDevice != nil {
		panic("net.ActiveDevice is already set")
	}
	ActiveDevice = driver
}
