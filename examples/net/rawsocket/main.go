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
	"net"
	"strconv"
	"time"

	"tinygo.org/x/drivers"
)

var (
	ssid string
	pass string
	addr string = "10.0.0.100:8080"
)

var buf = &bytes.Buffer{}

func main() {

	waitSerial()

	if err := netdev.NetConnect(); err != nil {
		log.Fatal(err)
	}

	for {
		sendBatch()
		time.Sleep(500 * time.Millisecond)
	}
}

func sendBatch() {

	host, sport, _ := net.SplitHostPort(addr)
	ip := net.ParseIP(host).To4()
	port, _ := strconv.Atoi(sport)

	// make TCP connection
	message("---------------\r\nDialing TCP connection")
	fd, _ := netdev.Socket(drivers.AF_INET, drivers.SOCK_STREAM, drivers.IPPROTO_TCP)
	err := netdev.Connect(fd, "", ip, port)
	for ; err != nil; err = netdev.Connect(fd, "", ip, port) {
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
		if w, err = netdev.Send(fd, buf.Bytes(), 0, time.Time{}); err != nil {
			println("error:", err.Error(), "\r")
			break
		}
		n += w
	}

	buf.Reset()
	ms := time.Now().Sub(start).Milliseconds()
	fmt.Fprint(buf, "\nWrote ", n, " bytes in ", ms, " ms\r\n")
	message(buf.String())

	if _, err := netdev.Send(fd, buf.Bytes(), 0, time.Time{}); err != nil {
		println("error:", err.Error(), "\r")
	}

	println("Disconnecting TCP...")
	netdev.Close(fd)
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
