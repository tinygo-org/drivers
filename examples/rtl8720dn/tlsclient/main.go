package main

import (
	"fmt"
	"strings"
	"time"

	"tinygo.org/x/drivers/net"
	"tinygo.org/x/drivers/net/tls"
	"tinygo.org/x/drivers/rtl8720dn"
)

// You can override the setting with the init() in another source code.
// func init() {
//    ssid = "your-ssid"
//    password = "your-password"
//    debug = true
//    server = "tinygo.org"
//    test_root_ca = "..."
// }

var (
	ssid     string
	password string
	server   string = "www.example.com"
	debug           = false
)

// Set the test_root_ca created by the following command
// $ openssl s_client -showcerts -verify 5 -connect www.example.com:443 < /dev/null
var test_root_ca = `-----BEGIN CERTIFICATE-----
MIIDrzCCApegAwIBAgIQCDvgVpBCRrGhdWrJWZHHSjANBgkqhkiG9w0BAQUFADBh
MQswCQYDVQQGEwJVUzEVMBMGA1UEChMMRGlnaUNlcnQgSW5jMRkwFwYDVQQLExB3
d3cuZGlnaWNlcnQuY29tMSAwHgYDVQQDExdEaWdpQ2VydCBHbG9iYWwgUm9vdCBD
QTAeFw0wNjExMTAwMDAwMDBaFw0zMTExMTAwMDAwMDBaMGExCzAJBgNVBAYTAlVT
MRUwEwYDVQQKEwxEaWdpQ2VydCBJbmMxGTAXBgNVBAsTEHd3dy5kaWdpY2VydC5j
b20xIDAeBgNVBAMTF0RpZ2lDZXJ0IEdsb2JhbCBSb290IENBMIIBIjANBgkqhkiG
9w0BAQEFAAOCAQ8AMIIBCgKCAQEA4jvhEXLeqKTTo1eqUKKPC3eQyaKl7hLOllsB
CSDMAZOnTjC3U/dDxGkAV53ijSLdhwZAAIEJzs4bg7/fzTtxRuLWZscFs3YnFo97
nh6Vfe63SKMI2tavegw5BmV/Sl0fvBf4q77uKNd0f3p4mVmFaG5cIzJLv07A6Fpt
43C/dxC//AH2hdmoRBBYMql1GNXRor5H4idq9Joz+EkIYIvUX7Q6hL+hqkpMfT7P
T19sdl6gSzeRntwi5m3OFBqOasv+zbMUZBfHWymeMr/y7vrTC0LUq7dBMtoM1O/4
gdW7jVg/tRvoSSiicNoxBN33shbyTApOB6jtSj1etX+jkMOvJwIDAQABo2MwYTAO
BgNVHQ8BAf8EBAMCAYYwDwYDVR0TAQH/BAUwAwEB/zAdBgNVHQ4EFgQUA95QNVbR
TLtm8KPiGxvDl7I90VUwHwYDVR0jBBgwFoAUA95QNVbRTLtm8KPiGxvDl7I90VUw
DQYJKoZIhvcNAQEFBQADggEBAMucN6pIExIK+t1EnE9SsPTfrgT1eXkIoyQY/Esr
hMAtudXH/vTBH1jLuG2cenTnmCmrEbXjcKChzUyImZOMkXDiqw8cvpOp/2PV5Adg
06O/nVsJ8dWO41P0jmP6P6fbtGbfYmbW0W5BjfIttep3Sp+dWOIrWcBAI+0tKIJF
PnlUkiaY4IBIqDfv8NZ5YBberOgOzW6sRBc4L0na4UU+Krk2U886UAb3LujEV0ls
YSEY1QSteDwsOoBrp+uvFRTp2InBuThs4pFsiv9kuXclVzDAGySj4dzp30d8tbQk
CAUw7C29C79Fv1C5qfPrmAESrciIxpg0X40KPMbp1ZWVbd4=
-----END CERTIFICATE-----
`

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
	rtl.SetRootCA(&test_root_ca)
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
			makeHTTPSRequest()
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

func makeHTTPSRequest() {

	var err error
	if conn != nil {
		conn.Close()
	}

	message("\r\n---------------\r\nDialing TCP connection")
	conn, err = tls.Dial("tcp", server, nil)
	for ; err != nil; conn, err = tls.Dial("tcp", server, nil) {
		message("Connection failed: " + err.Error())
		time.Sleep(5 * time.Second)
	}
	println("Connected!\r")

	print("Sending HTTPS request...")
	fmt.Fprintln(conn, "GET / HTTP/1.1")
	fmt.Fprintln(conn, "Host:", strings.Split(server, ":")[0])
	fmt.Fprintln(conn, "User-Agent: TinyGo")
	fmt.Fprintln(conn, "Connection: close")
	fmt.Fprintln(conn)
	println("Sent!\r\n\r")

	lastRequestTime = time.Now()
}

func message(msg string) {
	println(msg, "\r")
}
