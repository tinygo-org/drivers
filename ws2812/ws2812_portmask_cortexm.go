//go:build cortexm && !mk66f18

package ws2812

// portMasks returns the registers and masks that set and clear the pin.
func (d Device) portMasks() (portSet *uint32, maskSet uint32, portClear *uint32, maskClear uint32) {
	portSet, maskSet = d.Pin.PortMaskSet()
	portClear, maskClear = d.Pin.PortMaskClear()
	return
}
