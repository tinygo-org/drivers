// Package esp8266 implements TCP/UDP communication over serial
// with a separate Wifi ESP8266 board using the Espressif AT command set
// across a UART interface.
//
// More information at:
// https://github.com/espressif/ESP8266_AT/wiki
//
// In order to use this driver, the ESP8266 must be flashed with firmware
// supporting the AT command set. Many ESP8266 chips already have this firmware
// installed by default. You will need to install this firmware if you have an
// ESP8266 that has been flashed with NodeMCU (Lua) or Arduino firmware.
//
// Datasheet:
// https://www.espressif.com/sites/default/files/documentation/0a-esp8266ex_datasheet_en.pdf
//
package esp8266

import (
	"machine"
	"time"
)

// Device wraps UART connection to the ESP8266.
type Device struct {
	bus      machine.UART
	response []byte
}

// New returns a new esp8266-wifi driver. Pass in a fully configured UART bus.
func New(b machine.UART) Device {
	return Device{bus: b, response: make([]byte, 1024)}
}

// Configure sets up the device for communication.
func (d Device) Configure() {
}

// Connected checks if there is communication with the ESP8266.
func (d Device) Connected() bool {
	d.Execute(Test)

	// TODO: handle response here, should include "OK"
	// aka strings.Contains(string(r), "OK\r\n")
	r := d.Response()
	if len(r) > 0 {
		return true
	}
	return false
}

// Write raw bytes to the UART.
func (d Device) Write(b []byte) (n int, err error) {
	return d.bus.Write(b)
}

// Read raw bytes from the UART.
func (d Device) Read(b []byte) (n int, err error) {
	return d.bus.Read(b)
}

// Response gets the next response bytes from the ESP8266.
func (d Device) Response() []byte {
	i := 0

	retries := 0
	for {
		for d.bus.Buffered() > 0 {
			data, _ := d.bus.ReadByte()
			d.response[i] = data
			i++
		}
		retries++
		if retries > 2 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	return d.response[:i]
}
