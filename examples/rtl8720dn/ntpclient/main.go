// This is an example of using the rtl8720dn driver to implement a NTP client.
// It creates a UDP connection to request the current time and parse the
// response from a NTP server.
package main

import (
	"machine"

	"errors"
	"fmt"
	"runtime"
	"time"

	"tinygo.org/x/drivers/net"
	"tinygo.org/x/drivers/rtl8720dn"
)

// IP address of the server aka "hub". Replace with your own info.
// You can override the setting with the init() in another source code.
// func init() {
//    ssid = "your-ssid"
//    pass = "your-password"
//    ntpHost = "129.6.15.29"
//    debug = true
// }

var (
	ssid    string
	pass    string
	ntpHost = "129.6.15.29"
	debug   = false
)

const NTP_PACKET_SIZE = 48

var b = make([]byte, NTP_PACKET_SIZE)

func main() {
	err := run()
	for err != nil {
		fmt.Printf("error: %s\r\n", err.Error())
		time.Sleep(5 * time.Second)
	}
}

func run() error {
	adaptor := rtl8720dn.New(machine.UART3, machine.PB24, machine.PC24, machine.RTL8720D_CHIP_PU)
	adaptor.Debug(debug)
	adaptor.Configure()

	err := adaptor.ConnectToAccessPoint(ssid, pass, 10*time.Second)
	if err != nil {
		return err
	}

	ip, subnet, gateway, err := adaptor.GetIP()
	if err != nil {
		return err
	}
	fmt.Printf("IP Address : %s\r\n", ip)
	fmt.Printf("Mask       : %s\r\n", subnet)
	fmt.Printf("Gateway    : %s\r\n", gateway)

	// now make UDP connection
	hip := net.ParseIP(ntpHost)
	raddr := &net.UDPAddr{IP: hip, Port: 123}
	laddr := &net.UDPAddr{Port: 2390}
	conn, err := net.DialUDP("udp", laddr, raddr)
	if err != nil {
		return err
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

func message(format string, args ...interface{}) {
	println(fmt.Sprintf(format, args...), "\r")
}
