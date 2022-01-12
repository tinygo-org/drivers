package drivers

import (
	"errors"
	"time"
)

var (
	ErrWiFiMissingSSID    = errors.New("missing SSID")
	ErrWiFiConnectTimeout = errors.New("WiFi connect timeout")
)

type WiFiDriver interface {
	ConnectToAccessPoint(ssid, pass string, timeout time.Duration) error
	Disconnect() error
}
