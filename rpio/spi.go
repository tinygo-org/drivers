// Implemenmts drivers.SPI interface for the Raspberry Pi.
// Depends on the go-rpio library.

package rpio // import "tinygo.org/x/drivers/rpio"

import go_rpio "github.com/stianeikeland/go-rpio/v4"

type SPI struct {
}

func NewSPI() (s *SPI) {
	if err := go_rpio.SpiBegin(go_rpio.Spi0); err != nil {
		panic(err)
	}
	go_rpio.SpiSpeed(25_000_000) // 25 MHz
	go_rpio.SpiChipSelect(0)
	return &SPI{}
}

func (s *SPI) Tx(w, r []byte) error {
	data := make([]byte, len(w))
	copy(data, w)
	go_rpio.SpiExchange(data)
	copy(r, data)
	return nil
}

func (s *SPI) Transfer(b byte) (byte, error) {
	w := []byte{b}
	r := make([]byte, len(w))
	err := s.Tx(w, r)
	return r[0], err
}
