// This example is a websocket server.  It listens for websocket clients
// to connect and echos messages back to the client.  For client, see
//
//     https://pkg.go.dev/golang.org/x/net/websocket#example-Dial
//
// Note: It may be necessary to increase the stack size when using
// "golang.org/x/net/websocket".  Use the -stack-size=4KB command line option.

//go:build ninafw || wioterminal

package main

import (
	"io"
	"log"
	"machine"
	"net/http"
	"time"

	"golang.org/x/net/websocket"
	"tinygo.org/x/drivers/netlink"
	"tinygo.org/x/drivers/netlink/probe"
)

var (
	ssid string
	pass string
	port string = ":8080"
)

// Echo the data received on the WebSocket.
func EchoServer(ws *websocket.Conn) {
	io.Copy(ws, ws)
}

// Wait for user to open serial console
func waitSerial() {
	for !machine.Serial.DTR() {
		time.Sleep(100 * time.Millisecond)
	}
}

// This example demonstrates a trivial echo server.
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

	http.Handle("/echo", websocket.Handler(EchoServer))
	err = http.ListenAndServe(port, nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
