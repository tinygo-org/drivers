// This example opens a TCP connection and sends some data, for the purpose of
// testing speed and connectivity.
//
// You can open a server to accept connections from this program using:
//
// nc -lk 8080

//go:build ninafw || wioterminal || challenger_rp2040 || pico

package main

import (
	"bytes"
	"fmt"
	"log"
	"machine"
	"net"
	"time"

	"tinygo.org/x/drivers/netlink"
	"tinygo.org/x/drivers/netlink/probe"
)

var (
	ssid string
	pass string
	addr string = "10.0.0.100:8080"
)

var buf = &bytes.Buffer{}

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

	for {
		sendBatch()
		time.Sleep(500 * time.Millisecond)
	}
}

func sendBatch() {

	// make TCP connection
	message("---------------\r\nDialing TCP connection")
	conn, err := net.Dial("tcp", addr)
	for ; err != nil; conn, err = net.Dial("tcp", addr) {
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
		if w, err = conn.Write(buf.Bytes()); err != nil {
			println("error:", err.Error(), "\r")
			break
		}
		n += w
	}

	buf.Reset()
	ms := time.Now().Sub(start).Milliseconds()
	fmt.Fprint(buf, "\nWrote ", n, " bytes in ", ms, " ms\r\n")
	message(buf.String())

	if _, err := conn.Write(buf.Bytes()); err != nil {
		println("error:", err.Error(), "\r")
	}

	println("Disconnecting TCP...")
	conn.Close()
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
