package flash

import "machine"

type Command byte

const (
	CmdRead            Command = 0x03 // Single Read
	CmdQuadRead                = 0x6B // 1 line address, 4 line data
	CmdReadJedecID             = 0x9f
	CmdPageProgram             = 0x02
	CmdQuadPageProgram         = 0x32 // 1 line address, 4 line data
	CmdReadStatus              = 0x05
	CmdReadStatus2             = 0x35
	CmdWriteStatus             = 0x01
	CmdWriteStatus2            = 0x31
	CmdEnableReset             = 0x66
	CmdReset                   = 0x99
	CmdWriteEnable             = 0x06
	CmdWriteDisable            = 0x04
	CmdEraseSector             = 0x20
	CmdEraseBlock              = 0xD8
	CmdEraseChip               = 0xC7
)

type Transport struct {
	SPI  machine.SPI
	MOSI machine.Pin
	MISO machine.Pin
	SCK  machine.Pin
	SS   machine.Pin
}

func (tr *Transport) Begin() {

	// Configure SPI bus
	tr.SPI.Configure(machine.SPIConfig{
		Frequency: 50000000,
		MISO:      tr.MISO,
		MOSI:      tr.MOSI,
		SCK:       tr.SCK,
		LSBFirst:  false,
		Mode:      0,
	})

	// Configure chip select pin
	tr.SS.Configure(machine.PinConfig{Mode: machine.PinOutput})
	tr.SS.High()

}

func (tr *Transport) RunCommand(cmd Command) (err error) {
	tr.SS.Low()
	_, err = tr.SPI.Transfer(byte(cmd))
	tr.SS.High()
	return
}

func (tr *Transport) ReadCommand(cmd Command, rsp []byte) (err error) {
	tr.SS.Low()
	if _, err := tr.SPI.Transfer(byte(cmd)); err == nil {
		err = tr.readInto(rsp)
	}
	tr.SS.High()
	return
}

func (tr *Transport) ReadCommandByte(cmd Command) (rsp byte, err error) {
	tr.SS.Low()
	if _, err := tr.SPI.Transfer(byte(cmd)); err == nil {
		rsp, err = tr.SPI.Transfer(0xFF)
	}
	tr.SS.High()
	return
}

func (tr *Transport) WriteCommand(cmd Command, data []byte) (err error) {
	tr.SS.Low()
	if _, err := tr.SPI.Transfer(byte(cmd)); err == nil {
		err = tr.writeFrom(data)
	}
	tr.SS.High()
	return
}

func (tr *Transport) EraseCommand(cmd Command, address uint32) (err error) {
	tr.SS.Low()
	err = tr.sendAddress(cmd, address)
	tr.SS.High()
	return
}

func (tr *Transport) ReadMemory(addr uint32, rsp []byte) (err error) {
	tr.SS.Low()
	if err = tr.sendAddress(CmdRead, addr); err == nil {
		err = tr.readInto(rsp)
	}
	tr.SS.High()
	return
}

func (tr *Transport) WriteMemory(addr uint32, data []byte) (err error) {
	tr.SS.Low()
	if err = tr.sendAddress(CmdPageProgram, addr); err == nil {
		err = tr.writeFrom(data)
	}
	tr.SS.High()
	return
}

func (tr *Transport) sendAddress(cmd Command, addr uint32) error {
	_, err := tr.SPI.Transfer(byte(cmd))
	if err == nil {
		_, err = tr.SPI.Transfer(byte((addr >> 16) & 0xFF))
	}
	if err == nil {
		_, err = tr.SPI.Transfer(byte((addr >> 8) & 0xFF))
	}
	if err == nil {
		_, err = tr.SPI.Transfer(byte(addr & 0xFF))
	}
	return err
}

func (tr *Transport) readInto(rsp []byte) (err error) {
	for i, c := 0, len(rsp); i < c && err == nil; i++ {
		rsp[i], err = tr.SPI.Transfer(0xFF)
	}
	return
}

func (tr *Transport) writeFrom(data []byte) (err error) {
	for i, c := 0, len(data); i < c && err == nil; i++ {
		_, err = tr.SPI.Transfer(data[i])
	}
	return
}
