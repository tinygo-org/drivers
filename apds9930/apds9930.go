package apds9930

import "tinygo.org/x/drivers"

type Dev struct {
	bus    drivers.I2C
	_txerr error
	addr   uint16
	buf    [8]byte
}

func New(bus drivers.I2C, addr uint8) Dev {
	return Dev{bus: bus, addr: uint16(addr)}
}

func (d *Dev) Init() error {
	d.txNew()
	d.txWrite8(regENABLE, 0x00) // disable all features.
	d.txWrite8(regATIME, 0xee)  // set default integration time.
	d.txWrite8(regPPULSE, 0x04)
	d.txWrite8(regWTIME, 0xee) // set default wait time.
	d.txWrite8(regPTIME, 0xff) // set default pulse count.
	d.txWrite8(regCONTROL, 0x20)
	return d.txErr()
}

func (d *Dev) EnableProximity() error {
	d.txNew()
	d.txWrite8(regENABLE, 8|4|2|1)
	return d.txErr()
}

func (d *Dev) ProximityAvailable() bool {
	d.txNew()
	return d.txRead8(regSTATUS)&0x20 == 1
}

func (d *Dev) ReadProximity() uint8 {
	d.txNew()
	h := d.txRead8(regPDATAH)
	l := d.txRead8(regPDATAL)
	if d.txErr() != nil {
		return 0
	}
	return (h << 8) | l
}

func (d *Dev) txRead8(reg uint8) uint8 {
	if d.txErr() != nil {
		return 0
	}
	d.buf[0] = reg | 0xa0
	d._txerr = d.bus.Tx(d.addr, d.buf[:1], d.buf[1:2])
	return d.buf[1]
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
