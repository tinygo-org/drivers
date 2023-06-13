// package sdcard provides a TinyGo driver for sdcard/mmc devices
// using a SPI connection.
//
// To use a file system on the SDcard, please see the TinyFS repo:
//
// https://github.com/tinygo-org/tinyfs
package sdcard

import (
	"fmt"
	"machine"
	"time"
)

const (
	_CMD_TIMEOUT = 100

	_R1_IDLE_STATE           = 1 << 0
	_R1_ERASE_RESET          = 1 << 1
	_R1_ILLEGAL_COMMAND      = 1 << 2
	_R1_COM_CRC_ERROR        = 1 << 3
	_R1_ERASE_SEQUENCE_ERROR = 1 << 4
	_R1_ADDRESS_ERROR        = 1 << 5
	_R1_PARAMETER_ERROR      = 1 << 6

	// card types
	SD_CARD_TYPE_SD1  = 1 // Standard capacity V1 SD card
	SD_CARD_TYPE_SD2  = 2 // Standard capacity V2 SD card
	SD_CARD_TYPE_SDHC = 3 // High Capacity SD card
)

var (
	dummy [512]byte
)

type Device struct {
	bus        *machine.SPI
	sck        machine.Pin
	sdo        machine.Pin
	sdi        machine.Pin
	cs         machine.Pin
	cmdbuf     []byte
	dummybuf   []byte
	tokenbuf   []byte
	sdCardType byte
	CID        *CID
	CSD        *CSD
}

func New(b *machine.SPI, sck, sdo, sdi, cs machine.Pin) Device {
	return Device{
		bus:        b,
		cs:         cs,
		sck:        sck,
		sdo:        sdo,
		sdi:        sdi,
		cmdbuf:     make([]byte, 6),
		dummybuf:   make([]byte, 512),
		tokenbuf:   make([]byte, 1),
		sdCardType: 0,
	}
}

func (d *Device) Configure() error {
	return d.initCard()
}

func (d *Device) initCard() error {
	d.bus.Configure(machine.SPIConfig{
		SCK:       d.sck,
		SDO:       d.sdo,
		SDI:       d.sdi,
		Frequency: 250000,
		LSBFirst:  false,
		Mode:      0, // phase=0, polarity=0
	})

	// set pin modes
	d.cs.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.cs.High()

	for i := range dummy {
		dummy[i] = 0xFF
	}

	// clock card at least 100 cycles with cs high
	d.bus.Tx(dummy[:10], nil)

	d.cs.Low()
	d.bus.Tx(dummy[:], nil)

	// CMD0: init card; sould return _R1_IDLE_STATE (allow 5 attempts)
	ok := false
	tm := setTimeout(0, 2*time.Second)
	for !tm.expired() {
		// Wait up to 2 seconds to be the same as the Arduino
		if d.cmd(CMD0_GO_IDLE_STATE, 0, 0x95) == _R1_IDLE_STATE {
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("no SD card")
	}

	// CMD8: determine card version
	r := d.cmd(CMD8_SEND_IF_COND, 0x01AA, 0x87)
	if (r & _R1_ILLEGAL_COMMAND) == _R1_ILLEGAL_COMMAND {
		d.sdCardType = SD_CARD_TYPE_SD1
		return fmt.Errorf("init_card_v1 not impl\r\n")
	} else {
		// r7 response
		status := byte(0)
		for i := 0; i < 3; i++ {
			var err error
			status, err = d.bus.Transfer(byte(0xFF))
			if err != nil {
				return err
			}
		}
		if (status & 0x0F) != 0x01 {
			return fmt.Errorf("SD_CARD_ERROR_CMD8 %02X", status)
		}

		for i := 3; i < 4; i++ {
			var err error
			status, err = d.bus.Transfer(byte(0xFF))
			if err != nil {
				return err
			}
		}
		if status != 0xAA {
			return fmt.Errorf("SD_CARD_ERROR_CMD8 %02X", status)
		}
		d.sdCardType = SD_CARD_TYPE_SD2
	}

	// initialize card and send host supports SDHC if SD2
	arg := uint32(0)
	if d.sdCardType == SD_CARD_TYPE_SD2 {
		arg = 0x40000000
	}

	// check for timeout
	ok = false
	tm = setTimeout(0, 2*time.Second)
	for !tm.expired() {
		if d.acmd(ACMD41_SD_APP_OP_COND, arg) == 0 {
			ok = true
			break
		}
	}

	if !ok {
		return fmt.Errorf("SD_CARD_ERROR_ACMD41")
	}

	// if SD2 read OCR register to check for SDHC card
	if d.sdCardType == SD_CARD_TYPE_SD2 {
		if d.cmd(CMD58_READ_OCR, 0, 0xFF) != 0 {
			return fmt.Errorf("SD_CARD_ERROR_CMD58")
		}

		status, err := d.bus.Transfer(byte(0xFF))
		if err != nil {
			return err
		}
		if (status & 0xC0) == 0xC0 {
			d.sdCardType = SD_CARD_TYPE_SDHC
		}
		// discard rest of ocr - contains allowed voltage range
		for i := 1; i < 4; i++ {
			d.bus.Transfer(byte(0xFF))
		}
	}

	if d.cmd(CMD16_SET_BLOCKLEN, 0x0200, 0xFF) != 0 {
		return fmt.Errorf("SD_CARD_ERROR_CMD16")
	}

	var buf [16]byte
	// read CID
	err := d.ReadCID(buf[:])
	if err != nil {
		return err
	}
	d.CID = NewCID(buf[:])

	// read CSD
	err = d.ReadCSD(buf[:])
	if err != nil {
		return err
	}
	d.CSD = NewCSD(buf[:])

	d.cs.High()

	d.bus.Configure(machine.SPIConfig{
		SCK:       d.sck,
		SDO:       d.sdo,
		SDI:       d.sdi,
		Frequency: 4000000,
		LSBFirst:  false,
		Mode:      0, // phase=0, polarity=0
	})

	return nil
}

func (d Device) acmd(cmd byte, arg uint32) byte {
	d.cmd(CMD55_APP_CMD, 0, 0xFF)
	return d.cmd(cmd, arg, 0xFF)
}

func (d Device) cmd(cmd byte, arg uint32, crc byte) byte {
	d.cs.Low()

	if cmd != 12 {
		d.waitNotBusy(300 * time.Millisecond)
	}

	// create and send the command
	buf := d.cmdbuf
	buf[0] = 0x40 | cmd
	buf[1] = byte(arg >> 24)
	buf[2] = byte(arg >> 16)
	buf[3] = byte(arg >> 8)
	buf[4] = byte(arg)
	buf[5] = crc
	d.bus.Tx(buf, nil)

	if cmd == 12 {
		// skip 1 byte
		d.bus.Transfer(byte(0xFF))
	}

	// wait for the response (response[7] == 0)
	for i := 0; i < 0xFFFF; i++ {
		d.bus.Tx([]byte{0xFF}, d.tokenbuf)
		response := d.tokenbuf[0]
		if (response & 0x80) == 0 {
			return response
		}
	}

	// TODO
	//// timeout
	d.cs.High()
	d.bus.Transfer(byte(0xFF))

	return 0xFF // -1
}

func (d Device) waitNotBusy(timeout time.Duration) error {
	tm := setTimeout(1, timeout)
	for !tm.expired() {
		r, err := d.bus.Transfer(byte(0xFF))
		if err != nil {
			return err
		}
		if r == 0xFF {
			return nil
		}
	}
	return nil
}

func (d Device) waitStartBlock() error {
	status := byte(0xFF)

	tm := setTimeout(0, 300*time.Millisecond)
	for !tm.expired() {
		var err error
		status, err = d.bus.Transfer(byte(0xFF))
		if err != nil {
			d.cs.High()
			return err
		}
		if status != 0xFF {
			break
		}
	}

	if status != 254 {
		d.cs.High()
		return fmt.Errorf("SD_CARD_START_BLOCK")
	}

	return nil
}

// ReadCSD reads the CSD using CMD9.
func (d Device) ReadCSD(csd []byte) error {
	return d.readRegister(CMD9_SEND_CSD, csd)
}

// ReadCID reads the CID using CMD10
func (d Device) ReadCID(csd []byte) error {
	return d.readRegister(CMD10_SEND_CID, csd)
}

func (d Device) readRegister(cmd uint8, dst []byte) error {
	if d.cmd(cmd, 0, 0xFF) != 0 {
		return fmt.Errorf("SD_CARD_ERROR_READ_REG")
	}
	if err := d.waitStartBlock(); err != nil {
		return err
	}
	// transfer data
	for i := uint16(0); i < 16; i++ {
		r, err := d.bus.Transfer(byte(0xFF))
		if err != nil {
			return err
		}
		dst[i] = r
	}
	d.bus.Transfer(byte(0xFF))
	d.bus.Transfer(byte(0xFF))
	d.cs.High()

	return nil
}

// ReadData reads 512 bytes from sdcard into dst.
func (d Device) ReadData(block uint32, dst []byte) error {
	if len(dst) < 512 {
		return fmt.Errorf("len(dst) must be greater than or equal to 512")
	}

	// use address if not SDHC card
	if d.sdCardType != SD_CARD_TYPE_SDHC {
		block <<= 9
	}
	if d.cmd(CMD17_READ_SINGLE_BLOCK, block, 0xFF) != 0 {
		return fmt.Errorf("CMD17 error")
	}
	if err := d.waitStartBlock(); err != nil {
		return fmt.Errorf("waitStartBlock()")
	}

	err := d.bus.Tx(dummy[:512], dst)
	if err != nil {
		return err
	}

	// skip CRC (2byte)
	d.bus.Transfer(byte(0xFF))
	d.bus.Transfer(byte(0xFF))

	// TODO: probably not necessary
	d.cs.High()

	return nil
}

// WriteMultiStart starts the continuous write mode using CMD25.
func (d Device) WriteMultiStart(block uint32) error {
	// use address if not SDHC card
	if d.sdCardType != SD_CARD_TYPE_SDHC {
		block <<= 9
	}
	if d.cmd(CMD25_WRITE_MULTIPLE_BLOCK, block, 0xFF) != 0 {
		return fmt.Errorf("CMD25 error")
	}

	// skip 1 byte
	d.bus.Transfer(byte(0xFF))

	return nil
}

// WriteMulti performs continuous writing. It is necessary to call
// WriteMultiStart() in prior.
func (d Device) WriteMulti(buf []byte) error {
	// send Data Token for CMD25
	d.bus.Transfer(byte(0xFC))

	for i := 0; i < 512; i++ {
		_, err := d.bus.Transfer(buf[i])
		if err != nil {
			return err
		}
	}

	// send dummy CRC (2 byte)
	d.bus.Transfer(byte(0xFF))
	d.bus.Transfer(byte(0xFF))

	// Data Resp.
	r, err := d.bus.Transfer(byte(0xFF))
	if err != nil {
		return err
	}
	if (r & 0x1F) != 0x05 {
		return fmt.Errorf("SD_CARD_ERROR_WRITE")
	}

	// wait no busy
	err = d.waitNotBusy(600 * time.Millisecond)
	if err != nil {
		return fmt.Errorf("SD_CARD_ERROR_WRITE_TIMEOUT")
	}

	return nil
}

// WriteMultiStop exits the continuous write mode.
func (d Device) WriteMultiStop() error {
	defer d.cs.High()

	// Stop Tran token for CMD25
	d.bus.Transfer(0xFD)

	// skip 1 byte
	d.bus.Transfer(byte(0xFF))

	err := d.waitNotBusy(600 * time.Millisecond)
	if err != nil {
		return nil
	}

	return nil
}

// WriteData writes 512 bytes from dst to sdcard.
func (d Device) WriteData(block uint32, src []byte) error {
	if len(src) < 512 {
		return fmt.Errorf("len(src) must be greater than or equal to 512")
	}

	// use address if not SDHC card
	if d.sdCardType != SD_CARD_TYPE_SDHC {
		block <<= 9
	}
	if d.cmd(CMD24_WRITE_BLOCK, block, 0xFF) != 0 {
		return fmt.Errorf("CMD24 error")
	}

	// wait 1 byte?
	token := byte(0xFE)
	d.bus.Transfer(token)

	err := d.bus.Tx(src[:512], nil)
	if err != nil {
		return err
	}

	// send dummy CRC (2 byte)
	d.bus.Transfer(byte(0xFF))
	d.bus.Transfer(byte(0xFF))

	// Data Resp.
	r, err := d.bus.Transfer(byte(0xFF))
	if err != nil {
		return err
	}
	if (r & 0x1F) != 0x05 {
		return fmt.Errorf("SD_CARD_ERROR_WRITE")
	}

	// wait no busy
	err = d.waitNotBusy(600 * time.Millisecond)
	if err != nil {
		return fmt.Errorf("SD_CARD_ERROR_WRITE_TIMEOUT")
	}

	// TODO: probably not necessary
	d.cs.High()
	return nil
}

// ReadAt reads the given number of bytes from the sdcard.
func (dev *Device) ReadAt(buf []byte, addr int64) (int, error) {
	block := uint64(addr)
	// use address if not SDHC card
	if dev.sdCardType == SD_CARD_TYPE_SDHC {
		block >>= 9
	}

	idx := uint32(0)

	start := uint32(addr % 512)
	end := uint32(0)
	remain := uint32(len(buf))

	// If data starts in the middle
	if 0 < start {
		if start+remain <= 512 {
			end = start + remain
		} else {
			end = 512
		}

		err := dev.ReadData(uint32(block), dev.dummybuf)
		if err != nil {
			return 0, err
		}
		copy(buf[idx:], dev.dummybuf[start:end])

		remain -= end - start
		idx += end - start
		block++
	}

	// If more than 512 bytes left
	for 512 <= remain {
		start = 0
		end = 512

		err := dev.ReadData(uint32(block), dev.dummybuf)
		if err != nil {
			return 0, err
		}
		copy(buf[idx:], dev.dummybuf[start:end])

		remain -= end - start
		idx += end - start
		block++
	}

	// Read to the end
	if 0 < remain {
		start = 0
		end = remain

		err := dev.ReadData(uint32(block), dev.dummybuf)
		if err != nil {
			return 0, err
		}
		copy(buf[idx:], dev.dummybuf[start:end])

		remain -= end - start
		idx += end - start
		block++
	}

	return int(idx), nil
}

// WriteAt writes the given number of bytes to sdcard.
func (dev *Device) WriteAt(buf []byte, addr int64) (n int, err error) {
	block := uint64(addr)
	// use address if not SDHC card
	if dev.sdCardType == SD_CARD_TYPE_SDHC {
		block >>= 9
	}

	idx := uint32(0)

	start := uint32(addr % 512)
	end := uint32(0)
	remain := uint32(len(buf))

	// If data starts in the middle
	if 0 < start {
		if start+remain <= 512 {
			end = start + remain
		} else {
			end = 512
		}

		err := dev.ReadData(uint32(block), dev.dummybuf)
		if err != nil {
			return 0, err
		}
		copy(dev.dummybuf[start:end], buf[idx:])

		err = dev.WriteData(uint32(block), dev.dummybuf)
		if err != nil {
			return 0, err
		}

		remain -= end - start
		idx += end - start
		block++
	}

	// If more than 512 bytes left
	for 512 <= remain {
		start = 0
		end = 512

		err := dev.WriteData(uint32(block), buf[idx:idx+512])
		if err != nil {
			return 0, err
		}

		remain -= end - start
		idx += end - start
		block++
	}

	// Write to the end
	if 0 < remain {
		start = 0
		end = remain

		err := dev.ReadData(uint32(block), dev.dummybuf)
		if err != nil {
			return 0, err
		}
		copy(dev.dummybuf[start:end], buf[idx:])

		err = dev.WriteData(uint32(block), dev.dummybuf)
		if err != nil {
			return 0, err
		}

		remain -= end - start
		idx += end - start
		block++
	}

	return int(idx), nil
}

// Size returns the number of bytes in this sdcard.
func (dev *Device) Size() int64 {
	return int64(dev.CSD.Size())
}

// WriteBlockSize returns the block size in which data can be written to
// memory.
func (dev *Device) WriteBlockSize() int64 {
	return 512
}

// EraseBlockSize returns the smallest erasable area on this sdcard in bytes.
func (dev *Device) EraseBlockSize() int64 {
	return 512
}

// EraseBlocks erases the given number of blocks.
func (dev *Device) EraseBlocks(start, len int64) error {
	dev.WriteMultiStart(uint32(start))

	for i := range dev.dummybuf {
		dev.dummybuf[i] = 0
	}

	for i := 0; i < int(len); i++ {
		dev.WriteMulti(dev.dummybuf)
	}

	dev.WriteMultiStop()
	return nil
}
