package main

import (
	"fmt"
	"time"

	"tinygo.org/x/drivers/net"
	"tinygo.org/x/drivers/rtl8720dn"
)

// You can override the setting with the init() in another source code.
// func init() {
//    ssid = "your-ssid"
//    password = "your-password"
//    debug = true
//    server = "tinygo.org"
// }

var (
	ssid     string
	password string
	server   string = "tinygo.org"
	debug           = false
)

var buf [0x400]byte

var lastRequestTime time.Time
var conn net.Conn
var adaptor *rtl8720dn.RTL8720DN

func main() {
	err := run()
	for err != nil {
		fmt.Printf("error: %s\r\n", err.Error())
		time.Sleep(5 * time.Second)
	}
}

func run() error {
	rtl, err := setupRTL8720DN()
	if err != nil {
		return err
	}
	net.UseDriver(rtl)

	err = rtl.ConnectToAP(ssid, password)
	if err != nil {
		return err
	}

	ip, subnet, gateway, err := rtl.GetIP()
	if err != nil {
		return err
	}
	fmt.Printf("IP Address : %s\r\n", ip)
	fmt.Printf("Mask       : %s\r\n", subnet)
	fmt.Printf("Gateway    : %s\r\n", gateway)

	cnt := 0
	for {
		readConnection()
		if time.Now().Sub(lastRequestTime).Milliseconds() >= 10000 {
			makeHTTPRequest()
			cnt++
			fmt.Printf("-------- %d --------\r\n", cnt)
		}
	}
}

func readConnection() {
	if conn != nil {
		for n, err := conn.Read(buf[:]); n > 0; n, err = conn.Read(buf[:]) {
			if err != nil {
				println("Read error: " + err.Error())
			} else {
				print(string(buf[0:n]))
			}
		}
	}
}

func makeHTTPRequest() {

	var err error
	if conn != nil {
		conn.Close()
	}

	// make TCP connection
	ip := net.ParseIP(server)
	raddr := &net.TCPAddr{IP: ip, Port: 80}
	laddr := &net.TCPAddr{Port: 8080}

	message("\r\n---------------\r\nDialing TCP connection")
	conn, err = net.DialTCP("tcp", laddr, raddr)
	for ; err != nil; conn, err = net.DialTCP("tcp", laddr, raddr) {
		message("Connection failed: " + err.Error())
		time.Sleep(5 * time.Second)
	}
	println("Connected!\r")

	print("Sending HTTP request...")
	fmt.Fprintln(conn, "GET / HTTP/1.1")
	fmt.Fprintln(conn, "Host:", server)
	fmt.Fprintln(conn, "User-Agent: TinyGo")
	fmt.Fprintln(conn, "Connection: close")
	fmt.Fprintln(conn)
	println("Sent!\r\n\r")

	lastRequestTime = time.Now()
}

func message(msg string) {
	println(msg, "\r")
}
