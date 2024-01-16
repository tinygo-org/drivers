package sd

import (
	"errors"
	"time"

	"tinygo.org/x/drivers"
)

// See rustref.go for the new implementation.

var (
	errBadCSDCID            = errors.New("sd:bad CSD/CID in CRC or always1")
	errNoSDCard             = errors.New("sd:no card")
	errCardNotSupported     = errors.New("sd:card not supported")
	errWaitStartBlock       = errors.New("sd:did not find start block token")
	errNeedBlockLenMultiple = errors.New("sd:need blocksize multiple for I/O")
	errWrite                = errors.New("sd:write")
	errWriteTimeout         = errors.New("sd:write timeout")
	errReadTimeout          = errors.New("sd:read timeout")
	errBusyTimeout          = errors.New("sd:busy card timeout")
	errOOB                  = errors.New("sd:oob block access")
	errNoblocks             = errors.New("sd:no readable blocks")
)

type digitalPinout = func(b bool)

type SPICard struct {
	bus drivers.SPI
	cs  digitalPinout

	timers  [2]timer
	timeout time.Duration
	wait    time.Duration
	// Card Identification Register.
	cid CID
	// Card Specific Register.
	csd    CSD
	bufcmd [6]byte
	kind   CardKind
	// block indexing helper based on block size.
	blk     blkIdxer
	lastCRC uint16
}

func NewSPICard(spi drivers.SPI, cs digitalPinout) *SPICard {
	const defaultTimeout = 300 * time.Millisecond
	s := &SPICard{
		bus: spi,
		cs:  cs,
	}
	s.setTimeout(defaultTimeout)
	return s
}

// setTimeout sets the timeout for all operations and the wait time between each yield during busy spins.
func (c *SPICard) setTimeout(timeout time.Duration) {
	if timeout <= 0 {
		panic("timeout must be positive")
	}
	c.timeout = timeout
	c.wait = timeout / 512
}

func (c *SPICard) csEnable(b bool) {
	c.cs(!b)
}

// LastReadCRC returns the CRC for the last ReadBlock operation.
func (c *SPICard) LastReadCRC() uint16 { return c.lastCRC }

// Init initializes the SD card. This routine should be performed with a SPI clock
// speed of around 100..400kHz. One may increase the clock speed after initialization.
func (d *SPICard) Init() error {
	return d.initRs()
}

func (d *SPICard) NumberOfBlocks() int64 {
	return d.csd.NumberOfBlocks()
}

// CID returns a copy of the Card Identification Register value last read.
func (d *SPICard) CID() CID { return d.cid }

// CSD returns a copy of the Card Specific Data Register value last read.
func (d *SPICard) CSD() CSD { return d.csd }

func (d *SPICard) yield() { time.Sleep(d.wait) }

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
