package netif

import (
	"net"
	"time"
)

func AutoProbe(timeout time.Duration) (Stack, error) {
	// This function automatically gets the first available network device
	// and returns a stack for it.
	// It is intended to be used by the user as a convenience function.
	// Will be guarded by build tags and specific for whether the device
	// is a StackWifi, InterfaceEthPollWifi or InterfaceEthPoller.
	return nil, nil
}

func ProbeStackWifi() StackWifi {
	return nil
}
func ProbeEthPollWifi() InterfaceEthPollWifi {
	return nil
}
func ProbeEthPoller() InterfaceEthPoller {
	return nil
}

func StackForEthPoll(dev InterfaceEthPoller) Stack {
	// Use seqs to generate a stack.
	return nil
}

// OSI layer 4 enabled WIFI chip. i.e.: ESP32
func ExampleProbeStackWifi() {
	dev := ProbeStackWifi()
	wifiparams := WifiParams{
		SSID:        "myssid",
		ConnectMode: ConnectModeSTA,
		Passphrase:  "mypassphrase",
		Auth:        AuthTypeWPA2,
		CountryCode: "US",
	}
	err := StartWifiAutoconnect(dev, WifiAutoconnectParams{
		WifiParams: wifiparams,
	})
	if err != nil {
		panic(err)
	}
	for dev.NetFlags()&net.FlagRunning == 0 {
		time.Sleep(100 * time.Millisecond)
	}
	UseStack(dev)
	net.Dial("tcp", "192.168.1.1:33")
}

// Simplest case, OSI layer 2 enabled WIFI chip, i.e: CYW43439
func ExampleProbeEthPollWifi() {
	dev := ProbeEthPollWifi()
	wifiparams := WifiParams{
		SSID:        "myssid",
		ConnectMode: ConnectModeSTA,
		Passphrase:  "mypassphrase",
		Auth:        AuthTypeWPA2,
		CountryCode: "US",
	}
	err := StartWifiAutoconnect(dev, WifiAutoconnectParams{
		WifiParams: wifiparams,
	})
	if err != nil {
		panic(err)
	}
	for dev.NetFlags()&net.FlagRunning == 0 {
		time.Sleep(100 * time.Millisecond)
	}
	stack := StackForEthPoll(dev)
	UseStack(stack)
	net.Dial("tcp", "192.168.1.1:33")
}

// OSI level 2 chip with wired connection, i.e. ENC28J60.
func ExampleProbeEthPoller() {
	dev := ProbeEthPoller()
	if dev.NetFlags()&net.FlagRunning == 0 {
		panic("ethernet not connected")
	}
	stack := StackForEthPoll(dev)
	UseStack(stack)
	net.Dial("tcp", "192.168.1.1:33")
}

func ExampleAutoProbe() {
	stack, err := AutoProbe(10 * time.Second)
	if err != nil {
		panic(err)
	}
	UseStack(stack)
	net.Dial("tcp", "192.168.1.1:33")
}
