package bme68x

import (
	"fmt"
	"time"

	"tinygo.org/x/drivers"
)

const (
	// REG_MEM_PAGE is the MEM_PAGE address
	REG_MEM_PAGE uint8 = 0xF3
	// MEM_PAGE_MSK is mask for SPI memory page
	MEM_PAGE_MSK uint8 = 0x10
	// SPI_RD_MSK is the mask for reading a register in SPI
	SPI_RD_MSK uint8 = 0x80
	// SPI_WR_MSK is the mask for writing a register in SPI
	SPI_WR_MSK uint8 = 0x7F
	// MEM_PAGE0 is the SPI memory page 0
	MEM_PAGE0 uint8 = 0x10
	// MEM_PAGE1 is the SPI memory page 1
	MEM_PAGE1 uint8 = 0x00
)

type spi struct {
	bus drivers.SPI
	// memoryPage is the current memory page
	memoryPage uint8
}

// Reset performs a soft reset of the BME68x sensor.
func (s *spi) Reset(_ uint16) error {
	if err := s.readMemoryPage(); err != nil {
		return fmt.Errorf("failed to read memory page: %w", err)
	}

	if err := s.Write(0, []uint8{REG_SOFT_RESET}, []byte{CMD_RESET}); err != nil {
		return fmt.Errorf("failed to soft reset command: %w", err)
	}

	// wait for 10ms
	time.Sleep(time.Duration(PeriodReset) * time.Microsecond)

	// after reset get the memory page
	if err := s.readMemoryPage(); err != nil {
		return fmt.Errorf("failed to read memory page: %w", err)
	}

	return nil
}

// Read reads data from the BME68x sensor over SPI.
func (s *spi) Read(_ uint16, reg uint8, data []byte) error {
	if err := s.setMemoryPage(reg); err != nil {
		return fmt.Errorf("failed to set memory page: %w", err)
	}

	return s.read(reg, data[:])
}

func (s *spi) read(reg uint8, data []byte) error {
	reg |= SPI_RD_MSK

	return s.bus.Tx([]byte{reg}, data)
}

// Write writes data to the BME68x sensor over SPI.
func (s *spi) Write(_ uint16, reg []uint8, data []byte) error {
	buf := make([]uint8, LEN_INTERLEAVE_BUFF)

	if len(data) > 0 && len(data) <= int(LEN_INTERLEAVE_BUFF/2) {
		for i := 0; i < len(data); i++ {
			if err := s.setMemoryPage(reg[i]); err != nil {
				return fmt.Errorf("failed to set memory page: %w", err)
			}

			buf[2*i] = reg[i] & SPI_WR_MSK
			buf[2*i+1] = data[i]
		}

		if err := s.bus.Tx(buf, nil); err != nil {
			return err
		}
	}

	return nil
}

func (s *spi) setMemoryPage(reg uint8) error {
	memoryPage := MEM_PAGE0

	if reg > 0x7F {
		memoryPage = MEM_PAGE1
	}

	if memoryPage == s.memoryPage {
		return nil
	}

	s.memoryPage = memoryPage

	var data [1]byte
	if err := s.read(REG_MEM_PAGE|SPI_RD_MSK, data[:]); err != nil {
		return fmt.Errorf("failed to read memory page: %w", err)
	}

	data[0] &^= MEM_PAGE_MSK
	data[0] |= (memoryPage & MEM_PAGE_MSK)

	return s.bus.Tx([]byte{REG_MEM_PAGE | SPI_WR_MSK}, []byte{data[0]})
}

func (s *spi) readMemoryPage() error {
	var reg [1]byte
	if err := s.read(REG_MEM_PAGE|SPI_RD_MSK, reg[:]); err != nil {
		return err
	}

	s.memoryPage = reg[0] & MEM_PAGE_MSK

	return nil
}
