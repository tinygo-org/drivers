// This example gets an URL using http.Get().  URL scheme can be http or https.
//
// Note: It may be necessary to increase the stack size when using "net/http".
// Use the -stack-size=4KB command line option.
//
// Some targets (Arduino Nano33 IoT) don't have enough SRAM to run http.Get().
// Use the following for those targets:
//
//     examples/net/webclient (for HTTP)
//     examples/net/tlsclient (for HTTPS)

//go:build ninafw || wioterminal

package main

import (
	"fmt"
	"io"
	"log"
	"machine"
	"net/http"
	"net/url"
	"strings"
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

	name := "John Doe"
	occupation := "gardener"

	params := "name=" + url.QueryEscape(name) + "&" +
		"occupation=" + url.QueryEscape(occupation)

	path := fmt.Sprintf("https://httpbin.org/get?%s", params)

	cnt := 0
	for {
		fmt.Printf("Getting %s\r\n\r\n", path)
		resp, err := http.Get(path)
		if err != nil {
			fmt.Printf("%s\r\n", err.Error())
			time.Sleep(10 * time.Second)
			continue
		}

		fmt.Printf("%s %s\r\n", resp.Proto, resp.Status)
		for k, v := range resp.Header {
			fmt.Printf("%s: %s\r\n", k, strings.Join(v, " "))
		}
		fmt.Printf("\r\n")

		body, err := io.ReadAll(resp.Body)
		println(string(body))
		resp.Body.Close()

		cnt++
		fmt.Printf("-------- %d --------\r\n", cnt)
		time.Sleep(10 * time.Second)
	}
}

// Wait for user to open serial console
func waitSerial() {
	for !machine.Serial.DTR() {
		time.Sleep(100 * time.Millisecond)
	}
}
