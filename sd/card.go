package sd

import (
	"encoding/binary"
	"errors"
	"math"
	"strconv"
	"time"

	"tinygo.org/x/drivers"
)

var (
	errBadCSDCID        = errors.New("sd:bad CSD/CID in CRC or always1")
	errNoSDCard         = errors.New("sd:no card")
	errCardNotSupported = errors.New("sd:card not supported")
	errCmd8             = errors.New("sd:cmd8")
	errCmdOCR           = errors.New("sd:cmd_ocr")
	errCmdBlkLen        = errors.New("sd:cmd_blklen")
	errAcmdAppCond      = errors.New("sd:acmd_appOrCond")
	errWaitStartBlock   = errors.New("sd:did not find start block token")
	errNeed512          = errors.New("sd:need 512 bytes for I/O")
	errWrite            = errors.New("sd:write")
	errWriteTimeout     = errors.New("sd:write timeout")
	errBusyTimeout      = errors.New("sd:busy card timeout")
	errOOB              = errors.New("sd:oob block access")
	errNoblocks         = errors.New("sd:no readable blocks")
	errCmdGeneric       = errors.New("sd:command error")
)

type digitalPinout func(b bool)

type SPICard struct {
	bus       drivers.SPI
	cs        digitalPinout
	bufcmd    [6]byte
	buf       [512]byte
	bufTok    [1]byte
	kind      CardKind
	cid       CID
	csd       CSD
	lastCRC   uint16
	timers    [2]timer
	numblocks int64
	timeout   time.Duration
	// relative card address.
	rca uint32
}

func NewSPICard(spi drivers.SPI, cs digitalPinout) *SPICard {
	return &SPICard{bus: spi, cs: cs, timeout: 300 * time.Millisecond}
}

func (c *SPICard) csEnable(b bool) { c.cs(!b) }

// LastReadCRC returns the CRC for the last ReadBlock operation.
func (c *SPICard) LastReadCRC() uint16 { return c.lastCRC }

func (d *SPICard) Init() error {
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
	tm := d.timers[0].setTimeout(2 * time.Second)
	for !tm.expired() {
		// Wait up to 2 seconds to be the same as the Arduino
		result, err := d.cmd(cmdGoIdleState, 0, 0x95)
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
	r1, err := d.cmd(cmdSendIfCond, 0x01AA, 0x87)
	if err != nil {
		return err
	}
	if r1.IllegalCmdError() {
		d.kind = TypeSD1
		return errCardNotSupported
	}
	// r7 response
	status := byte(0)
	for i := 0; i < 3; i++ {
		var err error
		status, err = d.bus.Transfer(0xFF)
		if err != nil {
			return err
		}
	}
	if (status & 0x0F) != 0x01 {
		return makeResponseError(response1(status))
	}

	for i := 3; i < 4; i++ {
		var err error
		status, err = d.bus.Transfer(0xFF)
		if err != nil {
			return err
		}
	}
	if status != 0xAA {
		return makeResponseError(response1(status))
	}
	d.kind = TypeSD2

	// initialize card and send host supports SDHC if SD2
	arg := uint32(0)
	if d.kind == TypeSD2 {
		arg = 0x40000000
	}

	// check for timeout
	ok = false
	tm = tm.setTimeout(2 * time.Second)
	for !tm.expired() {
		r1, err = d.appCmd(acmdSD_APP_OP_COND, arg)
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
		err := d.cmdEnsure0Status(cmdReadOCR, 0, 0xFF)
		if err != nil {
			return err
		}

		statusb, err := d.bus.Transfer(0xFF)
		if err != nil {
			return err
		}
		if (statusb & 0xC0) == 0xC0 {
			d.kind = TypeSDHC
		}
		// discard rest of ocr - contains allowed voltage range
		for i := 1; i < 4; i++ {
			d.bus.Transfer(0xFF)
		}
	}
	err = d.cmdEnsure0Status(cmdSetBlocklen, 0x0200, 0xff)
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
	nb := d.csd.NumberOfBlocks()
	if nb > math.MaxUint32 {
		return errCardNotSupported
	} else if nb == 0 {
		return errNoblocks
	}
	d.numblocks = int64(nb)

	err = d.readRegister(cmdSendRelativeAddr, d.buf[:4])
	if err != nil {
		return err
	}
	d.rca = binary.BigEndian.Uint32(d.buf[:4])
	return nil
}

func (d *SPICard) NumberOfBlocks() uint64 {
	return uint64(d.numblocks)
}

// ReadBlock reads 512 bytes from sdcard into dst.
func (d *SPICard) ReadBlock(block int64, dst []byte) error {
	if len(dst) != 512 {
		return errNeed512
	} else if block >= d.numblocks {
		return errOOB
	}

	// use address if not SDHC card
	if d.kind != TypeSDHC {
		block <<= 9
	}

	err := d.cmdEnsure0Status(cmdReadSingleBlock, uint32(block), 0xff)
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
	hi, _ := d.bus.Transfer(0xFF)
	lo, _ := d.bus.Transfer(0xFF)
	d.lastCRC = uint16(hi)<<8 | uint16(lo)
	return nil
}

// WriteBlock writes 512 bytes from dst to sdcard.
func (d *SPICard) WriteBlock(block int64, src []byte) error {
	if len(src) != 512 {
		return errNeed512
	} else if block >= d.numblocks {
		return errOOB
	}

	// use address if not SDHC card
	if d.kind != TypeSDHC {
		block <<= 9
	}
	err := d.cmdEnsure0Status(cmdWriteBlock, uint32(block), 0xFF)
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
	d.bus.Transfer(0xFF)
	d.bus.Transfer(0xFF)

	// Data Resp.
	r, err := d.bus.Transfer(0xFF)
	if err != nil {
		return err
	}
	if (r & 0x1F) != 0x05 {
		return errWrite
	}

	err = d.waitNotBusy(2 * d.timeout)
	if err != nil {
		return errWriteTimeout
	}

	return nil
}

func (d *SPICard) ReadStatus() (response1, error) {
	if err := d.readRegister(cmdSendStatus, d.buf[:4]); err != nil {
		return 0, err
	}
	return response1(binary.BigEndian.Uint32(d.buf[:4])), nil
}

// CID returns a copy of the Card Identification Register value last read.
func (d *SPICard) CID() CID { return d.cid }

// CSD returns a copy of the Card Specific Data Register value last read.
func (d *SPICard) CSD() CSD { return d.csd }

func (d *SPICard) readCID() (CID, error) {
	buf := d.buf[len(d.buf)-16:]
	if err := d.readRegister(cmdSendCID, buf); err != nil {
		return CID{}, err
	}
	return DecodeCID(buf)
}

func (d *SPICard) readCSD() (CSD, error) {
	buf := d.buf[len(d.buf)-16:]
	if err := d.readRegister(cmdSendCSD, buf); err != nil {
		return CSD{}, err
	}
	return DecodeCSD(buf)
}

func (d *SPICard) readRegister(cmd command, dst []byte) error {
	err := d.cmdEnsure0Status(cmd, 0, 0xFF)
	if err != nil {
		return err
	}
	if err := d.waitStartBlock(); err != nil {
		return err
	}
	// transfer data
	for i := uint16(0); i < 16; i++ {
		r, err := d.bus.Transfer(0xFF)
		if err != nil {
			return err
		}
		dst[i] = r
	}
	// skip CRC.
	d.bus.Transfer(0xFF)
	d.bus.Transfer(0xFF)
	d.csEnable(false)
	return nil
}

func (d *SPICard) appCmd(cmd appcommand, arg uint32) (response1, error) {
	status, err := d.cmd(cmdAppCmd, 0, 0xFF)
	if err != nil {
		return status, err
	}
	return d.cmd(command(cmd), arg, 0xFF)
}

func (d *SPICard) cmdEnsure0Status(cmd command, arg uint32, crc byte) error {
	status, err := d.cmd(cmd, arg, crc)
	if err != nil {
		return err
	}
	if status != 0 {
		return makeResponseError(status)
	}
	return nil
}

func (d *SPICard) cmd(cmd command, arg uint32, precalculatedCRC byte) (response1, error) {
	const transmitterBit = 1 << 6
	if cmd >= transmitterBit {
		panic("invalid SD command")
	}
	d.csEnable(true)

	if cmd != cmdStopTransmission {
		err := d.waitNotBusy(d.timeout)
		if err != nil {
			return 0, err
		}
	}

	// create and send the command
	buf := d.bufcmd[:6]
	// Start bit is always zero; transmitter bit is one since we are Host.

	buf[0] = transmitterBit | byte(cmd)
	binary.BigEndian.PutUint32(buf[1:5], arg)
	if precalculatedCRC != 0 {
		buf[5] = precalculatedCRC
	} else {
		// CRC and end bit which is always 1.
		buf[5] = crc7noshift(buf[:5]) | 1
	}

	err := d.bus.Tx(buf, nil)
	if err != nil {
		return 0, err
	}
	if cmd == 12 {
		// skip 1 byte
		d.bus.Transfer(0xFF)
	}

	for i := 0; i < 0xFFFF; i++ {
		tok, _ := d.bus.Transfer(0xff)
		response := response1(tok)
		if (response & 0x80) == 0 {
			return response, nil
		}
	}

	d.csEnable(false)
	d.bus.Transfer(0xFF)
	return 0xFF, errCmdGeneric
}

func (d *SPICard) waitNotBusy(timeout time.Duration) error {
	if d.waitToken(timeout, 0xff) {
		return nil
	}
	return errBusyTimeout
}

func (d *SPICard) waitStartBlock() error {
	if d.waitToken(d.timeout, tokSTART_BLOCK) {
		return nil
	}
	d.csEnable(false)
	return errWaitStartBlock
}

// waitToken transmits over SPI waiting to read a given byte token. If argument tok
// is 0xff then waitToken will wait for a token that does NOT match 0xff.
func (d *SPICard) waitToken(timeout time.Duration, tok byte) bool {
	tm := d.timers[1].setTimeout(timeout)
	for {
		received, err := d.bus.Transfer(0xFF)
		if err != nil {
			return false
		}
		matchTok := received == tok
		if matchTok || (!matchTok && tok == 0xff) {
			return true
		} else if tm.expired() {
			return false
		}
	}
}

type response1Err struct {
	context string
	status  response1
}

func (e response1Err) Error() string {
	return e.status.Response()
	if e.context != "" {
		return "sd:" + e.context + " " + strconv.Itoa(int(e.status))
	}
	return "sd:status " + strconv.Itoa(int(e.status))
}

func (e response1) Response() string {
	b := make([]byte, 0, 8)
	return string(e.appendf(b))
}

func (r response1) appendf(b []byte) []byte {
	b = append(b, '[')
	if r.IsIdle() {
		b = append(b, "idle,"...)
	}
	if r.EraseReset() {
		b = append(b, "erase-rst,"...)
	}
	if r.EraseSeqError() {
		b = append(b, "erase-seq,"...)
	}
	if r.CRCError() {
		b = append(b, "crc-err,"...)
	}
	if r.AddressError() {
		b = append(b, "addr-err,"...)
	}
	if r.ParamError() {
		b = append(b, "param-err,"...)
	}
	if r.IllegalCmdError() {
		b = append(b, "illegal-cmd,"...)
	}
	if len(b) > 1 {
		b = b[:len(b)-1]
	}
	b = append(b, ']')
	return b
}

func makeResponseError(status response1) error {
	return response1Err{
		status: status,
	}
}

var timeoutTimer [2]timer

type timer struct {
	deadline time.Time
}

func (t *timer) setTimeout(timeout time.Duration) *timer {
	t.deadline = time.Now().Add(timeout)
	return t
}

func (t timer) expired() bool {
	return time.Since(t.deadline) >= 0
}
