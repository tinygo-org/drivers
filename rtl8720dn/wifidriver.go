package rtl8720dn

import (
	"time"

	"tinygo.org/x/drivers"
)

func (r *RTL8720DN) ConnectToAccessPoint(ssid, pass string, timeout time.Duration) error {
	if len(ssid) == 0 {
		return drivers.ErrWiFiMissingSSID
	}

	return r.ConnectToAP(ssid, pass)
}

func (r *RTL8720DN) Disconnect() error {
	_, err := r.Rpc_wifi_disconnect()
	return err
}
