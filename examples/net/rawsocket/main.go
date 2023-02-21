// This example opens a TCP connection and sends some data using raw netdev sockets.
//
// You can open a server to accept connections from this program using:
//
// nc -lk 8080

package main

import (
	"bytes"
	"fmt"
	"log"
	"machine"
	"net/netdev"
	"strconv"
	"strings"
	"time"
)

var (
	ssid string
	pass string
	addr string = "10.0.0.100:8080"
)

var buf = &bytes.Buffer{}

func main() {

	waitSerial()

	if err := NetConnect(); err != nil {
		log.Fatal(err)
	}

	for {
		sendBatch()
		time.Sleep(500 * time.Millisecond)
	}
}

func sendBatch() {

	parts := strings.Split(addr, ":")
	ip := netdev.ParseIP(parts[0])
	port, _ := strconv.Atoi(parts[1])
	sockAddr := netdev.NewSockAddr("", netdev.Port(port), ip)

	// make TCP connection
	message("---------------\r\nDialing TCP connection")
	sock, _ := dev.Socket(netdev.AF_INET, netdev.SOCK_STREAM, netdev.IPPROTO_TCP)
	err := dev.Connect(sock, sockAddr)
	for ; err != nil; err = dev.Connect(sock, sockAddr) {
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
		if w, err = dev.Send(sock, buf.Bytes(), 0, 0); err != nil {
			println("error:", err.Error(), "\r")
			break
		}
		n += w
	}

	buf.Reset()
	ms := time.Now().Sub(start).Milliseconds()
	fmt.Fprint(buf, "\nWrote ", n, " bytes in ", ms, " ms\r\n")
	message(buf.String())

	if _, err := dev.Send(sock, buf.Bytes(), 0, 0); err != nil {
		println("error:", err.Error(), "\r")
	}

	println("Disconnecting TCP...")
	dev.Close(sock)
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
