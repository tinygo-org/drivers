package seesaw

import (
	"testing"

	"github.com/frankban/quicktest"
)

func TestDevice_SoftReset_success(t *testing.T) {
	mocked := newMockDev(t, 0x49,
		when([]byte{0x00, 0x7F, 0xFF}, nil, nil),
		when([]byte{0x00, 0x01}, nil, nil),
		when(nil, []byte{0x55}, nil),
	)

	sut := New(mocked)

	err := sut.SoftReset()
	qt := quicktest.New(t)
	qt.Assert(err, quicktest.IsNil)
}

func TestDevice_WriteRegister(t *testing.T) {
	write := byte(0x1F)
	mocked := newMockDev(t, 0x49,
		when([]byte{0x01, 0x04, write}, nil, nil),
	)

	sut := New(mocked)

	err := sut.WriteRegister(ModuleGpioBase, FunctionGpioBulk, write)
	qt := quicktest.New(t)
	qt.Assert(err, quicktest.IsNil)
}

func TestDevice_ReadRegister(t *testing.T) {
	read := byte(0x23)
	mocked := newMockDev(t, 0x49,
		when([]byte{0x0F, 0x10}, nil, nil),
		when(nil, []byte{read}, nil),
	)

	sut := New(mocked)

	r, err := sut.ReadRegister(ModuleTouchBase, FunctionTouchChannelOffset)
	qt := quicktest.New(t)
	qt.Assert(err, quicktest.IsNil)
	qt.Assert(r, quicktest.Equals, r)
}

func TestDevice_Read(t *testing.T) {
	expectedRead := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	mocked := newMockDev(t, 0x49,
		when([]byte{0x0F, 0x10}, nil, nil),
		when(nil, expectedRead, nil),
	)

	sut := New(mocked)
	sut.ReadDelay = 0

	var buf [5]byte
	err := sut.Read(ModuleTouchBase, FunctionTouchChannelOffset, buf[:])

	qt := quicktest.New(t)
	qt.Assert(err, quicktest.IsNil)
	qt.Assert(buf[:], quicktest.DeepEquals, expectedRead)
}

func TestDevice_Write(t *testing.T) {
	expectedWrite := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	mocked := newMockDev(t, 0x49,
		when(append([]byte{0x0E, 0x04}, expectedWrite...), nil, nil),
	)

	sut := New(mocked)

	err := sut.Write(ModuleNeoPixelBase, FunctionNeopixelBuf, expectedWrite)
	qt := quicktest.New(t)
	qt.Assert(err, quicktest.IsNil)
}
