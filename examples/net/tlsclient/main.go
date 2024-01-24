// This example uses TLS to send an HTTPS request to retrieve a webpage
//
// You shall see "strict-transport-security" header in the response,
// this confirms communication is indeed over HTTPS
//
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Strict-Transport-Security

//go:build ninafw || wioterminal

package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"machine"
	"net"
	"strings"
	"time"

	"tinygo.org/x/drivers/netlink"
	"tinygo.org/x/drivers/netlink/probe"
)

var (
	ssid string
	pass string
	// HTTPS server address to hit with a GET / request
	address string = "httpbin.org:443"
)

var conn net.Conn

// Wait for user to open serial console
func waitSerial() {
	for !machine.Serial.DTR() {
		time.Sleep(100 * time.Millisecond)
	}
}

func check(err error) {
	if err != nil {
		println("Hit an error:", err.Error())
		panic("BYE")
	}
}

func readResponse() {
	r := bufio.NewReader(conn)
	resp, err := io.ReadAll(r)
	check(err)
	println(string(resp))
}

func closeConnection() {
	conn.Close()
}

func dialConnection() {
	var err error

	println("\r\n---------------\r\nDialing TLS connection")
	conn, err = tls.Dial("tcp", address, nil)
	for ; err != nil; conn, err = tls.Dial("tcp", address, nil) {
		println("Connection failed:", err.Error())
		time.Sleep(5 * time.Second)
	}
	println("Connected!\r")
}

func makeRequest() {
	print("Sending HTTPS request...")
	w := bufio.NewWriter(conn)
	fmt.Fprintln(w, "GET /get HTTP/1.1")
	fmt.Fprintln(w, "Host:", strings.Split(address, ":")[0])
	fmt.Fprintln(w, "User-Agent: TinyGo")
	fmt.Fprintln(w, "Connection: close")
	fmt.Fprintln(w)
	check(w.Flush())
	println("Sent!\r\n\r")
}

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

	for i := 0; ; i++ {
		dialConnection()
		makeRequest()
		readResponse()
		closeConnection()
		println("--------", i, "--------\r\n")
		time.Sleep(10 * time.Second)
	}
}
