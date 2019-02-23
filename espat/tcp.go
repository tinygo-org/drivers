package espat

import (
	"strconv"
	"time"
)

const (
	TCPMuxSingle   = 0
	TCPMuxMultiple = 1

	TCPTransferModeNormal      = 0
	TCPTransferModeUnvarnished = 1
)

// ConnectTCPSocket creates a new TCP socket connection for the ESP8266/ESP32.
// Currently only supports single connection mode.
func (d *Device) ConnectTCPSocket(addr, port string) error {
	protocol := "TCP"
	val := "\"" + protocol + "\",\"" + addr + "\"," + port
	d.Set(TCPConnect, val)
	time.Sleep(100 * time.Millisecond)
	d.Response()
	return nil
}

// ConnectUDPSocket creates a new UDP connection for the ESP8266/ESP32.
func (d *Device) ConnectUDPSocket(addr, sendport, listenport string) error {
	protocol := "UDP"
	val := "\"" + protocol + "\",\"" + addr + "\"," + sendport + "," + listenport + ",2"
	d.Set(TCPConnect, val)
	time.Sleep(pause * time.Millisecond)
	d.Response()
	return nil
}

// DisconnectSocket disconnects the ESP8266/ESP32 from the current TCP/UDP connection.
func (d *Device) DisconnectSocket() error {
	d.Execute(TCPClose)
	time.Sleep(pause * time.Millisecond)
	d.Response()
	return nil
}

// SetMux sets the ESP8266/ESP32 current client TCP/UDP configuration for concurrent connections
// either single TCPMuxSingle or multiple TCPMuxMultiple (up to 4).
func (d *Device) SetMux(mode int) error {
	val := strconv.Itoa(mode)
	d.Set(TCPMultiple, val)
	time.Sleep(pause * time.Millisecond)
	d.Response()
	return nil
}

// GetMux returns the ESP8266/ESP32 current client TCP/UDP configuration for concurrent connections.
func (d *Device) GetMux() ([]byte, error) {
	d.Query(TCPMultiple)
	return d.Response(), nil
}

// SetTCPTransferMode sets the ESP8266/ESP32 current client TCP/UDP transfer mode.
// Either TCPTransferModeNormal or TCPTransferModeUnvarnished.
func (d *Device) SetTCPTransferMode(mode int) error {
	val := strconv.Itoa(mode)
	d.Set(TransmissionMode, val)
	time.Sleep(pause * time.Millisecond)
	d.Response()
	return nil
}

// GetTCPTransferMode returns the ESP8266/ESP32 current client TCP/UDP transfer mode.
func (d *Device) GetTCPTransferMode() []byte {
	d.Query(TransmissionMode)
	return d.Response()
}

// StartSocketSend gets the ESP8266/ESP32 ready to receive TCP/UDP socket data.
func (d *Device) StartSocketSend(size int) error {
	val := strconv.Itoa(size)
	d.Set(TCPSend, val)

	// TODO: wait until ">" is received, which indicates
	// ready to receive data
	d.Response()
	return nil
}

// EndSocketSend tell the ESP8266/ESP32 the TCP/UDP socket data sending is complete,
// and to return to command mode. This is only used in "unvarnished" raw mode.
func (d *Device) EndSocketSend() error {
	d.Write([]byte("+++"))

	// TODO: wait until ">" is received, which indicates
	// ready to receive data
	d.Response()
	return nil
}
