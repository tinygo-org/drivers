package main

import (
	"machine"
	"net"
	"net/netip"
	"time"

	"tinygo.org/x/drivers/netdev"
	"tinygo.org/x/drivers/w5500"
)

func main() {
	machine.SPI0.Configure(machine.SPIConfig{
		Frequency: 33 * machine.MHz,
	})
	machine.GPIO17.Configure(machine.PinConfig{Mode: machine.PinOutput})

	eth := w5500.New(machine.SPI0, machine.GPIO17)
	eth.Configure(w5500.Config{
		MAC:        net.HardwareAddr{0xee, 0xbe, 0xe9, 0xa9, 0xb6, 0x4f},
		IP:         netip.AddrFrom4([4]byte{192, 168, 1, 2}),
		SubnetMask: netip.AddrFrom4([4]byte{255, 255, 255, 0}),
		Gateway:    netip.AddrFrom4([4]byte{192, 168, 1, 1}),
	})
	netdev.UseNetdev(eth)

	for {
		if eth.LinkStatus() != w5500.LinkStatusUp {
			println("Waiting for link to be up")

			time.Sleep(1 * time.Second)
			continue
		}
		break
	}
}
