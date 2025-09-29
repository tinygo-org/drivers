//go:build !tinygo

package legacy

import "tinygo.org/x/drivers/internal/pin"

// This file compiles for non-tinygo builds
// for use with "big" or "upstream" Go where
// there is no machine package.

func configurePinOut(p pin.Output)          {}
func configurePinInput(p pin.Input)         {}
func configurePinInputPulldown(p pin.Input) {}
func configurePinInputPullup(p pin.Input)   {}
func pinIsNoPin(a any) bool                 { return false }
