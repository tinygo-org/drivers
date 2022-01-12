package wifinina

import (
	"time"

	"tinygo.org/x/drivers"
)

func (d *Device) ConnectToAccessPoint(ssid, pass string, timeout time.Duration) error {
	if len(ssid) == 0 {
		return drivers.ErrWiFiMissingSSID
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

	return drivers.ErrWiFiConnectTimeout
}
