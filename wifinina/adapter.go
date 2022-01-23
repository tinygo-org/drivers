package wifinina

import (
	"time"

	"tinygo.org/x/drivers/net"
)

func (d *Device) ConnectToAccessPoint(ssid, pass string, timeout time.Duration) error {
	if len(ssid) == 0 {
		return net.ErrWiFiMissingSSID
	}

	start := time.Now()
	d.SetPassphrase(ssid, pass)

	for time.Since(start) < timeout {
		st, _ := d.GetConnectionStatus()
		if st == StatusConnected {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	return net.ErrWiFiConnectTimeout
}

func (d *Device) GetClientIP() (string, error) {
	ip, _, _, err := d.GetIP()
	return ip.String(), err
}
