package as7262

import (
	"encoding/binary"
	"math"
	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/internal/legacy"
)

type Error uint8

const (
	ErrInvalidID Error = 0x1
	TxValid      byte  = 0x02
	RxValid      byte  = 0x01
)

func (e Error) Error() string {
	switch e {
	case ErrInvalidID:
		return "Invalid chip ID"
	default:
		return "Unknown error"
	}
}

type Device struct {
	bus            drivers.I2C
	buf            []byte
	Address        uint8
	vControlReg    *vControlReg
	vLedControlReg *vLedControlReg
}

// New returns pointer of new as7262 device
func New(i2c drivers.I2C) *Device {
	return &Device{
		bus:            i2c,
		buf:            []byte{0},
		Address:        DefaultAddress,
		vControlReg:    newVControlReg(),
		vLedControlReg: newVLedControlReg(),
	}
}

// deviceStatus returns StatusReg of as7262
func (d *Device) deviceStatus() byte {
	d.buf[0] = 0b00000000
	legacy.ReadRegister(d.bus, DefaultAddress, StatusReg, d.buf)
	return d.buf[0]
}

// writeReady returns true if as7262 is ready to write write-register
func (d *Device) writeReady() bool {
	return d.deviceStatus()&TxValid == 0
}

// readReady return true if as7262 is ready to read read-register
func (d *Device) readReady() bool {
	return d.deviceStatus()&RxValid != 0
}

func (d *Device) readByte(reg byte) byte {
	d.buf[0] = 0b00000000
	for {
		if d.writeReady() {
			break
		}
	}

	legacy.WriteRegister(d.bus, d.Address, WriteReg, []byte{reg})

	for {
		if d.readReady() {
			break
		}
	}

	legacy.ReadRegister(d.bus, d.Address, ReadReg, d.buf)
	return d.buf[0]
}

func (d *Device) writeByte(reg byte, value byte) {
	for {
		if d.writeReady() {
			break
		}
	}

	legacy.WriteRegister(d.bus, d.Address, WriteReg, []byte{reg | 0x80})

	for {
		if d.writeReady() {
			break
		}
	}

	legacy.WriteRegister(d.bus, d.Address, WriteReg, []byte{value})
}

func (d *Device) read32Bit(reg byte) float32 {
	var bytes [4]byte

	for i := 0; i < 4; i++ {
		bytes[3-i] = d.readByte(reg + byte(i))
	}
	floatValue := math.Float32frombits(binary.BigEndian.Uint32(bytes[:]))
	return floatValue
}

// Temperature returns sensor temperature
func (d *Device) Temperature() byte {
	return d.readByte(TempRegister)
}

// GetColors set pointer array: V, B, G, Y, O, R
func (d *Device) GetColors(arr *[6]float32) {
	arr[0] = d.GetViolet()
	arr[1] = d.GetBlue()
	arr[2] = d.GetGreen()
	arr[3] = d.GetYellow()
	arr[4] = d.GetOrange()
	arr[5] = d.GetRed()
}

// GetRGB set pointer array: R, G, B
func (d *Device) GetRGB(arr *[3]float32) {
	arr[0] = d.GetRed()
	arr[1] = d.GetGreen()
	arr[2] = d.GetBlue()
}

// GetViolet returns violet value
func (d *Device) GetViolet() float32 {
	return d.read32Bit(VCalReg)
}

// GetBlue returns blue value
func (d *Device) GetBlue() float32 {
	return d.read32Bit(BCalReg)
}

// GetGreen returns green value
func (d *Device) GetGreen() float32 {
	return d.read32Bit(GCalReg)
}

// GetYellow returns yellow value
func (d *Device) GetYellow() float32 {
	return d.read32Bit(YCalReg)
}

// GetOrange returns orange value
func (d *Device) GetOrange() float32 {
	return d.read32Bit(OCalReg)
}

// GetRed returns red value
func (d *Device) GetRed() float32 {
	return d.read32Bit(RCalReg)
}
