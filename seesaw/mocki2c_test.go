package seesaw

import (
	"encoding/hex"
	"fmt"
	"testing"
)

type I2CHandleFunc func(t *testing.T, w, r []byte) error

// mocki2c implements the drivers.I2C interface and matches a list
// of handlers against actual invocations. Useful to test command/reply style I2C devices.
type mocki2c struct {
	addr     uint16
	handlers []I2CHandleFunc
	t        *testing.T
}

func (m *mocki2c) Tx(addr uint16, w, r []byte) error {
	assertEquals(m.t, addr, m.addr)
	if len(m.handlers) == 0 {
		ws := hex.EncodeToString(w)
		rs := hex.EncodeToString(r)
		panic(fmt.Sprintf("no handlers for: addr='%02x' w='%s' r='%s'", byte(addr), ws, rs))
	}
	h := m.handlers[0]
	m.handlers = m.handlers[1:]
	return h(m.t, w, r)
}

func newMockDev(t *testing.T, addr uint16, handlers ...I2CHandleFunc) *mocki2c {
	return &mocki2c{
		addr:     addr,
		handlers: handlers,
		t:        t,
	}
}

func when(expectedWrite, returningRead []byte, returningError error) I2CHandleFunc {
	return func(t *testing.T, w, r []byte) error {
		assertEquals(t, w, expectedWrite)
		assertEquals(t, len(r), len(returningRead))
		if r != nil {
			copy(r, returningRead)
		}
		return returningError
	}
}
