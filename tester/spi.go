package tester

import (
	"bytes"
)

type spiExpectation struct {
	write []byte
	read  []byte
}

type SPIBus struct {
	c            Failer
	expectations []spiExpectation
}

func NewSPIBus(c Failer) *SPIBus {
	return &SPIBus{
		c:            c,
		expectations: []spiExpectation{},
	}
}

func (s *SPIBus) Tx(w, r []byte) error {
	if len(s.expectations) == 0 {
		s.c.Fatalf("unexpected SPI exchange")
	}
	ex := s.expectations[0]
	if !bytes.Equal(ex.write, w) {
		s.c.Fatalf("unexpected SPI write: got %#v, expecting %#v", w, ex.write)
	}
	copy(r, ex.read)
	s.expectations = s.expectations[1:]
	return nil
}

func (s *SPIBus) Transfer(b byte) (byte, error) {
	buf := make([]byte, 1)
	err := s.Tx([]byte{b}, buf)
	return buf[0], err
}

func (s *SPIBus) Expect(in []byte, out []byte) {
	if len(in) != len(out) {
		s.c.Fatalf("Expect: input and output slices must be the same length")
	}
	s.expectations = append(s.expectations, spiExpectation{in, out})
}
