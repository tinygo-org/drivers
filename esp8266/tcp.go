package esp8266

import "time"

// ConnectSocket creates a new TCP or UDP connection for the ESP8266.
// Currently only supports single connection mode.
func (d Device) ConnectSocket(protocol, addr, port string) error {
	val := "\"" + protocol + "\",\"" + addr + "\",\"" + port + "\""
	d.Set(TCPConnect, val)
	time.Sleep(100 * time.Millisecond)
	d.Response()
	return nil
}

// DisconnectSocket disconnects the ESP8266 from the current TCP/UDP connection.
func (d Device) DisconnectSocket() error {
	d.Execute(TCPClose)
	time.Sleep(100 * time.Millisecond)
	d.Response()
	return nil
}
