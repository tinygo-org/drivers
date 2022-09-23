package main

import (
	"machine"

	"fmt"
	"strconv"
	"time"

	"tinygo.org/x/drivers/net"
	"tinygo.org/x/drivers/net/http"
	"tinygo.org/x/drivers/rtl8720dn"
)

// IP address of the server aka "hub". Replace with your own info.
// You can override the setting with the init() in another source code.
// func init() {
//    ssid = "your-ssid"
//    pass = "your-password"
//    hubIP = "192.168.1.118"
//    debug = true
// }

var (
	ssid  string
	pass  string
	hubIP = ""
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

	// now make UDP connection
	hip := net.ParseIP(hubIP)
	raddr := &net.UDPAddr{IP: hip, Port: 2222}
	laddr := &net.UDPAddr{Port: 2222}

	println("Dialing UDP connection...")
	conn, err := net.DialUDP("udp", laddr, raddr)
	if err != nil {
		return err
	}

	for {
		// send data
		println("Sending data...")
		for i := 0; i < 25; i++ {
			conn.Write([]byte("hello " + strconv.Itoa(i) + "\r\n"))
		}
		time.Sleep(1000 * time.Millisecond)
	}

	// Right now this code is never reached. Need a way to trigger it...
	println("Disconnecting UDP...")
	conn.Close()
	println("Done.")

	return nil
}

func message(msg string) {
	println(msg, "\r")
}
