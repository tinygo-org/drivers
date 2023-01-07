package net

import (
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestIPAddressString(t *testing.T) {
	c := qt.New(t)
	ipaddr := ParseIP("127.0.0.1")

	c.Assert(ipaddr.String(), qt.Equals, "127.0.0.1")
}
