// Package cap1166 controls the Microchip CAP1166 touch inputs and LED outputs.
// Datasheet: https://ww1.microchip.com/downloads/en/DeviceDoc/00001621B.pdf
package cap1166 // import "tinygo.org/x/drivers/cap1166"

import (
	"fmt"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/internal/legacy"
)

// DefaultAddress is the I2C address on the Display-o-Tron HAT.
// Other boards can use 0x28.
const DefaultAddress = 0x2C

var sensitivity = map[uint8]uint8{
	128: 0b000,
	64:  0b001,
	32:  0b010,
	16:  0b011,
	8:   0b100,
	4:   0b101,
	2:   0b110,
	1:   0b111,
}

// Device is a handle to a CAP1166 capacitive touch controller.
type Device struct {
	bus     drivers.I2C
	Address uint8
}

// New returns a new cap1166 driver on the given I2C bus and address.
func New(bus drivers.I2C, address uint8) Device {
	return Device{
		bus:     bus,
		Address: address,
	}
}

// Configure checks the product ID and enables all six touch inputs.
// It uses the settings from https://github.com/pimoroni/cap1xxx-python.
func (d *Device) Configure() error {
	id, err := d.readByte(regProductID)
	if err != nil {
		return err
	}
	if id != ProductID {
		return fmt.Errorf("cap1166: unexpected product id 0x%02X", id)
	}

	if err := d.writeByte(regInputEnable, 0x3F); err != nil {
		return err
	}
	if err := d.writeByte(regInterruptEnable, 0x3F); err != nil {
		return err
	}
	if err := d.EnableRepeat(0x00); err != nil {
		return err
	}
	if err := d.EnableMultitouch(true); err != nil {
		return err
	}
	if err := d.writeByte(regSamplingConfig, 0b00001000); err != nil {
		return err
	}
	if err := d.SetSensitivity(2); err != nil {
		return err
	}
	if err := d.writeByte(regGeneralConfig, 0b00111000); err != nil {
		return err
	}
	return d.writeByte(regConfiguration2, 0b01100000)
}

// Connected reports whether a CAP1166 answers at Address.
func (d *Device) Connected() bool {
	id, err := d.readByte(regProductID)
	return err == nil && id == ProductID
}

// InputStatus returns the six touch input bits. Bit n is set for input n.
// The channel constants give bit positions, not masks.
func (d *Device) InputStatus() (uint8, error) {
	status, err := d.readByte(regInputStatus)
	return status & 0x3F, err
}

// Pressed reports whether the given input channel is currently touched.
func (d *Device) Pressed(channel uint8) (bool, error) {
	status, err := d.InputStatus()
	if err != nil {
		return false, err
	}
	return status&(1<<channel) != 0, nil
}

// ClearInterrupt clears the interrupt flag in the main control register.
// The input status bits will not update again until this is called.
func (d *Device) ClearInterrupt() error {
	v, err := d.readByte(regMainControl)
	if err != nil {
		return err
	}
	return d.writeByte(regMainControl, v&^0x01)
}

// EnableRepeat enables touch-and-hold repeat events for the inputs set in
// mask, one bit per input.
func (d *Device) EnableRepeat(mask uint8) error {
	return d.writeByte(regRepeatEnable, mask)
}

// EnableMultitouch allows, or blocks, more than one simultaneous touch.
func (d *Device) EnableMultitouch(enable bool) error {
	v, err := d.readByte(regMultiTouchConf)
	if err != nil {
		return err
	}
	if enable {
		v &^= 0x80
	} else {
		v |= 0x80
	}
	return d.writeByte(regMultiTouchConf, v)
}

// SetSensitivity sets the touch sensitivity multiplier. Valid values are
// 128, 64, 32, 16, 8, 4, 2 and 1, where higher is more sensitive.
func (d *Device) SetSensitivity(multiplier uint8) error {
	bits, ok := sensitivity[multiplier]
	if !ok {
		return fmt.Errorf("cap1166: invalid sensitivity %d", multiplier)
	}
	v, err := d.readByte(regSensitivity)
	if err != nil {
		return err
	}
	v = (v &^ (0b111 << 4)) | (bits << 4)
	return d.writeByte(regSensitivity, v)
}

const (
	graphNumLEDs    = NumInputs
	graphStepValue  = 16
	graphTotalValue = graphStepValue * graphNumLEDs
)

// SetGraph sets the six LED outputs from 0.0 (all off) to 1.0 (all on).
// The boundary LED uses the direct duty cycle.
func (d *Device) SetGraph(percentage float32) error {
	if percentage < 0 {
		percentage = 0
	}
	if percentage > 1 {
		percentage = 1
	}

	if err := d.writeByte(regLEDDirectRamp, 0x00); err != nil {
		return err
	}
	if err := d.writeByte(regLEDBehaviour1, 0x00); err != nil {
		return err
	}
	if err := d.writeByte(regLEDBehaviour2, 0x00); err != nil {
		return err
	}

	actualValue := int(float32(graphTotalValue) * percentage)
	var polarity, state, duty uint8
	for x := 0; x < graphNumLEDs; x++ {
		if actualValue >= graphStepValue {
			polarity |= 1 << uint(graphNumLEDs-1-x)
		}
		if actualValue > 0 && actualValue < graphStepValue {
			state |= 1 << uint(graphNumLEDs-1-x)
			duty = uint8(actualValue << 4)
		}
		actualValue -= graphStepValue
	}

	if err := d.writeByte(regLEDDirectDuty, duty); err != nil {
		return err
	}
	if err := d.writeByte(regLEDPolarity, polarity); err != nil {
		return err
	}
	return d.writeByte(regLEDOutputCon, state)
}

// GraphOff turns off all 6 bar graph LEDs.
func (d *Device) GraphOff() error {
	if err := d.writeByte(regLEDPolarity, 0x00); err != nil {
		return err
	}
	return d.writeByte(regLEDOutputCon, 0x00)
}

func (d *Device) writeByte(reg, value uint8) error {
	return legacy.WriteRegister(d.bus, d.Address, reg, []byte{value})
}

func (d *Device) readByte(reg uint8) (uint8, error) {
	buf := make([]byte, 1)
	err := legacy.ReadRegister(d.bus, d.Address, reg, buf)
	return buf[0], err
}
