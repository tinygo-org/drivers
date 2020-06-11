// +build tinygo

package wifinina

import (
	"machine"
	"time"
)

func NewSPI(spi machine.SPI, cs, ack, gpio0, reset machine.Pin) *Device {
	return &Device{
		Transport: &SPITransport{
			SPI:   spi,
			CS:    cs,
			ACK:   ack,
			GPIO0: gpio0,
			RESET: reset,
		},
	}
}

type SPITransport struct {
	SPI   machine.SPI
	CS    machine.Pin
	ACK   machine.Pin
	GPIO0 machine.Pin
	RESET machine.Pin
}

var _ Transport = (*SPITransport)(nil)

func (d *SPITransport) Configure() {
	d.CS.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.ACK.Configure(machine.PinConfig{Mode: machine.PinInput})
	d.RESET.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.GPIO0.Configure(machine.PinConfig{Mode: machine.PinOutput})
}

// TODO: eventually replace this with an interrupt
func (d *SPITransport) GetACK(level bool, timeout time.Duration) bool {
	for now := time.Now(); time.Since(now) < timeout; {
		if d.ACK.Get() == level {
			return true
		}
	}
	return false
}

func (d *SPITransport) SetCS(level bool) {
	d.CS.Set(level)
}

func (d *SPITransport) SetGPIO0(level bool) {
	d.GPIO0.Set(level)
}

func (d *SPITransport) SetReset(level bool) {
	d.RESET.Set(level)
}

func (d *SPITransport) Transfer(b byte) (byte, error) {
	return d.SPI.Transfer(b)
}

func (d *SPITransport) Tx(w []byte, r []byte) error {
	return d.SPI.Tx(w, r)
}
