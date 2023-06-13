// This is an example of using the wifinina driver to implement a NTP client.
// It creates a UDP connection to request the current time and parse the
// response from a NTP server.
package main

import (
	"errors"
	"fmt"
	"machine"
	"runtime"
	"time"

	"tinygo.org/x/drivers/net"
	"tinygo.org/x/drivers/wifinina"
)

var (
	// access point info
	ssid string
	pass string
)

// IP address of the server aka "hub". Replace with your own info.
const ntpHost = "129.6.15.29"

const NTP_PACKET_SIZE = 48

// these are the default pins for the Arduino Nano33 IoT.
// change these to connect to a different UART or pins for the ESP8266/ESP32
var (

	// these are the default pins for the Arduino Nano33 IoT.
	spi = machine.NINA_SPI

	// this is the ESP chip that has the WIFININA firmware flashed on it
	adaptor *wifinina.Device
	b       = make([]byte, NTP_PACKET_SIZE)
)

func setup() {
	// Configure SPI for 8Mhz, Mode 0, MSB First
	spi.Configure(machine.SPIConfig{
		Frequency: 8 * 1e6,
		SDO:       machine.NINA_SDO,
		SDI:       machine.NINA_SDI,
		SCK:       machine.NINA_SCK,
	})

	adaptor = wifinina.New(spi,
		machine.NINA_CS,
		machine.NINA_ACK,
		machine.NINA_GPIO0,
		machine.NINA_RESETN)
	adaptor.Configure()
}

func main() {

	setup()

	waitSerial()

	connectToAP()
	displayIP()

	// now make UDP connection
	ip := net.ParseIP(ntpHost)
	raddr := &net.UDPAddr{IP: ip, Port: 123}
	laddr := &net.UDPAddr{Port: 2390}
	conn, err := net.DialUDP("udp", laddr, raddr)
	if err != nil {
		for {
			time.Sleep(time.Second)
			println(err)
		}
	}

	for {
		// send data
		println("Requesting NTP time...")
		t, err := getCurrentTime(conn)
		if err != nil {
			message("Error getting current time: %v", err)
		} else {
			message("NTP time: %v", t)
		}
		runtime.AdjustTimeOffset(-1 * int64(time.Since(t)))
		for i := 0; i < 10; i++ {
			message("Current time: %v", time.Now())
			time.Sleep(1 * time.Second)
		}
	}

}

// Wait for user to open serial console
func waitSerial() {
	for !machine.Serial.DTR() {
		time.Sleep(100 * time.Millisecond)
	}
}

func getCurrentTime(conn *net.UDPSerialConn) (time.Time, error) {
	if err := sendNTPpacket(conn); err != nil {
		return time.Time{}, err
	}
	clearBuffer()
	for now := time.Now(); time.Since(now) < time.Second; {
		time.Sleep(5 * time.Millisecond)
		if n, err := conn.Read(b); err != nil {
			return time.Time{}, fmt.Errorf("error reading UDP packet: %w", err)
		} else if n == 0 {
			continue // no packet received yet
		} else if n != NTP_PACKET_SIZE {
			return time.Time{}, fmt.Errorf("expected NTP packet size of %d: %d", NTP_PACKET_SIZE, n)
		}
		return parseNTPpacket(), nil
	}
	return time.Time{}, errors.New("no packet received after 1 second")
}

func sendNTPpacket(conn *net.UDPSerialConn) error {
	clearBuffer()
	b[0] = 0b11100011 // LI, Version, Mode
	b[1] = 0          // Stratum, or type of clock
	b[2] = 6          // Polling Interval
	b[3] = 0xEC       // Peer Clock Precision
	// 8 bytes of zero for Root Delay & Root Dispersion
	b[12] = 49
	b[13] = 0x4E
	b[14] = 49
	b[15] = 52
	if _, err := conn.Write(b); err != nil {
		return err
	}
	return nil
}

func parseNTPpacket() time.Time {
	// the timestamp starts at byte 40 of the received packet and is four bytes,
	// this is NTP time (seconds since Jan 1 1900):
	t := uint32(b[40])<<24 | uint32(b[41])<<16 | uint32(b[42])<<8 | uint32(b[43])
	const seventyYears = 2208988800
	return time.Unix(int64(t-seventyYears), 0)
}

func clearBuffer() {
	for i := range b {
		b[i] = 0
	}
}

const retriesBeforeFailure = 3

// connect to access point
func connectToAP() {
	time.Sleep(2 * time.Second)
	var err error
	for i := 0; i < retriesBeforeFailure; i++ {
		println("Connecting to " + ssid)
		err = adaptor.ConnectToAccessPoint(ssid, pass, 10*time.Second)
		if err == nil {
			println("Connected.")

			return
		}
	}

	// error connecting to AP
	failMessage(err.Error())
}

func displayIP() {
	ip, _, _, err := adaptor.GetIP()
	for ; err != nil; ip, _, _, err = adaptor.GetIP() {
		message(err.Error())
		time.Sleep(1 * time.Second)
	}
	message("IP address: " + ip.String())
}

func message(format string, args ...interface{}) {
	println(fmt.Sprintf(format, args...), "\r")
}

func failMessage(msg string) {
	for {
		println(msg)
		time.Sleep(1 * time.Second)
	}
}
