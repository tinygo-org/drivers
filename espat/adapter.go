package espat

import (
	"time"

	"tinygo.org/x/drivers/net"
)

func (d *Device) ConnectToAccessPoint(ssid, pass string, timeout time.Duration) error {
	if len(ssid) == 0 {
		return net.ErrWiFiMissingSSID
	}

	if err := d.SetWifiMode(WifiModeClient); err != nil {
		return err
	}

	return d.ConnectToAP(ssid, pass, 10)
}

func (d *Device) Disconnect() error {
	return d.DisconnectFromAP()
}
