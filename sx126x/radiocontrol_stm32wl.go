//go:build stm32wlx

package sx126x

import (
	"device/stm32"

	"runtime/interrupt"
)

// STM32RadioControl helps implement the RadioController interface
type STM32RadioControl struct {
	irqHandler func()
}

// SetNss sets the NSS line aka chip select for SPI.
func (rc *STM32RadioControl) SetNss(state bool) error {
	if state {
		stm32.PWR.SUBGHZSPICR.SetBits(stm32.PWR_SUBGHZSPICR_NSS)
	} else {
		stm32.PWR.SUBGHZSPICR.ClearBits(stm32.PWR_SUBGHZSPICR_NSS)
	}

	return nil
}

// WaitWhileBusy wait until the radio is no longer busy
func (rc *STM32RadioControl) WaitWhileBusy() error {
	count := 100
	var rfbusyms, rfbusys bool
	for count > 0 {
		rfbusyms = stm32.PWR.SR2.HasBits(stm32.PWR_SR2_RFBUSYMS)
		rfbusys = stm32.PWR.SR2.HasBits(stm32.PWR_SR2_RFBUSYS)

		if !(rfbusyms && rfbusys) {
			return nil
		}
		count--
	}
	return errWaitWhileBusyTimeout
}

// init() configures whatever needed for sx126x radio control
func init() {
	// Enable APB3 Periph clock and delay
	stm32.RCC.APB3ENR.SetBits(stm32.RCC_APB3ENR_SUBGHZSPIEN)
	_ = stm32.RCC.APB3ENR.Get()

	// Disable radio reset and wait it's ready
	stm32.RCC.CSR.ClearBits(stm32.RCC_CSR_RFRST)
	for stm32.RCC.CSR.HasBits(stm32.RCC_CSR_RFRSTF) {
	}

	// Set NSS line low
	stm32.PWR.SUBGHZSPICR.SetBits(stm32.PWR_SUBGHZSPICR_NSS)

	// Enable radio busy wakeup from Standby for CPU
	stm32.PWR.CR3.SetBits(stm32.PWR_CR3_EWRFBUSY)

	// Clear busy flag
	stm32.PWR.SCR.Set(stm32.PWR_SCR_CWRFBUSYF)

	// Enable SUBGHZ Spi
	// - /8 Prescaler
	// - Software Slave Management (NSS)
	// - FIFO Threshold and 8bit size
	stm32.SPI3.CR1.ClearBits(stm32.SPI_CR1_SPE)
	stm32.SPI3.CR1.Set(stm32.SPI_CR1_MSTR | stm32.SPI_CR1_SSI | (0b010 << 3) | stm32.SPI_CR1_SSM)
	stm32.SPI3.CR2.Set(stm32.SPI_CR2_FRXTH | (0b111 << 8))
	stm32.SPI3.CR1.SetBits(stm32.SPI_CR1_SPE)
}

func (rc *STM32RadioControl) SetupInterrupts(handler func()) error {
	irqHandler = handler
	intr := interrupt.New(stm32.IRQ_Radio_IRQ_Busy, handleInterrupt)
	intr.Enable()

	return nil
}

var irqHandler func()

func handleInterrupt(interrupt.Interrupt) {
	irqHandler()
}
