//go:build arduino_uno_r4_wifi

package probe

import (
	"machine"

	"tinygo.org/x/drivers/netdev"
	"tinygo.org/x/drivers/netlink"
	"tinygo.org/x/drivers/unor4wifi"
)

func Probe() (netlink.Netlinker, netdev.Netdever) {

	cfg := unor4wifi.Config{
		// UART3 is the SCI1 link to the ESP32-S3 AT interface.
		Uart:     machine.UART3,
		Tx:       machine.UART3_TX_PIN,
		Rx:       machine.UART3_RX_PIN,
		BaudRate: 115200,
	}

	dev := unor4wifi.New(&cfg)
	netdev.UseNetdev(dev)

	return dev, dev
}
