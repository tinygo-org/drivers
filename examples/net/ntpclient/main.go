// This is an example of an NTP client.
//
// It creates a UDP connection to request the current time and parse the
// response from a NTP server.  The system time is set to NTP time.

//go:build ninafw || wioterminal || challenger_rp2040

package main

import (
	"fmt"
	"io"
	"log"
	"machine"
	"net"
	"runtime"
	"time"

	"tinygo.org/x/drivers/netlink"
	"tinygo.org/x/drivers/netlink/probe"
)

var (
	ssid string
	pass string
	// IP address of the server aka "hub". Replace with your own info.
	ntpHost string = "0.pool.ntp.org:123"
)

const NTP_PACKET_SIZE = 48

var response = make([]byte, NTP_PACKET_SIZE)

func main() {

	waitSerial()

	link, _ := probe.Probe()

	err := link.NetConnect(&netlink.ConnectParams{
		Ssid:       ssid,
		Passphrase: pass,
	})
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.Dial("udp", ntpHost)
	if err != nil {
		log.Fatal(err)
	}

	println("Requesting NTP time...")

	t, err := getCurrentTime(conn)
	if err != nil {
		log.Fatal(fmt.Sprintf("Error getting current time: %v", err))
	} else {
		message("NTP time: %v", t)
	}

	conn.Close()
	link.NetDisconnect()

	runtime.AdjustTimeOffset(-1 * int64(time.Since(t)))

	for {
		message("Current time: %v", time.Now())
		time.Sleep(time.Minute)
	}
}

// Wait for user to open serial console
func waitSerial() {
	for !machine.Serial.DTR() {
		time.Sleep(100 * time.Millisecond)
	}
}

func getCurrentTime(conn net.Conn) (time.Time, error) {
	if err := sendNTPpacket(conn); err != nil {
		return time.Time{}, err
	}

	n, err := conn.Read(response)
	if err != nil && err != io.EOF {
		return time.Time{}, err
	}
	if n != NTP_PACKET_SIZE {
		return time.Time{}, fmt.Errorf("expected NTP packet size of %d: %d", NTP_PACKET_SIZE, n)
	}

	return parseNTPpacket(response), nil
}

func sendNTPpacket(conn net.Conn) error {
	var request = [48]byte{
		0xe3,
	}

	_, err := conn.Write(request[:])
	return err
}

func parseNTPpacket(r []byte) time.Time {
	// the timestamp starts at byte 40 of the received packet and is four bytes,
	// this is NTP time (seconds since Jan 1 1900):
	t := uint32(r[40])<<24 | uint32(r[41])<<16 | uint32(r[42])<<8 | uint32(r[43])
	const seventyYears = 2208988800
	return time.Unix(int64(t-seventyYears), 0)
}

func message(format string, args ...interface{}) {
	println(fmt.Sprintf(format, args...), "\r")
}
