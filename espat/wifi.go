package espat

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

// GetWifiMode returns the ESP8266/ESP32 wifi mode.
func (d *Device) GetWifiMode() []byte {
	d.Query(WifiMode)
	return d.Response()
}

// SetWifiMode sets the ESP8266/ESP32 wifi mode.
func (d *Device) SetWifiMode(mode int) error {
	val := strconv.Itoa(mode)
	d.Set(WifiMode, val)
	time.Sleep(pause * time.Millisecond)
	d.Response()
	return nil
}

// Wifi Client

// GetConnectedAP returns the ESP8266/ESP32 is currently connected to as a client.
func (d *Device) GetConnectedAP() []byte {
	d.Query(ConnectAP)
	return d.Response()
}

// ConnectToAP connects the ESP8266/ESP32 to an access point.
// ws is the number of seconds to wait for connection.
func (d *Device) ConnectToAP(ssid, pwd string, ws int) error {
	val := "\"" + ssid + "\",\"" + pwd + "\""
	d.Set(ConnectAP, val)
	// TODO: a better way to wait for connect and check for up to ws seconds.
	time.Sleep(time.Duration(ws) * time.Second)
	d.Response()
	return nil
}

// DisconnectFromAP disconnects the ESP8266/ESP32 from the current access point.
func (d *Device) DisconnectFromAP() error {
	d.Execute(Disconnect)
	time.Sleep(1000 * time.Millisecond)
	d.Response()
	return nil
}

// GetClientIP returns the ESP8266/ESP32 current client IP addess when connected to an Access Point.
func (d *Device) GetClientIP() string {
	d.Query(SetStationIP)
	return string(d.Response())
}

// SetClientIP sets the ESP8266/ESP32 current client IP addess when connected to an Access Point.
func (d *Device) SetClientIP(ipaddr string) []byte {
	val := "\"" + ipaddr + "\""
	d.Set(ConnectAP, val)
	time.Sleep(500 * time.Millisecond)
	d.Response()
	return nil
}

// Access Point

// GetAPConfig returns the ESP8266/ESP32 current configuration when acting as an Access Point.
func (d *Device) GetAPConfig() string {
	d.Query(SoftAPConfigCurrent)
	return string(d.Response())
}

// SetAPConfig sets the ESP8266/ESP32 current configuration when acting as an Access Point.
// ch indicates which radiochannel to use. security should be one of the const values
// such as WifiAPSecurityOpen etc.
func (d *Device) SetAPConfig(ssid, pwd string, ch, security int) error {
	chval := strconv.Itoa(ch)
	ecnval := strconv.Itoa(security)
	val := "\"" + ssid + "\",\"" + pwd + "\"," + chval + "," + ecnval
	d.Set(SoftAPConfigCurrent, val)
	time.Sleep(1000 * time.Millisecond)
	d.Response()
	return nil
}

// GetAPClients returns the ESP8266/ESP32 current clients when acting as an Access Point.
func (d *Device) GetAPClients() string {
	d.Query(ListConnectedIP)
	return string(d.Response())
}

// GetAPIP returns the ESP8266/ESP32 current IP addess when configured as an Access Point.
func (d *Device) GetAPIP() string {
	d.Query(SetSoftAPIPCurrent)
	return string(d.Response())
}

// SetAPIP sets the ESP8266/ESP32 current IP addess when configured as an Access Point.
func (d *Device) SetAPIP(ipaddr string) error {
	val := "\"" + ipaddr + "\""
	d.Set(SetSoftAPIPCurrent, val)
	time.Sleep(500 * time.Millisecond)
	d.Response()
	return nil
}

// GetAPConfigFlash returns the ESP8266/ESP32 current configuration acting as an Access Point
// from flash storage. These settings are those used after a reset.
func (d *Device) GetAPConfigFlash() string {
	d.Query(SoftAPConfigFlash)
	return string(d.Response())
}

// SetAPConfigFlash sets the ESP8266/ESP32 current configuration acting as an Access Point,
// and saves them to flash storage. These settings will be used after a reset.
// ch indicates which radiochannel to use. security should be one of the const values
// such as WifiAPSecurityOpen etc.
func (d *Device) SetAPConfigFlash(ssid, pwd string, ch, security int) error {
	chval := strconv.Itoa(ch)
	ecnval := strconv.Itoa(security)
	val := "\"" + ssid + "\",\"" + pwd + "\"," + chval + "," + ecnval
	d.Set(SoftAPConfigFlash, val)
	time.Sleep(1000 * time.Millisecond)
	d.Response()
	return nil
}

// GetAPIPFlash returns the ESP8266/ESP32 IP address as saved to flash storage.
// This is the IP address that will be used after a reset.
func (d *Device) GetAPIPFlash() string {
	d.Query(SetSoftAPIPFlash)
	return string(d.Response())
}

// SetAPIPFlash sets the ESP8266/ESP32 current IP addess when configured as an Access Point.
// The IP will be saved to flash storage, and will be used after a reset.
func (d *Device) SetAPIPFlash(ipaddr string) error {
	val := "\"" + ipaddr + "\""
	d.Set(SetSoftAPIPFlash, val)
	time.Sleep(500 * time.Millisecond)
	d.Response()
	return nil
}
