// This example posts an URL using http.Post().  URL scheme can be http or https.
//
// Note: It may be necessary to increase the stack size when using "net/http".
// Use the -stack-size=4KB command line option.
//
// Some targets (Arduino Nano33 IoT) don't have enough SRAM to run http.Post().
// Use the following for those targets:
//
//     examples/net/webclient (for HTTP)
//     examples/net/tlsclient (for HTTPS)

//go:build ninafw || wioterminal

package main

import (
	"bytes"
	"fmt"
	"io"
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

	path := "https://httpbin.org/post"
	data := []byte("{\"name\":\"John Doe\",\"occupation\":\"gardener\"}")

	resp, err := http.Post(path, "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(body))

	link.NetDisconnect()
}

// Wait for user to open serial console
func waitSerial() {
	for !machine.Serial.DTR() {
		time.Sleep(100 * time.Millisecond)
	}
}
