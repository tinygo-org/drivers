// This example posts an URL using http.PostForm().  URL scheme can be http or https.
//
// Note: It may be necessary to increase the stack size when using "net/http".
// Use the -stack-size=4KB command line option.
//
// Some targets (Arduino Nano33 IoT) don't have enough SRAM to run
// http.PostForm().  Use the following for those targets:
//
//     examples/net/webclient (for HTTP)
//     examples/net/tlsclient (for HTTPS)

package main

import (
	"fmt"
	"io"
	"log"
	"machine"
	"net/http"
	"net/url"
	"time"
)

var (
	ssid string
	pass string
)

func main() {

	waitSerial()

	if err := netdev.NetConnect(); err != nil {
		log.Fatal(err)
	}

	path := "https://httpbin.org/post"
	data := url.Values{
		"name":       {"John Doe"},
		"occupation": {"gardener"},
	}

	resp, err := http.PostForm(path, data)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(body))
}

// Wait for user to open serial console
func waitSerial() {
	for !machine.Serial.DTR() {
		time.Sleep(100 * time.Millisecond)
	}
}
