package seesaw

import (
	"errors"
)

var errInvalidEncoderNumber = errors.New("invalid encoder choice, 0-15 are supported")

// GetEncoderPosition returns the absolute position (or delta since the previous call) of the specified rotary encoder.
func (d *Device) GetEncoderPosition(encoder uint, asDelta bool) (int32, error) {
	if encoder >= 16 {
		return 0, errInvalidEncoderNumber
	}

	// The function address' upper nibble is the function, the lower nibble selects which encoder to communicate with
	fnAddr := FunctionAddress(encoder)
	if asDelta {
		fnAddr |= FunctionEncoderDelta
	} else {
		fnAddr |= FunctionEncoderPosition
	}

	var buf [4]byte
	err := d.Read(ModuleEncoderBase, fnAddr, buf[:])
	if err != nil {
		return 0, err
	}

	return int32(buf[0])<<24 | int32(buf[1])<<16 | int32(buf[2])<<8 | int32(buf[3]), nil
}

// SetEncoderPosition calibrate's the encoder's current absolute position to be whatever the provided position is.
func (d *Device) SetEncoderPosition(encoder uint, position int32) error {
	if encoder >= 16 {
		return errInvalidEncoderNumber
	}

	// The function address' upper nibble is the function, the lower nibble selects which encoder to communicate with
	fnAddr := FunctionEncoderPosition | FunctionAddress(encoder)

	buf := [4]byte{
		byte(position >> 24),
		byte(position >> 16),
		byte(position >> 8),
		byte(position),
	}

	return d.Write(ModuleEncoderBase, fnAddr, buf[:])
}
