package flash

import (
	"machine"
)

type transport interface {
	configure(config *DeviceConfig)
	supportQuadMode() bool
	setClockSpeed(hz uint32) (err error)
	runCommand(cmd byte) (err error)
	readCommand(cmd byte, rsp []byte) (err error)
	writeCommand(cmd byte, data []byte) (err error)
	eraseCommand(cmd byte, address uint32) (err error)
	readMemory(addr uint32, rsp []byte) (err error)
	writeMemory(addr uint32, data []byte) (err error)
}

// NewSPI returns a pointer to a flash device that uses a SPI peripheral to
// communicate with a serial memory chip.
func NewSPI(spi *machine.SPI, sdo, sdi, sck, cs machine.Pin) *Device {
	return &Device{
		trans: &spiTransport{
			spi: spi,
			sdo: sdo,
			sdi: sdi,
			sck: sck,
			ss:  cs,
		},
	}
}

type spiTransport struct {
	spi *machine.SPI
	sdo machine.Pin
	sdi machine.Pin
	sck machine.Pin
	ss  machine.Pin
}

func (tr *spiTransport) configure(config *DeviceConfig) {
	// Configure spi bus
	tr.setClockSpeed(5000000)

	// Configure chip select pin
	tr.ss.Configure(machine.PinConfig{Mode: machine.PinOutput})
	tr.ss.High()
}

func (tr *spiTransport) setClockSpeed(hz uint32) error {
	// TODO: un-hardcode this max speed; it is probably a sensible
	//       default maximum for atsamd and nrf at least
	if hz > 24*1e6 {
		hz = 24 * 1e6
	}
	tr.spi.Configure(machine.SPIConfig{
		Frequency: hz,
		SDI:       tr.sdi,
		SDO:       tr.sdo,
		SCK:       tr.sck,
		LSBFirst:  false,
		Mode:      0,
	})
	return nil
}

func (tr *spiTransport) supportQuadMode() bool {
	return false
}

func (tr *spiTransport) runCommand(cmd byte) (err error) {
	tr.ss.Low()
	_, err = tr.spi.Transfer(byte(cmd))
	tr.ss.High()
	return
}

func (tr *spiTransport) readCommand(cmd byte, rsp []byte) (err error) {
	tr.ss.Low()
	if _, err := tr.spi.Transfer(byte(cmd)); err == nil {
		err = tr.readInto(rsp)
	}
	tr.ss.High()
	return
}

func (tr *spiTransport) readCommandByte(cmd byte) (rsp byte, err error) {
	tr.ss.Low()
	if _, err := tr.spi.Transfer(byte(cmd)); err == nil {
		rsp, err = tr.spi.Transfer(0xFF)
	}
	tr.ss.High()
	return
}

func (tr *spiTransport) writeCommand(cmd byte, data []byte) (err error) {
	tr.ss.Low()
	if _, err := tr.spi.Transfer(byte(cmd)); err == nil {
		err = tr.writeFrom(data)
	}
	tr.ss.High()
	return
}

func (tr *spiTransport) eraseCommand(cmd byte, address uint32) (err error) {
	tr.ss.Low()
	err = tr.sendAddress(cmd, address)
	tr.ss.High()
	return
}

func (tr *spiTransport) readMemory(addr uint32, rsp []byte) (err error) {
	tr.ss.Low()
	if err = tr.sendAddress(cmdRead, addr); err == nil {
		err = tr.readInto(rsp)
	}
	tr.ss.High()
	return
}

func (tr *spiTransport) writeMemory(addr uint32, data []byte) (err error) {
	tr.ss.Low()
	if err = tr.sendAddress(cmdPageProgram, addr); err == nil {
		err = tr.writeFrom(data)
	}
	tr.ss.High()
	return
}

func (tr *spiTransport) sendAddress(cmd byte, addr uint32) error {
	_, err := tr.spi.Transfer(byte(cmd))
	if err == nil {
		_, err = tr.spi.Transfer(byte((addr >> 16) & 0xFF))
	}
	if err == nil {
		_, err = tr.spi.Transfer(byte((addr >> 8) & 0xFF))
	}
	if err == nil {
		_, err = tr.spi.Transfer(byte(addr & 0xFF))
	}
	return err
}

func (tr *spiTransport) readInto(rsp []byte) (err error) {
	for i, c := 0, len(rsp); i < c && err == nil; i++ {
		rsp[i], err = tr.spi.Transfer(0xFF)
	}
	return
}

func (tr *spiTransport) writeFrom(data []byte) (err error) {
	for i, c := 0, len(data); i < c && err == nil; i++ {
		_, err = tr.spi.Transfer(data[i])
	}
	return
}
