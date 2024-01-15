package sd

import (
	"encoding/binary"
	"errors"
	"io"
	"math"
	"math/bits"
	"time"
)

// Reference for this implementation:
// https://github.com/embassy-rs/embedded-sdmmc-rs/blob/master/src/sdmmc.rs

// Not used currently. We'd want to switch over to one way of doing things, Rust way.
func (d *SPICard) initRs() error {
	// Supply minimum of 74 clock cycles with CS high.
	d.csEnable(true)
	for i := 0; i < 10; i++ {
		d.send(0xff)
	}
	d.csEnable(false)
	for i := 0; i < 512; i++ {
		d.receive()
	}
	d.csEnable(true)
	defer d.csEnable(false)
	// Enter SPI mode
	const maxRetries = 32
	retries := maxRetries
	tm := d.timers[0].setTimeout(2 * time.Second)
	for retries > 0 {
		stat, err := d.card_command(cmdGoIdleState, 0) // CMD0.
		if err != nil {
			if isTimeout(err) {
				retries--
				continue // Try again!
			}
			return err
		}
		if stat == _R1_IDLE_STATE {
			break
		} else if tm.expired() {
			retries = 0
			break
		}
		retries--
	}
	if retries <= 0 {
		return errNoSDCard
	}
	const enableCRC = true
	if enableCRC {
		stat, err := d.card_command(cmdCRCOnOff, 1) // CMD59.
		if err != nil {
			return err
		} else if stat != _R1_IDLE_STATE {
			return errors.New("sd:cant enable CRC")
		}
	}

	tm.setTimeout(time.Second)
	for {
		stat, err := d.card_command(cmdSendIfCond, 0x1AA) // CMD8.
		if err != nil {
			return err
		} else if stat == (_R1_ILLEGAL_COMMAND | _R1_IDLE_STATE) {
			d.kind = TypeSD1
			break
		}
		d.receive()
		d.receive()
		d.receive()
		status, err := d.receive()
		if err != nil {
			return err
		}
		if status == 0xaa {
			d.kind = TypeSD2
			break
		}
		d.yield()
	}

	var arg uint32
	if d.kind != TypeSD1 {
		arg = 0x4000_0000
	}
	tm.setTimeout(time.Second)
	for !tm.expired() {
		stat, err := d.card_acmd(acmdSD_APP_OP_COND, arg)
		if err != nil {
			return err
		} else if stat == 0 { // READY state.
			break
		}
		d.yield()
	}
	err := d.updateCSDCID()
	if err != nil {
		return err
	}

	if d.kind != TypeSD2 {
		return nil // Done if not SD2.
	}

	// Discover if card is high capacity.
	stat, err := d.card_command(cmdReadOCR, 0)
	if err != nil {
		return err
	} else if stat != 0 {
		return makeResponseError(response1(stat))
	}
	ocr, err := d.receive()
	if err != nil {
		return err
	} else if ocr&0xc0 == 0xc0 {
		d.kind = TypeSDHC
	}
	// Discard next 3 bytes.
	d.receive()
	d.receive()
	d.receive()
	return nil
}

func (d *SPICard) updateCSDCID() (err error) {
	// read CID
	d.cid, err = d.read_cid()
	if err != nil {
		return err
	}
	d.csd, err = d.read_csd()
	if err != nil {
		return err
	}
	blockshift := d.csd.ReadBlockLenShift()
	blocklen := uint16(1) << blockshift
	capacity := d.csd.DeviceCapacity()
	if blocklen == 0 || capacity < uint64(blocklen) {
		return errNoblocks
	}
	nb := capacity / uint64(blocklen)
	if nb > math.MaxUint32 {
		return errCardNotSupported
	}
	d.blockshift = blockshift
	d.numblocks = int64(nb)
	return nil
}

// ReadBlock reads to a buffer multiple of 512 bytes from sdcard into dst starting at block `startBlockIdx`.
func (d *SPICard) ReadBlocks(dst []byte, startBlockIdx int64) error {
	numblocks, err := d.checkBounds(startBlockIdx, len(dst))
	if err != nil {
		return err
	}
	if d.kind != TypeSDHC {
		startBlockIdx <<= 9 // Multiply by 512 for non high capacity SD cards.
	}
	d.csEnable(true)
	defer d.csEnable(false)

	if numblocks == 1 {
		_, err = d.card_command(cmdReadSingleBlock, uint32(startBlockIdx))
		if err != nil {
			return err
		}
		return d.read_data(dst)

	} else if numblocks > 1 {
		blocksize := 1 << d.blockshift
		_, err = d.card_command(cmdReadMultipleBlock, uint32(startBlockIdx))
		if err != nil {
			return err
		}

		for i := 0; i < numblocks; i++ {
			offset := i * blocksize
			err = d.read_data(dst[offset : offset+blocksize])
			if err != nil {
				return err
			}
		}
		_, err = d.card_command(cmdStopTransmission, 0)
		return err
	}
	panic("unreachable numblocks<=0")
}

// WriteBlocks writes to sdcard from a buffer multiple of 512 bytes from src starting at block `startBlockIdx`.
func (d *SPICard) WriteBlocks(data []byte, startBlockIdx int64) error {
	numblocks, err := d.checkBounds(startBlockIdx, len(data))
	if err != nil {
		return err
	}
	if d.kind != TypeSDHC {
		startBlockIdx <<= 9 // Multiply by 512 for non high capacity SD cards.
	}
	d.csEnable(true)
	defer d.csEnable(false)

	if numblocks == 1 {
		_, err = d.card_command(cmdWriteBlock, uint32(startBlockIdx))
		if err != nil {
			return err
		}
		err = d.write_data(tokSTART_BLOCK, data)
		if err != nil {
			return err
		}
		err = d.wait_not_busy()
		if err != nil {
			return err
		}
		status, err := d.card_command(cmdSendStatus, 0)
		if err != nil {
			return err
		} else if status != 0 {
			return makeResponseError(response1(status))
		}
		status, err = d.receive()
		if err != nil {
			return err
		} else if status != 0 {
			return errWrite
		}
		return nil

	} else if numblocks > 1 {
		// Start multi block write.
		blocksize := 1 << d.blockshift
		_, err = d.card_command(cmdWriteMultipleBlock, uint32(startBlockIdx))
		if err != nil {
			return err
		}

		for i := 0; i < numblocks; i++ {
			offset := i * blocksize
			err = d.waitNotBusy(d.timeout)
			if err != nil {
				return err
			}
			err = d.write_data(tokWRITE_MULT, data[offset:offset+blocksize])
			if err != nil {
				return err
			}
		}
		// Stop the multi write operation.
		err = d.waitNotBusy(d.timeout)
		if err != nil {
			return err
		}
		return d.send(tokSTOP_TRAN)
	}
	panic("unreachable numblocks<=0")
}

func (d *SPICard) checkBounds(startBlockIdx int64, datalen int) (numblocks int, err error) {
	if startBlockIdx >= d.numblocks {
		return 0, errOOB
	} else if startBlockIdx > math.MaxUint32 {
		return 0, errCardNotSupported
	}
	tz := bits.TrailingZeros(uint(datalen))
	if tz < int(d.blockshift) {
		return 0, errNeedBlockLenMultiple
	}
	numblocks = datalen >> d.blockshift
	if numblocks == 0 {
		return 0, io.ErrShortBuffer
	}
	return numblocks, nil
}

func (d *SPICard) read_cid() (csd CID, err error) {
	err = d.cmd_read(cmdSendCID, 0, d.buf[:16]) // CMD10.
	if err != nil {
		return csd, err
	}
	return DecodeCID(d.buf[:16])
}

func (d *SPICard) read_csd() (csd CSD, err error) {
	err = d.cmd_read(cmdSendCSD, 0, d.buf[:16]) // CMD9.
	if err != nil {
		return csd, err
	}
	return DecodeCSD(d.buf[:16])
}

func (d *SPICard) cmd_read(cmd command, args uint32, buf []byte) error {
	status, err := d.card_command(cmd, args)
	if err != nil {
		return err
	} else if status != 0 {
		return makeResponseError(response1(status))
	}
	return d.read_data(buf)
}

func (d *SPICard) card_acmd(acmd appcommand, args uint32) (uint8, error) {
	_, err := d.card_command(cmdAppCmd, 0)
	if err != nil {
		return 0, err
	}
	return d.card_command(command(acmd), args)
}

func (d *SPICard) card_command(cmd command, args uint32) (uint8, error) {
	const transmitterBit = 1 << 6
	err := d.wait_not_busy()
	if err != nil {
		return 0, err
	}
	buf := d.bufcmd[:6]
	// Start bit is always zero; transmitter bit is one since we are Host.

	buf[0] = transmitterBit | byte(cmd)
	binary.BigEndian.PutUint32(buf[1:5], args)
	buf[5] = crc7noshift(buf[:5]) | 1 // CRC and end bit which is always 1.

	err = d.bus.Tx(buf, nil)
	if cmd == cmdStopTransmission {
		d.receive() // skip stuff byte for stop read.
	}

	for i := 0; i < 512; i++ {
		result, err := d.receive()
		if err != nil {
			return 0, err
		}
		if result&0x80 == 0 {
			return result, nil
		}
	}
	return 0, errReadTimeout
}

func (d *SPICard) read_data(data []byte) (err error) {
	var status uint8
	for {
		status, err = d.receive()
		if err != nil {
			return err
		}
		if status != 0xff {
			break
		}
		d.yield()
	}
	if status != tokSTART_BLOCK {
		return errWaitStartBlock
	}
	err = d.bus.Tx(nil, data)
	if err != nil {
		return err
	}
	// CRC16 is always sent on a data block.
	crchi, _ := d.receive()
	crclo, _ := d.receive()
	d.lastCRC = uint16(crclo) | uint16(crchi)<<8
	return nil
}

func (s *SPICard) wait_not_busy() error {
	tm := s.timers[1].setTimeout(s.timeout)
	for {
		tok, err := s.receive()
		if err != nil {
			return err
		} else if tok == 0xff {
			break
		} else if tm.expired() {
			return errBusyTimeout
		}
		s.yield()
	}
	return nil
}

func (s *SPICard) write_data(tok byte, data []byte) error {
	if len(data) > 512 {
		return errors.New("data too long for write_data")
	}
	crc := CRC16(data)
	err := s.send(tok)
	if err != nil {
		return err
	}
	err = s.bus.Tx(data, nil)
	if err != nil {
		return err
	}
	err = s.send(byte(crc >> 8))
	if err != nil {
		return err
	}
	err = s.send(byte(crc))
	if err != nil {
		return err
	}
	status, err := s.receive()
	if err != nil {
		return err
	}
	if status&_DATA_RES_MASK != _DATA_RES_ACCEPTED {
		return makeResponseError(response1(status))
	}
	return nil
}

func (s *SPICard) receive() (byte, error) {
	return s.bus.Transfer(0xFF)
}

func (s *SPICard) send(b byte) error {
	_, err := s.bus.Transfer(b)
	return err
}
