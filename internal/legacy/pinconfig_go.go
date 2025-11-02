//go:build !tinygo

package legacy

import "tinygo.org/x/drivers/internal/pin"

// This file compiles for non-tinygo builds
// for use with "big" or "upstream" Go where
// there is no machine package.

func configurePinOut(p pin.OutputInterface)          {}
func configurePinInput(p pin.InputInterface)         {}
func configurePinInputPulldown(p pin.InputInterface) {}
func configurePinInputPullup(p pin.InputInterface)   {}
func pinIsNoPin(a any) bool                          { return false }
