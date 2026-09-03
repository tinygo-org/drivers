package main

import "tinygo.org/x/drivers/espnet"

func main() {
	err := espnet.WiFi.Configure(espnet.Config{})
	if err != nil {
		println("failed to configure:", err.Error())
	}
	mac, err := espnet.WiFi.AccessPointMAC()
	if err != nil {
		println("failed to read MAC address:", err.Error())
		return
	}
	print("MAC address:")
	for _, b := range mac {
		print(" ", b)
	}
	println()
}
