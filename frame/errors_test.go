package frame_test

import (
	"reflect"
	"testing"
	"unsafe"

	"tinygo.org/x/drivers/frame"
)

func TestErrorReflect(t *testing.T) {
	for i := uint8(1); i < uint8(frame.ErrCodeMax); i++ {
		err := errorer(i)
		// println(unsafe.Sizeof(err))
		code := codeFromError(err)
		// t.Errorf("sizeof %v", unsafe.Sizeof(err))
		if i != code {
			t.Errorf("expect errcode %v. got %v", i, code)
		}
		codeu := codeFromErrorUnsafe(err)
		if i != codeu {
			t.Errorf("expect errcode %v. got %v (unsafe)", i, codeu)
		}
	}
}

func errorer(ec uint8) error {

	return frame.ErrorCode(ec)
}
func codeFromError(err error) uint8 {
	if err == nil {
		panic("arg is nil")
	}
	v := reflect.ValueOf(err)
	return uint8(v.Uint())
}

// type eface struct {
// 	typ uintptr
// 	val *uint8
// }

func codeFromErrorUnsafe(err error) uint8 {
	if err == nil {
		panic("arg is nil")
	}
	type eface struct {
		typ uintptr
		val *uint8
	}
	ptr := unsafe.Pointer(&err)
	val := (*uint8)(unsafe.Pointer((*eface)(ptr).val))
	return *val
}

func printError(err error) {
	if err != nil {
		type eface struct {
			typ, val unsafe.Pointer
		}
		passed_value := (*eface)(unsafe.Pointer(&err)).val
		println("error #", *(*uint8)(passed_value))
	}
}
