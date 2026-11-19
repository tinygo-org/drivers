//go:build !baremetal

package legacy

import "tinygo.org/x/drivers/internal/pin"

// This file compiles for builds that have no hardware pins. That is
// "big" or "upstream" Go, and TinyGo targets that are not bare metal
// such as wasm and wasi. The functions do nothing.

func configurePinOut(p pin.Output)          {}
func configurePinInput(p pin.Input)         {}
func configurePinInputPulldown(p pin.Input) {}
func configurePinInputPullup(p pin.Input)   {}
func pinIsNoPin(a any) bool                 { return false }
