//go:build baremetal && !fe310

package legacy

import "machine"

// If you are getting a build error here you then we missed adding
// your CPU build tag to the list of CPUs that do not have pulldown/pullups.
// Add it above and in pinhal_nopulls! You should also add a smoketest for it :)
const (
	pulldown = machine.PinInputPulldown
	pullup   = machine.PinInputPullup
)
