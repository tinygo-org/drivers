// This is xample of writing and reading E220 Configuration
package main

import (
	"fmt"
	"machine"
	"time"

	"tinygo.org/x/drivers/e220"
)

var (
	uart = machine.DefaultUART
	m0   = machine.D4
	m1   = machine.D5
	tx   = machine.UART_TX_PIN
	rx   = machine.UART_RX_PIN
	aux  = machine.D6
)

func main() {
	device := e220.New(uart, m0, m1, tx, rx, aux, 9600)
	err := device.Configure(e220.Mode3)
	if err != nil {
		fail(err.Error())
	}
	cfg := e220.Config{
		// Default Parameters
		ModuleAddr:         0x0000,
		UartSerialPortRate: e220.UartSerialPortRate9600Bps,
		AirDataRate:        e220.AirDataRate6250Bps,
		SubPacket:          e220.SubPacket200Bytes,
		RssiAmbient:        e220.RSSIAmbientDisable,
		TransmitPower:      e220.TransmitPowerUnavailable,
		Channel:            15,
		RssiByte:           e220.RSSIByteDisable,
		TransmitMethod:     e220.TransmitMethodTransparent,
		WorCycleSetting:    e220.WorCycleSetting2000ms,
		EncryptionKey:      0x0000,
		Version:            0x00,
	}
	err = device.WriteConfig(cfg)
	if err != nil {
		fail(err.Error())
	}
	// The EncryptionKey(byte6-7) is write-only, so a 0 value is read
	response, err := device.ReadConfig()
	if err != nil {
		fail(err.Error())
	}

	for {
		fmt.Println(response)
		time.Sleep(time.Millisecond * 1000)
	}
}

func fail(msg string) {
	for {
		println(msg)
		time.Sleep(1 * time.Second)
	}
}
