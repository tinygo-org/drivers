// This example opens a TCP connection and sends some data using netdev sockets.
//
// You can open a server to accept connections from this program using:
//
// nc -lk 8080

//go:build ninafw || wioterminal || challenger_rp2040

package main

import (
	"bytes"
	"fmt"
	"log"
	"machine"
	"net/netip"
	"time"

	"tinygo.org/x/drivers/netdev"
	"tinygo.org/x/drivers/netlink"
	"tinygo.org/x/drivers/netlink/probe"
)

var (
	ssid string
	pass string
	addr string = "10.0.0.100:8080"
)

var buf = &bytes.Buffer{}
var link netlink.Netlinker
var dev netdev.Netdever

func main() {

	waitSerial()

	link, dev = probe.Probe()

	err := link.NetConnect(&netlink.ConnectParams{
		Ssid:       ssid,
		Passphrase: pass,
	})
	if err != nil {
		log.Fatal(err)
	}

	for {
		sendBatch()
		time.Sleep(500 * time.Millisecond)
	}
}

func sendBatch() {

	addrPort, _ := netip.ParseAddrPort(addr)

	// make TCP connection
	message("---------------\r\nDialing TCP connection")
	fd, _ := dev.Socket(netdev.AF_INET, netdev.SOCK_STREAM, netdev.IPPROTO_TCP)
	err := dev.Connect(fd, "", addrPort)
	for ; err != nil; err = dev.Connect(fd, "", addrPort) {
		message(err.Error())
		time.Sleep(5 * time.Second)
	}

	n := 0
	w := 0
	start := time.Now()

	// send data
	message("Sending data")

	for i := 0; i < 1000; i++ {
		buf.Reset()
		fmt.Fprint(buf,
			"\r---------------------------- i == ", i, " ----------------------------"+
				"\r---------------------------- i == ", i, " ----------------------------")
		if w, err = dev.Send(fd, buf.Bytes(), 0, time.Time{}); err != nil {
			println("error:", err.Error(), "\r")
			break
		}
		n += w
	}

	buf.Reset()
	ms := time.Now().Sub(start).Milliseconds()
	fmt.Fprint(buf, "\nWrote ", n, " bytes in ", ms, " ms\r\n")
	message(buf.String())

	if _, err := dev.Send(fd, buf.Bytes(), 0, time.Time{}); err != nil {
		println("error:", err.Error(), "\r")
	}

	println("Disconnecting TCP...")
	dev.Close(fd)
}

func message(msg string) {
	println(msg, "\r")
}

// Wait for user to open serial console
func waitSerial() {
	for !machine.Serial.DTR() {
		time.Sleep(100 * time.Millisecond)
	}
}
