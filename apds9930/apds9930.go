package apds9930

import (
	"errors"

	"tinygo.org/x/drivers"
)

var errInvalidParam = errors.New("apds9930: invalid param")

type Dev struct {
	bus    drivers.I2C
	_txerr error
	addr   uint16
	buf    [3]byte
}

func New(bus drivers.I2C, addr uint8) Dev {
	return Dev{bus: bus, addr: uint16(addr)}
}

// Status contains info on:
//
//	AVALID: Indicates that the ALS Ch0/Ch1 channels have completed an integration cycle.
//	PSValid. Indicates that the PS has completed an integration cycle.
//	AINTL ALS Interrupt. Indicates that the device is asserting an ALS interrupt
//	PINT: Proximity Interrupt. Indicates that the device is asserting a proximity interrupt.
//	PSAT: Proximity Saturation. Indicates that the proximity measurement is saturated
type Status uint8

func (s Status) ALSAvailable() bool         { return s&(1<<0) != 0 } // AVALID
func (s Status) ProximityAvailable() bool   { return s&(1<<1) != 0 } // PVALID
func (s Status) HasALSInterrupt() bool      { return s&(1<<4) != 0 } // AINT
func (s Status) HasProxInterrupt() bool     { return s&(1<<5) != 0 } // PINT
func (s Status) IsProximitySaturated() bool { return s&(1<<6) != 0 } // PSAT

type Enable uint8

const (
	EnPower Enable = 1 << iota
	EnALS
	EnProx
	EnWait
	EnALSInt
	EnProxInt
	EnSleepAfterInt
)

// Luminic control gain.
type ALSGain uint8

const (
	AGain1 ALSGain = iota
	AGain8
	AGain16
	AGain120
)

type ProxGain uint8

const (
	PGain1 ProxGain = iota
	PGain2
	PGain4
	PGain8
)

type Drive uint8

const (
	Drive100mA Drive = iota
	Drive50mA
	Drive25mA
	Drive12_5mA
)

type Config struct {
	ProxGain ProxGain
	ALSGain  ALSGain
	LEDDrive Drive
}

func (d *Dev) Init(cfg Config) error {
	if cfg.LEDDrive > Drive100mA {
		return errInvalidParam
	}
	d.txNew()
	d.txWrite8(regENABLE, 0x00) // disable all features.
	d.txWrite8(regATIME, 0xee)  // set default integration time.
	d.txWrite8(regPPULSE, 0x04)
	d.txWrite8(regWTIME, 0xee) // set default wait time.
	d.txWrite8(regPTIME, 0xff) // set default pulse count.

	var ctlval uint8 = 0b10 << 4 // Use Channel 1 diode.
	ctlval |= uint8(cfg.LEDDrive&0b11) << 6
	ctlval |= uint8(cfg.ProxGain&0b11) << 2
	ctlval |= uint8(cfg.ALSGain & 0b11)
	d.txWrite8(regCONTROL, ctlval)
	return d.txErr()
}

func (d *Dev) Status() (Status, error) {
	d.txNew()
	v := d.txRead8(regSTATUS)
	return Status(v), d.txErr()
}

// Enable sets the ENABLE register used primarily to
// power the APDS-9930 device on/off, enable functions, and interrupts.
// Arguments must be ORed, i.e: d.Enable(EnPower|EnProx); to enable proximity.
func (d *Dev) Enable(en Enable) error {
	en &= 0b01111111 // Seventh bit reserved.
	d.txNew()
	d.txWrite8(regENABLE, uint8(en))
	return d.txErr()
}

func (d *Dev) enableLightSensor(withInterrupts bool) error {
	return nil
}

func (d *Dev) setAmbientLightGain() {
}

func (d *Dev) EnableProximity() error {
	return d.Enable(EnPower | EnALS | EnProx | EnWait)
}

func (d *Dev) proxIntLowThresh() (uint16, error) {
	d.txNew()
	return d.txRead16(regPILTL), d.txErr()
}

func (d *Dev) setProxIntLowThresh(loThresh uint16) error {
	d.txNew()
	d.txWrite16(regPILTL, loThresh)
	return d.txErr()
}

func (d *Dev) proxIntHighThresh() (uint16, error) {
	d.txNew()
	val := d.txRead16(regPIHTL)
	return val, d.txErr()
}

func (d *Dev) setProxIntHighThresh(hiThresh uint16) error {
	d.txNew()
	d.txWrite16(regPIHTL, hiThresh)
	return d.txErr()
}

func (d *Dev) LEDDrive() (Drive, error) {
	d.txNew()
	val := (d.txRead8(regCONTROL) >> 6) & 0b11
	return Drive(val), d.txErr()
}

// SetLEDDrive drive strength for proximity and ALS
//
//	Value    LED Current
//	  3         100 mA
//	  2         50 mA
//	  1         25 mA
//	  0         12.5 mA
func (d *Dev) SetLEDDrive(drive Drive) error {
	if drive > 3 {
		return errInvalidParam
	}
	current, err := d.LEDDrive()
	if err != nil {
		return err
	}
	// Replace LED bits in Control register.
	current &= 0b00111111
	current |= drive << 6
	d.txNew()
	d.txWrite8(regCONTROL, uint8(current))
	return d.txErr()
}

func (d *Dev) proxGain() (uint8, error) {
	val := d.txRead8(regCONTROL)
	return (val >> 2) & 0b11, d.txErr()
}

// ReadProximity returns a 10-bit value (0..1023), the higher the value the closer the object
func (d *Dev) ReadProximity() uint16 {
	d.txNew()
	v := d.txRead16(regPDATAL)
	if d.txErr() != nil {
		return 0
	}
	return v
}

func (d *Dev) txRead16(addr uint8) uint16 {
	if d.txErr() != nil {
		return 0
	}
	d.buf[0] = addr | protoAutoInc
	d._txerr = d.bus.Tx(d.addr, d.buf[:1], d.buf[1:3])
	return uint16(d.buf[1]) | uint16(d.buf[2])<<8

}

func (d *Dev) txRead8(addr uint8) uint8 {
	if d.txErr() != nil {
		return 0
	}
	d.buf[0] = addr | protoAutoInc
	d._txerr = d.bus.Tx(d.addr, d.buf[:1], d.buf[1:2])
	return d.buf[1]
}

func (d *Dev) txWrite16(addr uint8, val uint16) {
	d.txWrite8(addr, uint8(val))
	d.txWrite8(addr+1, uint8(val>>8))
}

func (d *Dev) txWrite8(reg uint8, val uint8) {
	if d.txErr() != nil {
		return
	}
	d.buf[0] = reg | 0x80
	d.buf[1] = val
	d._txerr = d.bus.Tx(d.addr, d.buf[:2], nil)
}

func (d *Dev) txNew() { d._txerr = nil }

func (d *Dev) txErr() error { return d._txerr }
