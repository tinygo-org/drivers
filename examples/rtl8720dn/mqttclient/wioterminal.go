//go:build wioterminal
// +build wioterminal

package main

import (
	"device/sam"
	"machine"
	"runtime/interrupt"
	"time"

	"tinygo.org/x/drivers/rtl8720dn"
)

var (
	uart UARTx
)

func handleInterrupt(interrupt.Interrupt) {
	// should reset IRQ
	uart.Receive(byte((uart.Bus.DATA.Get() & 0xFF)))
	uart.Bus.INTFLAG.SetBits(sam.SERCOM_USART_INT_INTFLAG_RXC)
}

func setupRTL8720DN() (*rtl8720dn.RTL8720DN, error) {
	machine.RTL8720D_CHIP_PU.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.RTL8720D_CHIP_PU.Low()
	time.Sleep(100 * time.Millisecond)
	machine.RTL8720D_CHIP_PU.High()
	time.Sleep(1000 * time.Millisecond)
	if debug {
		waitSerial()
	}

	uart = UARTx{
		UART: &machine.UART{
			Buffer: machine.NewRingBuffer(),
			Bus:    sam.SERCOM0_USART_INT,
			SERCOM: 0,
		},
	}

	uart.Interrupt = interrupt.New(sam.IRQ_SERCOM0_2, handleInterrupt)
	uart.Configure(machine.UARTConfig{TX: machine.PB24, RX: machine.PC24, BaudRate: 614400})

	rtl := rtl8720dn.New(uart)
	rtl.Debug(debug)

	_, err := rtl.Rpc_tcpip_adapter_init()
	if err != nil {
		return nil, err
	}

	return rtl, nil
}

// Wait for user to open serial console
func waitSerial() {
	for !machine.Serial.DTR() {
		time.Sleep(100 * time.Millisecond)
	}
}

type UARTx struct {
	*machine.UART
}

func (u UARTx) Read(p []byte) (n int, err error) {
	if u.Buffered() == 0 {
		time.Sleep(1 * time.Millisecond)
		return 0, nil
	}
	return u.UART.Read(p)
}
