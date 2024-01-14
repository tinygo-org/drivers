package sd

import (
	"encoding/binary"
	"errors"
	"runtime"
	"strconv"
	"time"

	"tinygo.org/x/drivers"
)

var (
	errNoSDCard         = errors.New("sd:no card")
	errCardNotSupported = errors.New("sd:card not supported")
	errCmd8             = errors.New("sd:cmd8")
	errCmdOCR           = errors.New("sd:cmd_ocr")
	errCmdBlkLen        = errors.New("sd:cmd_blklen")
	errAcmdAppCond      = errors.New("sd:acmd_appOrCond")
	errWaitStartBlock   = errors.New("sd:wait start block")
	errNeed512          = errors.New("sd:need 512 bytes for I/O")
	errWrite            = errors.New("sd:write")
	errWriteTimeout     = errors.New("sd:write timeout")
)

type digitalPinout func(b bool)

type Card struct {
	bus     drivers.SPI
	cs      digitalPinout
	bufcmd  [6]byte
	buf     [512]byte
	bufTok  [1]byte
	kind    CardKind
	cid     CID
	csd     CSD
	lastCRC uint16
}

func NewCard(spi drivers.SPI, cs digitalPinout) *Card {
	return &Card{bus: spi, cs: cs}
}

func (c *Card) csEnable(b bool) { c.cs(!b) }

// LastReadCRC returns the CRC for the last ReadBlock operation.
func (c *Card) LastReadCRC() uint16 { return c.lastCRC }

func (d *Card) Init() error {
	dummy := d.buf[:]
	for i := range dummy {
		dummy[i] = 0xFF
	}
	defer d.csEnable(false)

	d.csEnable(true)
	// clock card at least 100 cycles with cs high
	d.bus.Tx(dummy[:10], nil)
	d.csEnable(false)

	d.bus.Tx(dummy[:], nil)

	// CMD0: init card; sould return _R1_IDLE_STATE (allow 5 attempts)
	ok := false
	tm := setTimeout(0, 2*time.Second)
	for !tm.expired() {
		// Wait up to 2 seconds to be the same as the Arduino
		result, err := d.cmd(CMD0_GO_IDLE_STATE, 0, 0x95)
		if err != nil {
			return err
		}
		if result == _R1_IDLE_STATE {
			ok = true
			break
		}
	}
	if !ok {
		return errNoSDCard
	}

	// CMD8: determine card version
	r1, err := d.cmd(CMD8_SEND_IF_COND, 0x01AA, 0x87)
	if err != nil {
		return err
	}
	if r1.IllegalCmdError() {
		d.kind = TypeSD1
		return errCardNotSupported
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
			return makeResponseError(response1(status))
		}

		for i := 3; i < 4; i++ {
			var err error
			status, err = d.bus.Transfer(byte(0xFF))
			if err != nil {
				return err
			}
		}
		if status != 0xAA {
			return makeResponseError(response1(status))
		}
		d.kind = TypeSD2
	}

	// initialize card and send host supports SDHC if SD2
	arg := uint32(0)
	if d.kind == TypeSD2 {
		arg = 0x40000000
	}

	// check for timeout
	ok = false
	tm = setTimeout(0, 2*time.Second)
	for !tm.expired() {
		r1, err = d.appCmd(ACMD41_SD_APP_OP_COND, arg)
		if err != nil {
			return err
		}
		if r1 == 0 {
			break
		}
	}
	if r1 != 0 {
		return makeResponseError(r1)
	}

	// if SD2 read OCR register to check for SDHC card
	if d.kind == TypeSD2 {
		err := d.cmdEnsure0Status(CMD58_READ_OCR, 0, 0xFF)
		if err != nil {
			return err
		}

		statusb, err := d.bus.Transfer(byte(0xFF))
		if err != nil {
			return err
		}
		if (statusb & 0xC0) == 0xC0 {
			d.kind = TypeSDHC
		}
		// discard rest of ocr - contains allowed voltage range
		for i := 1; i < 4; i++ {
			d.bus.Transfer(byte(0xFF))
		}
	}
	err = d.cmdEnsure0Status(CMD16_SET_BLOCKLEN, 0x0200, 0xff)
	if err != nil {
		return err
	}

	// read CID
	d.cid, err = d.readCID()
	if err != nil {
		return err
	}
	d.csd, err = d.readCSD()
	if err != nil {
		return err
	}
	return nil
}

// ReadData reads 512 bytes from sdcard into dst.
func (d *Card) ReadBlock(block uint32, dst []byte) error {
	if len(dst) != 512 {
		return errNeed512
	}

	// use address if not SDHC card
	if d.kind != TypeSDHC {
		block <<= 9
	}
	err := d.cmdEnsure0Status(CMD17_READ_SINGLE_BLOCK, block, 0xFF)
	if err != nil {
		return err
	}
	defer d.csEnable(false)

	if err := d.waitStartBlock(); err != nil {
		return err
	}
	buf := d.buf[:]
	err = d.bus.Tx(buf, dst)
	if err != nil {
		return err
	}

	// skip CRC (2byte)
	hi, _ := d.bus.Transfer(byte(0xFF))
	lo, _ := d.bus.Transfer(byte(0xFF))
	d.lastCRC = uint16(hi)<<8 | uint16(lo)
	return nil
}

// WriteBlock writes 512 bytes from dst to sdcard.
func (d *Card) WriteBlock(block uint32, src []byte) error {
	if len(src) != 512 {
		return errNeed512
	}

	// use address if not SDHC card
	if d.kind != TypeSDHC {
		block <<= 9
	}
	err := d.cmdEnsure0Status(CMD24_WRITE_BLOCK, block, 0xFF)
	if err != nil {
		return err
	}
	defer d.csEnable(false)
	// wait 1 byte?
	token := byte(0xFE)
	d.bus.Transfer(token)

	err = d.bus.Tx(src[:512], nil)
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
		return errWrite
	}

	// wait no busy
	err = d.waitNotBusy(600 * time.Millisecond)
	if err != nil {
		return errWriteTimeout
	}

	return nil
}

// CID returns a copy of the Card Identification Register value last read.
func (d *Card) CID() CID { return d.cid }

// CSD returns a copy of the Card Specific Data Register value last read.
func (d *Card) CSD() CSD { return d.csd }

func (d *Card) readCID() (CID, error) {
	buf := d.buf[len(d.buf)-16:]
	if err := d.readRegister(CMD10_SEND_CID, buf); err != nil {
		return CID{}, err
	}
	return DecodeCID(buf)
}

func (d *Card) readCSD() (CSD, error) {
	buf := d.buf[len(d.buf)-16:]
	if err := d.readRegister(CMD9_SEND_CSD, buf); err != nil {
		return CSD{}, err
	}
	return DecodeCSD(buf)
}

func (d *Card) readRegister(cmd uint8, dst []byte) error {
	err := d.cmdEnsure0Status(cmd, 0, 0xFF)
	if err != nil {
		return err
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
	// skip CRC.
	d.bus.Transfer(byte(0xFF))
	d.bus.Transfer(byte(0xFF))
	d.csEnable(false)
	return nil
}

func (d *Card) appCmd(cmd byte, arg uint32) (response1, error) {
	status, err := d.cmd(CMD55_APP_CMD, 0, 0xFF)
	if err != nil {
		return status, err
	}
	return d.cmd(cmd, arg, 0xFF)
}

func (d *Card) cmdEnsure0Status(cmd byte, arg uint32, crc byte) error {
	status, err := d.cmd(cmd, arg, crc)
	if err != nil {
		return err
	}
	if status != 0 {
		return makeResponseError(status)
	}
	return nil
}

func (d *Card) cmd(cmd byte, arg uint32, crc byte) (response1, error) {
	d.csEnable(true)

	if cmd != 12 {
		d.waitNotBusy(300 * time.Millisecond)
	}

	// create and send the command
	buf := d.bufcmd[:]
	buf[0] = 0x40 | cmd
	binary.BigEndian.PutUint32(buf[1:5], arg)
	buf[5] = crc
	d.bus.Tx(buf, nil)

	if cmd == 12 {
		// skip 1 byte
		d.bus.Transfer(byte(0xFF))
	}

	// wait for the response (response[7] == 0)
	d.buf[0] = 0xFF
	dummy := d.buf[:1]
	for i := 0; i < 0xFFFF; i++ {
		d.bus.Tx(dummy, d.bufTok[:])
		response := response1(d.bufTok[0])
		if (response & 0x80) == 0 {
			return response, nil
		}
	}

	// TODO
	//// timeout
	d.csEnable(false)
	d.bus.Transfer(byte(0xFF))

	return 0xFF, nil // -1
}

func (d *Card) waitNotBusy(timeout time.Duration) error {
	tm := setTimeout(1, timeout)
	for !tm.expired() {
		r, err := d.bus.Transfer(byte(0xFF))
		if err != nil {
			return err
		}
		if r == 0xFF {
			return nil
		}
		runtime.Gosched()
	}
	return nil
}

func (d *Card) waitStartBlock() error {
	status := byte(0xFF)

	tm := setTimeout(0, 300*time.Millisecond)
	for !tm.expired() {
		var err error
		status, err = d.bus.Transfer(byte(0xFF))
		if err != nil {
			d.csEnable(false)
			return err
		}
		if status != 0xFF {
			break
		}
	}

	if status != 254 {
		d.csEnable(false)
		return errWaitStartBlock
	}

	return nil
}

type response1Err struct {
	context string
	status  response1
}

func (e response1Err) Error() string {
	if e.context != "" {
		return "sd:" + e.context + " " + strconv.Itoa(int(e.status))
	}
	return "sd:status " + strconv.Itoa(int(e.status))
}

func makeResponseError(status response1) error {
	return response1Err{
		status: status,
	}
}
