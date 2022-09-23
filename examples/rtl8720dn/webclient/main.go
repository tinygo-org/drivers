package main

import (
	"machine"

	"bufio"
	"fmt"
	"strings"
	"time"

	"tinygo.org/x/drivers/net/http"
	"tinygo.org/x/drivers/rtl8720dn"
)

// You can override the setting with the init() in another source code.
// func init() {
//    ssid = "your-ssid"
//    pass = "your-password"
//    url = "http://tinygo.org/"
//    debug = true
// }

var (
	ssid  string
	pass  string
	url   = "http://tinygo.org/"
	debug = false
)

var buf [0x400]byte

func main() {
	err := run()
	for err != nil {
		fmt.Printf("error: %s\r\n", err.Error())
		time.Sleep(5 * time.Second)
	}
}

func run() error {
	adaptor := rtl8720dn.New(machine.UART3, machine.PB24, machine.PC24, machine.RTL8720D_CHIP_PU)
	adaptor.Debug(debug)
	adaptor.Configure()

	http.UseDriver(adaptor)
	http.SetBuf(buf[:])

	err := adaptor.ConnectToAccessPoint(ssid, pass, 10*time.Second)
	if err != nil {
		return err
	}

	ip, subnet, gateway, err := adaptor.GetIP()
	if err != nil {
		return err
	}
	fmt.Printf("IP Address : %s\r\n", ip)
	fmt.Printf("Mask       : %s\r\n", subnet)
	fmt.Printf("Gateway    : %s\r\n", gateway)

	// You can send and receive cookies in the following way
	// 	import "tinygo.org/x/drivers/net/http/cookiejar"
	// 	jar, err := cookiejar.New(nil)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	client := &http.Client{Jar: jar}
	// 	http.DefaultClient = client

	cnt := 0
	for {
		// Various examples are as follows
		//
		// -- Get
		// 	resp, err := http.Get(url)
		//
		// -- Post
		// 	body := `cnt=12`
		// 	resp, err = http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(body))
		//
		// -- Post with JSON
		// 	body := `{"msg": "hello"}`
		// 	resp, err := http.Post(url, "application/json", strings.NewReader(body))

		resp, err := http.Get(url)
		if err != nil {
			return err
		}

		fmt.Printf("%s %s\r\n", resp.Proto, resp.Status)
		for k, v := range resp.Header {
			fmt.Printf("%s: %s\r\n", k, strings.Join(v, " "))
		}
		fmt.Printf("\r\n")

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			fmt.Printf("%s\r\n", scanner.Text())
		}
		resp.Body.Close()

		cnt++
		fmt.Printf("-------- %d --------\r\n", cnt)
		time.Sleep(10 * time.Second)
	}
}
