// connects the Go crypto/rand package to the random number generation
// on the ATECCx08 cryptographic processor.
package main

import (
	"crypto/rand"
	"errors"
)

var (
	errNoATECC = errors.New("no ATECCx08")
)

func init() {
	rand.Reader = &reader{}
}

type reader struct{}

func (r *reader) Read(b []byte) (n int, err error) {
	if len(b) == 0 {
		return
	}

	if atecc == nil {
		return 0, errNoATECC
	}

	if !atecc.IsLocked() {
		panic("ATECCx08 is not locked and cannot produce random numbers!")
	}

	rnds, err := atecc.Random()
	if err != nil {
		return 0, err
	}

	for i := 0; i < len(b); i += 32 {
		if i+32 > len(b) {
			copy(b[i:], rnds[:(len(b)-i)])
			break
		}

		copy(b[i:], rnds[:])
		rnds, err = atecc.Random()
		if err != nil {
			return 0, err
		}
	}

	return len(b), nil
}
