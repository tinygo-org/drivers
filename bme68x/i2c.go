package bme68x

import (
	"fmt"
	"time"

	"tinygo.org/x/drivers"
)

type i2c struct {
	bus drivers.I2C
}

// Reset performs a soft reset of the BME68x sensor.
func (i *i2c) Reset(addr uint16) error {
	if err := i.Write(addr, []uint8{REG_SOFT_RESET}, []byte{CMD_RESET}); err != nil {
		return fmt.Errorf("failed to soft reset command: %w", err)
	}

	time.Sleep(time.Duration(PeriodReset) * time.Microsecond)

	return nil
}

// Read reads data from the BME68x sensor over I2C.
func (i *i2c) Read(addr uint16, reg uint8, data []byte) error {
	return i.bus.Tx(addr, []byte{reg}, data)
}

// Write writes data to the BME68x sensor over I2C.
func (i *i2c) Write(addr uint16, reg []uint8, data []byte) error {
	buf := make([]uint8, LEN_INTERLEAVE_BUFF)

	if len(data) > 0 && len(data) <= int(LEN_INTERLEAVE_BUFF/2) {
		for i := 0; i < len(data); i++ {
			buf[2*i] = reg[i]
			buf[2*i+1] = data[i]
		}

		if err := i.bus.Tx(uint16(addr), buf, nil); err != nil {
			return err
		}
	}

	return nil
}
