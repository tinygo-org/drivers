package tester

// SPIBus implements the SPI interface in memory for testing.
type SPIBus struct {
	c Failer
	//devices []*I2CDevice
}

// NewSPIBus returns an SPIBus mock SPI instance that uses c to flag errors
// if they happen. After creating a SPI instance, add devices
// to it with addDevice before using NewSPIBus interface.
func NewSPIBus(c Failer) *SPIBus {
	return &SPIBus{
		c: c,
	}
}

// Tx is a mock implementation of Tx for testing.
func (s *SPIBus) Tx(w, r []byte) error {
	return nil
}

// Transfer is a mock implementation of Transfer for testing.
func (s *SPIBus) Transfer(b byte) (byte, error) {
	return 0, nil
}
