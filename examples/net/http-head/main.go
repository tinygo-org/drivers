// This example gets an URL using http.Head().  URL scheme can be http or https.
//
// Note: It may be necessary to increase the stack size when using "net/http".
// Use the -stack-size=4KB command line option.
//
// Some targets (Arduino Nano33 IoT) don't have enough SRAM to run http.Head().
// Use the following for those targets:
//
//     examples/net/webclient (for HTTP)
//     examples/net/tlsclient (for HTTPS)

//go:build ninafw || wioterminal

package main

import (
	"bytes"
	"fmt"
	"log"
	"machine"
	"net/http"
	"time"

	"tinygo.org/x/drivers/netlink"
	"tinygo.org/x/drivers/netlink/probe"
)

var (
	ssid string
	pass string
	url  string = "https://httpbin.org"
)

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

	resp, err := http.Head(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	var buf bytes.Buffer
	if err := resp.Write(&buf); err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(buf.Bytes()))

	link.NetDisconnect()
}

// Wait for user to open serial console
func waitSerial() {
	for !machine.Serial.DTR() {
		time.Sleep(100 * time.Millisecond)
	}
}
