package rtl8720dn

import (
	"time"

	"tinygo.org/x/drivers/net"
)

func (r *RTL8720DN) ConnectToAccessPoint(ssid, pass string, timeout time.Duration) error {
	if len(ssid) == 0 {
		return net.ErrWiFiMissingSSID
	}

	return r.ConnectToAP(ssid, pass)
}

func (r *RTL8720DN) Disconnect() error {
	_, err := r.Rpc_wifi_disconnect()
	return err
}

func (r *RTL8720DN) GetClientIP() (string, error) {
	ip, _, _, err := r.GetIP()
	return ip.String(), err
}
