package espat

import (
	"time"

	"tinygo.org/x/drivers/net"
)

func (d *Device) ConnectToAccessPoint(ssid, pass string, timeout time.Duration) error {
	if len(ssid) == 0 {
		return net.ErrWiFiMissingSSID
	}

	d.SetWifiMode(WifiModeClient)
	return d.ConnectToAP(ssid, pass, 10)
}

func (d *Device) Disconnect() error {
	return d.DisconnectFromAP()
}
