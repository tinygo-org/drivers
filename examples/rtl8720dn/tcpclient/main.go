// This example opens a TCP connection using a device with RTL8720DN firmware
// and sends some data, for the purpose of testing speed and connectivity.
//
// You can open a server to accept connections from this program using:
//
// nc -w 5 -lk 8080
package main

import (
	"machine"

	"bytes"
	"fmt"
	"time"

	"tinygo.org/x/drivers/net"
	"tinygo.org/x/drivers/rtl8720dn"
)

// You can override the setting with the init() in another source code.
// func init() {
//    ssid = "your-ssid"
//    pass = "your-password"
//    serverIP = "192.168.1.119"
//    debug = true
// }

var (
	ssid     string
	pass     string
	serverIP = ""
	debug    = false
)

var buf = &bytes.Buffer{}

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

	for {
		sendBatch()
		time.Sleep(500 * time.Millisecond)
	}
	println("Done.")

	return nil
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

func message(msg string) {
	println(msg, "\r")
}
