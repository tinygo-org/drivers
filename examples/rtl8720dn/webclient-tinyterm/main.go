package main

import (
	"machine"

	"bufio"
	"fmt"
	"image/color"
	"strings"
	"time"

	"tinygo.org/x/drivers/ili9341"
	"tinygo.org/x/drivers/net/http"
	"tinygo.org/x/drivers/rtl8720dn"

	"tinygo.org/x/tinyfont/proggy"
	"tinygo.org/x/tinyterm"
)

// You can override the setting with the init() in another source code.
// If debug is enabled, a serial connection is required.
// func init() {
//    ssid = "your-ssid"
//    pass = "your-password"
//    debug = false // true
//    server = "tinygo.org"
// }

var (
	ssid  string
	pass  string
	url   = "http://tinygo.org/"
	debug = false
)

var (
	display = ili9341.NewSPI(
		machine.SPI3,
		machine.LCD_DC,
		machine.LCD_SS_PIN,
		machine.LCD_RESET,
	)

	backlight = machine.LCD_BACKLIGHT

	terminal = tinyterm.NewTerminal(display)

	black = color.RGBA{0, 0, 0, 255}
	white = color.RGBA{255, 255, 255, 255}
	red   = color.RGBA{255, 0, 0, 255}
	blue  = color.RGBA{0, 0, 255, 255}
	green = color.RGBA{0, 255, 0, 255}

	font = &proggy.TinySZ8pt7b
)

var buf [0x400]byte

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

	adaptor := rtl8720dn.New(machine.UART3, machine.PB24, machine.PC24, machine.RTL8720D_CHIP_PU)
	adaptor.Debug(debug)
	adaptor.Configure()

	http.UseDriver(adaptor)
	http.SetBuf(buf[:])

	fmt.Fprintf(terminal, "ConnectToAP()\r\n")
	err := adaptor.ConnectToAccessPoint(ssid, pass, 10*time.Second)
	if err != nil {
		return err
	}
	fmt.Fprintf(terminal, "connected\r\n\r\n")

	ip, subnet, gateway, err := adaptor.GetIP()
	if err != nil {
		return err
	}
	fmt.Fprintf(terminal, "IP Address : %s\r\n", ip)
	fmt.Fprintf(terminal, "Mask       : %s\r\n", subnet)
	fmt.Fprintf(terminal, "Gateway    : %s\r\n", gateway)

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

		fmt.Fprintf(terminal, "%s %s\r\n", resp.Proto, resp.Status)
		for k, v := range resp.Header {
			fmt.Fprintf(terminal, "%s: %s\r\n", k, strings.Join(v, " "))
		}
		fmt.Printf("\r\n")

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			fmt.Fprintf(terminal, "%s\r\n", scanner.Text())
		}
		resp.Body.Close()

		cnt++
		fmt.Fprintf(terminal, "-------- %d --------\r\n", cnt)
		time.Sleep(10 * time.Second)
	}
}

func init() {
	machine.SPI3.Configure(machine.SPIConfig{
		SCK:       machine.LCD_SCK_PIN,
		SDO:       machine.LCD_SDO_PIN,
		SDI:       machine.LCD_SDI_PIN,
		Frequency: 40000000,
	})
	display.Configure(ili9341.Config{})

	backlight.Configure(machine.PinConfig{machine.PinOutput})
}
