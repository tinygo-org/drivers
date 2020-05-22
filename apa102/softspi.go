package apa102

import "machine"

// bbSPI is a dumb bit-bang implementation of SPI protocol that is hardcoded
// to mode 0 and ignores trying to receive data. Just enough for the APA102.
// Note: making this unexported for now because it is probable not suitable
// most purposes other than the APA102 package. It might be desirable to make
// this more generic and include it in the TinyGo "machine" package instead.
type bbSPI struct {
	SCK   machine.Pin
	MOSI  machine.Pin
	Delay uint32
}

// Configure sets up the SCK and MOSI pins as outputs and sets them low
func (s *bbSPI) Configure() {
	s.SCK.Configure(machine.PinConfig{Mode: machine.PinOutput})
	s.MOSI.Configure(machine.PinConfig{Mode: machine.PinOutput})
	s.SCK.Low()
	s.MOSI.Low()
	if s.Delay == 0 {
		s.Delay = 1
	}
}

// Tx matches signature of machine.SPI.Tx() and is used to send multiple bytes.
// The r slice is ignored and no error will ever be returned.
func (s *bbSPI) Tx(w []byte, r []byte) error {
	s.Configure()
	for _, b := range w {
		s.Transfer(b)
	}
	return nil
}

// delay represents a quarter of the clock cycle
func (s *bbSPI) delay() {
	for i := uint32(0); i < s.Delay; {
		i++
	}
}

// Transfer matches signature of machine.SPI.Transfer() and is used to send a
// single byte. The received data is ignored and no error will ever be returned.
func (s *bbSPI) Transfer(b byte) (byte, error) {
	for i := uint8(0); i < 8; i++ {

		// half clock cycle high to start
		s.SCK.High()
		s.delay()

		// write the value to MOSI (MSB first)
		if b&(1<<(7-i)) == 0 {
			s.MOSI.Low()
		} else {
			s.MOSI.High()
		}
		s.delay()

		// half clock cycle low
		s.SCK.Low()
		s.delay()

		// for actual SPI would try to read the MISO value here
		s.delay()
	}

	return 0, nil
}
