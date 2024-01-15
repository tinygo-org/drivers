// This example uses TCP to send an HTTP request to retrieve a webpage.  The
// HTTP request is hand-rolled to avoid the overhead of using http.Get() from
// the "net/http" package.  See example/net/http-get for the full http.Get()
// functionality.
//
// Example HTTP server:
// ---------------------------------------------------------------------------
// package main
//
// import "net/http"
//
// func main() {
//        http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
//                w.Write([]byte("hello"))
//        })
//        http.ListenAndServe(":8080", nil)
// }
// ---------------------------------------------------------------------------

//go:build ninafw || wioterminal

package main

import (
	"bufio"
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
	// HTTP server address to hit with a GET / request
	address string = "10.0.0.100:8080"
)

var conn net.Conn

// Wait for user to open serial console
func waitSerial() {
	for !machine.Serial.DTR() {
		time.Sleep(100 * time.Millisecond)
	}
}

func dialConnection() {
	var err error

	println("\r\n---------------\r\nDialing TCP connection")
	conn, err = net.Dial("tcp", address)
	for ; err != nil; conn, err = net.Dial("tcp", address) {
		println("Connection failed:", err.Error())
		time.Sleep(5 * time.Second)
	}
	println("Connected!\r")
}

func check(err error) {
	if err != nil {
		println("Hit an error:", err.Error())
		panic("BYE")
	}
}

func makeRequest() {
	println("Sending HTTP request...")
	w := bufio.NewWriter(conn)
	fmt.Fprintln(w, "GET / HTTP/1.1")
	fmt.Fprintln(w, "Host:", strings.Split(address, ":")[0])
	fmt.Fprintln(w, "User-Agent: TinyGo")
	fmt.Fprintln(w, "Connection: close")
	fmt.Fprintln(w)
	check(w.Flush())
	println("Sent!\r\n\r")
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
