# unor4wifi

WiFi network driver for the Arduino UNO R4 WiFi.

The board has a Renesas RA4M1 main MCU and an ESP32-S3 WiFi module. The
ESP32-S3 runs the Arduino [uno-r4-wifi-usb-bridge][bridge] firmware. The
RA4M1 talks to it over UART (SCI1, `machine.UART3`) with an AT style
protocol. This driver speaks that protocol and implements the
[netlink](../netlink) and [netdev](../netdev) interfaces, so the standard
`net` package works.

Keep the stock bridge firmware on the ESP32-S3. Do not flash the ESP32-S3.

## Supported

- WiFi station mode (`ConnectModeSTA`) with WPA2 or open networks
- WiFi access point mode (`ConnectModeAP`)
- TCP client and TCP server
- UDP
- DNS lookup

## Not supported

- TLS offload in the ESP32-S3. Use the Go `crypto/tls` package instead.
- Socket options (`SetSockOpt` returns `netdev.ErrNotSupported`).
- BLE. The bridge has HCI pass-through commands, but this driver does not
  use them.

## Usage

```go
import (
	"tinygo.org/x/drivers/netlink"
	"tinygo.org/x/drivers/netlink/probe"
)

link, _ := probe.Probe()
err := link.NetConnect(&netlink.ConnectParams{
	Ssid:       "my-ssid",
	Passphrase: "my-passphrase",
})
```

Then use `net.Dial`, `net.Listen` and so on.

The RA4M1 has 256 KB of flash. The `net/http` package does not fit.
Use the `net` package directly, or a small client such as the `ntpclient`
or `tcpclient` examples.

Build with `-target=arduino-uno-r4-wifi -stack-size 2kb`. The RA4M1 has
32 KB of RAM and each goroutine gets a stack of the given size, so a larger
stack size leaves too little heap. The network examples that fit in flash
(`socket`, `tcpclient`, `tcpecho`, `ntpclient`, `tlsclient`, `webclient`)
carry the `arduino_uno_r4_wifi` build tag.

Do not use `io.Copy` with a `net.TCPConn` on this board. It allocates a
32 KB buffer. Use a `Read`/`Write` loop with a small buffer instead.

## Protocol

Commands and response formats are defined in the bridge firmware:

- [commands.h][commands]
- [cmds_wifi_netif.h][netif] (TCP client and server)
- [cmds_wifi_udp.h][udp]

Text responses look like `+CMD:<data>\r\nOK\r\n`. Binary responses from
`+CLIENTRECEIVE` and `+UDPREAD` look like `+CMD:<len>|<len bytes>OK\r\n`.
Payload for `+CLIENTSEND` and `+UDPWRITE` follows the command line as raw
bytes.

[bridge]: https://github.com/arduino/uno-r4-wifi-usb-bridge
[commands]: https://github.com/arduino/uno-r4-wifi-usb-bridge/blob/main/UNOR4USBBridge/commands.h
[netif]: https://github.com/arduino/uno-r4-wifi-usb-bridge/blob/main/UNOR4USBBridge/cmds_wifi_netif.h
[udp]: https://github.com/arduino/uno-r4-wifi-usb-bridge/blob/main/UNOR4USBBridge/cmds_wifi_udp.h
