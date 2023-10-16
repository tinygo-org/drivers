package seesaw

import (
	"reflect"
	"testing"
	"time"
)

func TestDevice_SoftReset_success(t *testing.T) {

	mocked := newMockDev(t, 0x49,
		when([]byte{0x00, 0x7F, 0xFF}, nil, nil),
		when([]byte{0x00, 0x01}, nil, nil),
		when(nil, []byte{0x55}, nil),
	)

	sut := New(mocked)

	err := sut.SoftReset()
	assertEquals(t, err, nil)
}

func TestDevice_WriteRegister(t *testing.T) {

	write := byte(0x1F)
	mocked := newMockDev(t, 0x49,
		when([]byte{0x01, 0x04, write}, nil, nil),
	)

	sut := New(mocked)

	err := sut.WriteRegister(ModuleGpioBase, FunctionGpioBulk, write)
	assertEquals(t, err, nil)
}

func TestDevice_ReadRegister(t *testing.T) {

	read := byte(0x23)
	mocked := newMockDev(t, 0x49,
		when([]byte{0x0F, 0x10}, nil, nil),
		when(nil, []byte{read}, nil),
	)

	sut := New(mocked)

	r, err := sut.ReadRegister(ModuleTouchBase, FunctionTouchChannelOffset)
	assertEquals(t, err, nil)
	assertEquals(t, r, read)
}

func TestDevice_Read(t *testing.T) {

	expectedRead := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	mocked := newMockDev(t, 0x49,
		when([]byte{0x0F, 0x10}, nil, nil),
		when(nil, expectedRead, nil),
	)

	sut := New(mocked)

	var buf [5]byte
	err := sut.Read(ModuleTouchBase, FunctionTouchChannelOffset, buf[:], time.Nanosecond)
	assertEquals(t, err, nil)
	assertEquals(t, buf[:], expectedRead)
}

func TestDevice_Write(t *testing.T) {

	expectedWrite := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	mocked := newMockDev(t, 0x49,
		when(append([]byte{0x0E, 0x04}, expectedWrite...), nil, nil),
	)

	sut := New(mocked)

	err := sut.Write(ModuleNeoPixelBase, FunctionNeopixelBuf, expectedWrite)
	assertEquals(t, err, nil)
}

func assertEquals[e any](t *testing.T, actual, expected e) {
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("actual %+v != %+v not equals expected", actual, expected)
	}
}
