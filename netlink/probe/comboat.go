//go:build comboat_fw

package probe

import (
	"machine"

	"tinygo.org/x/drivers/comboat"
	"tinygo.org/x/drivers/netdev"
	"tinygo.org/x/drivers/netlink"
)

func Probe() (netlink.Netlinker, netdev.Netdever) {

	cfg := comboat.Config{
		BaudRate: 115200,
		Uart:     machine.UART1,
		Tx:       machine.UART1_TX_PIN,
		Rx:       machine.UART1_RX_PIN,
	}

	combo := comboat.NewDevice(&cfg)
	netdev.UseNetdev(combo)

	return combo, combo
}
