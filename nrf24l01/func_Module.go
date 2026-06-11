package nrf24l01

import (
	"errors"
	"time"
)

// initialize a new NRF24L01 Module
func Module(config Config) (NRF24L01, error) {
	if len(config.SPIConfig.SenderAddress) != 5 || len(config.SPIConfig.ListenerAddress) != 5 {
		return nil, errors.New("error: len(address) != 5")
	}

	s := nrf24l01{
		packageSize:     int(config.SPIConfig.PackageSize),
		channel:         config.SPIConfig.Channel,
		senderAddress:   config.SPIConfig.SenderAddress,
		listenerAddress: config.SPIConfig.ListenerAddress,
		addressWidth:    5,
		txCMD:           make([]uint8, 2),
		rxCMD:           make([]uint8, 2),
		tx:              make([]uint8, int(config.SPIConfig.PackageSize)+1),
		rx:              make([]uint8, int(config.SPIConfig.PackageSize)+1),
		spi:             config.SPI,
		ce:              config.CE,
		csn:             config.CSN,
		irq:             config.IRQ,
	}

	s.ce.Low()
	s.csn.High()
	// RESET
	s.WriteRegister(CONFIG, 0x00)
	time.Sleep(time.Millisecond * 10)
	// CLEAR
	s.WriteRegister(STATUS, 0x70)
	s.FlushTX()
	s.FlushRX()
	// BASIC CONFIG
	s.WriteRegister(SETUP_AW, 0x03)
	if config.SPIConfig.ACK {
		s.WriteRegister(EN_AA, 0x01)
	} else {
		s.WriteRegister(EN_AA, 0x00)
	}
	s.WriteRegister(EN_RXADDR, 0x01)
	s.WriteRegister(RF_SETUP, s.getRFSetup(config.SPIConfig))
	s.WriteRegister(RF_CH, s.channel)
	s.WriteRegister(RX_PW_P0, config.SPIConfig.PackageSize)
	s.WriteRegister(DYNPD, 0x00)
	s.WriteRegister(CONFIG, 0x3E)
	if config.SPIConfig.Sender {
		s.WriteRegister(CONFIG, 0x0A)
		time.Sleep(time.Millisecond * 10)
		s.WriteMultiRegister(TX_ADDR, s.senderAddress)
		s.WriteMultiRegister(RX_ADDR_P0, s.listenerAddress)
	} else {
		s.WriteRegister(CONFIG, 0x0B)
		time.Sleep(time.Millisecond * 10)
		s.WriteMultiRegister(TX_ADDR, s.listenerAddress)
		s.WriteMultiRegister(RX_ADDR_P0, s.senderAddress)
		s.ce.High()
	}
	return &s, nil
}
