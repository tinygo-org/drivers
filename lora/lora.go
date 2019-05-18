// Package lora provides a driver for SX127x LoRa transceivers.
//
// Datasheet:
// https://www.semtech.com/uploads/documents/DS_SX1276-7-8-9_W_APP_V6.pdf
//
package lora

import (
	"errors"
	"fmt"
	"machine"
	"time"
)

// Device wraps an SPI connection to a SX127x device.
type Device struct {
	bus     machine.SPI
	csPin   machine.GPIO
	rstPin  machine.GPIO
	dio0Pin machine.GPIO
}

type Config struct {
	Frequency       int32
	SpreadingFactor int8
	Bandwidth       int32
	CodingRate      int8
	TxPower         int8
}

// New creates a new SX127x connection. The SPI bus must already be configured.
func New(b machine.SPI, csPin machine.GPIO, rstPin machine.GPIO, dio0Pin machine.GPIO) Device {
	return Device{bus: b,
		csPin:   csPin,
		rstPin:  rstPin,
		dio0Pin: dio0Pin,
	}
}

// Configure initializes the display with default configuration
func (d *Device) Configure(cfg Config) (err error) {
	d.rstPin.Low()
	time.Sleep(10 * time.Millisecond)
	d.rstPin.High()
	time.Sleep(10 * time.Millisecond)

	if d.readRegister(REG_VERSION) != 0x12 {
		err = errors.New("SX127x module not found")
	}
	return err
}

// ReadTemperature returns the temperature in celsius milli degrees (ºC/1000).
func (d *Device) SendPacket(packet []byte) {
}

func (d *Device) PrintRegisters() {
	for i := 0; i < 128; i++ {
		fmt.Printf("% 02x: ", i)
		var b byte = d.readRegister(uint8(i))
		fmt.Printf("% 02x\n", b)
	}
}

func (d *Device) readRegister(reg uint8) uint8 {
	var data [1]byte
	d.csPin.High()
	var err = d.bus.Tx([]byte{reg}, data[:])
	d.csPin.Low()
	if err != nil {
		print("readRegister err: ")
		fmt.Println(err)
	}
	return data[0]
}
