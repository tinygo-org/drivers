// Package espat implements TCP/UDP wireless communication over serial
// with a separate ESP8266 or ESP32 board using the Espressif AT command set
// across a UART interface.
//
// In order to use this driver, the ESP8266/ESP32 must be flashed with firmware
// supporting the AT command set. Many ESP8266/ESP32 chips already have this firmware
// installed by default. You will need to install this firmware if you have an
// ESP8266 that has been flashed with NodeMCU (Lua) or Arduino firmware.
//
// AT Command Core repository:
// https://github.com/espressif/esp32-at
//
// Datasheet:
// https://www.espressif.com/sites/default/files/documentation/0a-esp8266ex_datasheet_en.pdf
//
// AT command set:
// https://www.espressif.com/sites/default/files/documentation/4a-esp8266_at_instruction_set_en.pdf
package espat // import "tinygo.org/x/drivers/espat"

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/net"
)

// Device wraps UART connection to the ESP8266/ESP32.
type Device struct {
	bus drivers.UART

	// command responses that come back from the ESP8266/ESP32
	response []byte

	// data received from a TCP/UDP connection forwarded by the ESP8266/ESP32
	socketdata []byte
}

// ActiveDevice is the currently configured Device in use. There can only be one.
var ActiveDevice *Device

// New returns a new espat driver. Pass in a fully configured UART bus.
func New(b drivers.UART) *Device {
	return &Device{bus: b, response: make([]byte, 512), socketdata: make([]byte, 0, 1024)}
}

// Configure sets up the device for communication.
func (d Device) Configure() {
	ActiveDevice = &d
	net.ActiveDevice = ActiveDevice
}

// Connected checks if there is communication with the ESP8266/ESP32.
func (d *Device) Connected() bool {
	d.Execute(Test)

	// handle response here, should include "OK"
	_, err := d.Response(100)
	if err != nil {
		return false
	}
	return true
}

// Write raw bytes to the UART.
func (d *Device) Write(b []byte) (n int, err error) {
	return d.bus.Write(b)
}

// Read raw bytes from the UART.
func (d *Device) Read(b []byte) (n int, err error) {
	return d.bus.Read(b)
}

// how long in milliseconds to pause after sending AT commands
const pause = 300

// Execute sends an AT command to the ESP8266/ESP32.
func (d Device) Execute(cmd string) error {
	_, err := d.Write([]byte("AT" + cmd + "\r\n"))
	return err
}

// Query sends an AT command to the ESP8266/ESP32 that returns the
// current value for some configuration parameter.
func (d Device) Query(cmd string) (string, error) {
	_, err := d.Write([]byte("AT" + cmd + "?\r\n"))
	return "", err
}

// Set sends an AT command with params to the ESP8266/ESP32 for a
// configuration value to be set.
func (d Device) Set(cmd, params string) error {
	_, err := d.Write([]byte("AT" + cmd + "=" + params + "\r\n"))
	return err
}

// Version returns the ESP8266/ESP32 firmware version info.
func (d Device) Version() []byte {
	d.Execute(Version)
	r, err := d.Response(100)
	if err != nil {
		return []byte("unknown")
	}
	return r
}

// Echo sets the ESP8266/ESP32 echo setting.
func (d Device) Echo(set bool) {
	if set {
		d.Execute(EchoConfigOn)
	} else {
		d.Execute(EchoConfigOff)
	}
	// TODO: check for success
	d.Response(100)
}

// Reset restarts the ESP8266/ESP32 firmware. Due to how the baud rate changes,
// this messes up communication with the ESP8266/ESP32 module. So make sure you know
// what you are doing when you call this.
func (d Device) Reset() {
	d.Execute(Restart)
	d.Response(100)
}

// ReadSocket returns the data that has already been read in from the responses.
func (d *Device) ReadSocket(b []byte) (n int, err error) {
	// make sure no data in buffer
	d.Response(300)

	count := len(b)
	if len(b) >= len(d.socketdata) {
		// copy it all, then clear socket data
		count = len(d.socketdata)
		copy(b, d.socketdata[:count])
		d.socketdata = d.socketdata[:0]
	} else {
		// copy all we can, then keep the remaining socket data around
		copy(b, d.socketdata[:count])
		copy(d.socketdata, d.socketdata[count:])
		d.socketdata = d.socketdata[:len(d.socketdata)-count]
	}

	return count, nil
}

// Response gets the next response bytes from the ESP8266/ESP32.
// The call will retry for up to timeout milliseconds before returning nothing.
func (d *Device) Response(timeout int) ([]byte, error) {
	// read data
	var size int
	var start, end int
	pause := 100 // pause to wait for 100 ms
	retries := timeout / pause

	for {
		size = d.bus.Buffered()

		if size > 0 {
			end += size
			d.bus.Read(d.response[start:end])

			// if "+IPD" then read socket data
			if strings.Contains(string(d.response[:end]), "+IPD") {
				// handle socket data
				return nil, d.parseIPD(end)
			}

			// if "OK" then the command worked
			if strings.Contains(string(d.response[:end]), "OK") {
				return d.response[start:end], nil
			}

			// if "Error" then the command failed
			if strings.Contains(string(d.response[:end]), "ERROR") {
				return d.response[start:end], errors.New("response error:" + string(d.response[start:end]))
			}

			// if anything else, then keep reading data in?
			start = end
		}

		// wait longer?
		retries--
		if retries == 0 {
			return nil, errors.New("response timeout error:" + string(d.response[start:end]))
		}

		time.Sleep(time.Duration(pause) * time.Millisecond)
	}
}

func (d *Device) parseIPD(end int) error {
	// find the "+IPD," to get length
	s := strings.Index(string(d.response[:end]), "+IPD,")

	// find the ":"
	e := strings.Index(string(d.response[:end]), ":")

	// find the data length
	val := string(d.response[s+5 : e])

	// TODO: verify count
	_, err := strconv.Atoi(val)
	if err != nil {
		// not expected data here. what to do?
		return err
	}

	// load up the socket data
	d.socketdata = append(d.socketdata, d.response[e+1:end]...)
	return nil
}

// IsSocketDataAvailable returns of there is socket data available
func (d *Device) IsSocketDataAvailable() bool {
	return len(d.socketdata) > 0 || d.bus.Buffered() > 0
}
