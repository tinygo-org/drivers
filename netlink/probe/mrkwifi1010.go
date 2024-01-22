//go:build arduino_mkrwifi1010

package probe

import (
	"machine"

	"tinygo.org/x/drivers/netdev"
	"tinygo.org/x/drivers/netlink"
	"tinygo.org/x/drivers/wifinina"
)

func Probe() (netlink.Netlinker, netdev.Netdever) {

	cfg := wifinina.Config{
		// Configure SPI for 8Mhz, Mode 0, MSB First
		Spi:  machine.NINA_SPI,
		Freq: 8 * 1e6,
		Sdo:  machine.NINA_SDO,
		Sdi:  machine.NINA_SDI,
		Sck:  machine.NINA_SCK,
		// Device pins
		Cs:     machine.NINA_CS,
		Ack:    machine.NINA_ACK,
		Gpio0:  machine.NINA_GPIO0,
		Resetn: machine.NINA_RESETN,
		// mMKR 1010 resets High
		ResetIsHigh: true,
	}

	nina := wifinina.New(&cfg)
	netdev.UseNetdev(nina)

	return nina, nina
}
