// Package semihosting implements parts of the ARM semihosting specification,
// for communicating over a debug connection.
//
// If you want to use it in OpenOCD, you have to enable it first with the
// following command:
//
//	arm semihosting enable
package semihosting

import (
	"device/arm"
	"unsafe"
)

// IOError is returned by I/O operations when they fail.
type IOError struct {
	BytesWritten int
}

func (e *IOError) Error() string {
	return "semihosting: I/O error"
}

// Write writes the given data to the given file descriptor. It returns an
// *IOError if the write was not successful.
func Write(fd uintptr, data []byte) error {
	if len(data) == 0 {
		return nil
	}
	params := struct {
		fd   uintptr
		data unsafe.Pointer
		len  int
	}{
		fd:   fd,
		data: unsafe.Pointer(&data[0]),
		len:  len(data),
	}
	unwritten := arm.SemihostingCall(arm.SemihostingWrite, uintptr(unsafe.Pointer(&params)))
	if unwritten != 0 {
		// Error: unwritten is the number of bytes not written.
		return &IOError{
			BytesWritten: len(data) - unwritten,
		}
	}
	return nil
}
