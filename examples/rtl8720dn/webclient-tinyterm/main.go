package main

import (
	"fmt"
	"image/color"
	"time"

	"tinygo.org/x/drivers/net"
	"tinygo.org/x/drivers/rtl8720dn"
	"tinygo.org/x/tinyfont/proggy"
	"tinygo.org/x/tinyterm"
)

// You can override the setting with the init() in another source code.
// If debug is enabled, a serial connection is required.
// func init() {
//    ssid = "your-ssid"
//    password = "your-password"
//    debug = false // true
//    server = "tinygo.org"
// }

var (
	ssid     string
	password string
	server   string = "tinygo.org"
	debug           = false
)

var (
	terminal = tinyterm.NewTerminal(display)

	black = color.RGBA{0, 0, 0, 255}
	white = color.RGBA{255, 255, 255, 255}
	red   = color.RGBA{255, 0, 0, 255}
	blue  = color.RGBA{0, 0, 255, 255}
	green = color.RGBA{0, 255, 0, 255}

	font = &proggy.TinySZ8pt7b
)

var buf [0x400]byte

var lastRequestTime time.Time
var conn net.Conn
var adaptor *rtl8720dn.RTL8720DN

func main() {
	display.FillScreen(black)
	backlight.High()

	terminal.Configure(&tinyterm.Config{
		Font:       font,
		FontHeight: 10,
		FontOffset: 6,
	})

	err := run()
	for err != nil {
		fmt.Fprintf(terminal, "error: %s\r\n", err.Error())
		time.Sleep(5 * time.Second)
	}
}

func run() error {
	fmt.Fprintf(terminal, "setupRTL8720DN()\r\n")
	if debug {
		fmt.Fprintf(terminal, "Running in debug mode.\r\n")
		fmt.Fprintf(terminal, "A serial connection is required to continue execution.\r\n")
	}
	rtl, err := setupRTL8720DN()
	if err != nil {
		return err
	}
	net.UseDriver(rtl)

	fmt.Fprintf(terminal, "ConnectToAP()\r\n")
	err = rtl.ConnectToAP(ssid, password)
	if err != nil {
		return err
	}
	fmt.Fprintf(terminal, "connected\r\n\r\n")

	ip, subnet, gateway, err := rtl.GetIP()
	if err != nil {
		return err
	}
	fmt.Fprintf(terminal, "IP Address : %s\r\n", ip)
	fmt.Fprintf(terminal, "Mask       : %s\r\n", subnet)
	fmt.Fprintf(terminal, "Gateway    : %s\r\n", gateway)

	cnt := 0
	for {
		readConnection()
		if time.Now().Sub(lastRequestTime).Milliseconds() >= 10000 {
			makeHTTPRequest()
			cnt++
			fmt.Fprintf(terminal, "-------- %d --------\r\n", cnt)
		}
	}
}

func readConnection() {
	if conn != nil {
		for n, err := conn.Read(buf[:]); n > 0; n, err = conn.Read(buf[:]) {
			if err != nil {
				fmt.Fprintf(terminal, "Read error: "+err.Error()+"\r\n")
			} else {
				fmt.Fprintf(terminal, string(buf[0:n]))
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
	fmt.Fprintf(terminal, "Connected!\r\n")

	fmt.Fprintf(terminal, "Sending HTTP request...")
	fmt.Fprintln(conn, "GET / HTTP/1.1")
	fmt.Fprintln(conn, "Host:", server)
	fmt.Fprintln(conn, "User-Agent: TinyGo")
	fmt.Fprintln(conn, "Connection: close")
	fmt.Fprintln(conn)
	fmt.Fprintf(terminal, "Sent!\r\n\r\n")

	lastRequestTime = time.Now()
}

func message(msg string) {
	fmt.Fprintf(terminal, "%s\r\n", msg)
}
