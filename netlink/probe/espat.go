//go:build challenger_rp2040

package probe

import (
	"machine"

	"tinygo.org/x/drivers/espat"
	"tinygo.org/x/drivers/netdev"
	"tinygo.org/x/drivers/netlink"
)

func Probe() (netlink.Netlinker, netdev.Netdever) {

	cfg := espat.Config{
		// UART
		Uart: machine.UART1,
		Tx:   machine.UART1_TX_PIN,
		Rx:   machine.UART1_RX_PIN,
	}

	esp := espat.NewDevice(&cfg)
	netdev.UseNetdev(esp)

	return esp, esp
}
