package espat

import (
	"strconv"
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
func (d *Device) GetWifiMode() ([]byte, error) {
	d.Query(WifiMode)
	return d.Response(100)
}

// SetWifiMode sets the ESP8266/ESP32 wifi mode.
func (d *Device) SetWifiMode(mode int) error {
	val := strconv.Itoa(mode)
	d.Set(WifiMode, val)
	_, err := d.Response(pause)
	return err
}

// Wifi Client

// GetConnectedAP returns the ESP8266/ESP32 is currently connected to as a client.
func (d *Device) GetConnectedAP() ([]byte, error) {
	d.Query(ConnectAP)
	return d.Response(100)
}

// ConnectToAP connects the ESP8266/ESP32 to an access point.
// ws is the number of seconds to wait for connection.
func (d *Device) ConnectToAP(ssid, pwd string, ws int) error {
	val := "\"" + ssid + "\",\"" + pwd + "\""
	d.Set(ConnectAP, val)

	_, err := d.Response(ws * 1000)
	return err
}

// DisconnectFromAP disconnects the ESP8266/ESP32 from the current access point.
func (d *Device) DisconnectFromAP() error {
	d.Execute(Disconnect)
	_, err := d.Response(1000)
	return err
}

// GetClientIP returns the ESP8266/ESP32 current client IP addess when connected to an Access Point.
func (d *Device) GetClientIP() (string, error) {
	d.Query(SetStationIP)
	r, err := d.Response(1000)
	return string(r), err
}

// SetClientIP sets the ESP8266/ESP32 current client IP addess when connected to an Access Point.
func (d *Device) SetClientIP(ipaddr string) error {
	val := "\"" + ipaddr + "\""
	d.Set(ConnectAP, val)
	_, err := d.Response(500)
	return err
}

// Access Point

// GetAPConfig returns the ESP8266/ESP32 current configuration when acting as an Access Point.
func (d *Device) GetAPConfig() (string, error) {
	d.Query(SoftAPConfigCurrent)
	r, err := d.Response(100)
	return string(r), err
}

// SetAPConfig sets the ESP8266/ESP32 current configuration when acting as an Access Point.
// ch indicates which radiochannel to use. security should be one of the const values
// such as WifiAPSecurityOpen etc.
func (d *Device) SetAPConfig(ssid, pwd string, ch, security int) error {
	chval := strconv.Itoa(ch)
	ecnval := strconv.Itoa(security)
	val := "\"" + ssid + "\",\"" + pwd + "\"," + chval + "," + ecnval
	d.Set(SoftAPConfigCurrent, val)
	_, err := d.Response(1000)
	return err
}

// GetAPClients returns the ESP8266/ESP32 current clients when acting as an Access Point.
func (d *Device) GetAPClients() (string, error) {
	d.Query(ListConnectedIP)
	r, err := d.Response(100)
	return string(r), err
}

// GetAPIP returns the ESP8266/ESP32 current IP addess when configured as an Access Point.
func (d *Device) GetAPIP() (string, error) {
	d.Query(SetSoftAPIPCurrent)
	r, err := d.Response(100)
	return string(r), err
}

// SetAPIP sets the ESP8266/ESP32 current IP addess when configured as an Access Point.
func (d *Device) SetAPIP(ipaddr string) error {
	val := "\"" + ipaddr + "\""
	d.Set(SetSoftAPIPCurrent, val)
	_, err := d.Response(500)
	return err
}

// GetAPConfigFlash returns the ESP8266/ESP32 current configuration acting as an Access Point
// from flash storage. These settings are those used after a reset.
func (d *Device) GetAPConfigFlash() (string, error) {
	d.Query(SoftAPConfigFlash)
	r, err := d.Response(100)
	return string(r), err
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
	_, err := d.Response(1000)
	return err
}

// GetAPIPFlash returns the ESP8266/ESP32 IP address as saved to flash storage.
// This is the IP address that will be used after a reset.
func (d *Device) GetAPIPFlash() (string, error) {
	d.Query(SetSoftAPIPFlash)
	r, err := d.Response(100)
	return string(r), err
}

// SetAPIPFlash sets the ESP8266/ESP32 current IP addess when configured as an Access Point.
// The IP will be saved to flash storage, and will be used after a reset.
func (d *Device) SetAPIPFlash(ipaddr string) error {
	val := "\"" + ipaddr + "\""
	d.Set(SetSoftAPIPFlash, val)
	_, err := d.Response(500)
	return err
}
