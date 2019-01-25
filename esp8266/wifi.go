package esp8266

import (
	"strconv"
	"time"
)

const (
	WifiModeClient = 1
	WifiModeAP     = 2
	WifiModeDual   = 3

	WifiAPSecurityOpen         = 1
	WifiAPSecurityWPA_PSK      = 2
	WifiAPSecurityWPA2_PSK     = 3
	WifiAPSecurityWPA_WPA2_PSK = 4
)

// GetWifiMode returns the ESP8266 wifi mode.
func (d *Device) GetWifiMode() []byte {
	d.Query(WifiMode)
	return d.Response()
}

// SetWifiMode sets the ESP8266 wifi mode.
func (d *Device) SetWifiMode(mode int) error {
	val := strconv.Itoa(mode)
	d.Set(WifiMode, val)
	time.Sleep(pause * time.Millisecond)
	d.Response()
	return nil
}

// Wifi Client

// GetConnectedAP returns the ESP8266 is currently connected to as a client.
func (d *Device) GetConnectedAP() []byte {
	d.Query(ConnectAP)
	return d.Response()
}

// ConnectToAP connects the ESP8266 to an access point.
// ws is the number of seconds to wait for connection.
func (d *Device) ConnectToAP(ssid, pwd string, ws int) error {
	val := "\"" + ssid + "\",\"" + pwd + "\""
	d.Set(ConnectAP, val)
	// TODO: a better way to wait for connect and check for up to ws seconds.
	time.Sleep(time.Duration(ws) * time.Second)
	d.Response()
	return nil
}

// DisconnectFromAP disconnects the ESP8266 from the current access point.
func (d *Device) DisconnectFromAP() error {
	d.Execute(Disconnect)
	time.Sleep(1000 * time.Millisecond)
	d.Response()
	return nil
}

// GetClientIP returns the ESP8266's current client IP addess when connected to an Access Point.
func (d *Device) GetClientIP() []byte {
	d.Query(SetStationIP)
	return d.Response()
}

// SetClientIP sets the ESP8266's current client IP addess when connected to an Access Point.
func (d *Device) SetClientIP(ipaddr string) []byte {
	val := "\"" + ipaddr + "\""
	d.Set(ConnectAP, val)
	time.Sleep(500 * time.Millisecond)
	d.Response()
	return nil
}

// Access Point

// GetSoftAPConfig returns the ESP8266 current configuration acting as an Access Point.
func (d *Device) GetSoftAPConfig() []byte {
	d.Query(SoftAPConfig)
	return d.Response()
}

// SetSoftAPConfig sets the ESP8266 current configuration acting as an Access Point.
func (d *Device) SetSoftAPConfig(ssid, pwd string, ch, security int) error {
	chval := strconv.Itoa(ch)
	ecnval := strconv.Itoa(security)
	val := "\"" + ssid + "\",\"" + pwd + "\"," + chval + "," + ecnval
	d.Set(ConnectAP, val)
	time.Sleep(1000 * time.Millisecond)
	d.Response()
	return nil
}

// GetSoftAPClients returns the ESP8266's current clients when acting as an Access Point.
func (d *Device) GetSoftAPClients() []byte {
	d.Query(ListConnectedIP)
	return d.Response()
}

// GetSoftAPIP returns the ESP8266's current IP addess when configured as an Access Point.
func (d *Device) GetSoftAPIP() []byte {
	d.Query(SetSoftAPIP)
	return d.Response()
}

// SetSoftAPIP sets the ESP8266's current IP addess when configured as an Access Point.
func (d *Device) SetSoftAPIP(ipaddr string) []byte {
	val := "\"" + ipaddr + "\""
	d.Set(SetSoftAPIP, val)
	time.Sleep(500 * time.Millisecond)
	d.Response()
	return nil
}
