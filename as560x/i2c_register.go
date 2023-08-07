package as560x // import tinygo.org/x/drivers/ams560x

import (
	"encoding/binary"
	"errors"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/internal/legacy"
)

// registerAttributes is a bitfield of attributes for a register
type registerAttributes uint8

const (
	// reg_read indicates that the register is readable
	reg_read registerAttributes = 1 << iota
	// reg_write indicates that the register is writeable
	reg_write
	// reg_program indicates that the register can be permanently programmed ('BURNed')
	reg_program
)

var (
	errRegisterNotReadable  = errors.New("Register is not readable")
	errRegisterNotWriteable = errors.New("Register is not writeable")
)

// i2cRegister encapsulates the address, structure and read/write logic for a register on a AS560x device
type i2cRegister struct {
	// host is the 'host register' for virtual registers. Physical/root registers have this set to self
	host *i2cRegister
	// address is the i2c address of the register. For 2-byte (word) addresses it's the low byte which holds the MSBs
	address uint8
	// shift is the number of bits the value is 'left shifted' into the register byte/word (0-15)
	shift uint16
	// mask is a bitwise mask applied to the register AFTER 'right shifting' to mask the register value
	mask uint16
	// num_bytes is the width of the register in bytes, 1 or 2.
	num_bytes uint8
	// attributes holds the register attributes. A bitfield of REG_xyz constants
	attributes registerAttributes
	// cached indicates whether we are holding a cached value of the register in value
	cached bool
	// value can be used as a 'cache' of the register's value for writeable registers.
	value uint16
}

// newI2CRegister returns a pointer to a new i2cRegister with no cached value
func newI2CRegister(address uint8, shift uint16, mask uint16, num_bytes uint8, attributes registerAttributes) *i2cRegister {
	reg := &i2cRegister{
		address:    address,
		shift:      shift,
		mask:       mask,
		num_bytes:  num_bytes,
		attributes: attributes,
	}
	// root registers host themselves
	reg.host = reg
	return reg
}

// newVirtualRegister returns a pointer to a new i2cRegister with the given host register and shift/mask.
func newVirtualRegister(host *i2cRegister, shift uint16, mask uint16) *i2cRegister {
	return &i2cRegister{
		host:       host,
		address:    host.address,
		shift:      shift,
		mask:       mask,
		num_bytes:  host.num_bytes,
		attributes: host.attributes,
	}
}

// invalidate invalidates any cached value for the register and forces an I2C read on the next read()
func (r *i2cRegister) invalidate() {
	r.host.cached = false
	r.host.value = 0
}

// readShiftAndMask is an internal method to read a value for the register over the given I2C bus from the device with the given address applying the given shift and mask
func (r *i2cRegister) readShiftAndMask(bus drivers.I2C, deviceAddress uint8, shift uint16, mask uint16) (uint16, error) {
	if r.host.attributes&reg_read == 0 {
		return 0, errRegisterNotReadable
	}

	// Only read over I2C if we don't have the host register value cached
	var val uint16 = r.host.value
	if !r.host.cached {
		// To avoid an alloc we always use an array of 2 bytes
		var buffer [2]byte
		var buf []byte
		if r.host.num_bytes < 2 {
			buf = buffer[:1]
		} else {
			buf = buffer[:]
		}
		// Read the host register over I2C
		err := legacy.ReadRegister(bus, deviceAddress, r.host.address, buf)
		if nil != err {
			return 0, err
		}
		// Unpack data from I2C
		if r.host.num_bytes > 1 {
			val = binary.BigEndian.Uint16(buf)
		} else {
			val = uint16(buf[0])
		}
		// cache this value if the host register is writeable. Note we cache the entire buffer without applying shift/mask
		if r.host.attributes&reg_write != 0 {
			r.host.value = val
			r.host.cached = true
		}
	}
	// Shift and mask the value before returning
	val >>= shift
	val &= mask
	return val, nil
}

// read reads a value for the register over the given I2C bus from the device with the given address.
func (r *i2cRegister) read(bus drivers.I2C, deviceAddress uint8) (uint16, error) {
	return r.readShiftAndMask(bus, deviceAddress, r.shift, r.mask)
}

// write writes a value for the register over the given I2C bus to the device with the given address.
func (r *i2cRegister) write(bus drivers.I2C, deviceAddress uint8, value uint16) error {
	if r.host.attributes&reg_write == 0 {
		return errRegisterNotWriteable
	}
	var newValue uint16 = 0
	// Data sheet tells us to do a read first, modify only the desired bits and then write back
	// since (quote:) 'Blank fields may contain factory settings'
	// We will also need to do this anyway to support virtualRegister mappings on some registers
	// (e.g. CONF/STATUS)
	if (r.host.attributes & reg_read) > 0 { // not all registers are readable, e.g. BURN
		// read the host register's entire host byte/word, regardless of shift & mask
		readValue, error := r.readShiftAndMask(bus, deviceAddress, 0, 0xffff)
		if error != nil {
			return error
		}
		// Zero-out ONLY the relevant bits in newValue we just read
		readValue &= (0xffff ^ (r.mask << r.shift))
		newValue = readValue
	}
	// Mask the new value and shift it into place
	value &= r.mask
	value <<= r.shift
	// OR the masked & shifted value back into newValue to be written
	newValue |= value
	// Pack newValue into a byte buffer to write. To avoid an alloc we always use an array of 2 bytes
	var buffer [2]byte
	var buf []byte
	if r.host.num_bytes < 2 {
		buf = buffer[:1]
		buf[0] = uint8(newValue & 0xff)
	} else {
		buf = buffer[:]
		binary.BigEndian.PutUint16(buf, newValue)
	}

	// Write the register from the buffer over I2C
	err := legacy.WriteRegister(bus, deviceAddress, r.host.address, buf)
	// after successful I2C write, cache this value if the host register (if also readable)
	// Note we cache the entire buffer without applying shift/mask
	if nil == err && r.host.attributes&reg_read != 0 {
		r.host.value = newValue
		r.host.cached = true
	}
	return err
}
