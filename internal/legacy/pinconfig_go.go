//go:build !tinygo

package legacy

import "tinygo.org/x/drivers"

// This file compiles for non-tinygo builds
// for use with "big" or "upstream" Go where
// there is no machine package.

func configurePinOut(p drivers.OutputPin)          {}
func configurePinInput(p drivers.InputPin)         {}
func configurePinInputPulldown(p drivers.InputPin) {}
func configurePinInputPullup(p drivers.InputPin)   {}
func pinIsNoPin(a drivers.Pin) bool                { return false }
