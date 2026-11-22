// Package conf provides the interface used to modify STUSB4500 configuration.
package conf // import "tinygo.org/x/drivers/stusb4500/conf"

import (
	"machine"
)

type EventCallback func()
type ErrorCallback func(error)

type Configuration struct {
	ResetPin  machine.Pin
	AlertPin  machine.Pin
	AttachPin machine.Pin

	OnInitFail     ErrorCallback // I2C initialization failure
	OnResetFail    ErrorCallback // I2C reconnection to STUSB4500 failed
	OnConnect      EventCallback // I2C connection to STUSB4500 succeeded
	OnConnectFail  ErrorCallback // I2C connection to STUSB4500 failed
	OnError        ErrorCallback // I2C or STUSB4500 runtime error
	OnCableAttach  EventCallback // USB Type-C cable connected
	OnCableDetach  EventCallback // USB Type-C cable disconnected
	OnCapabilities EventCallback // USB PD capabilities received from source

	USBPDTimeout int // -1 = unlimited (never timeout)
}

const (
	DefaultUSBPDTimeout = 200 // messages sent
)
