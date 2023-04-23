package ndir

import (
	"errors"
	"fmt"
	"runtime"
	"time"

	"tinygo.org/x/drivers"
)

// Addr returns the I2C address given the solder pad configuration on the Sandbox Electronics i2c/uart converter.
// When the resistor is connected between the left and middle pads the bit is said to be set
// and a0 or a1 should be passed in as true.
func Addr(a0, a1 bool) uint8 {
	return 0b1001000 | b2u8(a0) | b2u8(a1)<<2
}

func b2u8(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}

// See https://github.com/SandboxElectronics/NDIR/blob/master/NDIR_I2C/NDIR_I2C.cpp

// General Registers
const (
	AddrRHR       = 0x00
	AddrTHR       = 0x00
	AddrIER       = 0x01
	AddrFCR       = 0x02
	AddrIIR       = 0x02
	AddrLCR       = 0x03
	AddrMCR       = 0x04
	AddrLSR       = 0x05
	AddrMSR       = 0x06
	AddrSPR       = 0x07
	AddrTCR       = 0x06
	AddrTLR       = 0x07
	AddrTXLVL     = 0x08
	AddrRXLVL     = 0x09
	AddrIODIR     = 0x0A
	AddrIOSTATE   = 0x0B
	AddrIOINTENA  = 0x0C
	AddrIOCONTROL = 0x0E // This addr fails on write of 0x08?
	AddrEFCR      = 0x0F
)

// Special registers
const (
	AddrDLL = 0x00
	AddrDLH = 1
)

const (
	shortTxCooldown = time.Millisecond
	longTxCooldown  = 10 * time.Millisecond
	rxTimeout       = 100 * time.Millisecond
)

var (
	cmd_readCO2                = [...]byte{0xFF, 0x01, 0x86, 0x00, 0x00, 0x00, 0x00, 0x00, 0x79}
	cmd_measure                = [...]byte{0xFF, 0x01, 0x9C, 0x00, 0x00, 0x00, 0x00, 0x00, 0x63}
	cmd_calibrateZero          = [...]byte{0xFF, 0x01, 0x87, 0x00, 0x00, 0x00, 0x00, 0x00, 0x78}
	cmd_enableAutoCalibration  = [...]byte{0xFF, 0x01, 0x79, 0xA0, 0x00, 0x00, 0x00, 0x00, 0xE6}
	cmd_disableAutoCalibration = [...]byte{0xFF, 0x01, 0x79, 0x00, 0x00, 0x00, 0x00, 0x00, 0x86}
)

// DevI2C is a handle to a MH-Z16 NDIR CO2 Sensorusing the I2C interface.
type DevI2C struct {
	bus             drivers.I2C
	addr            uint8
	nextAvail       time.Time
	initTime        time.Time
	lastMeasurement int32
}

// NewDevI2C returns a new NDIR device ready for use. It performs no I/O.
func NewDevI2C(bus drivers.I2C, addr uint8) *DevI2C {
	return &DevI2C{
		bus:             bus,
		addr:            addr,
		lastMeasurement: -1,
	}
}

// PPM returns the CO2 parts per million read in the last Update call.
func (d *DevI2C) PPM() int32 {
	return d.lastMeasurement
}

var errInitWait = errors.New("ndir: must wait 12 seconds after init before reading concentration")

// Update reads the PPM from the NDIR and stores it ready for the PPM() method.
func (d *DevI2C) Update() (err error) {
	if time.Since(d.initTime) < 12*time.Second {
		// Wait 12 seconds before performing first read.
		return nil
	}
	err = d.writeRegister(AddrFCR, 0x07)
	if err != nil {
		return err
	}
	err = d.send(cmd_measure[:])
	if err != nil {
		return fmt.Errorf("sending cmd_measure: %w", err)
	}
	time.Sleep(11 * time.Millisecond)
	var buf [9]byte
	buf, err = d.receive()

	if err != nil {
		return fmt.Errorf("receiving during measure: %w", err)
	}
	if buf[0] != 0xff && buf[1] != 0x9c {
		return fmt.Errorf("buffer rx bad values: %q", string(buf[:]))
	}
	var sum uint16
	for i := 0; i < len(buf); i++ {
		sum += uint16(buf[i])
	}
	mod := sum % 256
	if mod != 0xff {
		return fmt.Errorf("ndir checksum modulus got %#x, expected 0xff", mod)
	}
	ppm := uint32(buf[2])<<24 | uint32(buf[3])<<16 | uint32(buf[4])<<8 | uint32(buf[5])
	d.lastMeasurement = int32(ppm)
	return nil
}

func (d *DevI2C) Hardware() string {
	return fmt.Sprintf("NDIR %#x@%s", d.addr, d.bus)
}

func (d *DevI2C) Init() (err error) {
	// AddrIOCONTROL write is always NACKed so ignore.
	err = d.writeRegister(AddrIOCONTROL, 0x08)
	if err != nil {
		println(err.Error())
	}
	err = d.writeRegister(AddrFCR, 0x07)
	if err != nil {
		return err
	}
	err = d.writeRegister(AddrLCR, 0x83)
	if err != nil {
		return err
	}
	err = d.writeRegister(AddrDLL, 0x60)
	if err != nil {
		return err
	}
	err = d.writeRegister(AddrDLH, 0x00)
	if err != nil {
		return err
	}
	err = d.writeRegister(AddrLCR, 0x03)
	if err != nil {
		return err
	}
	d.initTime = time.Now()
	return nil
}

// CalibrateZero calibrates the NDIR to around 412ppm.
func (d *DevI2C) CalibrateZero() error {
	return d.enactCommand(cmd_calibrateZero[:])
}

// SetAutoCalibration can enable or disable the NDIR's auto calibration mode.
func (d *DevI2C) SetAutoCalibration(enable bool) (err error) {
	if enable {
		err = d.enactCommand(cmd_enableAutoCalibration[:])
	} else {
		err = d.enactCommand(cmd_disableAutoCalibration[:])
	}
	return err
}

func (d *DevI2C) send(cmd []byte) error {
	txlvl, err := d.ReadRegister(AddrTXLVL)
	if err != nil {
		return err
	}
	if int(txlvl) < len(cmd) {
		return fmt.Errorf("txlvl=%d less than length of command %d", txlvl, len(cmd))
	}
	return d.tx(append([]byte{AddrTHR}, cmd...), nil)
}

func (d *DevI2C) receive() (cmd [9]byte, err error) {
	start := time.Now()
	n := uint8(9)
	for n > 0 {
		if time.Since(start) > rxTimeout {
			return [9]byte{}, errors.New("NDIR rx timeout")
		}
		rxlvl, err := d.ReadRegister(AddrRXLVL)
		if err != nil {
			return [9]byte{}, err
		}
		if rxlvl > n {
			rxlvl = n
		}
		ptr := 9 - n
		err = d.tx([]byte{AddrRHR << 3}, cmd[ptr:ptr+rxlvl])
		n -= rxlvl
		if err != nil {
			return [9]byte{}, err
		}
	}
	return cmd, nil
}

func (d *DevI2C) enactCommand(cmd []byte) error {
	if len(cmd) > 31 {
		return errors.New("ndir: command too long")
	}
	// Most commands always start with the same FCR write here.
	err := d.writeRegister(AddrFCR, 0x07)
	if err != nil {
		return err
	}
	time.Sleep(longTxCooldown)

	// C++ send method begins here.
	got, err := d.ReadRegister(AddrTXLVL)
	if err != nil {
		return err
	}
	if got < uint8(len(cmd)) {
		return fmt.Errorf("ndir: txlevel=%d too low for command of length %d", got, len(cmd))
	}
	var buf [32]byte
	buf[0] = AddrTHR
	n := 1 + copy(buf[1:], cmd)
	err = d.tx(buf[:n], nil)
	if err != nil {
		return err
	}
	d.nextAvail.Add(longTxCooldown) // add some extra time.
	return nil
}

func (d *DevI2C) writeRegister(addr, val uint8) (err error) {
	return d.WriteRegisters(addr, []byte{val})
}

func (d *DevI2C) WriteRegisters(addr uint8, vals []byte) (err error) {
	var buf [32]byte
	if len(vals) > 31 {
		return errors.New("can only write up to 31 bytes")
	}
	buf[0] = addr << 3
	n := copy(buf[1:], vals)
	err = d.tx(buf[:n+1], nil)
	if err != nil {
		err = fmt.Errorf("NDIR write %#x (%d) to %#x: %w", buf[1], len(vals), buf[0], err)
	}
	return err
}

func (d *DevI2C) ReadRegister(addr uint8) (uint8, error) {
	var buf [2]byte
	buf[0] = addr << 3
	err := d.tx(buf[:1], buf[1:2])
	if err != nil {
		err = fmt.Errorf("NDIR read from %#x: %w", buf[0], err)
	}
	return buf[1], err
}

func (d *DevI2C) tx(w, r []byte) error {
	wait := time.Until(d.nextAvail)
	if wait > 0 {
		// Try yielding process first, maybe there's a short time to wait and a schedule call is enough delay.
		runtime.Gosched()
		wait = time.Until(d.nextAvail)
		if wait > 0 {
			// If yielding did not work then perform sleep
			time.Sleep(wait)
		}
	}
	err := d.bus.Tx(uint16(d.addr), w, r)
	d.nextAvail = time.Now().Add(shortTxCooldown)
	return err
}
