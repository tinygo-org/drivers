// This example opens a TCP connection and sends some data,
// for the purpose of testing speed and connectivity.
//
// You can open a server to accept connections from this program using:
//
// nc -w 5 -lk 8080
package main

import (
	"bytes"
	"fmt"
	"time"

	"tinygo.org/x/drivers/net"
)

var (
	// access point info
	ssid string
	pass string
)

// IP address of the server aka "hub". Replace with your own info.
const serverIP = ""

var buf = &bytes.Buffer{}

func main() {
	initAdaptor()

	connectToAP()

	for {
		sendBatch()
		time.Sleep(500 * time.Millisecond)
	}
	println("Done.")
}

func sendBatch() {

	// make TCP connection
	ip := net.ParseIP(serverIP)
	raddr := &net.TCPAddr{IP: ip, Port: 8080}
	laddr := &net.TCPAddr{Port: 8080}

	message("---------------\r\nDialing TCP connection")
	conn, err := net.DialTCP("tcp", laddr, raddr)
	for ; err != nil; conn, err = net.DialTCP("tcp", laddr, raddr) {
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
			continue
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

	// Right now this code is never reached. Need a way to trigger it...
	println("Disconnecting TCP...")
	conn.Close()
}

// connect to access point
func connectToAP() {
	time.Sleep(2 * time.Second)
	println("Connecting to " + ssid)
	err := adaptor.ConnectToAccessPoint(ssid, pass, 10*time.Second)
	if err != nil { // error connecting to AP
		for {
			println(err)
			time.Sleep(1 * time.Second)
		}
	}

	println("Connected.")

	time.Sleep(2 * time.Second)
	ip, err := adaptor.GetClientIP()
	for ; err != nil; ip, err = adaptor.GetClientIP() {
		message(err.Error())
		time.Sleep(1 * time.Second)
	}
	message(ip)
}

func message(msg string) {
	println(msg, "\r")
}
