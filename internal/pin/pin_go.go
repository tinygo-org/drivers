//go:build !baremetal

package pin

func configureOutput(p Output)       {}
func configureInput(p Input)         {}
func configureInputPulldown(p Input) {}
func configureInputPullup(p Input)   {}
func isNotPin(a any) bool            { return false }
