package flash

import "machine"

func NewSPI(spi *machine.SPI, miso, mosi, sck, cs machine.Pin) *Device {
	return &Device{
		transport: &spiTransport{
			spi:  spi,
			mosi: mosi,
			miso: miso,
			sck:  sck,
			ss:   cs,
		},
	}
}

type spiTransport struct {
	spi  *machine.SPI
	mosi machine.Pin
	miso machine.Pin
	sck  machine.Pin
	ss   machine.Pin
}

func (tr *spiTransport) begin() {
	// Configure spi bus
	tr.setClockSpeed(5000000)

	// Configure chip select pin
	tr.ss.Configure(machine.PinConfig{Mode: machine.PinOutput})
	tr.ss.High()
}

func (tr *spiTransport) setClockSpeed(hz uint32) error {
	tr.spi.Configure(machine.SPIConfig{
		Frequency: hz,
		MISO:      tr.miso,
		MOSI:      tr.mosi,
		SCK:       tr.sck,
		LSBFirst:  false,
		Mode:      0,
	})
	return nil
}

func (tr *spiTransport) supportQuadMode() bool {
	return false
}

func (tr *spiTransport) runCommand(cmd Command) (err error) {
	tr.ss.Low()
	_, err = tr.spi.Transfer(byte(cmd))
	tr.ss.High()
	return
}

func (tr *spiTransport) readCommand(cmd Command, rsp []byte) (err error) {
	tr.ss.Low()
	if _, err := tr.spi.Transfer(byte(cmd)); err == nil {
		err = tr.readInto(rsp)
	}
	tr.ss.High()
	return
}

func (tr *spiTransport) readCommandByte(cmd Command) (rsp byte, err error) {
	tr.ss.Low()
	if _, err := tr.spi.Transfer(byte(cmd)); err == nil {
		rsp, err = tr.spi.Transfer(0xFF)
	}
	tr.ss.High()
	return
}

func (tr *spiTransport) writeCommand(cmd Command, data []byte) (err error) {
	tr.ss.Low()
	if _, err := tr.spi.Transfer(byte(cmd)); err == nil {
		err = tr.writeFrom(data)
	}
	tr.ss.High()
	return
}

func (tr *spiTransport) eraseCommand(cmd Command, address uint32) (err error) {
	tr.ss.Low()
	err = tr.sendAddress(cmd, address)
	tr.ss.High()
	return
}

func (tr *spiTransport) readMemory(addr uint32, rsp []byte) (err error) {
	tr.ss.Low()
	if err = tr.sendAddress(CmdRead, addr); err == nil {
		err = tr.readInto(rsp)
	}
	tr.ss.High()
	return
}

func (tr *spiTransport) writeMemory(addr uint32, data []byte) (err error) {
	tr.ss.Low()
	if err = tr.sendAddress(CmdPageProgram, addr); err == nil {
		err = tr.writeFrom(data)
	}
	tr.ss.High()
	return
}

func (tr *spiTransport) sendAddress(cmd Command, addr uint32) error {
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
