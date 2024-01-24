// This example is a websocket client.  It connects to a websocket server
// which echos messages back to the client.  For server, see
//
//     https://pkg.go.dev/golang.org/x/net/websocket#example-Handler
//
// Note: It may be necessary to increase the stack size when using
// "golang.org/x/net/websocket".  Use the -stack-size=4KB command line option.

//go:build ninafw || wioterminal

package main

import (
	"fmt"
	"log"
	"machine"
	"time"

	"golang.org/x/net/websocket"
	"tinygo.org/x/drivers/netlink"
	"tinygo.org/x/drivers/netlink/probe"
)

var (
	ssid string
	pass string
	url  string = "ws://10.0.0.100:8080/echo"
)

// Wait for user to open serial console
func waitSerial() {
	for !machine.Serial.DTR() {
		time.Sleep(100 * time.Millisecond)
	}
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

	origin := "http://localhost/"
	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := ws.Write([]byte("hello, world!\n")); err != nil {
		log.Fatal(err)
	}
	var msg = make([]byte, 512)
	var n int
	if n, err = ws.Read(msg); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Received: %s", msg[:n])
}
